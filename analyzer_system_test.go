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

	if !containsString(report.Warnings, "[PERFORMANCE] vm.dirty_bytes is set to 0") {
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

	if !containsString(report.Warnings, "System is using a short hostname instead of an FQDN") {
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

// Test: Verify invalid /var/lib/sss/ ownership for SSSD 2.10+ (message temporarily disabled)
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

	if containsString(report.Problems, "files or directories in /var/lib/sss/ are incorrectly owned") {
		t.Errorf("Expected no /var/lib/sss/ ownership message for SSSD 2.10+ (currently disabled), but one was emitted.")
	}
}

// Test: Verify SLES15 SP7 sssd-owned /var/lib/sss/ emits no ownership message (currently disabled)
func TestAnalyzePermissions_SLES15SP7_SssdOwnership(t *testing.T) {
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

	if containsString(report.Problems, "files or directories in /var/lib/sss/ are incorrectly owned") {
		t.Errorf("SLES15 SP7 should not emit the /var/lib/sss/ ownership message (currently disabled), but one was emitted.")
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

// Test: Verify SLES15 SP7 with the OFFICIAL SUSE documented ownership
// (root:sssd, 0640) is treated as VALID and emits no permission error.
func TestAnalyzePermissions_SLES15SP7_OfficialRootSssd_Valid(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"basic-environment.txt": "PRETTY_NAME=\"SUSE Linux Enterprise Server 15 SP7\"\n",
		"rpm.txt":               "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt":              "-rw-r----- 1 root sssd 1024 Mar 10 10:00 sssd.conf\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeOSAndHardware(dir, &report)
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if containsString(report.Problems, "Please change the permissions.") {
		t.Errorf("root:sssd with 0640 is the official SUSE-documented configuration and must NOT raise a permission error; got: %v", report.Problems)
	}
	if containsString(report.Problems, "sssd.conf") {
		t.Errorf("no sssd.conf permission problem expected for root:sssd 0640 on SP7; got: %v", report.Problems)
	}
}

// Test: Verify SLES15 SP7 sssd.conf ownership message (buggy build) suggests changing
// permissions AND updating sssd when the 2.10.2-150700.9.17.1 package is installed
// and the ownership is truly invalid (sssd:sssd).
func TestAnalyzePermissions_SLES15SP7_ConfOwnership_BuggyBuild(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"basic-environment.txt": "PRETTY_NAME=\"SUSE Linux Enterprise Server 15 SP7\"\n",
		"rpm.txt":               "sssd-2.10.2-150700.9.17.1.x86_64\n",
		"sssd.txt":              "-rw-r----- 1 sssd sssd 1024 Mar 10 10:00 sssd.conf\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeOSAndHardware(dir, &report)
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if !containsString(report.Problems, "Please change the permissions.") {
		t.Errorf("Expected 'Please change the permissions.' message for SLES15 SP7 sssd.conf ownership.")
	}
	if !containsString(report.Problems, "sssd 2.10.2-150700.9.17.1 is installed, please update sssd to the most recent package.") {
		t.Errorf("Expected update-sssd notification for buggy SLES15 SP7 build 2.10.2-150700.9.17.1.")
	}
}

// Test: Verify SLES15 SP7 sssd.conf ownership message (non-buggy build) only suggests
// changing permissions, without the update-sssd notification. The truly invalid
// case is sssd:sssd ownership.
func TestAnalyzePermissions_SLES15SP7_ConfOwnership_NonBuggyBuild(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"basic-environment.txt": "PRETTY_NAME=\"SUSE Linux Enterprise Server 15 SP7\"\n",
		"rpm.txt":               "sssd-2.10.2-150700.4.1.x86_64\n",
		"sssd.txt":              "-rw-r----- 1 sssd sssd 1024 Mar 10 10:00 sssd.conf\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeOSAndHardware(dir, &report)
	analyzePackages(dir, &report)
	analyzeSSSDFilePermissions(dir, &report)

	if !containsString(report.Problems, "Please change the permissions.") {
		t.Errorf("Expected 'Please change the permissions.' message for SLES15 SP7 sssd.conf ownership.")
	}
	if containsString(report.Problems, "2.10.2-150700.9.17.1 is installed, please update sssd to the most recent package.") {
		t.Errorf("Non-buggy SLES15 SP7 build should not emit the update-sssd notification.")
	}
}
