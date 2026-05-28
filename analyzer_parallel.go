// analyzer_parallel.go - Parallel analysis phases for SSSD Inspector
// Provides concurrent execution of independent analysis steps
// to improve performance on multi-core systems.

package main

import (
	"fmt"
	"sync"
)

// analyzePhase2Parallel runs the Phase 2 config analysis steps concurrently.
// These functions read different files and have no dependencies on each other,
// making them safe to parallelize.
func analyzePhase2Parallel(dirPath string, report *ReportData) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Helper to run analysis with mutex protection for the shared report
	run := func(fn func(string, *ReportData)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			fn(dirPath, report)
			mu.Unlock()
		}()
	}

	// These functions all read different files independently:
	// - analyzePAM reads pam.txt (or config file)
	// - analyzeNSSwitch reads nsswitch.conf
	// - analyzeHosts reads hosts file
	// - analyzeNSCD reads nscd.conf
	// - analyzePackages reads rpm.txt
	// - analyzeServices reads systemd.txt
	// - analyzeMACStatus reads boot.txt / security-*.txt
	// - analyzeDiskSpace reads fs-diskio.txt / storage.txt
	run(analyzePAM)
	run(analyzeNSSwitch)
	run(analyzeHosts)
	run(analyzeNSCD)
	run(analyzePackages)
	run(analyzeServices)
	run(func(dirPath string, report *ReportData) { analyzeMACStatus(dirPath, report) })
	run(analyzeDiskSpace)

	wg.Wait()

	// These depend on results from the parallel phase, so run sequentially afterwards
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

// init registers the parallel analysis functions
// These are used by analyzeData in analyzer_core.go when parallelism is beneficial.
var _ = fmt.Sprintf // ensure fmt import is used if needed elsewhere
