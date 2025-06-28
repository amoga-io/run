package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PackageLock manages locks for package operations
type PackageLock struct {
	locks map[string]*sync.Mutex
	mutex sync.RWMutex
}

// NewPackageLock creates a new package lock manager
func NewPackageLock() *PackageLock {
	return &PackageLock{
		locks: make(map[string]*sync.Mutex),
	}
}

// AcquireLock acquires a lock for a specific package
func (pl *PackageLock) AcquireLock(packageName string) error {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	// Validate package name
	if err := ValidatePackageName(packageName); err != nil {
		return fmt.Errorf("invalid package name for locking: %w", err)
	}

	// Get or create lock for this package
	lock, exists := pl.locks[packageName]
	if !exists {
		lock = &sync.Mutex{}
		pl.locks[packageName] = lock
	}

	// Try to acquire lock with timeout
	lockChan := make(chan bool, 1)
	go func() {
		lock.Lock()
		lockChan <- true
	}()

	select {
	case <-lockChan:
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for lock on package %s (another operation may be in progress)", packageName)
	}
}

// ReleaseLock releases a lock for a specific package
func (pl *PackageLock) ReleaseLock(packageName string) {
	pl.mutex.RLock()
	lock, exists := pl.locks[packageName]
	pl.mutex.RUnlock()

	if exists {
		lock.Unlock()
	}
}

// SystemLock manages system-wide locks for critical operations
type SystemLock struct {
	lockFile string
	file     *os.File
}

// NewSystemLock creates a new system lock
func NewSystemLock(lockName string) (*SystemLock, error) {
	// Create lock directory
	lockDir := filepath.Join(os.TempDir(), "run-locks")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	lockFile := filepath.Join(lockDir, fmt.Sprintf("%s.lock", lockName))

	return &SystemLock{
		lockFile: lockFile,
	}, nil
}

// Acquire acquires the system lock
func (sl *SystemLock) Acquire() error {
	// Try to create lock file
	file, err := os.OpenFile(sl.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("system lock already held by another process")
		}
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	// Write PID to lock file
	pid := os.Getpid()
	if _, err := file.WriteString(fmt.Sprintf("%d\n", pid)); err != nil {
		file.Close()
		os.Remove(sl.lockFile)
		return fmt.Errorf("failed to write to lock file: %w", err)
	}

	sl.file = file
	return nil
}

// Release releases the system lock
func (sl *SystemLock) Release() error {
	if sl.file != nil {
		sl.file.Close()
		sl.file = nil
	}

	if err := os.Remove(sl.lockFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	return nil
}

// Global package lock instance
var globalPackageLock = NewPackageLock()

// AcquirePackageLock acquires a lock for package operations
func AcquirePackageLock(packageName string) error {
	return globalPackageLock.AcquireLock(packageName)
}

// ReleasePackageLock releases a lock for package operations
func ReleasePackageLock(packageName string) {
	globalPackageLock.ReleaseLock(packageName)
}
