// main_test.go
//
// TestMain isolates every test from the machine's real definition roots:
// without this, a developer or CI host with rules in /etc/sssd-inspector or
// ~/.sssd-inspector would inject extra findings into tests that count them.
// Individual tests that exercise the user/system roots re-point the vars
// themselves (no test in this package runs in parallel, so mutating them is
// safe).
package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	userDefinitionsRoot = func() string { return "" }
	systemDefinitionsRoot = func() string { return "" }
	os.Exit(m.Run())
}
