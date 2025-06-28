package pkg

import (
	"os"
	"testing"
)

// TestMain sets up the test environment for CI
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("TESTING", "true")

	// Create temporary test directory if needed
	testDir := "/tmp/run-test"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		panic("Failed to create test directory")
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(testDir)
	os.Unsetenv("TESTING")

	os.Exit(code)
}
