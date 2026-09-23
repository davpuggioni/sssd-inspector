// cli_flags.go
//
// Single source of truth for the command-line flag surface shared by the
// static CLI binary (main_cli.go) and the hybrid GUI/CLI binary
// (main_gui.go).
//
// History note: a `-logdir` flag (raw SSSD log directory analysis) once lived
// in main_cli.go and was silently lost during a project restructure because
// the flag surface was duplicated across two main() functions. Every flag now
// comes from this registry, and cli_flags_test.go pins the registry against
// README.md, so a user-visible flag can no longer disappear unnoticed.
package main

import (
	"flag"

	"sssd-inspector/constants"
)

// cliOptions holds the values of the flags registered by registerCLIFlags.
// The pointers are owned by the flag.FlagSet passed to that function.
type cliOptions struct {
	Version   *bool
	Analyze   *string
	LogDir    *string
	TXT       *bool
	HTML      *bool
	JSON      *bool
	Anonymize *bool

	// Compare is only registered when includeCompare is true (CLI binary);
	// it is nil otherwise, so callers must not dereference it blindly.
	Compare *string
}

// registerCLIFlags registers the shared flag surface on fs and returns the
// parsed-value holders.
//
// includeCompare adds the CLI-only -compare flag (differential analysis).
// The hybrid GUI binary does not implement that path, so it must not
// advertise the flag.
func registerCLIFlags(fs *flag.FlagSet, includeCompare bool) *cliOptions {
	opts := &cliOptions{
		Version:   fs.Bool(constants.FlagVersion, false, constants.DescVersion),
		Analyze:   fs.String(constants.FlagAnalyze, "", constants.DescAnalyze),
		LogDir:    fs.String(constants.FlagLogDir, "", constants.DescLogDir),
		TXT:       fs.Bool(constants.FlagTXT, false, constants.DescTXT),
		HTML:      fs.Bool(constants.FlagHTML, false, constants.DescHTML),
		JSON:      fs.Bool(constants.FlagJSON, false, constants.DescJSON),
		Anonymize: fs.Bool(constants.FlagAnonymize, false, constants.DescAnonymize),
	}
	if includeCompare {
		opts.Compare = fs.String(constants.FlagCompare, "", constants.DescCompare)
	}
	return opts
}
