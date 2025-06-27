package system

import (
	"fmt"
	"os/exec"
)

// Dependency represents a system dependency
type Dependency struct {
	Command     string
	Package     string
	Description string
}

// RequiredDependencies returns the list of dependencies needed for the CLI
func RequiredDependencies() []Dependency {
	return []Dependency{
		{
			Command:     "git",
			Package:     "git",
			Description: "Git version control system",
		},
		{
			Command:     "go",
			Package:     "golang-go",
			Description: "Go programming language",
		},
		{
			Command:     "sudo",
			Package:     "sudo",
			Description: "Sudo privileges",
		},
	}
}

// CommandExists checks if a command is available in PATH
func CommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// CheckDependency checks if a single dependency is available
func CheckDependency(dep Dependency) bool {
	return CommandExists(dep.Command)
}

// CheckAllDependencies checks all required dependencies
func CheckAllDependencies() ([]Dependency, error) {
	var missing []Dependency

	for _, dep := range RequiredDependencies() {
		if !CheckDependency(dep) {
			missing = append(missing, dep)
		}
	}

	return missing, nil
}

// InstallSystemPackages installs system packages via apt (silent)
func InstallSystemPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Update package list silently
	updateCmd := exec.Command("sudo", "apt-get", "update", "-qq")
	updateCmd.Env = append(updateCmd.Env, "DEBIAN_FRONTEND=noninteractive")
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install packages silently
	args := append([]string{"apt-get", "install", "-y", "-qq", "--no-install-recommends"}, packages...)
	installCmd := exec.Command("sudo", args...)
	installCmd.Env = append(installCmd.Env, "DEBIAN_FRONTEND=noninteractive")

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages %v: %w", packages, err)
	}

	return nil
}

// ExecuteCommands executes a series of shell commands with proper error handling
func ExecuteCommands(commands [][]string) error {
	for _, cmdArgs := range commands {
		if len(cmdArgs) == 0 {
			continue
		}

		// Handle shell operators like || true
		if len(cmdArgs) >= 3 && cmdArgs[len(cmdArgs)-2] == "||" && cmdArgs[len(cmdArgs)-1] == "true" {
			// Execute command and ignore errors if || true
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)-2]...)
			cmd.Run() // Ignore error as intended
			continue
		}

		// Handle other shell operators like ||
		if len(cmdArgs) >= 3 && cmdArgs[len(cmdArgs)-2] == "||" {
			// Execute main command, if it fails, execute the fallback
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)-2]...)
			if err := cmd.Run(); err != nil {
				// Execute fallback command
				fallbackCmd := exec.Command(cmdArgs[len(cmdArgs)-1])
				fallbackCmd.Run()
			}
			continue
		}

		// Regular command execution
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			// Continue with other commands for removal operations but log the error
			fmt.Printf("Warning: command failed: %v (continuing...)\n", cmdArgs)
		}
	}
	return nil
}

// InstallDependencies installs missing dependencies on Ubuntu/Debian systems
func InstallDependencies(missing []Dependency) error {
	if len(missing) == 0 {
		return nil
	}

	fmt.Println("Installing missing dependencies...")

	// Update package list first
	updateCmd := exec.Command("sudo", "apt", "update")
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install each missing dependency
	for _, dep := range missing {
		fmt.Printf("Installing %s (%s)...\n", dep.Description, dep.Package)

		installCmd := exec.Command("sudo", "apt", "install", "-y", dep.Package)
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install %s: %w", dep.Package, err)
		}
	}

	return nil
}

// EnsureDependencies checks and installs all required dependencies
func EnsureDependencies() error {
	fmt.Println("Checking dependencies...")

	missing, err := CheckAllDependencies()
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	if len(missing) == 0 {
		fmt.Println("All dependencies are available.")
		return nil
	}

	fmt.Printf("Missing dependencies: ")
	for i, dep := range missing {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(dep.Command)
	}
	fmt.Println()

	return InstallDependencies(missing)
}
