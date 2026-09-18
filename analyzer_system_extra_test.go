// analyzer_system_extra_test.go
//
// Pin-down tests for the Phase 1 system analyzers that still had low
// coverage: analyzeBasicHealth, analyzeSCC, analyzePerformance (sar/iowait
// branch) and analyzeMACDenials (evidence collection).
package main

import (
	"os"
	"testing"
)

func TestAnalyzeBasicHealth_SupportCaseID(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"basic-health-check.txt": "SR#: 12345678901\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeBasicHealth(dir, &report)
	if report.SupportCaseID != "12345678901" {
		t.Errorf("expected SupportCaseID 12345678901, got %q", report.SupportCaseID)
	}
}

func TestAnalyzeSCC_RegisteredAndUnknown(t *testing.T) {
	registered := setupMockDir(t, map[string]string{
		"updates.txt": "Status: Active\n",
	})
	defer os.RemoveAll(registered)
	var r1 ReportData
	analyzeSCC(registered, &r1)
	if r1.SCCStatus != "Registered" {
		t.Errorf("expected Registered, got %q", r1.SCCStatus)
	}

	unknown := setupMockDir(t, map[string]string{})
	defer os.RemoveAll(unknown)
	var r2 ReportData
	analyzeSCC(unknown, &r2)
	if r2.SCCStatus != "Not Registered or Unknown" {
		t.Errorf("expected Not Registered or Unknown, got %q", r2.SCCStatus)
	}
}

func TestAnalyzePerformance_SarHighIOWait(t *testing.T) {
	// Fields: like `sar -u` output; the parser reads field len-3 as %iowait.
	dir := setupMockDir(t, map[string]string{
		"sar.txt": "12:00:01 AM     all      5.00      0.00      3.00     35.50      0.00     56.50\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzePerformance(dir, &report)
	if !containsString(report.Warnings, "[PERFORMANCE] High CPU %iowait") {
		t.Errorf("expected the high-iowait warning, got %v", report.Warnings)
	}
}

func TestAnalyzePerformance_SarLowIOWaitSilent(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sar.txt": "12:00:01 AM     all      5.00      0.00      3.00      2.10      0.00     89.90\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzePerformance(dir, &report)
	if containsString(report.Warnings, "%iowait") {
		t.Errorf("low iowait must stay silent, got %v", report.Warnings)
	}
}

func TestAnalyzeMACDenials_CollectsEvidence(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"security-selinux.txt": "SELinux status: enabled\ntype=AVC msg=audit(1): avc: denied { read } for sssd comm=\"sssd\" name=\"cache\"\ntype=AVC msg=audit(2): avc: denied { write } for sssd comm=\"sssd_be\"\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	report.MACType = "SELinux"
	analyzeMACDenials(dir, &report)
	if len(report.MACDenialExamples) != 2 {
		t.Fatalf("expected 2 denial examples, got %v", report.MACDenialExamples)
	}
	if !containsString(report.Warnings, "denials detected for SSSD") {
		t.Errorf("expected the MAC denial warning, got %v", report.Warnings)
	}
}

func TestAnalyzeMACDenials_NoDenialsSilent(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"security-selinux.txt": "SELinux status: enabled\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	report.MACType = "SELinux"
	analyzeMACDenials(dir, &report)
	if len(report.MACDenialExamples) != 0 || len(report.Warnings) != 0 {
		t.Errorf("clean MAC log must stay silent: %v / %v", report.MACDenialExamples, report.Warnings)
	}
}
