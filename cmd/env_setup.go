package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var envSetupCmd = &cobra.Command{
	Use:   "env-setup",
	Short: "Set up environment variables for version managers and reload the shell",
	Long:  `Automatically appends required environment variables for pyenv, nvm, sdkman, and phpenv to your shell config and reloads the shell.`,
	RunE:  runEnvSetup,
}

func runEnvSetup(cmd *cobra.Command, args []string) error {
	shell := os.Getenv("SHELL")
	var rcFile string
	if strings.Contains(shell, "zsh") {
		rcFile = ".zshrc"
	} else {
		rcFile = ".bashrc"
	}
	home := os.Getenv("HOME")
	configPath := fmt.Sprintf("%s/%s", home, rcFile)

	entries := []struct {
		name    string
		snippet string
	}{
		{"pyenv", "export PYENV_ROOT=\"$HOME/.pyenv\"\nexport PATH=\"$PYENV_ROOT/bin:$PATH\"\neval \"$(pyenv init --path)\"\neval \"$(pyenv virtualenv-init -)\""},
		{"nvm", "export NVM_DIR=\"$HOME/.nvm\"\n[ -s \"$NVM_DIR/nvm.sh\" ] && \\. \"$NVM_DIR/nvm.sh\""},
		{"sdkman", "export SDKMAN_DIR=\"$HOME/.sdkman\"\nsource $SDKMAN_DIR/bin/sdkman-init.sh"},
		{"phpenv", "export PHPENV_ROOT=\"$HOME/.phpenv\"\nexport PATH=\"$PHPENV_ROOT/bin:$PATH\"\neval \"$(phpenv init -)\""},
	}

	// Read current config
	contents, _ := os.ReadFile(configPath)
	config := string(contents)
	updated := false

	for _, entry := range entries {
		if !strings.Contains(config, entry.snippet) {
			f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", configPath, err)
			}
			defer f.Close()
			fmt.Fprintf(f, "\n# %s environment setup\n%s\n", entry.name, entry.snippet)
			updated = true
		}
	}

	if updated {
		fmt.Printf("\u2713 Environment variables added to %s. Reloading shell...\n", configPath)
		// Reload the shell
		execCmd := exec.Command("bash", "-c", fmt.Sprintf("source %s", configPath))
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Run()
	} else {
		fmt.Printf("No changes needed. Environment already set up in %s.\n", configPath)
	}
	return nil
}
