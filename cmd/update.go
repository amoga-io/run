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

	// Check and install required dependencies
	if err := system.EnsureDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Find the repository directory
	repoDir := filepath.Join(os.Getenv("HOME"), ".run")

	// Check if repository exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repository not found at %s. Please reinstall using curl command", repoDir)
	}

	fmt.Printf("Found repository at: %s\n", repoDir)

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

	// Pull latest changes
	fmt.Println("Pulling latest changes...")
	pullCmd := exec.Command("git", "pull", "origin", "main")
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

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
