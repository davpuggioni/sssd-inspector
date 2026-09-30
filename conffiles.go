// conffiles.go
//
// Locating /etc files inside a supportconfig bundle.
//
// supportconfig never ships an /etc file as a standalone file: its conf_files()
// helper (supportconfig.rc) appends every file to a per-topic bundle preceded by
// a marker line:
//
//	# /etc/hosts                     -> the file content follows
//	# /etc/hosts - File not found    -> the file did not exist on the host
//
// /etc/hosts in particular is written by net_info_namespace() into network.txt
// (etc_info() only collects files matching *conf/*cfg, so /etc/hosts never
// reaches etc.txt). That is why looking for a file literally named "hosts"
// finds nothing in a real supportconfig.
//
// extractSection() alone cannot tell "this file was absent on the analysed
// host" from "this supportconfig does not carry the section at all", because it
// discards everything that follows the marker on the same line. locateConfFile
// closes that gap by reporting the three distinct states below.
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// ConfFileStatus describes how an /etc file was found (or not) in a
// supportconfig.
type ConfFileStatus int

const (
	// ConfAbsent: no marker anywhere — the supportconfig does not carry this
	// file, so the tool must not claim anything about the analysed host.
	ConfAbsent ConfFileStatus = iota
	// ConfMissingHost: the bundle records "# <path> - File not found", i.e. the
	// file genuinely did not exist on the analysed host.
	ConfMissingHost
	// ConfPresent: the marker was found and the content follows it.
	ConfPresent
)

// ConfFileNotFoundSuffix is the suffix supportconfig appends to the marker of a
// file that does not exist on the analysed host.
const ConfFileNotFoundSuffix = "- File not found"

// hostsStatus* are the values recorded in ReportData.HostsFileStatus.
const (
	HostsStatusPresent      = "present"
	HostsStatusMissingHost  = "missing_on_host"
	HostsStatusNotCollected = "not_collected"
	// HostsStatusMalformed means the file was collected and read, but it does
	// not parse as /etc/hosts at all. It is deliberately distinct from
	// "present": a present file that simply lacks a loopback line is valid
	// syntax, and conflating the two tells the operator to look for corruption
	// that is not there.
	HostsStatusMalformed = "malformed"
)

// locateConfFile searches the given supportconfig bundles (in order) for the
// "# <path>" marker emitted by supportconfig's conf_files() and returns the
// block content, the status and the bundle it came from. A standalone file with
// the same base name is used as a fallback, which keeps raw directories, older
// supportconfig layouts and the unit-test fixtures working.
//
// The returned content excludes the marker line and any comment line that
// belongs to a following marker; comment lines inside the file are preserved so
// callers can decide whether to honour them.
func locateConfFile(dirPath string, bundles []string, path string) (string, ConfFileStatus, string) {
	marker := "# " + strings.TrimRight(path, "/")

	for _, bundle := range bundles {
		if bundle == "" {
			continue
		}
		var sb strings.Builder
		inSection := false
		found := false
		missing := false
		done := false

		scanFiles(dirPath, []string{bundle}, func(line string) {
			if done {
				return
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, marker) {
				rest := strings.TrimSpace(strings.TrimPrefix(trimmed, marker))
				if strings.EqualFold(rest, ConfFileNotFoundSuffix) {
					found, missing, inSection, done = true, true, false, true
					return
				}
				found, inSection = true, true
				return
			}
			if !inSection {
				return
			}
			// A new "#==[ ... ]" block or the marker of the next configuration
			// file ends this section.
			if strings.HasPrefix(trimmed, "#==[") || isConfFileMarker(trimmed) {
				inSection, done = false, true
				return
			}
			sb.WriteString(line + "\n")
		})

		if missing {
			return "", ConfMissingHost, bundle
		}
		if found {
			return sb.String(), ConfPresent, bundle
		}
	}

	// Fallback: a standalone file (raw directories, older supportconfigs, tests).
	base := filepath.Base(strings.TrimRight(path, "/"))
	if info, err := os.Stat(filepath.Join(dirPath, base)); err == nil && !info.IsDir() {
		return readFileSafe(dirPath, base), ConfPresent, base
	}
	return "", ConfAbsent, ""
}

// isConfFileMarker reports whether a line looks like the "# /etc/<file>" marker
// that opens the next configuration block inside a supportconfig bundle. Only
// comment lines that consist of a single absolute path (optionally followed by
// the "- File not found" suffix) qualify, so ordinary comments inside an /etc
// file (e.g. "# 127.0.0.1 localhost") are not mistaken for markers.
func isConfFileMarker(line string) bool {
	if !strings.HasPrefix(line, "# /") {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "# "))
	return isConfFileMarkerText(rest)
}

// isConfFileMarkerText is the marker test without the leading "# ".
func isConfFileMarkerText(rest string) bool {
	if !strings.HasPrefix(rest, "/") {
		return false
	}
	if !strings.ContainsAny(rest, " \t") {
		return true // e.g. "# /etc/hosts"
	}
	// e.g. "# /etc/hosts - File not found"
	if idx := strings.Index(rest, " "+ConfFileNotFoundSuffix); idx >= 0 {
		return strings.TrimSpace(rest[idx+1:]) == ConfFileNotFoundSuffix
	}
	return false
}
