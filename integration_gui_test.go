//go:build !cli

// Package main provides GUI-only integration tests for SSSD Inspector.
// These tests require the Wails framework and embedded frontend assets,
// which are only available in the GUI build (without the cli build tag).
package main

import (
	"testing"
)

// TestBackendFrontendIntegrationGUI verifies GUI-specific backend components
func TestBackendFrontendIntegrationGUI(t *testing.T) {
	// Test 1: Verify App struct can be created
	t.Run("AppCreation", func(t *testing.T) {
		app := NewApp()
		if app == nil {
			t.Fatal("Failed to create App instance")
		}
	})

	// Test 2: Verify frontend assets are embedded
	t.Run("FrontendAssets", func(t *testing.T) {
		// Try to read a known frontend file
		indexHTML, err := assets.ReadFile("frontend/dist/index.html")
		if err != nil {
			t.Error("Failed to read embedded index.html:", err)
		}

		if len(indexHTML) == 0 {
			t.Error("index.html is empty")
		}

		// Dynamically discover the JS file in the assets directory
		// to avoid hardcoding content-hash filenames that change on rebuild
		dirEntries, err := assets.ReadDir("frontend/dist/assets")
		if err != nil {
			t.Fatal("Failed to read embedded assets directory:", err)
		}

		var jsFile []byte
		for _, entry := range dirEntries {
			name := entry.Name()
			if len(name) > 3 && name[len(name)-3:] == ".js" {
				jsFile, err = assets.ReadFile("frontend/dist/assets/" + name)
				if err == nil {
					break
				}
			}
		}

		if jsFile == nil {
			t.Error("Failed to find or read any .js file in embedded assets")
		}

		if len(jsFile) == 0 {
			t.Error("JS file is empty")
		}
	})
}

// TestFrontendCompatibilityGUI verifies GUI-specific frontend-backend integration
func TestFrontendCompatibilityGUI(t *testing.T) {
	t.Run("BackendMethods", func(t *testing.T) {
		app := NewApp()

		// These methods should be callable from frontend
		// We can't check for nil on methods, but we can verify they exist
		// by checking their signatures (compile-time check)

		// Test that Analyze method exists
		_ = func() (interface{}, error) {
			return app.Analyze("/nonexistent/file.txz", false)
		}

		// Test that SaveTXT method exists
		_ = func() (string, error) {
			return app.SaveTXT(ReportData{})
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		app := NewApp()

		// Test with invalid file path
		_, err := app.Analyze("/nonexistent/file.txz", false)
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}
