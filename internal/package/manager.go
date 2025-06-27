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
	repoPath := filepath.Join(os.Getenv("HOME"), ".run")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository not found at %s. Please reinstall CLI", repoPath)
	}
	return &Manager{repoPath: repoPath}, nil
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
	scriptPath := filepath.Join(m.repoPath, pkg.ScriptPath)

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("installation script not found: %s", scriptPath)
	}

	fmt.Printf("Executing installation script for %s...\n", pkg.Name)

	// Make script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Set environment for silent installation
	cmd := exec.Command("bash", scriptPath)
	cmd.Env = append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
		"NEEDRESTART_MODE=a", // Automatic restart services
	)

	// Run silently but capture errors
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation script failed: %w\nOutput: %s", err, string(output))
	}

	// Verify installation
	if !m.isPackageInstalled(pkg) {
		return fmt.Errorf("package installation verification failed - commands not available: %s", strings.Join(pkg.Commands, ", "))
	}

	fmt.Printf("✓ Successfully installed %s\n", pkg.Name)
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
	commands := [][]string{
		{"sudo", "pkill", "-f", "python3", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "python3*", "python3-*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/python3*"},
		{"sudo", "rm", "-rf", "/usr/local/bin/python*"},
		{"sudo", "rm", "-rf", "/usr/local/bin/pip*"},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".cache/pip")},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".local/lib/python*")},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("Python", commands)
}

func (m *Manager) removeNode() error {
	fmt.Println("Stopping Node.js processes...")
	commands := [][]string{
		{"sudo", "pkill", "-f", "node", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "nodejs", "npm"},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".npm")},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".npm-global")},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".node-gyp")},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".node_repl_history")},
		{"sudo", "rm", "-rf", "/usr/local/lib/node_modules"},
		{"sudo", "rm", "-rf", "/usr/local/bin/node*"},
		{"sudo", "rm", "-rf", "/usr/local/bin/npm*"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("Node.js", commands)
}

func (m *Manager) removeDocker() error {
	fmt.Println("Stopping Docker services...")
	commands := [][]string{
		{"sudo", "systemctl", "stop", "docker", "||", "true"},
		{"sudo", "systemctl", "stop", "docker.socket", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
		{"sudo", "rm", "-rf", "/var/lib/docker"},
		{"sudo", "rm", "-rf", "/var/lib/containerd"},
		{"sudo", "rm", "-rf", "/etc/docker"},
		{"sudo", "rm", "-f", "/etc/apt/sources.list.d/docker.list"},
		{"sudo", "rm", "-f", "/etc/apt/keyrings/docker.gpg"},
		{"sudo", "groupdel", "docker", "||", "true"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("Docker", commands)
}

func (m *Manager) removeNginx() error {
	fmt.Println("Stopping Nginx services...")
	commands := [][]string{
		{"sudo", "systemctl", "stop", "nginx", "||", "true"},
		{"sudo", "systemctl", "disable", "nginx", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "nginx", "nginx-*"},
		{"sudo", "rm", "-rf", "/etc/nginx"},
		{"sudo", "rm", "-rf", "/var/log/nginx"},
		{"sudo", "rm", "-rf", "/var/lib/nginx"},
		{"sudo", "rm", "-rf", "/usr/share/nginx"},
		{"sudo", "userdel", "www-data", "||", "true"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("Nginx", commands)
}

func (m *Manager) removePostgres() error {
	fmt.Println("Stopping PostgreSQL services...")
	commands := [][]string{
		{"sudo", "systemctl", "stop", "postgresql", "||", "true"},
		{"sudo", "systemctl", "disable", "postgresql", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "postgresql*", "postgresql-*"},
		{"sudo", "rm", "-rf", "/etc/postgresql"},
		{"sudo", "rm", "-rf", "/var/lib/postgresql"},
		{"sudo", "rm", "-rf", "/var/log/postgresql"},
		{"sudo", "userdel", "postgres", "||", "true"},
		{"sudo", "groupdel", "postgres", "||", "true"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("PostgreSQL", commands)
}

func (m *Manager) removePHP() error {
	fmt.Println("Stopping PHP services...")
	commands := [][]string{
		{"sudo", "systemctl", "stop", "php*-fpm", "||", "true"},
		{"sudo", "apt-get", "purge", "-y", "-qq", "php*", "php*-*"},
		{"sudo", "rm", "-rf", "/etc/php"},
		{"sudo", "rm", "-rf", "/var/lib/php"},
		{"sudo", "rm", "-rf", "/var/log/php*"},
		{"sudo", "rm", "-rf", "/usr/share/php*"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("PHP", commands)
}

func (m *Manager) removeJava() error {
	fmt.Println("Removing Java...")
	commands := [][]string{
		{"sudo", "apt-get", "purge", "-y", "-qq", "openjdk-*", "default-jdk", "default-jre"},
		{"sudo", "rm", "-rf", "/usr/lib/jvm"},
		{"sudo", "rm", "-rf", "/usr/share/java"},
		{"sudo", "sed", "-i", "/JAVA_HOME/d", "/etc/environment"},
		{"sudo", "apt-get", "autoremove", "-y", "-qq"},
	}
	return m.executeRemovalCommands("Java", commands)
}

func (m *Manager) removePM2() error {
	fmt.Println("Stopping PM2 processes...")
	commands := [][]string{
		{"pm2", "kill", "||", "true"},
		{"npm", "uninstall", "-g", "pm2", "||", "true"},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".pm2")},
	}
	return m.executeRemovalCommands("PM2", commands)
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
	if err := system.ExecuteCommands(commands); err != nil {
		return fmt.Errorf("failed to remove %s: %w", packageName, err)
	}
	fmt.Printf("✓ %s removed completely\n", packageName)
	return nil
}
