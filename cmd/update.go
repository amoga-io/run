package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update devkit by pulling the latest changes from GitHub",
	Long: `Update devkit to the latest version by pulling changes from GitHub,
rebuilding the binary, and installing it. This is similar to 'brew upgrade'.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Checking for updates...")

		// Path to the cloned repository
		repoDir := filepath.Join(os.Getenv("HOME"), ".devkit")

		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ devkit repo not found in %s\n", repoDir)
			fmt.Fprintf(cmd.ErrOrStderr(), "💡 Try reinstalling with: bash <(curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/install.sh)\n")
			return
		}

		// Get current commit before update
		beforeCmd := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD")
		beforeCommit, _ := beforeCmd.Output()
		beforeHash := strings.TrimSpace(string(beforeCommit))

		// Run 'git pull' inside the repo
		fmt.Println("⬇️  Pulling latest changes...")
		gitPull := exec.Command("git", "-C", repoDir, "pull", "origin", "main")
		gitPull.Stdout = cmd.OutOrStdout()
		gitPull.Stderr = cmd.ErrOrStderr()

		if err := gitPull.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Error updating devkit: %v\n", err)
			return
		}

		// Get current commit after update
		afterCmd := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD")
		afterCommit, _ := afterCmd.Output()
		afterHash := strings.TrimSpace(string(afterCommit))

		// Check if there were actually any updates
		if beforeHash == afterHash {
			fmt.Println("✅ devkit is already up to date!")

			// Show current version information
			if version, commit, err := getVersionFromGit(repoDir); err == nil {
				fmt.Printf("📍 Current version: %s (%s)\n", version, commit)
			} else if currentCommit, err := exec.Command("git", "-C", repoDir, "log", "--oneline", "-1").Output(); err == nil {
				fmt.Printf("📍 Current version: %s", string(currentCommit))
			}
			return
		}

		// Show version information for the update
		beforeVersion, _, _ := getVersionFromGit(repoDir)
		if beforeVersion == "" {
			beforeVersion = beforeHash
		}

		fmt.Printf("📦 Updated from %s to %s\n", beforeVersion, afterHash)

		// Check if we now have a new version tag
		if afterVersion, _, err := getVersionFromGit(repoDir); err == nil && afterVersion != beforeVersion {
			fmt.Printf("🎉 New version available: %s\n", afterVersion)
		}

		// Show what changed
		changelogCmd := exec.Command("git", "-C", repoDir, "log", "--oneline", fmt.Sprintf("%s..%s", beforeHash, afterHash))
		if changelog, err := changelogCmd.Output(); err == nil && len(changelog) > 0 {
			fmt.Println("🔄 Changes:")
			fmt.Print(string(changelog))
		}

		fmt.Println("devkit repository updated successfully!")

		// Rebuild the CLI after pulling updates
		fmt.Println("🔨 Rebuilding devkit...")

		// Check if Makefile exists and use it for better version handling
		makefilePath := filepath.Join(repoDir, "Makefile")
		var buildCmd *exec.Cmd

		if _, err := os.Stat(makefilePath); err == nil {
			// Use make for building with version info
			buildCmd = exec.Command("make", "build")
		} else {
			// Fallback to go build
			buildCmd = exec.Command("go", "build", "-o", "devkit")
		}

		buildCmd.Dir = repoDir
		buildCmd.Stdout = cmd.OutOrStdout()
		buildCmd.Stderr = cmd.ErrOrStderr()

		if err := buildCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Error rebuilding devkit: %v\n", err)
			return
		}

		// Create a temporary binary name to avoid "text file busy" error
		tempBinary := filepath.Join("/usr/local/bin", "devkit.new")
		finalBinary := "/usr/local/bin/devkit"

		// First, copy to a temporary location
		fmt.Println("📦 Installing updated binary...")
		copyCmd := exec.Command("sudo", "cp", filepath.Join(repoDir, "devkit"), tempBinary)
		copyCmd.Stdout = cmd.OutOrStdout()
		copyCmd.Stderr = cmd.ErrOrStderr()

		if err := copyCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Error copying updated devkit: %v\n", err)
			return
		}

		// Make it executable
		chmodCmd := exec.Command("sudo", "chmod", "+x", tempBinary)
		if err := chmodCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Error setting permissions: %v\n", err)
			// Clean up the temporary file
			exec.Command("sudo", "rm", "-f", tempBinary).Run()
			return
		}

		// Atomically move the new binary to replace the old one
		// This avoids the "text file busy" error because mv is atomic
		moveCmd := exec.Command("sudo", "mv", tempBinary, finalBinary)
		moveCmd.Stdout = cmd.OutOrStdout()
		moveCmd.Stderr = cmd.ErrOrStderr()

		if err := moveCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Error installing updated devkit: %v\n", err)
			// Clean up the temporary file if move failed
			exec.Command("sudo", "rm", "-f", tempBinary).Run()
			return
		}

		fmt.Println("🚀 devkit update complete!")
		fmt.Println("✨ Run 'devkit version' to see the new version")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
