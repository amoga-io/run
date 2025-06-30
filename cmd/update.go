/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update CLI to latest version from Git",
	Long: `Pull latest changes from Git repository and rebuild the binary.

The update process:
  1. Fetches latest changes from the repository
  2. Handles any local changes gracefully
  3. Rebuilds the binary with latest features
  4. Installs the updated binary atomically

Requirements:
  • Git must be available
  • Go must be available for building
  • Sudo access for binary installation

Examples:
  run update`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Updating run CLI...")

	// Check for required dependencies
	if err := checkUpdateDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Find the repository directory
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return fmt.Errorf("HOME environment variable is not set")
	}
	repoDir := filepath.Join(homeDir, ".run")

	// Check if repository exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		// Repository doesn't exist, clone it
		fmt.Println("📥 Cloning repository...")
		cloneCmd := exec.Command("git", "clone", "https://github.com/amoga-io/run.git", repoDir)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		fmt.Println("✅ Repository cloned successfully")
	} else {
		fmt.Printf("📁 Found repository at: %s\n", repoDir)
	}

	// Change to repository directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(repoDir); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Ensure we return to original directory
	defer func() {
		os.Chdir(originalDir)
	}()

	// Check if it's a git repository
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository. Please reinstall using: curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/scripts/install.sh | bash")
	}

	// Update repository
	if err := updateRepository(); err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	// Build and install
	if err := buildAndInstall(); err != nil {
		return fmt.Errorf("failed to build and install: %w", err)
	}

	fmt.Println("🎉 Update completed successfully!")
	fmt.Println("✨ CLI has been updated to the latest version.")

	// Show current version
	if version := getCurrentVersion(); version != "" {
		fmt.Printf("📦 Current version: %s\n", version)
	}

	return nil
}

// checkUpdateDependencies checks if required tools are available
func checkUpdateDependencies() error {
	dependencies := []string{"git", "go"}

	for _, dep := range dependencies {
		if _, err := exec.LookPath(dep); err != nil {
			switch dep {
			case "git":
				return fmt.Errorf("git is required for updates. Install with: sudo apt-get install git")
			case "go":
				return fmt.Errorf("go is required for building. Install with: run install essentials")
			default:
				return fmt.Errorf("%s is required but not found", dep)
			}
		}
	}

	fmt.Println("✅ All dependencies available")
	return nil
}

// updateRepository updates the git repository
func updateRepository() error {
	fmt.Println("🔄 Pulling latest changes...")

	// Fetch latest changes
	fmt.Println("📡 Fetching from remote...")
	fetchCmd := exec.Command("git", "fetch", "origin", "main")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch latest changes: %w", err)
	}

	// Check if we have local changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, _ := statusCmd.Output()

	if len(statusOutput) > 0 {
		fmt.Println("⚠️  Local changes detected, stashing them...")
		// Stash any local changes
		stashCmd := exec.Command("git", "stash", "push", "-m", "Auto-stash before update")
		stashCmd.Run() // Don't fail if nothing to stash
	}

	// Hard reset to match remote (overwrites local changes)
	fmt.Println("🔄 Applying latest changes...")
	resetCmd := exec.Command("git", "reset", "--hard", "origin/main")
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to latest changes: %w", err)
	}

	// Clean any untracked files
	cleanCmd := exec.Command("git", "clean", "-fd")
	cleanCmd.Run() // Don't fail on this

	fmt.Println("✅ Repository updated to latest version")
	return nil
}

// buildAndInstall builds the binary and installs it
func buildAndInstall() error {
	// Prepare Go modules
	fmt.Println("📦 Preparing Go modules...")
	modCmd := exec.Command("go", "mod", "tidy")
	if err := modCmd.Run(); err != nil {
		return fmt.Errorf("failed to prepare Go modules: %w", err)
	}

	// Build new binary
	fmt.Println("🔨 Building new binary...")
	binaryName := "run"

	// Get version information for build
	version := getVersionInfo()
	commit := getCommitInfo()
	buildDate := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	fmt.Printf("📋 Building version: %s (commit: %s)\n", version, commit)

	// Build with version information embedded
	buildCmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf(`-X 'main.Version=%s' -X 'main.GitCommit=%s' -X 'main.BuildDate=%s'`,
			version, commit, buildDate),
		"-o", binaryName, ".")

	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build new binary: %w", err)
	}

	// Verify binary was built
	if _, err := os.Stat(binaryName); os.IsNotExist(err) {
		return fmt.Errorf("binary was not created successfully")
	}

	// Install the updated binary
	fmt.Println("📥 Installing updated binary...")
	if err := installBinary(binaryName); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	fmt.Println("✅ Binary installed successfully")
	return nil
}

// installBinary installs the binary atomically
func installBinary(binaryName string) error {
	installDir := "/usr/local/bin"
	finalBinary := filepath.Join(installDir, binaryName)

	// Use atomic replacement to avoid "text file busy" errors
	tempBinary := filepath.Join(installDir, binaryName+".new")

	// Copy to temporary location
	copyCmd := exec.Command("sudo", "cp", binaryName, tempBinary)
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable
	chmodCmd := exec.Command("sudo", "chmod", "+x", tempBinary)
	if err := chmodCmd.Run(); err != nil {
		// Clean up temp file on failure
		exec.Command("sudo", "rm", "-f", tempBinary).Run()
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Atomically replace the binary
	mvCmd := exec.Command("sudo", "mv", tempBinary, finalBinary)
	if err := mvCmd.Run(); err != nil {
		// Clean up temp file on failure
		exec.Command("sudo", "rm", "-f", tempBinary).Run()
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// getVersionInfo gets version information from git
func getVersionInfo() string {
	versionCmd := exec.Command("git", "describe", "--tags", "--always")
	versionOutput, err := versionCmd.Output()
	if err != nil {
		return "v0.0.0-dev"
	}
	return strings.TrimSpace(string(versionOutput))
}

// getCommitInfo gets commit information from git
func getCommitInfo() string {
	commitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	commitOutput, err := commitCmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(commitOutput))
}

// getCurrentVersion gets the current version of the installed binary
func getCurrentVersion() string {
	versionCmd := exec.Command("run", "--version")
	versionOutput, err := versionCmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(versionOutput))
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
