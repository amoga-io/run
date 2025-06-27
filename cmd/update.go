package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/amoga-io/run/internal/system"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update CLI to latest version from Git",
	Long:  "Pull latest changes from Git repository and rebuild the binary",
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("Updating run CLI...")

	// Check and install required dependencies for CLI operation
	if err := system.EnsureRuntimeRequirements(); err != nil {
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
		fmt.Println("Cloning repository...")
		cloneCmd := exec.Command("git", "clone", "https://github.com/amoga-io/run.git", repoDir)
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		fmt.Println("✓ Repository cloned")
	} else {
		fmt.Printf("Found repository at: %s\n", repoDir)
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
		return fmt.Errorf("not a git repository. Please reinstall using curl command")
	}

	// Force clean update to handle local changes
	fmt.Println("Pulling latest changes...")

	// Fetch latest changes
	fetchCmd := exec.Command("git", "fetch", "origin", "main")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch latest changes: %w", err)
	}

	// Check if we have local changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, _ := statusCmd.Output()

	if len(statusOutput) > 0 {
		fmt.Println("⚠️  Local changes detected, forcing clean update...")
		// Stash any local changes
		stashCmd := exec.Command("git", "stash", "push", "-m", "Auto-stash before update")
		stashCmd.Run() // Don't fail if nothing to stash
	}

	// Hard reset to match remote (overwrites local changes)
	resetCmd := exec.Command("git", "reset", "--hard", "origin/main")
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to latest changes: %w", err)
	}

	// Clean any untracked files
	cleanCmd := exec.Command("git", "clean", "-fd")
	cleanCmd.Run() // Don't fail on this

	fmt.Println("✓ Repository updated to latest version")

	// Prepare Go modules
	fmt.Println("Preparing Go modules...")
	modCmd := exec.Command("go", "mod", "tidy")
	modCmd.Stdout = os.Stdout
	modCmd.Stderr = os.Stderr

	if err := modCmd.Run(); err != nil {
		return fmt.Errorf("failed to prepare Go modules: %w", err)
	}

	// Build new binary
	fmt.Println("Building new binary...")
	binaryName := "run"

	// Get version information for build
	versionCmd := exec.Command("git", "describe", "--tags", "--always")
	versionOutput, err := versionCmd.Output()
	version := "v0.0.0-dev"
	if err == nil {
		version = strings.TrimSpace(string(versionOutput))
	}

	commitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	commitOutput, err := commitCmd.Output()
	commit := "unknown"
	if err == nil {
		commit = strings.TrimSpace(string(commitOutput))
	}

	buildDate := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	fmt.Printf("Building version: %s\n", version)

	// Build with version information embedded
	buildCmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf(`-X 'github.com/amoga-io/run/cmd.Version=%s' -X 'github.com/amoga-io/run/cmd.GitCommit=%s' -X 'github.com/amoga-io/run/cmd.BuildDate=%s'`,
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
	fmt.Println("Installing updated binary...")
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
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Atomically replace the binary
	mvCmd := exec.Command("sudo", "mv", tempBinary, finalBinary)
	if err := mvCmd.Run(); err != nil {
		// Clean up temp file on failure
		exec.Command("sudo", "rm", "-f", tempBinary).Run()
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Println("Update completed successfully!")
	fmt.Println("CLI has been updated to the latest version.")

	return nil
}
