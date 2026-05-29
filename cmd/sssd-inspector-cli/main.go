package main

import "sssd-inspector/internal/analyzer"

func main() {
	// Run the centralized engine. Force a exit/help screen if invoked with empty inputs.
	analyzer.HandleCommandLineArgs(true)
}
