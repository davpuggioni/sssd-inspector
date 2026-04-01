// analyzer_logs_test.go
package main

import (
	"os"
	"strings"
	"testing"
)

// Test: Verify the RC4 Downgrade bug is reported in Tuning
func TestAnalyzeKerberos_RC4DowngradeBug(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"messages": "Dec 10 12:00:00 server sssd: service key not available\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeSSSDConfigAndLogs(dir, &report)

	if !containsString(report.Warnings, "[AD CRYPTO BUG]") {
		t.Errorf("Failed to detect RC4 downgrade bug in streamed logs")
	}
}

// Test: Verify deprecated 'enumerate = true' is caught
func TestAnalyzeConfig_EnumerateTrue(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/ad]\nenumerate = true\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeSSSDConfigAndLogs(dir, &report)

	if !containsString(report.Problems, "[DEPRECATION] 'enumerate = true' is set") {
		t.Errorf("Failed to detect deprecated enumerate=true setting")
	}
}

// Test: Verify SELinux Log Warnings are suppressed on AppArmor
func TestAnalyzeSSSDConfigAndLogs_SELinuxSuppression(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"security-apparmor.txt": "apparmor module is loaded.\n",
		"sssd.txt":              "SELINUX_getpeercon failed\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeMACStatus(dir, &report)
	analyzeSSSDConfigAndLogs(dir, &report)

	foundError := false
	for _, err := range report.SSSDLogErrors {
		if strings.Contains(err.Description, "SELinux Warning") {
			foundError = true
			break
		}
	}

	if foundError {
		t.Errorf("Expected strict SELinux warnings to be suppressed when AppArmor is active.")
	}
}

// Test: Verify SELinux Log Warnings are preserved on SELinux
func TestAnalyzeSSSDConfigAndLogs_SELinuxActive(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"security-selinux.txt": "SELinux status:                 enabled\n",
		"sssd.txt":             "SELINUX_getpeercon failed\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeMACStatus(dir, &report)
	analyzeSSSDConfigAndLogs(dir, &report)

	foundError := false
	for _, err := range report.SSSDLogErrors {
		if strings.Contains(err.Description, "SELinux Warning") {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected SELinux warnings to be preserved when SELinux is active.")
	}
}
