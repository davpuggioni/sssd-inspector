// integration_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sssd-inspector/config"
	"sssd-inspector/constants"
)

// TestBackendFrontendIntegration verifies that the backend components
// work correctly and can be called from the frontend
func TestBackendFrontendIntegration(t *testing.T) {
	// Test 1: Verify App struct can be created
	t.Run("AppCreation", func(t *testing.T) {
		app := NewApp()
		if app == nil {
			t.Fatal("Failed to create App instance")
		}
	})

	// Test 2: Verify configuration loading
	t.Run("ConfigurationLoading", func(t *testing.T) {
		if appConfig == nil {
			t.Fatal("App configuration not loaded")
		}

		if appConfig.App.Name == "" {
			t.Fatal("App name not configured")
		}

		if appConfig.App.Version == "" {
			t.Fatal("App version not configured")
		}
	})

	// Test 3: Verify constants are properly defined
	t.Run("ConstantsDefined", func(t *testing.T) {
		if constants.AppName == "" {
			t.Error("App name constant not defined")
		}

		if constants.AppVersion == "" {
			t.Error("App version constant not defined")
		}

		if constants.TXZFormat == "" || constants.TarXZFormat == "" {
			t.Error("File extension constants not defined")
		}
	})

	// Test 4: Verify file validation works
	t.Run("FileValidation", func(t *testing.T) {
		// Test valid extensions
		validExts := []string{constants.TXZFormat, constants.TarXZFormat}

		// Test valid extension
		validExt := filepath.Ext("test.txz")
		isValid := false
		for _, ext := range validExts {
			if "."+ext == validExt {
				isValid = true
				break
			}
		}
		if !isValid {
			t.Error("Valid file extension not recognized")
		}

		// Test invalid extension
		invalidExt := filepath.Ext("test.txt")
		isValid = false
		for _, ext := range validExts {
			if "."+ext == invalidExt {
				isValid = true
				break
			}
		}
		if isValid {
			t.Error("Invalid file extension incorrectly accepted")
		}
	})

	// Test 5: Verify report generation structure
	t.Run("ReportStructure", func(t *testing.T) {
		report := ReportData{
			AppVersion:    constants.AppVersion,
			Timestamp:     time.Now().Format(constants.TimestampFormat),
			SupportCaseID: "TEST-123",
		}

		if report.AppVersion != constants.AppVersion {
			t.Error("Report app version not set correctly")
		}

		if report.Timestamp == "" {
			t.Error("Report timestamp not set")
		}
	})

	// Test 6: Verify utility functions work
	t.Run("UtilityFunctions", func(t *testing.T) {
		// Test file existence check
		testFile := "integration_test.go"
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("Test file should exist")
		}

		// Test path operations
		dir := filepath.Dir(testFile)
		if dir == "" {
			t.Error("Directory extraction failed")
		}

		base := filepath.Base(testFile)
		if base != "integration_test.go" {
			t.Error("Base name extraction failed")
		}
	})

	// Test 7: Verify frontend assets are embedded
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

// TestFrontendCompatibility verifies that the frontend can work with the backend
func TestFrontendCompatibility(t *testing.T) {
	// Test 1: Verify all backend methods that frontend expects exist
	t.Run("BackendMethods", func(t *testing.T) {
		app := NewApp()

		// These methods should be callable from frontend
		// We can't check for nil on methods, but we can verify they exist by calling them with safe parameters
		// OpenFileBrowser requires runtime context, so we can't test it directly here

		// Test that Analyze method exists by checking its signature
		// This will fail to compile if method doesn't exist
		_ = func() (interface{}, error) {
			return app.Analyze("/nonexistent/file.txz", false)
		}

		// Test that SaveTXT method exists
		_ = func() (string, error) {
			return app.SaveTXT(ReportData{})
		}
	})

	// Test 2: Verify data structures match frontend expectations
	t.Run("DataStructures", func(t *testing.T) {
		report := ReportData{
			AppVersion:    "2.0.0",
			Timestamp:     "2023-01-01T00:00:00Z",
			SupportCaseID: "TEST-123",
			SLESRlease:    "SLES 15 SP4",
			KernelVersion: "5.14.21-150400.24.44-default",
		}

		// Verify JSON tags match frontend expectations
		if report.AppVersion == "" {
			t.Error("AppVersion field not working")
		}

		if report.Timestamp == "" {
			t.Error("Timestamp field not working")
		}

		if report.SLESRlease == "" {
			t.Error("SLESRlease field not working")
		}
	})

	// Test 3: Verify error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		app := NewApp()

		// Test with invalid file path
		_, err := app.Analyze("/nonexistent/file.txz", false)
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}

// TestConfigurationIntegration verifies configuration system works end-to-end
func TestConfigurationIntegration(t *testing.T) {
	// Test 1: Verify default configuration
	t.Run("DefaultConfiguration", func(t *testing.T) {
		defaultConfig := config.DefaultConfig()

		if defaultConfig.App.Name == "" {
			t.Error("Default app name not set")
		}

		if defaultConfig.App.Version == "" {
			t.Error("Default app version not set")
		}

		if defaultConfig.GUI.Window.Width <= 0 {
			t.Error("Default window width invalid")
		}

		if defaultConfig.GUI.Window.Height <= 0 {
			t.Error("Default window height invalid")
		}
	})

	// Test 2: Verify configuration validation
	t.Run("ConfigurationValidation", func(t *testing.T) {
		validConfig := config.DefaultConfig()
		err := validConfig.Validate()
		if err != nil {
			t.Error("Default configuration should be valid:", err)
		}

		// Test invalid configuration
		invalidConfig := &config.Config{}
		invalidConfig.App.Name = ""
		err = invalidConfig.Validate()
		if err == nil {
			t.Error("Empty app name should be invalid")
		}
	})

	// Test 3: Verify configuration loading with non-existent path
	t.Run("ConfigurationLoading", func(t *testing.T) {
		// When a specific path is provided and the file doesn't exist,
		// LoadConfig should return an error and nil config.
		loadedConfig, err := config.LoadConfig("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Expected error for non-existent config path")
		}
		if loadedConfig != nil {
			t.Error("Expected nil config when loading specific non-existent path")
		}

		// When no path is provided, LoadConfig should return defaults
		defaultConfig, err := config.LoadConfig("")
		if err != nil {
			t.Error("Expected no error when loading default config, got:", err)
		}
		if defaultConfig == nil {
			t.Error("Expected default config when loading empty path")
		}
	})
}
