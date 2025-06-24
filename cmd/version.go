package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Version information - can be set via build flags
var (
	Version   = "v1.0.0"  // Default version
	GitCommit = "unknown" // Git commit hash
	BuildDate = "unknown" // Build date
)

// getVersionFromGit gets the version from git tags in the repo directory
func getVersionFromGit(repoDir string) (string, string, error) {
	// Try to get the latest git tag
	tagCmd := exec.Command("git", "-C", repoDir, "describe", "--tags", "--exact-match", "HEAD")
	if tagOutput, err := tagCmd.Output(); err == nil {
		tag := strings.TrimSpace(string(tagOutput))
		// Get commit hash
		commitCmd := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD")
		commitOutput, _ := commitCmd.Output()
		commit := strings.TrimSpace(string(commitOutput))
		return tag, commit, nil
	}

	// If no exact tag, get the latest tag and commit count
	describeCmd := exec.Command("git", "-C", repoDir, "describe", "--tags", "--always", "--dirty")
	if describeOutput, err := describeCmd.Output(); err == nil {
		version := strings.TrimSpace(string(describeOutput))
		commitCmd := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD")
		commitOutput, _ := commitCmd.Output()
		commit := strings.TrimSpace(string(commitOutput))
		return version, commit, nil
	}

	// Fallback to commit hash only
	commitCmd := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD")
	if commitOutput, err := commitCmd.Output(); err == nil {
		commit := strings.TrimSpace(string(commitOutput))
		return "v0.0.0-" + commit, commit, nil
	}

	return "", "", fmt.Errorf("unable to determine version")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current version of gocli",
	Long:  `Display version information including semantic version, git commit, and build details.`,
	Run: func(cmd *cobra.Command, args []string) {
		var repoDir string

		// First try current directory (for development)
		if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err == nil {
			if wd, err := os.Getwd(); err == nil {
				repoDir = wd
			}
		}

		// Fallback to installed location
		if repoDir == "" {
			repoDir = filepath.Join(os.Getenv("HOME"), ".gocli")
		}

		// Try to get version from git repository
		if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
			if gitVersion, gitCommit, err := getVersionFromGit(repoDir); err == nil {
				fmt.Printf("gocli version %s\n", gitVersion)
				fmt.Printf("Git commit: %s\n", gitCommit)

				// Show build date if available
				if BuildDate != "unknown" {
					fmt.Printf("Build date: %s\n", BuildDate)
				}

				// Check for uncommitted changes
				statusCmd := exec.Command("git", "-C", repoDir, "status", "--porcelain")
				if output, err := statusCmd.Output(); err == nil && len(output) > 0 {
					fmt.Println("Status: dirty (uncommitted changes)")
				} else {
					fmt.Println("Status: clean")
				}
				return
			}
		}

		// Fallback to compiled-in version
		fmt.Printf("gocli version %s\n", Version)
		if GitCommit != "unknown" {
			fmt.Printf("Git commit: %s\n", GitCommit)
		}
		if BuildDate != "unknown" {
			fmt.Printf("Build date: %s\n", BuildDate)
		}
		fmt.Println("Note: Repository not found, showing built-in version")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
