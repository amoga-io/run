package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var CLIName = "devkit"

func getScriptName(command, packageName string) (string, bool) {
	if command == "install" {
		script, exists := InstallPackageRegistry[packageName]
		return script, exists
	} else if command == "remove" {
		script, exists := RemovePackageRegistry[packageName]
		return script, exists
	}
	return "", false
}

func GetScriptPath(command, packageName string) (string, error) {
	script, exists := getScriptName(command, packageName)
	if !exists {
		return "", fmt.Errorf("no script found for command '%s' and package '%s'", command, packageName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}
	scriptDir := filepath.Join(home, "."+CLIName, "scripts")
	scriptPath := filepath.Join(scriptDir, script)

	return scriptPath, nil
}

func ExecuteScript(scriptPath string) error {
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// Make script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %v", err)
	}

	fmt.Printf("Executing script: %s\n", scriptPath)

	// Execute the script
	cmd := exec.Command(scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute script: %v", err)
	}

	return nil
}
