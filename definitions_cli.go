// definitions_cli.go
//
// The command-line face of the Definitions Studio. Both binaries — the static
// CLI (main_cli.go) and the hybrid GUI/CLI (main_gui.go) — route their
// definitions flags through runDefinitionsCommand, so the two front-ends cannot
// drift apart. That is the lesson of cli_flags.go: a duplicated dispatch once
// silently lost the -logdir flag, and the definitions surface is exactly the
// kind of feature that would have been added to one binary only.
package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"sssd-inspector/constants"
)

// definitionsAction selects which definitions command a run must perform.
type definitionsAction int

const (
	// definitionsNone means "no definitions flag was used".
	definitionsNone definitionsAction = iota
	definitionsInfo
	definitionsValidate
	definitionsTest
)

// providedFlags returns the names of the flags explicitly present in rawArgs.
// Commands whose value IS the flag (like -rules-test) need it to tell
// a "-rules-test" used with an empty value (a usage error) apart from the
// flag not being used at all.
//
// The shared flag registry is used purely as a name table and the scan is
// textual on purpose: flag.Parse stops at the first non-flag argument (the
// supportconfig path), so parsing alone would miss definitions flags written
// after it.
func providedFlags(rawArgs []string) map[string]bool {
	fs := flag.NewFlagSet("provided", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	registerCLIFlags(fs, true)

	provided := map[string]bool{}
	for _, arg := range rawArgs {
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		name := strings.TrimLeft(arg, "-")
		if i := strings.IndexByte(name, '='); i >= 0 {
			name = name[:i]
		}
		if name != "" && fs.Lookup(name) != nil {
			provided[name] = true
		}
	}
	return provided
}

// definitionsActionFromOpts maps the parsed flags onto the action to run. The
// flags are mutually exclusive and evaluated in the documented order
// (-definitions-info, then -validate-rules, then -rules-test). provided carries
// the flags explicitly present on the command line: it is the only way to tell
// a "-rules-test" used with an empty value (a usage error) from the flag
// not being used at all.
func definitionsActionFromOpts(opts *cliOptions, provided map[string]bool) (definitionsAction, string) {
	switch {
	case opts == nil:
		return definitionsNone, ""
	case opts.DefinitionsInfo != nil && *opts.DefinitionsInfo:
		return definitionsInfo, ""
	case opts.ValidateRules != nil && *opts.ValidateRules:
		return definitionsValidate, ""
	case opts.RulesTest != nil && (provided[constants.FlagRulesTest] || *opts.RulesTest != ""):
		return definitionsTest, *opts.RulesTest
	default:
		return definitionsNone, ""
	}
}

// runDefinitionsCommand executes the definitions command requested on the
// command line. It reports whether the invocation was handled and the process
// exit code: 0 success, 1 when problems were found or the command failed,
// 2 for a usage error.
//
// handled == false means "none of the definitions flags was used": the caller
// continues its normal dispatch (and, for the hybrid binary, may start the GUI).
func runDefinitionsCommand(opts *cliOptions, provided map[string]bool, stdout, stderr io.Writer) (bool, int) {
	action, target := definitionsActionFromOpts(opts, provided)
	switch action {
	case definitionsNone:
		return false, 0

	case definitionsInfo:
		fmt.Fprint(stdout, RenderDefinitionsInventory(ListDefinitions()))
		return true, 0

	case definitionsValidate:
		results, diags := ValidateDiscoveredRules()
		diags = append(diags, ValidateDiscoveredKBArticles()...)
		text, ok := RenderRuleValidation(results, diags)
		fmt.Fprint(stdout, text)
		if !ok {
			fmt.Fprintln(stderr, "Definition validation failed: the problems above are inputs the analysis would skip.")
			return true, 1
		}
		return true, 0

	case definitionsTest:
		if target == "" {
			fmt.Fprintf(stderr, "Error: -%s requires the path of a supportconfig directory or archive\n", constants.FlagRulesTest)
			return true, 2
		}
		res, err := TestRulesAgainst(target)
		if err != nil {
			fmt.Fprintf(stderr, "Rule dry-run failed: %v\n", err)
			return true, 1
		}
		fmt.Fprint(stdout, RenderRuleTest(res))
		return true, 0

	default:
		return false, 0
	}
}
