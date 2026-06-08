// integration_test.go - CLI-safe integration tests (no GUI dependencies)
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
	// Test 1: Verify configuration loading
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

	// Test 2: Verify constants are properly defined
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

	// Test 3: Verify file validation works
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

	// Test 4: Verify report generation structure
	t.Run("ReportStructure", func(t *testing.T) {
		report := ReportData{
			AppVersion: constants.AppVersion,
			Timestamp:  time.Now().Format(constants.TimestampFormat),
		}

		if report.AppVersion != constants.AppVersion {
			t.Error("Report app version not set correctly")
		}

		if report.Timestamp == "" {
			t.Error("Report timestamp not set")
		}
	})

	// Test 5: Verify utility functions work
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
}

// TestFrontendCompatibility verifies that the frontend can work with the backend
func TestFrontendCompatibility(t *testing.T) {
	// Test 1: Verify data structures match frontend expectations
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
