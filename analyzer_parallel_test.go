// analyzer_parallel_test.go
//
// Regression tests for the parallel phase helpers (analyzer_parallel.go, 0%
// coverage). These functions duplicate the sequential phase sequence used by
// analyzeData but run it concurrently; this test pins their equivalence so
// the two paths cannot silently diverge (the same class of drift that lost
// the -logdir feature). Also exercises them under -race.
package main

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

// parallelFixture returns a supportconfig payload that makes as many Phase
// 1/2 analyzers emit findings as possible, so the equivalence check is
// meaningful (not vacuously equal on empty reports).
func parallelFixture(t *testing.T) string {
	t.Helper()
	dir := setupMockDir(t, map[string]string{
		"sssd.conf":             "[domain/example.com]\nid_provider = ad\nad_domain = example.com\nenumerate = true\nuse_fully_qualified_names = True\n",
		"sssd.txt":              "-rw------- 1 root root 1024 Mar 10 10:00 sssd.conf\nDec 10 12:00:00 server sssd: Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed\n",
		"basic-environment.txt": "Hostname: host.example.com\n",
		"systemd.txt":           "\u25cf sssd.service - System Security Services Daemon\n   Active: active (running)\n",
		"nsswitch.conf":         "passwd: files sss\ngroup: sss files\n",
		"pam.txt":               "auth sufficient pam_sss.so\n",
		"nscd.conf":             "enable-cache passwd yes\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
		"fs-diskio.txt":         "/dev/sda1  100G  98G  2G  98% /\n",
		"memory.txt":            "vm.dirty_bytes = 0\n",
		"ntp.txt":               "System time     : 12.000 seconds fast of NTP time\nSystem clock synchronized: yes\n",
		"security-selinux.txt":  "SELinux status:                 enabled\n",
		"hosts":                 "127.0.0.1 localhost\n10.0.0.1 ad.example.com ad\n",
	})
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// seedReport mirrors the defaults analyzeData sets before the phases, so both
// runs start from identical baselines.
func seedReport() ReportData {
	return ReportData{
		SssdService:    "Not Running / Unknown",
		WinbindService: "Not Running / Unknown",
		NscdStatus:     "Not Running / Unknown",
		TimeService:    "Not Running / Unknown",
		KerberosRealm:  "Not configured",
	}
}

// runSequentialPhases executes the exact Phase 1 + Phase 2 call sequence of
// analyzeData (analyzer_core.go). If that sequence changes, this helper must
// be updated together with it — the compiler will not catch drift here, the
// reviewer must.
func runSequentialPhases(dirPath string, report *ReportData) {
	analyzeBasicHealth(dirPath, report)
	analyzeOSAndHardware(dirPath, report)
	analyzeHostnameAndFQDN(dirPath, report)
	analyzeSCC(dirPath, report)
	analyzeDNS(dirPath, report)
	analyzeTime(dirPath, report)
	analyzePerformance(dirPath, report)

	analyzePAM(dirPath, report)
	analyzeNSSwitch(dirPath, report)
	analyzeHosts(dirPath, report)
	analyzeNSCD(dirPath, report)
	analyzePackages(dirPath, report)
	analyzeSSSDVersionAge(report)
	analyzeServices(dirPath, report)
	analyzeMACStatus(dirPath, report)
	analyzeSSSDFilePermissions(dirPath, report)
	analyzeDiskSpace(dirPath, report)
}

// comparablePhase1_2 extracts the fields touched by Phase 1/2 analyzers so
// the equivalence check ignores later phases (log scanning, correlation,
// summary, graph).
type comparablePhase1_2 struct {
	Problems      []string
	Warnings      []string
	SssdInstalled bool
	NsswitchValid bool
	PamSss        bool
	PamGDPR       bool
	Realm         string
	Hostname      string
	AdDomain      string
	ADProvider    bool
	Keytab        bool
}

func snapshotPhases(r *ReportData) comparablePhase1_2 {
	p := append([]string{}, r.Problems...)
	w := append([]string{}, r.Warnings...)
	sort.Strings(p)
	sort.Strings(w)
	return comparablePhase1_2{
		Problems: p, Warnings: w,
		SssdInstalled: r.SssdInstalled,
		NsswitchValid: r.NsswitchValid,
		PamSss:        r.PamSssInstalled,
		PamGDPR:       r.PamGDPRRestricted,
		Realm:         r.KerberosRealm,
		Hostname:      r.Hostname,
		AdDomain:      r.AdDomain,
		ADProvider:    r.ADProviderMode,
		Keytab:        r.KeytabFound,
	}
}

// TestParallelPhases_EquivalentToSequential is the central regression guard
// for analyzer_parallel.go: concurrent and sequential execution must produce
// identical findings.
func TestParallelPhases_EquivalentToSequential(t *testing.T) {
	dir := parallelFixture(t)

	seq := seedReport()
	runSequentialPhases(dir, &seq)
	if len(seq.Problems)+len(seq.Warnings) == 0 {
		t.Fatalf("fixture produced no findings; the equivalence check would be vacuous")
	}

	par := seedReport()
	analyzePhase1Parallel(dir, &par)
	analyzePhase2Parallel(dir, &par)

	if got, want := snapshotPhases(&par), snapshotPhases(&seq); !reflect.DeepEqual(got, want) {
		t.Errorf("parallel phases diverged from sequential phases:\n got=%#v\nwant=%#v", got, want)
	}
}

// TestParallelPhases_Repeatable excludes order-dependent flakiness of the
// goroutine scheduling (3 consecutive runs must agree).
func TestParallelPhases_Repeatable(t *testing.T) {
	dir := parallelFixture(t)
	var first comparablePhase1_2
	for i := 0; i < 3; i++ {
		r := seedReport()
		analyzePhase1Parallel(dir, &r)
		analyzePhase2Parallel(dir, &r)
		snap := snapshotPhases(&r)
		if i == 0 {
			first = snap
			continue
		}
		if !reflect.DeepEqual(snap, first) {
			t.Fatalf("parallel run %d disagrees with run 0: nondeterministic findings", i)
		}
	}
}
