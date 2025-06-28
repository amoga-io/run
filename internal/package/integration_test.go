package pkg

import (
	"fmt"
	"testing"

	"github.com/amoga-io/run/internal/output"
)

// IntegrationTestSuite provides a test suite for integration tests
type IntegrationTestSuite struct {
	manager *Manager
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager for integration tests: %v", err)
	}

	return &IntegrationTestSuite{
		manager: manager,
	}
}

// TestPackageLifecycle tests the complete lifecycle of a package
func TestPackageLifecycle(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	// Test package validation
	t.Run("Package Validation", func(t *testing.T) {
		testPackageValidation(t, suite.manager)
	})

	// Test dependency checking
	t.Run("Dependency Checking", func(t *testing.T) {
		testDependencyChecking(t)
	})

	// Test package installation simulation
	t.Run("Installation Simulation", func(t *testing.T) {
		testInstallationSimulation(t, suite.manager)
	})

	// Test package removal simulation
	t.Run("Removal Simulation", func(t *testing.T) {
		testRemovalSimulation(t, suite.manager)
	})
}

// testPackageValidation tests package validation functionality
func testPackageValidation(t *testing.T, manager *Manager) {
	// Test valid packages
	validPackages := []string{"python", "node", "docker", "essentials"}
	for _, pkgName := range validPackages {
		pkg, err := manager.validatePackage(pkgName)
		if err != nil {
			t.Fatalf("Failed to validate package %s: %v", pkgName, err)
		}
		if pkg.Name != pkgName {
			t.Fatalf("Expected package name %s, got %s", pkgName, pkg.Name)
		}
	}

	// Test invalid packages
	invalidPackages := []string{"nonexistent", "invalid-package", ""}
	for _, pkgName := range invalidPackages {
		_, err := manager.validatePackage(pkgName)
		if err == nil {
			t.Fatalf("Expected error for invalid package %s", pkgName)
		}
	}
}

// testDependencyChecking tests dependency checking functionality
func testDependencyChecking(t *testing.T) {
	// Test packages with dependencies
	pythonPkg, _ := GetPackage("python")
	if len(pythonPkg.Dependencies) == 0 {
		t.Fatal("Python package should have dependencies")
	}

	// Test packages without dependencies
	essentialsPkg, _ := GetPackage("essentials")
	if len(essentialsPkg.Dependencies) != 0 {
		t.Fatal("Essentials package should not have dependencies")
	}

	// Test dependency validation
	for _, dep := range pythonPkg.Dependencies {
		if dep == "" {
			t.Fatal("Dependency should not be empty")
		}
	}
}

// testInstallationSimulation tests installation simulation
func testInstallationSimulation(t *testing.T, manager *Manager) {
	// Test rollback setup
	rollbackPoint, err := manager.setupRollback("test-package")
	if err != nil {
		t.Fatalf("Failed to setup rollback: %v", err)
	}
	if rollbackPoint == nil {
		t.Fatal("Rollback point should not be nil")
	}

	// Test rollback cleanup
	manager.cleanupRollback(rollbackPoint, nil)

	// Test error handling
	manager.cleanupRollback(rollbackPoint, fmt.Errorf("test error"))
}

// testRemovalSimulation tests removal simulation
func testRemovalSimulation(t *testing.T, manager *Manager) {
	// Test package checking
	pkg := Package{
		Name:     "test",
		Commands: []string{"ls"}, // ls should exist on most systems
	}

	if !manager.isPackageInstalled(pkg) {
		t.Fatal("ls command should be available")
	}

	// Test system version detection
	version := manager.getSystemVersion("python")
	_ = version // Use version to avoid unused variable warning
}

// TestSequentialOperations tests sequential package operations
func TestSequentialOperations(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	// Test sequential execution setup
	packages := []string{"python", "node", "docker"}

	// Create a simple operation function for testing
	testOperation := func(packageName string) error {
		// Simulate a quick operation
		_, exists := GetPackage(packageName)
		if !exists {
			return fmt.Errorf("package %s not found", packageName)
		}
		return nil
	}

	// Test sequential execution
	results := ExecutePackagesSequential(suite.manager, packages, testOperation, "Testing")

	if len(results) != len(packages) {
		t.Fatalf("Expected %d results, got %d", len(packages), len(results))
	}

	// Check that all operations succeeded
	for _, result := range results {
		if !result.Success {
			t.Fatalf("Expected success for package %s", result.Name)
		}
	}
}

// TestConfigurationManagement tests configuration management
func TestConfigurationManagement(t *testing.T) {
	// Test config manager creation
	configManager, err := NewConfigManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test package config retrieval
	config, exists := configManager.GetPackageConfig("python")
	if !exists {
		t.Fatal("Python config should exist")
	}

	if config.Name != "python" {
		t.Fatalf("Expected config name 'python', got '%s'", config.Name)
	}

	// Test package config listing
	configs := configManager.ListPackageConfigs()
	if len(configs) == 0 {
		t.Fatal("Should return at least some package configs")
	}

	// Test category filtering
	devConfigs := configManager.GetPackagesByCategory("development")
	if len(devConfigs) == 0 {
		t.Fatal("Should return at least some development packages")
	}

	for _, config := range devConfigs {
		if config.Category != "development" {
			t.Fatalf("Expected category 'development', got '%s'", config.Category)
		}
	}

	// Test build intensive packages
	buildPackages := configManager.BuildIntensivePackages()
	if len(buildPackages) == 0 {
		t.Fatal("Should return at least some build intensive packages")
	}

	// Test related packages
	relatedPackages := configManager.RelatedPackages()
	if len(relatedPackages) == 0 {
		t.Fatal("Should return at least some related packages")
	}
}

// TestOutputInterface tests the output interface
func TestOutputInterface(t *testing.T) {
	// Test console output creation
	consoleOutput := output.NewConsoleOutput()
	if consoleOutput == nil {
		t.Fatal("Console output should not be nil")
	}

	// Test output methods (these should not panic)
	consoleOutput.Info("Test info message")
	consoleOutput.Success("Test success message")
	consoleOutput.Error("Test error message")
	consoleOutput.Warning("Test warning message")
	consoleOutput.Progress("Test progress message")
	consoleOutput.Section("Test Section")

	// Test summary
	successful := []string{"python", "node"}
	failed := []string{"docker"}
	consoleOutput.Summary("test", successful, failed, 3, "run test")

	// Test global output functions
	output.Info("Global info message")
	output.Success("Global success message")
	output.Error("Global error message")
	output.Warning("Global warning message")
	output.Progress("Global progress message")
	output.Section("Global Section")
	output.Summary("global", successful, failed, 3, "run global")
}

// TestErrorHandling tests error handling throughout the system
func TestErrorHandling(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	// Test validation errors
	_, err := suite.manager.validatePackage("")
	if err == nil {
		t.Fatal("Expected error for empty package name")
	}

	_, err = suite.manager.validatePackage("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent package")
	}

	// Test rollback setup errors
	// This might fail if HOME is not set, which is expected
	_, _ = suite.manager.setupRollback("test")
	// We don't check for error here as it might fail in test environment

	// Test dependency error handling
	err = suite.manager.handleDependencyError(fmt.Errorf("test error"), nil)
	if err == nil {
		t.Fatal("Expected error from handleDependencyError")
	}
}
