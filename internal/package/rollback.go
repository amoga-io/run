package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/amoga-io/run/internal/logger"
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

	// Create backup
	if err := copyFile(originalPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
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
	log := logger.GetLogger().WithOperation("rollback").WithPackage(rp.PackageName)

	log.Info("Executing rollback: %s (%s)", rp.Operation, rp.ID)
	fmt.Printf("Executing rollback for %s (%s)...\n", rp.PackageName, rp.Operation)

	// Execute rollback commands in reverse order
	for i := len(rp.Commands) - 1; i >= 0; i-- {
		cmd := rp.Commands[i]
		log.Info("Executing rollback command: %s", cmd)
		fmt.Printf("Executing rollback command: %s\n", cmd)

		if err := executeRollbackCommand(cmd); err != nil {
			log.Error("Rollback command failed: %s - %v", cmd, err)
			fmt.Printf("Warning: rollback command failed: %v\n", err)
		}
	}

	// Restore backup files
	for originalPath, backupPath := range rp.BackupFiles {
		log.Info("Restoring file: %s from %s", originalPath, backupPath)
		fmt.Printf("Restoring file: %s\n", originalPath)
		if err := restoreBackupFile(originalPath, backupPath); err != nil {
			log.Error("Failed to restore file %s: %v", originalPath, err)
			fmt.Printf("Warning: failed to restore file %s: %v\n", originalPath, err)
		}
	}

	log.Info("Rollback completed successfully")
	fmt.Printf("✓ Rollback completed for %s\n", rp.PackageName)
	return nil
}

// executeRollbackCommand executes a rollback command
func executeRollbackCommand(command string) error {
	// Parse command into executable and arguments
	args := parseCommand(command)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("  Executing: %s\n", command)
	return cmd.Run()
}

// parseCommand parses a shell command into executable and arguments
func parseCommand(command string) []string {
	// Simple command parsing - in production, use a proper shell parser
	var args []string
	var current string
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(command); i++ {
		char := command[i]

		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if quoteChar == char {
				inQuotes = false
				quoteChar = 0
			} else {
				current += string(char)
			}
		case ' ':
			if !inQuotes {
				if current != "" {
					args = append(args, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		args = append(args, current)
	}

	return args
}

// restoreBackupFile restores a file from backup
func restoreBackupFile(originalPath, backupPath string) error {
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(originalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy backup to original location
	if err := copyFile(backupPath, originalPath); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	fmt.Printf("  Restored: %s from %s\n", originalPath, backupPath)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	_, err = destFile.ReadFrom(sourceFile)
	return err
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

	logger.Info("Rollback point cleaned up: %s (%s)", pointID, point.PackageName)
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
			logger.Error("Failed to cleanup old rollback point %s: %v", pointID, err)
			fmt.Printf("Warning: failed to cleanup old rollback point %s: %v\n", pointID, err)
		}
	}

	if len(toRemove) > 0 {
		logger.Info("Cleaned up %d old rollback points", len(toRemove))
		fmt.Printf("Cleaned up %d old rollback points\n", len(toRemove))
	}

	return nil
}
