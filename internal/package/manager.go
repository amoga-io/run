package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/amoga-io/run/internal/logger"
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

// InstallPackage installs a package with dependency checking and rollback support
func (m *Manager) InstallPackage(packageName string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Installing %s (%s)...\n", pkg.Name, pkg.Description)

	// Create rollback manager
	rollbackManager, err := NewRollbackManager()
	if err != nil {
		return fmt.Errorf("failed to create rollback manager: %w", err)
	}

	// Create rollback point
	rollbackPoint, err := rollbackManager.CreateRollbackPoint(packageName, "install")
	if err != nil {
		return fmt.Errorf("failed to create rollback point: %w", err)
	}

	// Defer rollback cleanup on success
	defer func() {
		if err == nil {
			rollbackManager.CleanupRollbackPoint(rollbackPoint.ID)
		}
	}()

	// Validate dependencies before installation
	if err := ValidateDependencies(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// Smart suggestions before installation
	m.provideSuggestions(pkg)

	// Step 1: Check and install dependencies
	if err := m.installDependencies(pkg); err != nil {
		// Execute rollback on dependency failure
		rollbackPoint.ExecuteRollback()
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Step 2: Check if package is already installed, remove if so
	if m.isPackageInstalled(pkg) {
		fmt.Printf("Package %s is already installed, removing first...\n", pkg.Name)
		if err := m.RemovePackage(packageName); err != nil {
			// Execute rollback on removal failure
			rollbackPoint.ExecuteRollback()
			return fmt.Errorf("failed to remove existing package: %w", err)
		}
	}

	// Step 3: Execute installation script
	if err := m.executeInstallScript(pkg); err != nil {
		// Execute rollback on installation failure
		rollbackPoint.ExecuteRollback()
		return fmt.Errorf("installation failed: %w", err)
	}

	return nil
}

// InstallPackageWithArgs installs a package and passes extra arguments to the install script
func (m *Manager) InstallPackageWithArgs(packageName string, extraArgs []string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Installing %s (%s)...\n", pkg.Name, pkg.Description)

	// Validate dependencies before installation
	if err := ValidateDependencies(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

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

// executeInstallScript executes the installation script for a package
func (m *Manager) executeInstallScript(pkg Package) error {
	log := logger.GetLogger().WithOperation("execute_install_script").WithPackage(pkg.Name)

	// Validate script path
	if err := ValidatePackageName(pkg.Name); err != nil {
		log.Error("Invalid package name: %v", err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	// Resolve script path
	safePath, err := ValidatePath(pkg.ScriptPath)
	if err != nil {
		log.Error("Invalid script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	resolvedScriptPath, err := safePath.Resolve(m.repoPath)
	if err != nil {
		log.Error("Failed to resolve script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	// Check if script exists
	if _, err := os.Stat(resolvedScriptPath); os.IsNotExist(err) {
		log.Error("Script file does not exist: %s", resolvedScriptPath)
		return fmt.Errorf("installation script not found: %s", resolvedScriptPath)
	}

	// Make script executable
	if err := os.Chmod(resolvedScriptPath, 0755); err != nil {
		log.Error("Failed to make script executable %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	log.Info("Executing installation script: %s", resolvedScriptPath)
	fmt.Printf("Executing installation script for %s...\n", pkg.Name)

	// Execute script
	cmd := exec.Command(resolvedScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = m.repoPath

	if err := cmd.Run(); err != nil {
		log.Error("Installation script failed %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("installation script failed: %w", err)
	}

	log.Info("Installation script completed successfully")
	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// executeInstallScriptWithArgs executes the installation script with extra arguments
func (m *Manager) executeInstallScriptWithArgs(pkg Package, extraArgs []string) error {
	log := logger.GetLogger().WithOperation("execute_install_script_with_args").WithPackage(pkg.Name)

	// Validate script path
	if err := ValidatePackageName(pkg.Name); err != nil {
		log.Error("Invalid package name: %v", err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	// Resolve script path
	safePath, err := ValidatePath(pkg.ScriptPath)
	if err != nil {
		log.Error("Invalid script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	resolvedScriptPath, err := safePath.Resolve(m.repoPath)
	if err != nil {
		log.Error("Failed to resolve script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	// Check if script exists
	if _, err := os.Stat(resolvedScriptPath); os.IsNotExist(err) {
		log.Error("Script file does not exist: %s", resolvedScriptPath)
		return fmt.Errorf("installation script not found: %s", resolvedScriptPath)
	}

	// Make script executable
	if err := os.Chmod(resolvedScriptPath, 0755); err != nil {
		log.Error("Failed to make script executable %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	log.Info("Executing installation script with arguments: %s %v", resolvedScriptPath, extraArgs)
	fmt.Printf("Executing installation script for %s with arguments...\n", pkg.Name)

	// Execute script with arguments
	cmd := exec.Command(resolvedScriptPath, extraArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = m.repoPath

	if err := cmd.Run(); err != nil {
		log.Error("Installation script failed %s %v: %v", resolvedScriptPath, extraArgs, err)
		return fmt.Errorf("installation script failed: %w", err)
	}

	log.Info("Installation script completed successfully")
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
		fmt.Printf("Package %s is not detected via CLI commands, attempting user-level removal...\n", pkg.Name)
	}

	switch packageName {
	case "python":
		return m.removePython()
	case "node":
		return m.removeNode()
	case "php":
		return m.removePHP()
	case "essentials":
		return m.removeEssentials()
	case "docker":
		return m.removeDocker()
	case "nginx":
		return m.removeNginx()
	case "postgres":
		return m.removePostgres()
	case "java":
		return m.removeJava()
	case "pm2":
		return m.removePM2()
	default:
		return fmt.Errorf("safe removal not implemented for package: %s", packageName)
	}
}

// getSystemVersion gets the system-installed version of a package
func (m *Manager) getSystemVersion(packageName string) string {
	switch packageName {
	case "python":
		if out, err := exec.Command("python3", "--version").Output(); err == nil {
			parts := strings.Fields(string(out))
			if len(parts) >= 2 {
				ver := parts[1]
				verParts := strings.Split(ver, ".")
				if len(verParts) >= 2 {
					return verParts[0] + "." + verParts[1]
				}
			}
		}
	case "node":
		if out, err := exec.Command("node", "--version").Output(); err == nil {
			ver := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
			verParts := strings.Split(ver, ".")
			if len(verParts) >= 2 {
				return verParts[0]
			}
		}
	case "php":
		if out, err := exec.Command("php", "-v").Output(); err == nil {
			parts := strings.Fields(string(out))
			if len(parts) >= 2 {
				ver := parts[1]
				verParts := strings.Split(ver, ".")
				if len(verParts) >= 2 {
					return verParts[0] + "." + verParts[1]
				}
			}
		}
	}
	return ""
}

// Only remove user-installed Python versions (never system python3)
func (m *Manager) removePython() error {
	fmt.Println("Stopping Python services...")

	systemVersion := m.getSystemVersion("python")
	userVersions := []string{"3.10", "3.11", "3.12"} // Add more as needed
	removedAny := false

	for _, v := range userVersions {
		if v == systemVersion {
			fmt.Printf("Refusing to remove system Python (%s). This would break your OS.\n", v)
			continue
		}
		// Only remove from /usr/local, never from /usr/bin
		fmt.Printf("Attempting to remove user-installed Python %s...\n", v)
		commands := [][]string{
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

// Only remove user-installed Node.js (never system nodejs)
func (m *Manager) removeNode() error {
	systemVersion := m.getSystemVersion("node")
	userVersions := []string{"18", "20", "21"} // Add more as needed
	removedAny := false

	for _, v := range userVersions {
		if v == systemVersion {
			fmt.Printf("Refusing to remove system Node.js (%s). This would break your OS.\n", v)
			continue
		}
		fmt.Printf("Attempting to remove user-installed Node.js %s...\n", v)
		commands := [][]string{
			{"sudo", "rm", "-rf", "/usr/local/lib/node_modules"},
			{"sudo", "rm", "-rf", "/usr/local/bin/node*"},
			{"sudo", "rm", "-rf", "/usr/local/bin/npm*"},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".npm")},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".npm-global")},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".node-gyp")},
			{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".node_repl_history")},
		}
		m.executeRemovalCommands("Node.js "+v, commands)
		removedAny = true
	}
	if !removedAny {
		fmt.Println("No user-installed Node.js versions found to remove.")
	}
	return nil
}

// Only remove user-installed PHP (never system php)
func (m *Manager) removePHP() error {
	systemVersion := m.getSystemVersion("php")
	userVersions := []string{"8.1", "8.2", "8.3"}
	removedAny := false

	for _, v := range userVersions {
		if v == systemVersion {
			fmt.Printf("Refusing to remove system PHP (%s). This would break your OS.\n", v)
			continue
		}
		fmt.Printf("Attempting to remove user-installed PHP %s...\n", v)
		commands := [][]string{
			{"sudo", "rm", "-rf", "/usr/local/bin/php" + v + "*"},
			{"sudo", "rm", "-rf", "/usr/local/lib/php" + v + "*"},
			{"sudo", "rm", "-rf", "/usr/local/etc/php" + v + "*"},
			{"sudo", "rm", "-rf", "/usr/local/pear"},
		}
		m.executeRemovalCommands("PHP "+v, commands)
		removedAny = true
	}
	if !removedAny {
		fmt.Println("No user-installed PHP versions found to remove.")
	}
	return nil
}

// Essentials: Only remove user-level tools, never system packages
func (m *Manager) removeEssentials() error {
	fmt.Println("Removing user-level system essentials...")
	commands := [][]string{
		{"sudo", "rm", "-rf", "/usr/local/bin/redis-server"},
		{"sudo", "rm", "-rf", "/usr/local/bin/gcc"},
		{"sudo", "rm", "-rf", "/usr/local/bin/make"},
		{"sudo", "rm", "-rf", "/usr/local/bin/g++"},
		{"sudo", "rm", "-rf", "/usr/local/bin/ncdu"},
		{"sudo", "rm", "-rf", "/usr/local/bin/jq"},
		{"sudo", "rm", "-rf", "/usr/local/bin/curl"},
		{"sudo", "rm", "-rf", "/usr/local/bin/wget"},
		{"sudo", "rm", "-rf", "/usr/local/lib/redis*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/gcc*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/make*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/g++*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/ncdu*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/jq*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/curl*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/wget*"},
	}
	return m.executeRemovalCommands("User-level system essentials", commands)
}

// executeRemovalCommands executes removal commands with proper error handling
func (m *Manager) executeRemovalCommands(packageName string, commands [][]string) error {
	log := logger.GetLogger().WithOperation("execute_removal_commands").WithPackage(packageName)

	log.Info("Executing removal commands: %d", len(commands))

	var errors []string

	for i, cmdArgs := range commands {
		if len(cmdArgs) == 0 {
			continue
		}

		log.Info("Executing removal command %d: %v", i, cmdArgs)

		// Handle shell operators like || true
		if len(cmdArgs) >= 3 && cmdArgs[len(cmdArgs)-2] == "||" && cmdArgs[len(cmdArgs)-1] == "true" {
			// Execute command and ignore errors if || true
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)-2]...)
			if err := cmd.Run(); err != nil {
				log.Warn("Removal command failed (ignored due to || true): %v - %v", cmdArgs, err)
			}
			continue
		}

		// Handle other shell operators like ||
		if len(cmdArgs) >= 3 && cmdArgs[len(cmdArgs)-2] == "||" {
			// Execute main command, if it fails, execute the fallback
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)-2]...)
			if err := cmd.Run(); err != nil {
				log.Warn("Main removal command failed, trying fallback: %v - %v", cmdArgs, err)
				// Execute fallback command
				fallbackCmd := exec.Command(cmdArgs[len(cmdArgs)-1])
				if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
					log.Error("Fallback removal command also failed: %v (fallback error: %v)", cmdArgs, fallbackErr)
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
			log.Warn("Removal command failed: %v - %v", cmdArgs, err)
			fmt.Printf("Warning: %s (continuing...)\n", errorMsg)
			errors = append(errors, errorMsg)
		}
	}

	// Return aggregated errors if any occurred
	if len(errors) > 0 {
		return fmt.Errorf("some removal commands failed: %s", strings.Join(errors, "; "))
	}

	log.Info("All removal commands completed")
	fmt.Printf("✓ %s removed completely\n", packageName)
	return nil
}

// Only remove user-level Docker files
func (m *Manager) removeDocker() error {
	fmt.Println("Removing user-level Docker files...")
	commands := [][]string{
		{"sudo", "rm", "-rf", "/usr/local/bin/docker"},
		{"sudo", "rm", "-rf", "/usr/local/bin/docker-compose"},
		{"sudo", "rm", "-rf", "/usr/local/bin/docker*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/docker*"},
		{"sudo", "rm", "-rf", filepath.Join(os.Getenv("HOME"), ".docker")},
		{"sudo", "rm", "-rf", "/usr/local/share/docker*"},
	}
	return m.executeRemovalCommands("User-level Docker", commands)
}

// Only remove user-level Nginx files
func (m *Manager) removeNginx() error {
	fmt.Println("Removing user-level Nginx files...")
	commands := [][]string{
		{"sudo", "rm", "-rf", "/usr/local/bin/nginx"},
		{"sudo", "rm", "-rf", "/usr/local/lib/nginx*"},
		{"sudo", "rm", "-rf", "/usr/local/share/nginx*"},
		{"sudo", "rm", "-rf", filepath.Join(os.Getenv("HOME"), ".nginx")},
	}
	return m.executeRemovalCommands("User-level Nginx", commands)
}

// Only remove user-level Postgres files
func (m *Manager) removePostgres() error {
	fmt.Println("Removing user-level Postgres files...")
	commands := [][]string{
		{"sudo", "rm", "-rf", "/usr/local/bin/psql"},
		{"sudo", "rm", "-rf", "/usr/local/lib/postgres*"},
		{"sudo", "rm", "-rf", "/usr/local/share/postgres*"},
		{"sudo", "rm", "-rf", filepath.Join(os.Getenv("HOME"), ".psql_history")},
	}
	return m.executeRemovalCommands("User-level Postgres", commands)
}

// Only remove user-level Java files
func (m *Manager) removeJava() error {
	fmt.Println("Removing user-level Java files...")
	commands := [][]string{
		{"sudo", "rm", "-rf", "/usr/local/bin/java*"},
		{"sudo", "rm", "-rf", "/usr/local/lib/jvm*"},
		{"sudo", "rm", "-rf", "/usr/local/share/java*"},
	}
	return m.executeRemovalCommands("User-level Java", commands)
}

// Only remove user-level PM2 files
func (m *Manager) removePM2() error {
	fmt.Println("Removing user-level PM2 files...")
	commands := [][]string{
		{"pm2", "kill", "||", "true"},
		{"npm", "uninstall", "-g", "pm2", "||", "true"},
		{"rm", "-rf", filepath.Join(os.Getenv("HOME"), ".pm2")},
	}
	return m.executeRemovalCommands("User-level PM2", commands)
}
