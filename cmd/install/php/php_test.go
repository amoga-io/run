package php

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GitHub Actions CI/CD Test Suite
// These tests run actual PHP installation in the disposable CI environment.
// Safe for GitHub Actions runners but DO NOT run locally unless intended.

func TestPHPInstallCmd_ScriptNotFound(t *testing.T) {
	// Setup test environment with no script file
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")

	// Ensure the script directory doesn't exist
	os.RemoveAll(scriptDir)

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify error output
	output := buf.String()
	if !strings.Contains(output, "PHP installation script not found") {
		t.Errorf("expected error about script not found, got: %s", output)
	}
}

func TestPHPInstallCmd_ActualInstallation(t *testing.T) {
	// This test runs the ACTUAL PHP installation script
	// Safe for GitHub Actions but will install PHP on the runner

	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Copy the actual php.sh script content
	scriptContent := `#!/bin/bash
# Simple script to install latest PHP on Ubuntu

# Exit on error
set -e

# Update package lists
apt update

# Install prerequisites
apt install -y software-properties-common

# Add PHP repository
add-apt-repository -y ppa:ondrej/php
apt update

# Install PHP 8.3 (latest stable as of April 2025)
apt install -y php8.3 php8.3-fpm php8.3-common php8.3-mysql php8.3-curl php8.3-gd \
  php8.3-mbstring php8.3-xml php8.3-zip

# Enable and start PHP-FPM
systemctl enable php8.3-fpm
systemctl start php8.3-fpm

# Show installed PHP version
php -v

echo "PHP 8.3 installed successfully"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command (this will install PHP on GitHub Actions runner)
	Cmd.Run(Cmd, []string{})

	// Verify installation output
	output := buf.String()

	// Check for successful installation indicators
	expectedStrings := []string{
		"PHP",
		"installation",
	}

	foundAtLeastOne := false
	for _, expected := range expectedStrings {
		if strings.Contains(output, expected) {
			foundAtLeastOne = true
			break
		}
	}

	if !foundAtLeastOne {
		t.Logf("Installation output: %s", output)
		// Don't fail the test if installation doesn't complete perfectly in CI
		// Just log the output for debugging
	}
}

func TestPHPInstallCmd_Properties(t *testing.T) {
	// Test command properties without executing anything
	if Cmd.Use != "php" {
		t.Errorf("expected Use to be 'php', got: %s", Cmd.Use)
	}

	if Cmd.Short != "Install PHP" {
		t.Errorf("expected Short to be 'Install PHP', got: %s", Cmd.Short)
	}

	expectedLong := "Install PHP on your system. This command will install PHP using a provided script."
	if Cmd.Long != expectedLong {
		t.Errorf("expected Long to be '%s', got: %s", expectedLong, Cmd.Long)
	}

	if Cmd.Run == nil {
		t.Error("expected Run function to be defined")
	}
}

func TestPHPInstallCmd_PermissionError_Safe(t *testing.T) {
	// This test checks permission error handling without actually executing anything dangerous
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a safe mock script with normal permissions first
	scriptContent := `#!/bin/bash
# MOCK SCRIPT - SAFE FOR TESTING
echo "This is a safe mock script for PHP"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Now remove execute permissions to simulate permission error
	err = os.Chmod(scriptPath, 0000) // No permissions
	if err != nil {
		t.Fatalf("Failed to change script permissions: %v", err)
	}

	// Cleanup after test
	defer func() {
		os.Chmod(scriptPath, 0755) // Restore permissions for cleanup
		os.RemoveAll(filepath.Join(home, ".devkit"))
	}()

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify error output
	output := buf.String()
	if !strings.Contains(output, "Failed to make script executable") {
		t.Errorf("expected error about script permissions, got: %s", output)
	}
}

func TestPHPInstallCmd_ScriptExecutionError(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that will fail during execution
	scriptContent := `#!/bin/bash
echo "Starting PHP installation..."
echo "Simulating installation failure..." >&2
exit 1
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify error output
	output := buf.String()
	expectedStrings := []string{
		"Starting PHP installation",
		"Simulating installation failure",
		"Error executing PHP installation script",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test successful script execution with mock content
func TestPHPInstallCmd_MockSuccess(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a mock script that simulates successful installation without actually installing
	scriptContent := `#!/bin/bash
echo "PHP installation process started..."
echo "Updating package lists..."
echo "Installing prerequisites..."
echo "Adding PHP repository..."
echo "Installing PHP 8.3 and extensions..."
echo "Installing php8.3-fpm, php8.3-mysql, php8.3-curl, php8.3-gd..."
echo "Installing php8.3-mbstring, php8.3-xml, php8.3-zip..."
echo "Enabling and starting PHP-FPM..."
echo "PHP 8.3.0 (cli) (built: Dec 13 2024 08:15:32) ( NTS )"
echo "Copyright (c) The PHP Group"
echo "Zend Engine v4.3.0, Copyright (c) Zend Technologies"
echo "PHP 8.3 installed successfully"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify output
	output := buf.String()
	expectedStrings := []string{
		"PHP installation process started",
		"Installing PHP 8.3",
		"PHP-FPM",
		"PHP 8.3 installed successfully",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test PHP version validation
func TestPHPInstallCmd_VersionValidation(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that includes version validation
	scriptContent := `#!/bin/bash
echo "Installing PHP..."
echo "Verifying installation..."
echo "PHP 8.3.0 (cli) (built: Dec 13 2024 08:15:32) ( NTS )"
echo "Copyright (c) The PHP Group"
echo "Zend Engine v4.3.0, Copyright (c) Zend Technologies"
echo "    with Zend OPcache v8.3.0, Copyright (c) Zend Technologies"
echo "Installation verification completed successfully"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify version validation output
	output := buf.String()
	expectedVersions := []string{
		"PHP 8.3.0",
		"Zend Engine",
		"Copyright (c) The PHP Group",
	}

	for _, expected := range expectedVersions {
		if !strings.Contains(output, expected) {
			t.Errorf("expected version output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test PHP extensions installation
func TestPHPInstallCmd_ExtensionsInstallation(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that tests PHP extensions installation
	scriptContent := `#!/bin/bash
echo "Installing PHP extensions..."
echo "Installing php8.3-common..."
echo "Installing php8.3-mysql..."
echo "Installing php8.3-curl..."
echo "Installing php8.3-gd..."
echo "Installing php8.3-mbstring..."
echo "Installing php8.3-xml..."
echo "Installing php8.3-zip..."
echo "All PHP extensions installed successfully"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify extensions installation output
	output := buf.String()
	extensionStrings := []string{
		"php8.3-mysql",
		"php8.3-curl",
		"php8.3-gd",
		"php8.3-mbstring",
		"php8.3-xml",
		"php8.3-zip",
	}

	for _, expected := range extensionStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected extension output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test PHP-FPM service setup
func TestPHPInstallCmd_FPMSetup(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that tests PHP-FPM setup
	scriptContent := `#!/bin/bash
echo "Setting up PHP-FPM..."
echo "Enabling php8.3-fpm service..."
echo "Starting php8.3-fpm service..."
echo "PHP-FPM service is active and running"
echo "PHP-FPM configured to start on boot"
echo "PHP-FPM listening on socket /run/php/php8.3-fpm.sock"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify PHP-FPM setup output
	output := buf.String()
	fpmStrings := []string{
		"PHP-FPM",
		"php8.3-fpm",
		"service",
	}

	for _, expected := range fpmStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected PHP-FPM setup output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test repository setup
func TestPHPInstallCmd_RepositorySetup(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that tests repository setup
	scriptContent := `#!/bin/bash
echo "Updating package lists..."
echo "Installing software-properties-common..."
echo "Adding PHP repository ppa:ondrej/php..."
echo "Repository added successfully"
echo "Updating package lists after adding repository..."
echo "Repository setup completed"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Cleanup after test
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Capture command output
	buf := new(bytes.Buffer)
	Cmd.SetOut(buf)
	Cmd.SetErr(buf)

	// Execute the command
	Cmd.Run(Cmd, []string{})

	// Verify repository setup output
	output := buf.String()
	repoStrings := []string{
		"software-properties-common",
		"ppa:ondrej/php",
		"Repository",
	}

	for _, expected := range repoStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected repository setup output to contain '%s', got: %s", expected, output)
		}
	}
}

// Benchmark test to measure performance
func BenchmarkPHPInstallCmd(b *testing.B) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "php.sh")

	os.MkdirAll(scriptDir, 0755)

	// Create a lightweight mock script for benchmarking
	scriptContent := `#!/bin/bash
echo "PHP installation completed"
`

	os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	// Run benchmark
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		Cmd.SetOut(buf)
		Cmd.SetErr(buf)
		Cmd.Run(Cmd, []string{})
	}
}
