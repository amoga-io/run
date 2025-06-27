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

// CheckDependency checks if a single dependency is available
func CheckDependency(dep Dependency) bool {
	_, err := exec.LookPath(dep.Command)
	return err == nil
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
