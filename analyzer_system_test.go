// analyzer_system_test.go
package main

import (
	"os"
	"testing"
)

// Test: Verify vm.dirty_bytes performance issue
func TestAnalyzePerformance_DirtyBytes(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"memory.txt": "vm.dirty_bytes = 0\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzePerformance(dir, &report)

	if !containsString(report.Problems, "[PERFORMANCE] vm.dirty_bytes is set to 0") {
		t.Errorf("Failed to detect vm.dirty_bytes = 0 misconfiguration")
	}
}

// Test: Verify SSSD >= 2.10 file permissions (Should catch root:root as an error on standard OS)
func TestAnalyzePermissions_NewVersion(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"rpm.txt":  "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt": "-rw------- 1 root root 1024 Mar 10 10:00 sssd.conf\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if !containsString(report.Problems, "strictly requires 'root:sssd'") {
		t.Errorf("Failed to detect incorrect file ownership for SSSD 2.10+")
	}
}

// Test: Verify Disk Space warning for full partitions
func TestAnalyzeDiskSpace_Critical(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"fs-diskio.txt": "/dev/sda1  100G  98G  2G  98% /\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeDiskSpace(dir, &report)

	if !containsString(report.Problems, "[CRITICAL] Partition / is 98% full") {
		t.Errorf("Failed to detect 98%% full root partition")
	}
}

// Test: Verify Systemd Exit Code parsing (Code 4 = Config Error)
func TestAnalyzeServices_CrashCode4(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"systemd.txt": "● sssd.service - System Security Services Daemon\n   Loaded: loaded\n   Active: failed (Result: exit-code)\n  Process: 1234 ExecStart=/usr/sbin/sssd (code=exited, status=4)\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeServices(dir, &report)

	if !containsString(report.Problems, "crashed with exit code 4 (Configuration Error)") {
		t.Errorf("Failed to parse systemd exit code 4")
	}
}

// Test: Verify FQDN Hostname check
func TestAnalyzeHostname_ShortName(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"basic-environment.txt": "Hostname: linux-server\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeHostnameAndFQDN(dir, &report)

	if !containsString(report.Problems, "System is using a short hostname instead of an FQDN") {
		t.Errorf("Failed to detect short hostname")
	}
}

// Test: Verify AppArmor MAC Status extraction
func TestAnalyzeMACStatus_AppArmor(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"security-apparmor.txt": "apparmor module is loaded.\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeMACStatus(dir, &report)

	if report.MACType != "AppArmor" {
		t.Errorf("Expected MACType 'AppArmor', got '%s'", report.MACType)
	}
}

// Test: Verify SELinux MAC Status extraction
func TestAnalyzeMACStatus_SELinux(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"security-selinux.txt": "SELinux status:                 enabled\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeMACStatus(dir, &report)

	if report.MACType != "SELinux" {
		t.Errorf("Expected MACType 'SELinux', got '%s'", report.MACType)
	}
}

// Test: Verify valid /var/lib/sss/ ownership for SSSD 2.10+ (Standard Behavior)
func TestAnalyzePermissions_VarLibSss_SSSD210_Valid(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"rpm.txt": "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt": `/var/lib/sss/:
total 20
drwx------ 2 sssd sssd   40 2026-03-01 10:39 db
drwxr-xr-x 2 sssd sssd 4096 2026-03-01 10:28 mc

/var/lib/sss/mc:
total 33900
-rw-rw-r-- 1 sssd sssd  6940392 2026-03-01 10:38 group
`,
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if containsString(report.Problems, "/var/lib/sss/") {
		t.Errorf("Expected no /var/lib/sss/ permissions errors for perfectly owned directories, but found one.")
	}
}

// Test: Verify invalid /var/lib/sss/ ownership for SSSD 2.10+ (Standard Behavior)
func TestAnalyzePermissions_VarLibSss_SSSD210_Invalid(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"rpm.txt": "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt": `/var/lib/sss/:
total 20
drwx------ 2 sssd sssd   40 2026-03-01 10:39 db
drwxr-xr-x 2 root root 4096 2026-03-01 10:28 mc

/var/lib/sss/mc:
total 33900
-rw-rw-r-- 1 root root  6940392 2026-03-01 10:38 group
`,
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if !containsString(report.Problems, "2 files or directories in /var/lib/sss/ are incorrectly owned") {
		t.Errorf("Failed to detect exactly 2 incorrectly owned files in /var/lib/sss/ for SSSD 2.10+")
	}
}

// Test: Verify SLES15 SP7 Bug Detection (sssd running unprivileged triggers warning)
func TestAnalyzePermissions_SLES15SP7_UnprivilegedBug(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
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
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeOSAndHardware(dir, &report)
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if !containsString(report.Problems, "upgrade to a version of sssd later than") && !containsString(report.Problems, "greater than version sssd-2.10.2") {
		t.Errorf("Failed to detect SLES15 SP7 unprivileged SSSD regression warning.")
	}
}

// Test: Verify SLES15 SP7 Valid Configuration (Reverted to root:root correctly)
func TestAnalyzePermissions_SLES15SP7_Valid(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
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
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeOSAndHardware(dir, &report)
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if containsString(report.Problems, "incorrectly owned") {
		t.Errorf("SLES15 SP7 with root:root should be valid, but raised an error.")
	}
}
