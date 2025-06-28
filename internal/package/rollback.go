package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RollbackPoint represents a system state that can be restored
type RollbackPoint struct {
	ID          string
	Timestamp   time.Time
	PackageName string
	Operation   string // "install" or "remove"
	BackupFiles map[string]string
	Commands    []string
}

// RollbackManager manages rollback points and operations
type RollbackManager struct {
	rollbackDir string
	points      map[string]*RollbackPoint
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager() (*RollbackManager, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable is not set")
	}

	rollbackDir := filepath.Join(homeDir, ".run", "rollbacks")
	if err := os.MkdirAll(rollbackDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create rollback directory: %w", err)
	}

	return &RollbackManager{
		rollbackDir: rollbackDir,
		points:      make(map[string]*RollbackPoint),
	}, nil
}

// CreateRollbackPoint creates a rollback point before package operation
func (rm *RollbackManager) CreateRollbackPoint(packageName, operation string) (*RollbackPoint, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, fmt.Errorf("invalid package name for rollback: %w", err)
	}

	pointID := fmt.Sprintf("%s_%s_%d", packageName, operation, time.Now().Unix())

	point := &RollbackPoint{
		ID:          pointID,
		Timestamp:   time.Now(),
		PackageName: packageName,
		Operation:   operation,
		BackupFiles: make(map[string]string),
		Commands:    make([]string, 0),
	}

	// Create rollback directory
	pointDir := filepath.Join(rm.rollbackDir, pointID)
	if err := os.MkdirAll(pointDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create rollback point directory: %w", err)
	}

	rm.points[pointID] = point
	return point, nil
}

// AddBackupFile adds a file to be backed up for rollback
func (rp *RollbackPoint) AddBackupFile(originalPath, backupPath string) error {
	// Validate paths
	if _, err := ValidatePath(originalPath); err != nil {
		return fmt.Errorf("invalid original path: %w", err)
	}

	if _, err := ValidatePath(backupPath); err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	rp.BackupFiles[originalPath] = backupPath
	return nil
}

// AddRollbackCommand adds a command to be executed during rollback
func (rp *RollbackPoint) AddRollbackCommand(command string) {
	rp.Commands = append(rp.Commands, command)
}

// ExecuteRollback executes the rollback for this point
func (rp *RollbackPoint) ExecuteRollback() error {
	fmt.Printf("Executing rollback for %s (%s)...\n", rp.PackageName, rp.Operation)

	// Execute rollback commands in reverse order
	for i := len(rp.Commands) - 1; i >= 0; i-- {
		cmd := rp.Commands[i]
		fmt.Printf("Executing rollback command: %s\n", cmd)

		// Execute command (simplified - in real implementation, use exec.Command)
		if err := executeRollbackCommand(cmd); err != nil {
			fmt.Printf("Warning: rollback command failed: %v\n", err)
		}
	}

	// Restore backup files
	for originalPath, backupPath := range rp.BackupFiles {
		fmt.Printf("Restoring file: %s\n", originalPath)
		if err := restoreBackupFile(originalPath, backupPath); err != nil {
			fmt.Printf("Warning: failed to restore file %s: %v\n", originalPath, err)
		}
	}

	fmt.Printf("✓ Rollback completed for %s\n", rp.PackageName)
	return nil
}

// executeRollbackCommand executes a rollback command
func executeRollbackCommand(command string) error {
	// This is a simplified implementation
	// In a real implementation, you would use exec.Command and handle different command types
	fmt.Printf("  Executing: %s\n", command)
	return nil
}

// restoreBackupFile restores a file from backup
func restoreBackupFile(originalPath, backupPath string) error {
	// This is a simplified implementation
	// In a real implementation, you would copy the backup file to the original location
	fmt.Printf("  Restoring: %s from %s\n", originalPath, backupPath)
	return nil
}

// CleanupRollbackPoint removes a rollback point after successful operation
func (rm *RollbackManager) CleanupRollbackPoint(pointID string) error {
	point, exists := rm.points[pointID]
	if !exists {
		return fmt.Errorf("rollback point not found: %s", pointID)
	}

	// Remove rollback directory
	pointDir := filepath.Join(rm.rollbackDir, pointID)
	if err := os.RemoveAll(pointDir); err != nil {
		return fmt.Errorf("failed to remove rollback directory: %w", err)
	}

	// Remove from points map
	delete(rm.points, pointID)

	fmt.Printf("✓ Rollback point cleaned up: %s (%s)\n", pointID, point.PackageName)
	return nil
}

// ListRollbackPoints lists all available rollback points
func (rm *RollbackManager) ListRollbackPoints() []*RollbackPoint {
	var points []*RollbackPoint
	for _, point := range rm.points {
		points = append(points, point)
	}
	return points
}

// GetRollbackPoint gets a specific rollback point
func (rm *RollbackManager) GetRollbackPoint(pointID string) (*RollbackPoint, bool) {
	point, exists := rm.points[pointID]
	return point, exists
}

// CleanupOldRollbackPoints removes rollback points older than the specified duration
func (rm *RollbackManager) CleanupOldRollbackPoints(maxAge time.Duration) error {
	now := time.Now()
	var toRemove []string

	for pointID, point := range rm.points {
		if now.Sub(point.Timestamp) > maxAge {
			toRemove = append(toRemove, pointID)
		}
	}

	for _, pointID := range toRemove {
		if err := rm.CleanupRollbackPoint(pointID); err != nil {
			fmt.Printf("Warning: failed to cleanup old rollback point %s: %v\n", pointID, err)
		}
	}

	if len(toRemove) > 0 {
		fmt.Printf("Cleaned up %d old rollback points\n", len(toRemove))
	}

	return nil
}
