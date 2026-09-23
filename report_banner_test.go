// report_banner_test.go
//
// Regression guard for the report footer. The banner previously hardcoded
// "Created by Davide M. Puggioni with Gemini Pro - 2026", which:
//   - attributed authorship in the tool output that was not mirrored in README.md
//   - carried a year (2026) that drifted from reality
//   - duplicated the version string, allowing it to drift from `-v`
//
// The footer is now built from constants.AppVersion via toolSignature(). These
// tests are build-tag-free because buildTextReport / buildJSONReport /
// toolSignature are tag-free and shared by both binaries.
package main

import (
	"encoding/json"
	"strings"
	"testing"

	"sssd-inspector/constants"
)

// TestReportBanner_NoAnachronisticAttribution pins the footer: it must not
// carry the stale AI attribution or the future year.
func TestReportBanner_NoAnachronisticAttribution(t *testing.T) {
	report := ReportData{AppVersion: constants.AppVersion}

	out := buildTextReport(report)
	if !strings.Contains(out, toolSignature(constants.AppVersion)) {
		t.Errorf("footer must end with the current tool signature %q", toolSignature(constants.AppVersion))
	}
	for _, bad := range []string{"Gemini Pro", "with Gemini", "2026 - Released"} {
		if strings.Contains(out, bad) {
			t.Errorf("footer leaks stale attribution %q; got:\n%s", bad, out)
		}
	}
}

// TestReportBanner_VersionStaysInSync: the footer string is built from
// constants.AppVersion, so it can never drift from `-v`.
func TestReportBanner_VersionStaysInSync(t *testing.T) {
	report := ReportData{AppVersion: "99.99.99"}
	out := buildTextReport(report)
	if !strings.Contains(out, "sssd-inspector v99.99.99") {
		t.Errorf("footer does not reflect the report version: %s", out)
	}
}

// TestReportBanner_JSONVersionSynced: the JSON export must carry the same
// version string shown in the text footer.
func TestReportBanner_JSONVersionSynced(t *testing.T) {
	report := ReportData{AppVersion: constants.AppVersion}
	data, err := buildJSONReport(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if v, _ := decoded["app_version"].(string); v != constants.AppVersion {
		t.Errorf("app_version = %q, want %q", v, constants.AppVersion)
	}
}
