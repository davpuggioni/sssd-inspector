package analyzer

import (
	"flag"
	"fmt"
	"os"

	"sssd-inspector/internal/config"
	"sssd-inspector/internal/constants"
)

// HandleCommandLineArgs is the single source of truth for parsing CLI flags.
// It returns 'true' if a CLI command was executed, or 'false' if it should fallback to GUI.
func HandleCommandLineArgs(exitOnNoArgs bool) bool {
	// 1. Define all application flags in one central place
	versionShort := flag.Bool(constants.FlagVersion, false, constants.DescVersion)
	cliPath := flag.String(constants.FlagAnalyze, "", constants.DescAnalyze)
	logDir := flag.String(constants.FlagLogDir, "", constants.DescLogDir)
	txtReport := flag.Bool(constants.FlagTXT, false, constants.DescTXT)
	htmlReport := flag.Bool(constants.FlagHTML, false, constants.DescHTML)
	anonymize := flag.Bool(constants.FlagAnonymize, false, constants.DescAnonymize)
	flag.Parse()

	// 2. Action: Version check
	if *versionShort {
		fmt.Printf("%s version %s (CLI)\n", constants.AppName, constants.AppVersion)
		os.Exit(0)
	}

	// 3. Action: Log directory processing
	if *logDir != "" {
		if err := RunLogDirAnalyze(*logDir, *txtReport, *htmlReport, *anonymize); err != nil {
			fmt.Fprintf(os.Stderr, "LogDir analysis failed: %v\n", err)
			os.Exit(1)
		}
		return true
	}

	// 4. Action: Archive package processing
	if *cliPath != "" {
		if err := RunCLI(*cliPath, *txtReport, *htmlReport, *anonymize); err != nil {
			fmt.Fprintf(os.Stderr, "CLI execution failed: %v\n", err)
			os.Exit(1)
		}
		return true
	}

	// 5. Action: Positional argument fallbacks (e.g., dragging a file onto the binary)
	if flag.NArg() > 0 {
		path := flag.Arg(0)
		genTxt := *txtReport
		genHtml := *htmlReport
		genAnonymize := *anonymize

		if !genTxt && !genHtml && config.Global != nil && config.Global.CLI.DefaultGenerateBothFormats {
			genTxt = true
			genHtml = true
		}

		if err := RunCLI(path, genTxt, genHtml, genAnonymize); err != nil {
			fmt.Fprintf(os.Stderr, "CLI execution failed: %v\n", err)
			os.Exit(1)
		}
		return true
	}

	// 6. Handling the difference between the Standalone CLI and Hybrid GUI apps
	if exitOnNoArgs {
		// The standalone server binary HAS to have arguments to run
		PrintCLIUsage()
		os.Exit(0)
	}

	// No flags detected, telling the root main.go it's safe to open the GUI window
	return false
}

// PrintCLIUsage houses your central terminal help documentation
func PrintCLIUsage() {
	fmt.Printf("SSSD Inspector version %s (CLI)\n", constants.AppVersion)
	fmt.Println("Usage: sssd-inspector [options] <path>")
	fmt.Println("\nOptions:")
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagVersion, "version", constants.DescVersion)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagAnalyze, "analyze", constants.DescAnalyze)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagLogDir, "logdir", constants.DescLogDir)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagTXT, "txt", constants.DescTXT)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagHTML, "html", constants.DescHTML)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagAnonymize, "anonymize", constants.DescAnonymize)
}
