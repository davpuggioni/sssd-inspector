// analyzer_auth_test.go
package main

import (
	"os"
	"testing"

	"sssd-inspector/pkg/analysis"
	"sssd-inspector/pkg/types"
)

// testAnalyzer returns a shared AnalyzerContext for use in test helpers
func testAnalyzer() *analysis.AnalyzerContext {
	return analysis.NewAnalyzerContext()
}

// Test: Verify the Kerberos 5-minute rule (Time Skew)
func TestAnalyzeTime_KerberosSkew(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"systemd.txt": "Active: active (running) chronyd.service\n",
		"ntp.txt":     "System time     : 312.450 seconds fast of NTP time\nSystem clock synchronized: yes\n",
	})
	defer os.RemoveAll(dir)

	var report types.ReportData
	testAnalyzer().AnalyzeTime(dir, &report)

	if !containsString(report.Problems, "System time offset is 312.45") {
		t.Errorf("Failed to detect Kerberos Time Skew > 300 seconds via streaming")
	}
}

// Test: Verify strict nsswitch.conf ordering (sss before files)
func TestAnalyzeNSSwitch_BadOrdering(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"nsswitch.conf": "passwd: sss files\ngroup: files sss\n",
	})
	defer os.RemoveAll(dir)

	var report types.ReportData
	testAnalyzer().AnalyzeNSSwitch(dir, &report)

	if !containsString(report.Problems, "'sss' is listed before 'files'/'compat' for 'passwd'") {
		t.Errorf("Failed to detect bad nsswitch.conf ordering")
	}
}
