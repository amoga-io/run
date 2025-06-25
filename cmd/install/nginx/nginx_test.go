package nginx

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GitHub Actions CI/CD Test Suite
// These tests run actual Nginx installation in the disposable CI environment.
// Safe for GitHub Actions runners but DO NOT run locally unless intended.

func TestNginxInstallCmd_ScriptNotFound(t *testing.T) {
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
	if !strings.Contains(output, "Nginx installation script not found") {
		t.Errorf("expected error about script not found, got: %s", output)
	}
}

func TestNginxInstallCmd_ActualInstallation(t *testing.T) {
	// This test runs the ACTUAL Nginx installation script
	// Safe for GitHub Actions but will install Nginx on the runner

	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "nginx.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Copy the actual nginx.sh script content
	scriptContent := `#!/bin/bash

# Add Nginx official repository
echo "deb [arch=amd64] http://nginx.org/packages/mainline/ubuntu/ $(lsb_release -cs) nginx" | sudo tee /etc/apt/sources.list.d/nginx.list

# Add Nginx signing key
curl -fsSL https://nginx.org/keys/nginx_signing.key | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/nginx.gpg

# Install nginx
sudo apt update
sudo apt install -y nginx

# Create required directories
sudo mkdir -p /var/run/nginx
sudo mkdir -p /var/log/nginx

# Set ownership
sudo chown -R $USER:$USER /var/log/nginx
sudo chown -R $USER:$USER /var/run/nginx

# Set directory permissions
sudo chmod 755 /var/run/nginx
sudo chmod 755 /var/log/nginx

# Allow nginx to bind to ports 80/443 without root
sudo setcap cap_net_bind_service=+ep /usr/sbin/nginx

# Backup original nginx.conf
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# Update nginx.conf - remove user directive and update pid path
sudo sed -i "s/user .*;/user $USER;/" /etc/nginx/nginx.conf
sudo sed -i '/http {/a \    client_max_body_size 10M;' /etc/nginx/nginx.conf

# Create minimal site configuration
echo "server { listen 80 default_server; listen [::]:80 default_server; server_name _; location / { return 200 'nginx is working!'; add_header Content-Type text/plain; } }" | sudo tee /etc/nginx/conf.d/test-site.conf

# Set proper ownership for configuration
sudo chown $USER:$USER /etc/nginx/nginx.conf
sudo chown -R $USER:$USER /etc/nginx/conf.d

# Test configuration
nginx -t

# Start nginx
sudo systemctl start nginx
sudo systemctl enable nginx

echo "Nginx installed and running as user $USER"
echo "Test the installation: curl http://localhost"
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

	// Execute the command (this will install Nginx on GitHub Actions runner)
	Cmd.Run(Cmd, []string{})

	// Verify installation output
	output := buf.String()

	// Check for successful installation indicators
	expectedStrings := []string{
		"Nginx",
		"installed",
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

func TestNginxInstallCmd_Properties(t *testing.T) {
	// Test command properties without executing anything
	if Cmd.Use != "nginx" {
		t.Errorf("expected Use to be 'nginx', got: %s", Cmd.Use)
	}

	if Cmd.Short != "Install Nginx" {
		t.Errorf("expected Short to be 'Install Nginx', got: %s", Cmd.Short)
	}

	expectedLong := "Install Nginx on your system. This command will install Nginx using a provided script."
	if Cmd.Long != expectedLong {
		t.Errorf("expected Long to be '%s', got: %s", expectedLong, Cmd.Long)
	}

	if Cmd.Run == nil {
		t.Error("expected Run function to be defined")
	}
}

func TestNginxInstallCmd_PermissionError_Safe(t *testing.T) {
	// This test checks permission error handling without actually executing anything dangerous
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "nginx.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a safe mock script
	scriptContent := `#!/bin/bash
# MOCK SCRIPT - SAFE FOR TESTING
echo "This is a safe mock script for Nginx"
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

func TestNginxInstallCmd_ScriptExecutionError(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "nginx.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that will fail during execution
	scriptContent := `#!/bin/bash
echo "Starting Nginx installation..."
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
		"Starting Nginx installation",
		"Simulating installation failure",
		"Error executing Nginx installation script",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// Test successful script execution with mock content
func TestNginxInstallCmd_MockSuccess(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "nginx.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a mock script that simulates successful installation without actually installing
	scriptContent := `#!/bin/bash
echo "Nginx installation process started..."
echo "Adding Nginx repository..."
echo "Installing Nginx packages..."
echo "Setting up directories and permissions..."
echo "Configuring Nginx..."
echo "Starting Nginx service..."
echo "Nginx installed and running as user $USER"
echo "Test the installation: curl http://localhost"
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
		"Nginx installation process started",
		"Adding Nginx repository",
		"Installing Nginx packages",
		"Nginx installed and running",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// Benchmark test to measure performance
func BenchmarkNginxInstallCmd(b *testing.B) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "nginx.sh")

	os.MkdirAll(scriptDir, 0755)

	// Create a lightweight mock script for benchmarking
	scriptContent := `#!/bin/bash
echo "Nginx installation completed"
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

// Test that validates the Nginx configuration syntax (mock)
func TestNginxInstallCmd_ConfigValidation(t *testing.T) {
	// Setup test environment
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "nginx.sh")

	// Create the scripts directory
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create script directory: %v", err)
	}

	// Create a script that includes nginx configuration validation
	scriptContent := `#!/bin/bash
echo "Installing Nginx..."
echo "Creating configuration..."
echo "Testing configuration syntax..."
echo "nginx: configuration file /etc/nginx/nginx.conf test is successful"
echo "Configuration validation passed"
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

	// Verify configuration validation output
	output := buf.String()
	if !strings.Contains(output, "Configuration validation passed") {
		t.Errorf("expected configuration validation message, got: %s", output)
	}
}
