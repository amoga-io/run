package pkg

import (
	"os"
)

// MockManager creates a manager that doesn't require the repository
func MockManager() *Manager {
	// Create a temporary directory for testing
	tempDir, _ := os.MkdirTemp("", "run-test-*")

	return &Manager{
		repoPath: tempDir,
	}
}

// CleanupMockManager cleans up the mock manager's temporary directory
func CleanupMockManager(manager *Manager) {
	if manager != nil && manager.repoPath != "" {
		os.RemoveAll(manager.repoPath)
	}
}
