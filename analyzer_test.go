// analyzer_test.go
package main

import (
	"strings"
	"testing"
)

// Helper function to check if a specific string exists in a slice
func containsString(slice []string, expected string) bool {
	for _, item := range slice {
		if strings.Contains(item, expected) {
			return true
		}
	}
	return false
}

// Test 1: Verify that SSSD version extraction works correctly
func TestGetSSSDVersion(t *testing.T) {
	packages := []string{"sssd-2.10.2-150700.9.17.1.x86_64", "sssd-ad-2.10.2"}
	major, minor := getSSSDVersion(packages)

	if major != 2 || minor != 10 {
		t.Errorf("Expected version 2.10, but got %d.%d", major, minor)
	}
}

// Test 2: Verify the Kerberos 5-minute rule (Time Skew)
func TestAnalyzeTime_KerberosSkew(t *testing.T) {
	fileMap := map[string]string{
		"systemd.txt": "Active: active (running) chronyd.service\n",
		"ntp.txt":     "System time     : 312.450 seconds fast of NTP time\nSystem clock synchronized: yes\n",
	}
	var report ReportData
	analyzeTime(fileMap, &report)

	if !containsString(report.Problems, "System time offset is 312.45") {
		t.Errorf("Failed to detect Kerberos Time Skew > 300 seconds")
	}
}

// Test 3: Verify the RC4 Downgrade bug is reported in Tuning
func TestAnalyzeKerberos_RC4DowngradeBug(t *testing.T) {
	fileMap := map[string]string{
		"messages": "Dec 10 12:00:00 server sssd: service key not available\n",
	}
	var report ReportData
	analyzeSSSDConfigAndLogs(fileMap, &report)

	if !containsString(report.Warnings, "[AD CRYPTO BUG]") {
		t.Errorf("Failed to detect RC4 downgrade bug in logs")
	}
}

// Test 4: Verify deprecated 'enumerate = true' is caught
func TestAnalyzeConfig_EnumerateTrue(t *testing.T) {
	fileMap := map[string]string{
		"sssd.conf": "[domain/ad]\nenumerate = true\n",
	}
	var report ReportData
	analyzeSSSDConfigAndLogs(fileMap, &report)

	if !containsString(report.Problems, "[DEPRECATION] 'enumerate = true' is set") {
		t.Errorf("Failed to detect deprecated enumerate=true setting")
	}
}

// Test 5: Verify vm.dirty_bytes performance issue
func TestAnalyzePerformance_DirtyBytes(t *testing.T) {
	fileMap := map[string]string{
		"memory.txt": "vm.dirty_bytes = 0\n",
	}
	var report ReportData
	analyzePerformance(fileMap, &report)

	if !containsString(report.Problems, "[PERFORMANCE] vm.dirty_bytes is set to 0") {
		t.Errorf("Failed to detect vm.dirty_bytes = 0 misconfiguration")
	}
}

// Test 6: Verify strict nsswitch.conf ordering (sss before files)
func TestAnalyzeNSSwitch_BadOrdering(t *testing.T) {
	fileMap := map[string]string{
		// 'sss' is mistakenly placed before 'files'
		"nsswitch.conf": "passwd: sss files\ngroup: files sss\n",
	}
	var report ReportData
	analyzeNSSwitch(fileMap, &report)

	if !containsString(report.Problems, "'sss' is listed before 'files'/'compat' for 'passwd'") {
		t.Errorf("Failed to detect bad nsswitch.conf ordering")
	}
}

// Test 7: Verify SSSD >= 2.10 file permissions (Should catch root:root as an error on standard OS)
func TestAnalyzePermissions_NewVersion(t *testing.T) {
	fileMap := map[string]string{
		"rpm.txt": "sssd-2.10.2-150700.9.17.1.x86_64\n",
		// Wrong permissions for >= 2.10 (should be root:sssd 0640)
		"sssd.txt": "-rw------- 1 root root 1024 Mar 10 10:00 sssd.conf\n",
	}
	var report ReportData
	analyzePackages(fileMap, &report) // This extracts the version 2.10
	analyzeSSSDFilePermissions(fileMap, &report)

	if !containsString(report.Problems, "strictly requires 'root:sssd'") {
		t.Errorf("Failed to detect incorrect file ownership for SSSD 2.10+")
	}
}

// Test 8: Verify Disk Space warning for full partitions
func TestAnalyzeDiskSpace_Critical(t *testing.T) {
	fileMap := map[string]string{
		"fs-diskio.txt": "/dev/sda1  100G  98G  2G  98% /\n",
	}
	var report ReportData
	analyzeDiskSpace(fileMap, &report)

	if !containsString(report.Problems, "[CRITICAL] Partition / is 98% full") {
		t.Errorf("Failed to detect 98%% full root partition")
	}
}

// Test 9: Verify Systemd Exit Code parsing (Code 4 = Config Error)
func TestAnalyzeServices_CrashCode4(t *testing.T) {
	fileMap := map[string]string{
		"systemd.txt": "● sssd.service - System Security Services Daemon\n   Loaded: loaded\n   Active: failed (Result: exit-code)\n  Process: 1234 ExecStart=/usr/sbin/sssd (code=exited, status=4)\n",
	}
	var report ReportData
	analyzeServices(fileMap, &report)

	if !containsString(report.Problems, "crashed with exit code 4 (Configuration Error)") {
		t.Errorf("Failed to parse systemd exit code 4")
	}
}

// Test 10: Verify FQDN Hostname check
func TestAnalyzeHostname_ShortName(t *testing.T) {
	fileMap := map[string]string{
		"basic-environment.txt": "Hostname: linux-server\n", // Missing the dot (e.g. .local)
	}
	var report ReportData
	analyzeHostnameAndFQDN(fileMap, &report)

	if !containsString(report.Problems, "System is using a short hostname instead of an FQDN") {
		t.Errorf("Failed to detect short hostname")
	}
}

// Test 11: Verify AppArmor MAC Status extraction
func TestAnalyzeMACStatus_AppArmor(t *testing.T) {
	fileMap := map[string]string{
		"security-apparmor.txt": "apparmor module is loaded.\n",
	}
	var report ReportData
	analyzeMACStatus(fileMap, &report)

	if report.MACType != "AppArmor" {
		t.Errorf("Expected MACType 'AppArmor', got '%s'", report.MACType)
	}
}

// Test 12: Verify SELinux MAC Status extraction
func TestAnalyzeMACStatus_SELinux(t *testing.T) {
	fileMap := map[string]string{
		"security-selinux.txt": "SELinux status:                 enabled\n",
	}
	var report ReportData
	analyzeMACStatus(fileMap, &report)

	if report.MACType != "SELinux" {
		t.Errorf("Expected MACType 'SELinux', got '%s'", report.MACType)
	}
}

// Test 13: Verify SELinux Log Warnings are suppressed on AppArmor
func TestAnalyzeSSSDConfigAndLogs_SELinuxSuppression(t *testing.T) {
	fileMap := map[string]string{
		"security-apparmor.txt": "apparmor module is loaded.\n",
		"sssd.txt":              "SELINUX_getpeercon failed\n",
	}
	var report ReportData
	analyzeMACStatus(fileMap, &report)
	analyzeSSSDConfigAndLogs(fileMap, &report)

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

// Test 14: Verify SELinux Log Warnings are preserved on SELinux
func TestAnalyzeSSSDConfigAndLogs_SELinuxActive(t *testing.T) {
	fileMap := map[string]string{
		"security-selinux.txt": "SELinux status:                 enabled\n",
		"sssd.txt":             "SELINUX_getpeercon failed\n",
	}
	var report ReportData
	analyzeMACStatus(fileMap, &report)
	analyzeSSSDConfigAndLogs(fileMap, &report)

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

// Test 15: Verify valid /var/lib/sss/ ownership for SSSD 2.10+ (Standard Behavior)
func TestAnalyzePermissions_VarLibSss_SSSD210_Valid(t *testing.T) {
	fileMap := map[string]string{
		"rpm.txt": "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt": `/var/lib/sss/:
total 20
drwx------ 2 sssd sssd   40 2026-03-01 10:39 db
drwxr-xr-x 2 sssd sssd 4096 2026-03-01 10:28 mc

/var/lib/sss/mc:
total 33900
-rw-rw-r-- 1 sssd sssd  6940392 2026-03-01 10:38 group
`,
	}
	var report ReportData
	analyzePackages(fileMap, &report)
	analyzeSSSDFilePermissions(fileMap, &report)

	if containsString(report.Problems, "/var/lib/sss/") {
		t.Errorf("Expected no /var/lib/sss/ permissions errors for perfectly owned directories, but found one.")
	}
}

// Test 16: Verify invalid /var/lib/sss/ ownership for SSSD 2.10+ (Standard Behavior)
func TestAnalyzePermissions_VarLibSss_SSSD210_Invalid(t *testing.T) {
	fileMap := map[string]string{
		"rpm.txt": "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt": `/var/lib/sss/:
total 20
drwx------ 2 sssd sssd   40 2026-03-01 10:39 db
drwxr-xr-x 2 root root 4096 2026-03-01 10:28 mc

/var/lib/sss/mc:
total 33900
-rw-rw-r-- 1 root root  6940392 2026-03-01 10:38 group
`,
	}
	var report ReportData
	analyzePackages(fileMap, &report)
	analyzeSSSDFilePermissions(fileMap, &report)

	if !containsString(report.Problems, "2 files or directories in /var/lib/sss/ are incorrectly owned") {
		t.Errorf("Failed to detect exactly 2 incorrectly owned files in /var/lib/sss/ for SSSD 2.10+")
	}
}

// Test 17: Verify SLES15 SP7 Bug Detection (sssd running unprivileged triggers warning)
func TestAnalyzePermissions_SLES15SP7_UnprivilegedBug(t *testing.T) {
	fileMap := map[string]string{
		"basic-environment.txt": "PRETTY_NAME=\"SUSE Linux Enterprise Server 15 SP7\"\n",
		"rpm.txt":               "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt": `/var/lib/sss/:
total 20
drwx------ 2 sssd sssd   40 2026-03-01 10:39 db
drwxr-xr-x 2 sssd sssd 4096 2026-03-01 10:28 mc

/var/lib/sss/mc:
total 33900
-rw-rw-r-- 1 sssd sssd  6940392 2026-03-01 10:38 group
`,
	}
	var report ReportData
	analyzeOSAndHardware(fileMap, &report)
	analyzePackages(fileMap, &report)
	analyzeSSSDFilePermissions(fileMap, &report)

	if !containsString(report.Problems, "it is recommended to upgrade to a version of sssd later than 2.10.2-150700.9.17.1") {
		t.Errorf("Failed to detect SLES15 SP7 unprivileged SSSD regression warning.")
	}
}

// Test 18: Verify SLES15 SP7 Valid Configuration (Reverted to root:root correctly)
func TestAnalyzePermissions_SLES15SP7_Valid(t *testing.T) {
	fileMap := map[string]string{
		"basic-environment.txt": "PRETTY_NAME=\"SUSE Linux Enterprise Server 15 SP7\"\n",
		"rpm.txt":               "sssd-2.10.2-150700.4.1.x86_64\n",
		"sssd.txt": `/var/lib/sss/:
total 20
drwx------ 2 root root   40 2026-03-01 10:39 db
drwxr-xr-x 2 root root 4096 2026-03-01 10:28 mc

/var/lib/sss/mc:
total 33900
-rw-rw-r-- 1 root root  6940392 2026-03-01 10:38 group
`,
	}
	var report ReportData
	analyzeOSAndHardware(fileMap, &report)
	analyzePackages(fileMap, &report)
	analyzeSSSDFilePermissions(fileMap, &report)

	if containsString(report.Problems, "incorrectly owned") {
		t.Errorf("SLES15 SP7 with root:root should be valid, but raised an error.")
	}
}
