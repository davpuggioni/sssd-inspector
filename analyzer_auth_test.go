// analyzer_auth_test.go
package main

import (
	"os"
	"testing"
)

// Test: Verify the Kerberos 5-minute rule (Time Skew)
func TestAnalyzeTime_KerberosSkew(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"systemd.txt": "Active: active (running) chronyd.service\n",
		"ntp.txt":     "System time     : 312.450 seconds fast of NTP time\nSystem clock synchronized: yes\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeTime(dir, &report)

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

	var report ReportData
	analyzeNSSwitch(dir, &report)

	if !containsString(report.Warnings, "'sss' is listed before 'files'/'compat' for 'passwd'") {
		t.Errorf("Failed to detect bad nsswitch.conf ordering")
	}
}
