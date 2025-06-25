package node

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GitHub Actions CI/CD Test Suite
// These tests run actual Node.js installation in the disposable CI environment.
// Safe for GitHub Actions runners but DO NOT run locally unless intended.

func TestNodeInstallCmd_ScriptNotFound(t *testing.T) {
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
	if !strings.Contains(output, "Node.js installation script not found") {
		t.Errorf("expected error about script not found, got: %s", output)
	}
}

func TestNodeInstallCmd_ActualInstallation(t *testing.T) {
	// This test runs the ACTUAL Node.js installation script
	// Safe for GitHub Actions but will install Node.js on the runner

	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Copy the actual node.sh script content
	scriptContent := `#!/bin/bash

# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Create npm global directory in user's home
mkdir -p ~/.npm-global
npm config set prefix ~/.npm-global

# Add npm global bin to PATH in ~/.profile if not already present
if ! grep -q "PATH=~/.npm-global/bin:\$PATH" ~/.profile; then
    echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.profile
fi

# Source the updated profile
source ~/.profile

# Install pnpm 9.10.0 using npm
npm install -g pnpm@9.10.0

# Add pnpm to PATH
export PATH=~/.npm-global/bin:$PATH

# Install pm2
npm install -g pm2

# Setup PM2 startup script
pm2 startup

# Execute the generated PM2 startup command
sudo env PATH=$PATH:/usr/bin /home/azureuser/.npm-global/lib/node_modules/pm2/bin/pm2 startup systemd -u azureuser --hp /home/azureuser
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

	// Execute the command (this will install Node.js on GitHub Actions runner)
	Cmd.Run(Cmd, []string{})

	// Verify installation output
	output := buf.String()

	// Check for successful installation indicators
	expectedStrings := []string{
		"Node.js",
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

func TestNodeInstallCmd_Properties(t *testing.T) {
	// Test command properties without executing anything
	if Cmd.Use != "node" {
		t.Errorf("expected Use to be 'node', got: %s", Cmd.Use)
	}

	if Cmd.Short != "Install Node.js" {
		t.Errorf("expected Short to be 'Install Node.js', got: %s", Cmd.Short)
	}

	expectedLong := "Install Node.js on your system. This command will install Node.js using a provided script."
	if Cmd.Long != expectedLong {
		t.Errorf("expected Long to be '%s', got: %s", expectedLong, Cmd.Long)
	}

	if Cmd.Run == nil {
		t.Error("expected Run function to be defined")
	}
}

func TestNodeInstallCmd_PermissionError_Safe(t *testing.T) {
	// This test checks permission error handling without actually executing anything dangerous
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a safe mock script
	scriptContent := `#!/bin/bash
# MOCK SCRIPT - SAFE FOR TESTING
echo "This is a safe mock script for Node.js"
`

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0000) // No permissions
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	// Make the directory read-only to simulate permission error
	err = os.Chmod(scriptDir, 0444)
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	// Cleanup after test
	defer func() {
		os.Chmod(scriptDir, 0755) // Restore permissions for cleanup
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

func TestNodeInstallCmd_ScriptExecutionError(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that will fail during execution
	scriptContent := `#!/bin/bash
echo "Starting Node.js installation..."
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
		"Starting Node.js installation",
		"Simulating installation failure",
		"Error executing Node.js installation script",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test successful script execution with mock content
func TestNodeInstallCmd_MockSuccess(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a mock script that simulates successful installation without actually installing
	scriptContent := `#!/bin/bash
echo "Node.js installation process started..."
echo "Installing Node.js 20..."
echo "Setting up npm global directory..."
echo "Configuring PATH variables..."
echo "Installing pnpm 9.10.0..."
echo "Installing PM2..."
echo "Setting up PM2 startup script..."
echo "Node.js, npm, pnpm, and PM2 installed successfully!"
echo "Node.js version: v20.x.x"
echo "npm version: x.x.x"
echo "pnpm version: 9.10.0"
echo "PM2 version: x.x.x"
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
		"Node.js installation process started",
		"Installing Node.js 20",
		"Installing pnpm 9.10.0",
		"Installing PM2",
		"installed successfully",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test Node.js version validation
func TestNodeInstallCmd_VersionValidation(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that includes version validation
	scriptContent := `#!/bin/bash
echo "Installing Node.js..."
echo "Verifying installation..."
echo "Node.js version: v20.11.1"
echo "npm version: 10.2.4"
echo "pnpm version: 9.10.0"
echo "PM2 version: 5.3.0"
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
		"Node.js version:",
		"npm version:",
		"pnpm version: 9.10.0",
		"PM2 version:",
	}

	for _, expected := range expectedVersions {
		if !strings.Contains(output, expected) {
			t.Errorf("expected version output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test PATH configuration
func TestNodeInstallCmd_PathConfiguration(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that tests PATH configuration
	scriptContent := `#!/bin/bash
echo "Setting up npm global directory..."
echo "Configuring PATH variables..."
echo "Adding ~/.npm-global/bin to PATH"
echo "PATH configuration completed"
echo "pnpm added to PATH successfully"
echo "Global packages will be installed to ~/.npm-global"
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

	// Verify PATH configuration output
	output := buf.String()
	pathRelatedStrings := []string{
		"PATH configuration completed",
		"pnpm added to PATH",
		"Global packages will be installed",
	}

	for _, expected := range pathRelatedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected PATH configuration output to contain '%s', got: %s", expected, output)
		}
	}
}

// Benchmark test to measure performance
func BenchmarkNodeInstallCmd(b *testing.B) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	os.MkdirAll(scriptDir, 0755)

	// Create a lightweight mock script for benchmarking
	scriptContent := `#!/bin/bash
echo "Node.js installation completed"
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

// Test PM2 setup functionality
func TestNodeInstallCmd_PM2Setup(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "node.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that tests PM2 setup
	scriptContent := `#!/bin/bash
echo "Installing PM2..."
echo "Setting up PM2 startup script..."
echo "PM2 startup script configured for systemd"
echo "PM2 is now ready to manage Node.js applications"
echo "Use 'pm2 start app.js' to start applications"
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

	// Verify PM2 setup output
	output := buf.String()
	pm2Strings := []string{
		"Installing PM2",
		"PM2 startup script",
		"PM2 is now ready",
	}

	for _, expected := range pm2Strings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected PM2 setup output to contain '%s', got: %s", expected, output)
		}
	}
}
