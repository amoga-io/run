package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// RequirementCategory defines different types of system requirements
type RequirementCategory string

const (
	Bootstrap   RequirementCategory = "bootstrap"   // Needed to build CLI
	Runtime     RequirementCategory = "runtime"     // Needed for CLI to work
	Development RequirementCategory = "development" // Nice to have for dev work
	Optional    RequirementCategory = "optional"    // Provided by essentials
)

// SystemRequirement represents what's needed for the CLI to function
type SystemRequirement struct {
	Name        string              `json:"name"`
	Commands    []string            `json:"commands"` // Commands that must exist
	Packages    []string            `json:"packages"` // Packages to install if missing
	Description string              `json:"description"`
	Category    RequirementCategory `json:"category"`
	Critical    bool                `json:"critical"` // Must have for CLI to work
}

// Dependency represents a system dependency (legacy compatibility)
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
	var errors []string

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
				if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
					errors = append(errors, fmt.Sprintf("command failed: %v (fallback also failed)", cmdArgs))
				}
			}
			continue
		}

		// Regular command execution
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			// Continue with other commands for removal operations but track the error
			errorMsg := fmt.Sprintf("command failed: %v", cmdArgs)
			fmt.Printf("Warning: %s (continuing...)\n", errorMsg)
			errors = append(errors, errorMsg)
		}
	}

	// Return aggregated errors if any occurred
	if len(errors) > 0 {
		return fmt.Errorf("some commands failed: %s", strings.Join(errors, "; "))
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

// EnsureDependencies checks and installs all required dependencies (legacy compatibility)
// Deprecated: Use EnsureRuntimeRequirements instead
func EnsureDependencies() error {
	return EnsureRuntimeRequirements()
}

// GetSystemRequirements returns all system requirements categorized
func GetSystemRequirements() []SystemRequirement {
	return []SystemRequirement{
		// Bootstrap requirements - absolutely needed to build CLI
		{
			Name:        "git",
			Commands:    []string{"git"},
			Packages:    []string{"git"},
			Description: "Git version control system",
			Category:    Bootstrap,
			Critical:    true,
		},
		{
			Name:        "golang",
			Commands:    []string{"go"},
			Packages:    []string{"golang-go"},
			Description: "Go programming language",
			Category:    Bootstrap,
			Critical:    true,
		},
		// Runtime requirements - needed for CLI to work properly
		{
			Name:        "sudo",
			Commands:    []string{"sudo"},
			Packages:    []string{"sudo"},
			Description: "Administrative privileges",
			Category:    Runtime,
			Critical:    true,
		},
		{
			Name:        "curl",
			Commands:    []string{"curl"},
			Packages:    []string{"curl"},
			Description: "HTTP client for downloads",
			Category:    Runtime,
			Critical:    false,
		},
		// Development requirements - nice to have for development
		{
			Name:        "build-tools",
			Commands:    []string{"gcc", "make"},
			Packages:    []string{"build-essential"},
			Description: "Essential build tools",
			Category:    Development,
			Critical:    false,
		},
		// Optional requirements - provided by essentials
		{
			Name:        "utilities",
			Commands:    []string{"jq", "ncdu"},
			Packages:    []string{"jq", "ncdu"},
			Description: "Development utilities",
			Category:    Optional,
			Critical:    false,
		},
	}
}

// GetRequirementsByCategory returns requirements filtered by category
func GetRequirementsByCategory(category RequirementCategory) []SystemRequirement {
	var filtered []SystemRequirement
	for _, req := range GetSystemRequirements() {
		if req.Category == category {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// CheckSystemRequirements checks what's missing based on category filter
func CheckSystemRequirements(categories ...RequirementCategory) ([]SystemRequirement, error) {
	var missing []SystemRequirement

	requirements := GetSystemRequirements()
	if len(categories) > 0 {
		// Filter by categories
		var filtered []SystemRequirement
		for _, req := range requirements {
			for _, cat := range categories {
				if req.Category == cat {
					filtered = append(filtered, req)
					break
				}
			}
		}
		requirements = filtered
	}

	for _, req := range requirements {
		allPresent := true
		for _, cmd := range req.Commands {
			if !CommandExists(cmd) {
				allPresent = false
				break
			}
		}

		if !allPresent {
			missing = append(missing, req)
		}
	}

	return missing, nil
}

// EnsureBootstrapRequirements ensures only bootstrap dependencies (git, go)
func EnsureBootstrapRequirements() error {
	missing, err := CheckSystemRequirements(Bootstrap)
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		fmt.Println("✓ Bootstrap dependencies available")
		return nil
	}

	fmt.Printf("Installing bootstrap dependencies...\n")
	var packages []string
	for _, req := range missing {
		fmt.Printf("  - %s: %s\n", req.Name, req.Description)
		packages = append(packages, req.Packages...)
	}

	return InstallSystemPackages(packages)
}

// EnsureRuntimeRequirements ensures runtime dependencies
func EnsureRuntimeRequirements() error {
	missing, err := CheckSystemRequirements(Runtime)
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		fmt.Println("✓ Runtime dependencies available")
		return nil
	}

	fmt.Printf("Installing runtime dependencies...\n")
	var packages []string
	for _, req := range missing {
		if req.Critical {
			fmt.Printf("  - %s: %s (critical)\n", req.Name, req.Description)
		} else {
			fmt.Printf("  - %s: %s\n", req.Name, req.Description)
		}
		packages = append(packages, req.Packages...)
	}

	return InstallSystemPackages(packages)
}
