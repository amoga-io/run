package python

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amoga-io/run/internals/utils"
)

func TestPythonScriptLocation(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) (string, func())
		expectScript bool
		expectError  bool
		expectedErr  string
	}{
		{
			name: "script exists and is executable",
			setupFunc: func(t *testing.T) (string, func()) {
				// Create a temporary directory structure
				tempDir := t.TempDir()
				scriptDir := filepath.Join(tempDir, ".devkit", "scripts")
				err := os.MkdirAll(scriptDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}

				// Create a mock script
				scriptPath := filepath.Join(scriptDir, "python.sh")
				scriptContent := `#!/bin/bash
# Mock python installation script
echo "This is a test script"
exit 0`
				err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
				if err != nil {
					t.Fatalf("Failed to create test script: %v", err)
				}

				// Set HOME to our temp directory
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				return tempDir, func() {
					os.Setenv("HOME", originalHome)
				}
			},
			expectScript: true,
			expectError:  false,
		},
		{
			name: "script file not found",
			setupFunc: func(t *testing.T) (string, func()) {
				// Create a temporary directory but no script
				tempDir := t.TempDir()

				// Set HOME to our temp directory (no .devkit/scripts directory)
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				return tempDir, func() {
					os.Setenv("HOME", originalHome)
				}
			},
			expectScript: false,
			expectError:  true,
			expectedErr:  "Python installation script not found",
		},
		{
			name: "script exists but not executable",
			setupFunc: func(t *testing.T) (string, func()) {
				// Create a temporary directory structure
				tempDir := t.TempDir()
				scriptDir := filepath.Join(tempDir, ".devkit", "scripts")
				err := os.MkdirAll(scriptDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}

				// Create a script without execute permissions
				scriptPath := filepath.Join(scriptDir, "python.sh")
				scriptContent := `#!/bin/bash
echo "This is a test script"
exit 0`
				err = os.WriteFile(scriptPath, []byte(scriptContent), 0644) // No execute permission
				if err != nil {
					t.Fatalf("Failed to create test script: %v", err)
				}

				// Set HOME to our temp directory
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tempDir)

				return tempDir, func() {
					os.Setenv("HOME", originalHome)
				}
			},
			expectScript: true,
			expectError:  false, // chmod in the command will make it executable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			_, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Test script location using utils function
			scriptPath := utils.GetScriptPath("python.sh")

			if tt.expectScript {
				// Verify script exists
				if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
					t.Errorf("Expected script to exist at %s, but it doesn't", scriptPath)
				}
			} else {
				// Verify script doesn't exist
				if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
					t.Errorf("Expected script to not exist at %s, but it does", scriptPath)
				}
			}

			// Test command behavior without execution
			// We don't need to capture output since we're not executing

			// Mock the command execution by testing the logic manually
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				if tt.expectError && !strings.Contains(tt.expectedErr, "Python installation script not found") {
					t.Errorf("Expected error message about script not found")
				}
			} else {
				// Script exists, check if we can make it executable
				if err := os.Chmod(scriptPath, 0755); err != nil {
					t.Errorf("Failed to make script executable: %v", err)
				}

				// Verify it's now executable
				info, err := os.Stat(scriptPath)
				if err != nil {
					t.Errorf("Failed to stat script: %v", err)
				} else if info.Mode().Perm()&0111 == 0 {
					t.Errorf("Script is not executable after chmod")
				}
			}
		})
	}
}

func TestPythonCommandProperties(t *testing.T) {
	t.Run("command has correct properties", func(t *testing.T) {
		if Cmd.Use != "python" {
			t.Errorf("Expected Use to be 'python', got '%s'", Cmd.Use)
		}
		if Cmd.Short != "Install python" {
			t.Errorf("Expected Short to be 'Install python', got '%s'", Cmd.Short)
		}
		if !strings.Contains(Cmd.Long, "Install python on your system") {
			t.Errorf("Expected Long to contain 'Install python on your system', got '%s'", Cmd.Long)
		}
		if Cmd.Run == nil {
			t.Error("Expected Run function to be set, but it's nil")
		}
	})
}

func TestPythonUtilsGetScriptPath(t *testing.T) {
	t.Run("GetScriptPath returns correct path format", func(t *testing.T) {
		scriptPath := utils.GetScriptPath("python.sh")

		// Check that the path ends with the expected structure
		if !strings.HasSuffix(scriptPath, "/.devkit/scripts/python.sh") {
			t.Errorf("Expected script path to end with '/.devkit/scripts/python.sh', got '%s'", scriptPath)
		}

		// Check that it's an absolute path
		if !filepath.IsAbs(scriptPath) {
			t.Errorf("Expected absolute path, got '%s'", scriptPath)
		}
	})
}

func TestPythonScriptPermissions(t *testing.T) {
	t.Run("script can be made executable", func(t *testing.T) {
		// Create a temporary script
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "test-python.sh")

		scriptContent := `#!/bin/bash
echo "test script"
exit 0`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		// Verify it's not executable initially
		info, err := os.Stat(scriptPath)
		if err != nil {
			t.Fatalf("Failed to stat script: %v", err)
		}
		if info.Mode().Perm()&0111 != 0 {
			t.Error("Script should not be executable initially")
		}

		// Make it executable (same as done in python.go)
		err = os.Chmod(scriptPath, 0755)
		if err != nil {
			t.Errorf("Failed to make script executable: %v", err)
		}

		// Verify it's now executable
		info, err = os.Stat(scriptPath)
		if err != nil {
			t.Fatalf("Failed to stat script after chmod: %v", err)
		}
		if info.Mode().Perm()&0111 == 0 {
			t.Error("Script should be executable after chmod")
		}
	})
}
