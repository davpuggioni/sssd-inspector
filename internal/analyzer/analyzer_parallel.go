// Package analyzer provides concurrent execution blocks and synchronized phase 
// orchestrators to safely leverage multi-core systems during ingestion.
package analyzer

import (
	"sync"
)

// analyzePhase2Parallel runs the Phase 2 config analysis steps concurrently.
// These functions read different files and have no dependencies on each other.
func analyzePhase2Parallel(dirPath string, report *ReportData) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Helper to run analysis with mutex protection for the shared report map
	run := func(fn func(string, *ReportData)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			fn(dirPath, report)
			mu.Unlock()
		}()
	}

	// These functions execute independently across localized files
	run(analyzePAM)
	run(analyzeNSSwitch)
	run(analyzeHosts)
	run(analyzeNSCD)
	run(analyzePackages)
	run(analyzeServices)
	run(func(dirPath string, report *ReportData) { analyzeMACStatus(dirPath, report) })
	run(analyzeDiskSpace)

	wg.Wait()

	// Runs sequentially downstream since they rely on combined multi-file outputs
	analyzeSSSDVersionAge(report)
	analyzeSSSDFilePermissions(dirPath, report)
}

// analyzePhase1Parallel runs Phase 1 system analysis steps concurrently.
func analyzePhase1Parallel(dirPath string, report *ReportData) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	basicFn := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			analyzeBasicHealth(dirPath, report)
			mu.Unlock()
		}()
	}
	osFn := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			analyzeOSAndHardware(dirPath, report)
			analyzeHostnameAndFQDN(dirPath, report)
			analyzeSCC(dirPath, report)
			mu.Unlock()
		}()
	}
	dnsFn := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			analyzeDNS(dirPath, report)
			mu.Unlock()
		}()
	}
	timeFn := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			analyzeTime(dirPath, report)
			mu.Unlock()
		}()
	}
	perfFn := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			analyzePerformance(dirPath, report)
			mu.Unlock()
		}()
	}

	basicFn()
	osFn()
	dnsFn()
	timeFn()
	perfFn()

	wg.Wait()
}
