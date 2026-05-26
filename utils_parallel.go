// Package main provides parallel file processing utilities for SSSD Inspector
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// workItem represents a single file processing job for the parallel processor
type workItem struct {
	filePath string
	fileName string
}

// ParallelFileProcessor provides concurrent file scanning capabilities
// using a worker pool pattern for efficient I/O-bound operations
type ParallelFileProcessor struct {
	maxWorkers    int
	bufferSize    int
	maxLineLength int
}

// NewParallelFileProcessor creates a new ParallelFileProcessor with the specified worker count
func NewParallelFileProcessor(maxWorkers int) *ParallelFileProcessor {
	if maxWorkers <= 0 {
		maxWorkers = 4 // Default number of workers
	}
	return &ParallelFileProcessor{
		maxWorkers:    maxWorkers,
		bufferSize:    64 * 1024,   // 64KB default buffer
		maxLineLength: 1024 * 1024, // 1MB max line length
	}
}

// ScanFilesParallel scans multiple files concurrently using a worker pool.
// Each file is processed by a separate goroutine, and the lineFunc callback
// is invoked for each line. The function blocks until all files are processed.
// Thread safety of lineFunc is the caller's responsibility.
func (pfp *ParallelFileProcessor) ScanFilesParallel(dirPath string, files []string, lineFunc func(line string)) {
	if len(files) == 0 {
		return
	}

	// Create a buffered channel for work items
	jobs := make(chan workItem, len(files))
	results := make(chan error, len(files))

	// Determine the number of workers (cap at number of files)
	numWorkers := pfp.maxWorkers
	if numWorkers > len(files) {
		numWorkers = len(files)
	}

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go pfp.worker(jobs, results, lineFunc, &wg)
	}

	// Send jobs
	for _, name := range files {
		jobs <- workItem{
			filePath: filepath.Join(dirPath, name),
			fileName: name,
		}
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Check for errors (non-fatal, just log them)
	for err := range results {
		if err != nil {
			fmt.Printf("Warning: parallel scan error: %v\n", err)
		}
	}
}

// worker processes files from the jobs channel
func (pfp *ParallelFileProcessor) worker(jobs <-chan workItem, results chan<- error, lineFunc func(line string), wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		err := pfp.scanFile(job.filePath, lineFunc)
		results <- err
	}
}

// scanFile processes a single file and applies the line function
func (pfp *ParallelFileProcessor) scanFile(filePath string, lineFunc func(line string)) error {
	f, err := os.Open(filePath)
	if err != nil {
		// If the file simply doesn't exist, silently ignore it
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, pfp.bufferSize)
	scanner.Buffer(buf, pfp.maxLineLength)

	for scanner.Scan() {
		lineFunc(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error for file %s: %w", filePath, err)
	}
	return nil
}

// BatchProcessor allows processing items in parallel batches
type BatchProcessor struct {
	maxWorkers int
}

// NewBatchProcessor creates a new BatchProcessor
func NewBatchProcessor(maxWorkers int) *BatchProcessor {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	return &BatchProcessor{
		maxWorkers: maxWorkers,
	}
}

// ProcessBatch processes items in parallel using a worker pool
// items: slice of items to process
// processFunc: function to process each item (must be thread-safe for writing to results)
// Returns a slice of results (order is not guaranteed to match input order)
func (bp *BatchProcessor) ProcessBatch(items []string, processFunc func(item string) string) []string {
	if len(items) == 0 {
		return nil
	}

	type result struct {
		value string
	}

	jobs := make(chan string, len(items))
	results := make(chan result, len(items))

	numWorkers := bp.maxWorkers
	if numWorkers > len(items) {
		numWorkers = len(items)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				results <- result{value: processFunc(item)}
			}
		}()
	}

	for _, item := range items {
		jobs <- item
	}
	close(jobs)

	wg.Wait()
	close(results)

	var output []string
	for r := range results {
		output = append(output, r.value)
	}
	return output
}

// globalParallelProcessor is the default parallel file processor instance
var globalParallelProcessor = NewParallelFileProcessor(4)
