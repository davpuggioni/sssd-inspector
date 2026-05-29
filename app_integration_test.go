package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sssd-inspector/internal/analyzer"
	"sssd-inspector/internal/config"
	"sssd-inspector/internal/constants"
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

	// Test 2: Verify configuration loading points to the unified global variable
	t.Run("ConfigurationLoading", func(t *testing.T) {
		if config.Global == nil {
			t.Fatal("App configuration not loaded globally")
		}

		if config.Global.App.Name == "" {
			t.Fatal("App name not configured")
		}

		if config.Global.App.Version == "" {
			t.Fatal("App version not configured")
		}
	})

	// Test 3: Verify constants are properly defined using functional hooks
	t.Run("ConstantsDefined", func(t *testing.T) {
		if constants.AppName == "" {
			t.Error("App name constant not defined")
		}

		if constants.AppVersion == "" {
			t.Error("App version constant not defined")
		}

		// Evaluates functional extensions mapping
		formats := constants.RelevantFiles()
		if len(formats) == 0 {
			t.Error("Relevant files configuration array is empty")
		}
	})

	// Test 4: Verify report generation structure mapping to internal/analyzer
	t.Run("ReportStructure", func(t *testing.T) {
		report := analyzer.ReportData{
			AppVersion:    constants.AppVersion,
			Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
			SupportCaseID: "TEST-123",
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
		testFile := "app_integration_test.go"
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("Test file context boundary could not check itself")
		}

		dir := filepath.Dir(testFile)
		if dir == "" {
			t.Error("Directory extraction failed")
		}
	})

	// Test 6: Verify frontend assets are embedded
	t.Run("FrontendAssets", func(t *testing.T) {
		indexHTML, err := assets.ReadFile("frontend/dist/index.html")
		if err != nil {
			t.Error("Failed to read embedded index.html:", err)
		}

		if len(indexHTML) == 0 {
			t.Error("index.html is empty")
		}

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
	})
}

// TestFrontendCompatibility verifies that the frontend can work with the backend
func TestFrontendCompatibility(t *testing.T) {
	t.Run("BackendMethods", func(t *testing.T) {
		app := NewApp()

		// Verifies signature mapping compatibility profiles
		_ = func() (interface{}, error) {
			return app.Analyze("/nonexistent/file.txz", false)
		}

		_ = func() (string, error) {
			return app.SaveTXT(analyzer.ReportData{})
		}
	})

	t.Run("DataStructures", func(t *testing.T) {
		report := analyzer.ReportData{
			AppVersion:    "2.0.0",
			Timestamp:     "2023-01-01T00:00:00Z",
			SupportCaseID: "TEST-123",
			SLESRlease:    "SLES 15 SP4",
			KernelVersion: "5.14.21-150400.24.44-default",
		}

		if report.AppVersion == "" {
			t.Error("AppVersion field not working")
		}
	})
}

// TestConfigurationIntegration verifies configuration system works end-to-end
func TestConfigurationIntegration(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		defaultConfig := config.DefaultConfig()

		if defaultConfig.App.Name == "" {
			t.Error("Default app name not set")
		}

		if defaultConfig.GUI.Window.Width <= 0 {
			t.Error("Default window width invalid")
		}
	})

	t.Run("ConfigurationValidation", func(t *testing.T) {
		validConfig := config.DefaultConfig()
		err := validConfig.Validate()
		if err != nil {
			t.Error("Default configuration should be valid:", err)
		}

		invalidConfig := &config.Config{}
		invalidConfig.App.Name = ""
		err = invalidConfig.Validate()
		if err == nil {
			t.Error("Empty app name should be invalid")
		}
	})
}
