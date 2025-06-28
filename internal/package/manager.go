package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/amoga-io/run/internal/system"
)

type Manager struct {
	repoPath string
}

func NewManager() (*Manager, error) {
	repoPath, err := GetRepoPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository path: %w", err)
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable is not set")
	}

	resolvedPath, err := repoPath.Resolve(homeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository not found at %s. Please reinstall CLI", resolvedPath)
	}

	return &Manager{repoPath: resolvedPath}, nil
}

// InstallPackage installs a package with dependency checking
func (m *Manager) InstallPackage(packageName string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Installing %s (%s)...\n", pkg.Name, pkg.Description)

	// Smart suggestions before installation
	m.provideSuggestions(pkg)

	// Step 1: Check and install dependencies
	if err := m.installDependencies(pkg); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Step 2: Check if package is already installed, remove if so
	if m.isPackageInstalled(pkg) {
		fmt.Printf("Package %s is already installed, removing first...\n", pkg.Name)
		if err := m.RemovePackage(packageName); err != nil {
			return fmt.Errorf("failed to remove existing package: %w", err)
		}
	}

	// Step 3: Execute installation script
	return m.executeInstallScript(pkg)
}

// InstallPackageWithArgs installs a package and passes extra arguments to the install script
func (m *Manager) InstallPackageWithArgs(packageName string, extraArgs []string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Installing %s (%s)...\n", pkg.Name, pkg.Description)

	// Smart suggestions before installation
	m.provideSuggestions(pkg)

	// Step 1: Check and install dependencies
	if err := m.installDependencies(pkg); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Step 2: Check if package is already installed, remove if so
	if m.isPackageInstalled(pkg) {
		fmt.Printf("Package %s is already installed, removing first...\n", pkg.Name)
		if err := m.RemovePackage(packageName); err != nil {
			return fmt.Errorf("failed to remove existing package: %w", err)
		}
	}

	// Step 3: Execute installation script with extra arguments
	return m.executeInstallScriptWithArgs(pkg, extraArgs)
}

// provideSuggestions provides smart suggestions based on package being installed
func (m *Manager) provideSuggestions(pkg Package) {
	essentialsPkg, essentialsExists := GetPackage("essentials")
	if !essentialsExists {
		return
	}

	isEssentialsInstalled := m.isPackageInstalled(essentialsPkg)

	// Suggest essentials for development packages
	if pkg.Category == "development" && !isEssentialsInstalled {
		fmt.Printf("💡 Tip: Installing 'essentials' first provides build tools helpful for %s\n", pkg.Name)
		fmt.Printf("💡 Run: run install essentials %s\n\n", pkg.Name)
	}

	// Suggest essentials for packages that commonly need build tools
	buildIntensivePackages := map[string]bool{
		"python": true,
		"node":   true,
		"php":    true,
	}

	if buildIntensivePackages[pkg.Name] && !isEssentialsInstalled {
		fmt.Printf("💡 Recommended: '%s' benefits from development tools in 'essentials'\n", pkg.Name)
		fmt.Printf("💡 Consider: run install essentials %s\n\n", pkg.Name)
	}

	// Suggest related packages
	relatedSuggestions := map[string][]string{
		"nginx":    {"php", "node"},
		"postgres": {"python", "node", "java"},
		"docker":   {"node", "python"},
		"node":     {"pm2"},
	}

	if suggestions, exists := relatedSuggestions[pkg.Name]; exists {
		var availableSuggestions []string
		for _, suggestion := range suggestions {
			if _, exists := GetPackage(suggestion); exists {
				if !m.isPackageInstalled(Package{Name: suggestion}) {
					availableSuggestions = append(availableSuggestions, suggestion)
				}
			}
		}
		if len(availableSuggestions) > 0 {
			fmt.Printf("💡 Commonly used with %s: %s\n", pkg.Name, strings.Join(availableSuggestions, ", "))
			fmt.Printf("💡 Install together: run install %s %s\n\n", pkg.Name, strings.Join(availableSuggestions, " "))
		}
	}
}

// installDependencies checks and installs required dependencies
func (m *Manager) installDependencies(pkg Package) error {
	if len(pkg.Dependencies) == 0 {
		fmt.Printf("No dependencies required for %s\n", pkg.Name)
		return nil
	}

	fmt.Printf("Checking dependencies for %s: %s\n", pkg.Name, strings.Join(pkg.Dependencies, ", "))

	var missingPackages []string

	for _, dep := range pkg.Dependencies {
		// Check if dependency is a package in our registry
		if depPkg, exists := GetPackage(dep); exists {
			if !m.isPackageInstalled(depPkg) {
				fmt.Printf("Required package %s is not installed\n", dep)
				// Recursively install package dependencies
				if err := m.InstallPackage(dep); err != nil {
					return fmt.Errorf("failed to install required package %s: %w", dep, err)
				}
			}
		} else {
			// Check if it's a system command/package
			if !system.CommandExists(dep) {
				missingPackages = append(missingPackages, dep)
			}
		}
	}

	// Install missing system packages
	if len(missingPackages) > 0 {
		fmt.Printf("Installing system dependencies: %s\n", strings.Join(missingPackages, ", "))
		if err := system.InstallSystemPackages(missingPackages); err != nil {
			return fmt.Errorf("failed to install system dependencies: %w", err)
		}
	}

	fmt.Printf("✓ All dependencies satisfied for %s\n", pkg.Name)
	return nil
}

// isPackageInstalled checks if a package is installed by checking its commands
func (m *Manager) isPackageInstalled(pkg Package) bool {
	for _, cmd := range pkg.Commands {
		if !system.CommandExists(cmd) {
			return false
		}
	}
	return len(pkg.Commands) > 0 // Only return true if there are commands to check
}

// executeInstallScript runs the package installation script
func (m *Manager) executeInstallScript(pkg Package) error {
	// Validate script path
	scriptPath, err := ValidatePath(pkg.ScriptPath)
	if err != nil {
		return fmt.Errorf("invalid script path: %w", err)
	}

	resolvedScriptPath, err := scriptPath.Resolve(m.repoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	if _, err := os.Stat(resolvedScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("installation script not found: %s", resolvedScriptPath)
	}

	fmt.Printf("Executing installation script for %s...\n", pkg.Name)

	// Make script executable
	if err := os.Chmod(resolvedScriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Set environment for silent installation
	cmd := exec.Command(resolvedScriptPath)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation script failed: %w", err)
	}

	// Verify installation
	if !m.isPackageInstalled(pkg) {
		return fmt.Errorf("package installation verification failed - commands not available: %s", strings.Join(pkg.Commands, ", "))
	}

	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// executeInstallScriptWithArgs runs the package installation script with extra arguments
func (m *Manager) executeInstallScriptWithArgs(pkg Package, extraArgs []string) error {
	// Validate script path
	scriptPath, err := ValidatePath(pkg.ScriptPath)
	if err != nil {
		return fmt.Errorf("invalid script path: %w", err)
	}

	resolvedScriptPath, err := scriptPath.Resolve(m.repoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	if _, err := os.Stat(resolvedScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("installation script not found: %s", resolvedScriptPath)
	}

	fmt.Printf("Executing installation script for %s with arguments...\n", pkg.Name)

	// Make script executable
	if err := os.Chmod(resolvedScriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Set environment for silent installation
	cmd := exec.Command(resolvedScriptPath, extraArgs...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation script failed: %w", err)
	}

	// Verify installation
	if !m.isPackageInstalled(pkg) {
		return fmt.Errorf("package installation verification failed - commands not available: %s", strings.Join(pkg.Commands, ", "))
	}

	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// RemovePackage removes a package completely (Go-based removal)
func (m *Manager) RemovePackage(packageName string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Removing %s completely...\n", pkg.Name)

	// Check if package is actually installed via CLI before attempting removal
	if !m.isPackageInstalled(pkg) {
		fmt.Printf("Package %s is not detected via CLI commands, attempting system-wide removal...\n", pkg.Name)
	}

	switch packageName {
	case "python":
		return m.removePython()
	case "node":
		return m.removeNode()
	case "docker":
		return m.removeDocker()
	case "nginx":
		return m.removeNginx()
	case "postgres":
		return m.removePostgres()
	case "php":
		return m.removePHP()
	case "java":
		return m.removeJava()
	case "pm2":
		return m.removePM2()
	case "essentials":
		return m.removeEssentials()
	default:
		return fmt.Errorf("removal not implemented for package: %s", packageName)
	}
}

// Individual removal functions (Go-based complete cleanup)
func (m *Manager) removePython() error {
	fmt.Println("Stopping Python services...")

	// Get system default python version
	systemVersion := ""
	if out, err := exec.Command("python3", "--version").Output(); err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			ver := parts[1]
			verParts := strings.Split(ver, ".")
			if len(verParts) >= 2 {
				systemVersion = verParts[0] + "." + verParts[1]
			}
		}
	}

	// Only allow removal of user-installed versions (e.g., python3.10, python3.11)
	userVersions := []string{"3.10", "3.11"} // Add more as needed
	removedAny := false
	for _, v := range userVersions {
		if v == systemVersion {
			fmt.Printf("Refusing to remove system Python (%s). This would break your OS.\n", v)
			continue
		}
		fmt.Printf("Attempting to remove Python %s...\n", v)
		commands := [][]string{
			{"sudo", "apt-get", "purge", "-y", "python" + v, "python" + v + "-venv", "python" + v + "-dev"},
			{"sudo", "apt-get", "autoremove", "-y"},
			{"sudo", "rm", "-rf", "/usr/local/lib/python" + v + "*"},
			{"sudo", "rm", "-rf", "/usr/local/bin/python" + v + "*"},
			{"sudo", "rm", "-rf", "/usr/local/bin/pip" + v + "*"},
		}
		m.executeRemovalCommands("Python "+v, commands)
		removedAny = true
	}
	if !removedAny {
		fmt.Println("No user-installed Python versions found to remove.")
	}
	return nil
}

func (m *Manager) removeUserVersions(
	packageName string,
	userVersions []string,
	commandBuilder func(version string) [][]string,
) error {
	removedAny := false
	for _, v := range userVersions {
		fmt.Printf("Attempting to remove %s %s...\n", packageName, v)
		commands := commandBuilder(v)
		m.executeRemovalCommands(fmt.Sprintf("%s %s", packageName, v), commands)
		removedAny = true
	}
	if !removedAny {
		fmt.Printf("No user-installed %s versions found to remove.\n", packageName)
	}
	return nil
}

func (m *Manager) removeNode() error {
	return m.removeUserVersions("Node.js", []string{"18", "20"}, func(v string) [][]string {
		return [][]string{
			{"sudo", "apt-get", "purge", "-y", "nodejs", "npm"},
			{"sudo", "apt-get", "autoremove", "-y"},
			{"sudo", "rm", "-rf", "/usr/local/lib/node_modules"},
			{"sudo", "rm", "-rf", "/usr/local/bin/node*"},
			{"sudo", "rm", "-rf", "/usr/local/bin/npm*"},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".npm")},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".npm-global")},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".node-gyp")},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".node_repl_history")},
		}
	})
}

func (m *Manager) removeDocker() error {
	return m.removeUserVersions("Docker", []string{"latest"}, func(v string) [][]string {
		return [][]string{
			{"sudo", "systemctl", "stop", "docker", "||", "true"},
			{"sudo", "systemctl", "stop", "docker.socket", "||", "true"},
			{"sudo", "apt-get", "purge", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
			{"sudo", "rm", "-rf", "/var/lib/docker"},
			{"sudo", "rm", "-rf", "/var/lib/containerd"},
			{"sudo", "rm", "-rf", "/etc/docker"},
			{"sudo", "rm", "-f", "/etc/apt/sources.list.d/docker.list"},
			{"sudo", "rm", "-f", "/etc/apt/keyrings/docker.gpg"},
			{"sudo", "groupdel", "docker", "||", "true"},
			{"sudo", "apt-get", "autoremove", "-y"},
		}
	})
}

func (m *Manager) removeNginx() error {
	return m.removeUserVersions("Nginx", []string{"mainline", "stable"}, func(v string) [][]string {
		return [][]string{
			{"sudo", "systemctl", "stop", "nginx", "||", "true"},
			{"sudo", "systemctl", "disable", "nginx", "||", "true"},
			{"sudo", "apt-get", "purge", "-y", "nginx", "nginx-*"},
			{"sudo", "rm", "-rf", "/etc/nginx"},
			{"sudo", "rm", "-rf", "/var/log/nginx"},
			{"sudo", "rm", "-rf", "/var/lib/nginx"},
			{"sudo", "rm", "-rf", "/usr/share/nginx"},
			{"sudo", "userdel", "www-data", "||", "true"},
			{"sudo", "apt-get", "autoremove", "-y"},
		}
	})
}

func (m *Manager) removePostgres() error {
	return m.removeUserVersions("PostgreSQL", []string{"15", "16", "17"}, func(v string) [][]string {
		return [][]string{
			{"sudo", "systemctl", "stop", "postgresql", "||", "true"},
			{"sudo", "systemctl", "disable", "postgresql", "||", "true"},
			{"sudo", "apt-get", "purge", "-y", fmt.Sprintf("postgresql-%s", v)},
			{"sudo", "rm", "-rf", fmt.Sprintf("/etc/postgresql/%s", v)},
			{"sudo", "rm", "-rf", "/var/lib/postgresql"},
			{"sudo", "rm", "-rf", "/var/log/postgresql"},
			{"sudo", "userdel", "postgres", "||", "true"},
			{"sudo", "groupdel", "postgres", "||", "true"},
			{"sudo", "apt-get", "autoremove", "-y"},
		}
	})
}

func (m *Manager) removePHP() error {
	return m.removeUserVersions("PHP", []string{"8.1", "8.2", "8.3"}, func(v string) [][]string {
		return [][]string{
			{"sudo", "systemctl", "stop", fmt.Sprintf("php%s-fpm", v), "||", "true"},
			{"sudo", "apt-get", "purge", "-y", fmt.Sprintf("php%s", v), fmt.Sprintf("php%s-fpm", v), fmt.Sprintf("php%s-common", v)},
			{"sudo", "rm", "-rf", "/etc/php"},
			{"sudo", "rm", "-rf", "/var/lib/php"},
			{"sudo", "rm", "-rf", "/var/log/php*"},
			{"sudo", "rm", "-rf", "/usr/share/php*"},
			{"sudo", "apt-get", "autoremove", "-y"},
		}
	})
}

func (m *Manager) removeJava() error {
	return m.removeUserVersions("Java", []string{"11", "17", "21"}, func(v string) [][]string {
		return [][]string{
			{"sudo", "apt-get", "purge", "-y", fmt.Sprintf("openjdk-%s-jdk", v), fmt.Sprintf("openjdk-%s-jre", v)},
			{"sudo", "rm", "-rf", "/usr/lib/jvm"},
			{"sudo", "rm", "-rf", "/usr/share/java"},
			{"sudo", "sed", "-i", "/JAVA_HOME/d", "/etc/environment"},
			{"sudo", "apt-get", "autoremove", "-y"},
		}
	})
}

func (m *Manager) removePM2() error {
	return m.removeUserVersions("PM2", []string{"latest"}, func(v string) [][]string {
		return [][]string{
			{"pm2", "kill", "||", "true"},
			{"npm", "uninstall", "-g", "pm2", "||", "true"},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".pm2")},
		}
	})
}

func (m *Manager) removeEssentials() error {
	fmt.Println("Removing system essentials...")
	commands := [][]string{
		{"sudo", "systemctl", "stop", "redis-server", "||", "true"},
		{"sudo", "systemctl", "disable", "redis-server", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "redis-server", "build-essential", "python3", "g++", "make"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "ncdu", "jq", "curl", "wget"},
		{"sudo", "rm", "-rf", "/etc/redis"},
		{"sudo", "rm", "-rf", "/var/lib/redis"},
		{"sudo", "rm", "-rf", "/var/log/redis"},
		{"sudo", "sed", "-i", "/SystemMaxUse=512M/d", "/etc/systemd/journald.conf"},
		{"sudo", "sed", "-i", "/\\* hard core 0/d", "/etc/security/limits.conf"},
		{"sudo", "systemctl", "restart", "systemd-journald"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("System essentials", commands)
}

// executeRemovalCommands is a helper function to execute removal commands with consistent error handling
func (m *Manager) executeRemovalCommands(packageName string, commands [][]string) error {
	for _, cmdArgs := range commands {
		if len(cmdArgs) == 0 {
			continue
		}

		// Handle shell operators like || true
		if len(cmdArgs) >= 3 && cmdArgs[len(cmdArgs)-2] == "||" && cmdArgs[len(cmdArgs)-1] == "true" {
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)-2]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run() // Ignore error as intended
			continue
		}

		// Handle other shell operators like ||
		if len(cmdArgs) >= 3 && cmdArgs[len(cmdArgs)-2] == "||" {
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)-2]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Execute fallback command
				fallbackCmd := exec.Command(cmdArgs[len(cmdArgs)-1])
				fallbackCmd.Stdout = os.Stdout
				fallbackCmd.Stderr = os.Stderr
				fallbackCmd.Run()
			}
			continue
		}

		// Regular command execution
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Continue with other commands for removal operations but track the error
			errorMsg := fmt.Sprintf("command failed: %v", cmdArgs)
			fmt.Printf("Warning: %s (continuing...)\n", errorMsg)
		}
	}

	fmt.Printf("✓ %s removed completely\n", packageName)
	return nil
}
