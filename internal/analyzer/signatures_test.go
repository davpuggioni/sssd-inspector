// analyzer_logs_test.go
package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// Test: Verify KB Article Matching Engine
func TestMatchKBArticles_Success(t *testing.T) {
	// 1. Setup a temporary directory for the mock logs
	dir := setupMockDir(t, map[string]string{
		"messages": "Dec 10 12:00:00 server sssd: Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed\n",
	})
	defer os.RemoveAll(dir)

	// 2. Setup a temporary directory for the mock KB JSON files
	exePath, _ := os.Executable()
	mockKBDir := filepath.Join(filepath.Dir(exePath), "kb_articles")
	os.MkdirAll(mockKBDir, 0755)
	defer os.RemoveAll(mockKBDir) // Clean up mock KB dir after test

	// Create a mock KB Article matching TID-000020793
	mockArticle := TIDArticle{
		TIDID:       "TID-TEST-01",
		Title:       "SSSD MEMORY Keytab Error",
		Description: "Keytab is corrupted.",
		LogPatterns: []string{"Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]"},
	}
	articleBytes, _ := json.Marshal(mockArticle)
	os.WriteFile(filepath.Join(mockKBDir, "test_article.json"), articleBytes, 0644)

	var report ReportData
	matchKBArticles(dir, &report)

	if len(report.MatchedTIDs) == 0 {
		t.Fatalf("Failed to match mock KB article against log evidence")
	}

	if report.MatchedTIDs[0].TIDID != "TID-TEST-01" {
		t.Errorf("Expected TID-TEST-01, got %s", report.MatchedTIDs[0].TIDID)
	}

	if len(report.MatchedTIDs[0].Evidence) == 0 {
		t.Errorf("Failed to extract log evidence for matched KB article")
	}
}

// Test: Verify KB Article AppArmor Suppression Logic
func TestMatchKBArticles_AppArmorSuppression(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"messages": "sssd: SELinux context evaluation failed\n",
	})
	defer os.RemoveAll(dir)

	exePath, _ := os.Executable()
	mockKBDir := filepath.Join(filepath.Dir(exePath), "kb_articles")
	os.MkdirAll(mockKBDir, 0755)
	defer os.RemoveAll(mockKBDir)

	// Create a mock SELinux KB Article
	mockArticle := TIDArticle{
		TIDID:       "TID-SELINUX-TEST",
		Title:       "SELinux mapping failed",
		Description: "SELinux maps were recently updated",
		LogPatterns: []string{"SELinux context evaluation failed"},
	}
	articleBytes, _ := json.Marshal(mockArticle)
	os.WriteFile(filepath.Join(mockKBDir, "test_selinux.json"), articleBytes, 0644)

	var report ReportData
	// Simulate an AppArmor environment
	report.MACType = "AppArmor"

	matchKBArticles(dir, &report)

	// The engine should explicitly ignore SELinux TIDs when AppArmor is active
	if len(report.MatchedTIDs) > 0 {
		t.Fatalf("KB Matching Engine failed to suppress SELinux article in an AppArmor environment")
	}
}

// Test: Verify extractSection accurately pulls the sssd.conf block via streaming
func TestExtractSection_Streaming(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": `Some junk data
# /etc/sssd/sssd.conf
[domain/ad]
id_provider = ad
enumerate = false
#==[ Command ]======================================
More junk data`,
	})
	defer os.RemoveAll(dir)

	extracted := extractSection(dir, "sssd.txt", "# /etc/sssd/sssd.conf")

	if !strings.Contains(extracted, "id_provider = ad") {
		t.Errorf("Failed to extract 'id_provider = ad' from sssd.txt")
	}

	if strings.Contains(extracted, "More junk data") {
		t.Errorf("extractSection leaked data past the boundary marker (#==[)")
	}
}

// Test: Verify Chronological Timeline Extraction and Sorting
func TestAnalyzeSSSDConfigAndLogs_TimelineExtraction(t *testing.T) {
	// Create mock files with mixed timestamps out of chronological order
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": `(2026-04-01 12:00:00) [sssd] [krb5_child] service key not available
(2026-04-01 10:00:00) [sssd] [watchdog] terminated by own WATCHDOG
(2026-04-01 11:00:00) [sssd] [sysdb] database disk image is malformed
`,
		"messages": "Dec 10 12:05:00 server sssd: Preauthentication failed\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeSSSDConfigAndLogs(dir, &report)

	if len(report.Timeline) != 4 {
		t.Fatalf("Expected 4 timeline events, got %d", len(report.Timeline))
	}

	// Verify sorting applied correctly (lexicographical check)
	// Expected Order:
	// 1. "2026-04-01 10:00:00"
	// 2. "2026-04-01 11:00:00"
	// 3. "2026-04-01 12:00:00"
	// 4. "Dec 10 12:05:00"
	if report.Timeline[0].Timestamp != "2026-04-01 10:00:00" {
		t.Errorf("Timeline sorting failed, expected '2026-04-01 10:00:00' first, got '%s'", report.Timeline[0].Timestamp)
	}
	if report.Timeline[1].Timestamp != "2026-04-01 11:00:00" {
		t.Errorf("Timeline sorting failed, expected '2026-04-01 11:00:00' second, got '%s'", report.Timeline[1].Timestamp)
	}
	if report.Timeline[2].Timestamp != "2026-04-01 12:00:00" {
		t.Errorf("Timeline sorting failed, expected '2026-04-01 12:00:00' third, got '%s'", report.Timeline[2].Timestamp)
	}
	if report.Timeline[3].Timestamp != "Dec 10 12:05:00" {
		t.Errorf("Timeline sorting failed, expected 'Dec 10 12:05:00' fourth, got '%s'", report.Timeline[3].Timestamp)
	}
}
