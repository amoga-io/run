package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VersionManager represents a version manager (nvm, pyenv, sdkman, etc.)
type VersionManager struct {
	Name     string
	HomePath string
	Paths    []string
}

// VersionManagerInfo contains information about a package installed via version manager
type VersionManagerInfo struct {
	Manager     string
	PackageName string
	Paths       []string
	Commands    []string
}

// GetVersionManagers returns all supported version managers
func GetVersionManagers() map[string]VersionManager {
	home := os.Getenv("HOME")

	return map[string]VersionManager{
		"nvm": {
			Name:     "nvm",
			HomePath: filepath.Join(home, ".nvm"),
			Paths: []string{
				filepath.Join(home, ".nvm"),
				filepath.Join(home, ".nvm/versions/node"),
			},
		},
		"pyenv": {
			Name:     "pyenv",
			HomePath: filepath.Join(home, ".pyenv"),
			Paths: []string{
				filepath.Join(home, ".pyenv"),
				filepath.Join(home, ".pyenv/versions"),
			},
		},
		"rbenv": {
			Name:     "rbenv",
			HomePath: filepath.Join(home, ".rbenv"),
			Paths: []string{
				filepath.Join(home, ".rbenv"),
				filepath.Join(home, ".rbenv/versions"),
			},
		},
		"nodenv": {
			Name:     "nodenv",
			HomePath: filepath.Join(home, ".nodenv"),
			Paths: []string{
				filepath.Join(home, ".nodenv"),
				filepath.Join(home, ".nodenv/versions"),
			},
		},
		"sdkman": {
			Name:     "sdkman",
			HomePath: filepath.Join(home, ".sdkman"),
			Paths: []string{
				filepath.Join(home, ".sdkman"),
				filepath.Join(home, ".sdkman/candidates"),
			},
		},
	}
}

// IsVersionManagerInstalled checks if a version manager is installed
func IsVersionManagerInstalled(managerName string) bool {
	managers := GetVersionManagers()
	manager, exists := managers[managerName]
	if !exists {
		return false
	}

	for _, path := range manager.Paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// IsPackageInstalledViaVersionManager checks if a specific package is installed via any version manager
func IsPackageInstalledViaVersionManager(packageName string) bool {
	managers := GetVersionManagers()

	for managerName := range managers {
		if IsPackageInstalledViaSpecificManager(managerName, packageName) {
			return true
		}
	}
	return false
}

// IsPackageInstalledViaSpecificManager checks if a package is installed via a specific version manager
func IsPackageInstalledViaSpecificManager(managerName, packageName string) bool {
	home := os.Getenv("HOME")

	switch managerName {
	case "nvm":
		if packageName == "node" || packageName == "npm" {
			versionsPath := filepath.Join(home, ".nvm/versions/node")
			if matches, _ := filepath.Glob(filepath.Join(versionsPath, "*")); len(matches) > 0 {
				return true
			}
		}
	case "pyenv":
		if packageName == "python" {
			versionsPath := filepath.Join(home, ".pyenv/versions")
			if matches, _ := filepath.Glob(filepath.Join(versionsPath, "*")); len(matches) > 0 {
				return true
			}
		}
	case "sdkman":
		if packageName == "java" {
			candidatesPath := filepath.Join(home, ".sdkman/candidates/java")
			if matches, _ := filepath.Glob(filepath.Join(candidatesPath, "*")); len(matches) > 0 {
				return true
			}
		}
	}

	return false
}

// GetVersionManagerInfo gets detailed information about a package installed via version manager
func GetVersionManagerInfo(packageName string) (*VersionManagerInfo, error) {
	managers := GetVersionManagers()
	home := os.Getenv("HOME")

	for managerName, manager := range managers {
		if !IsPackageInstalledViaSpecificManager(managerName, packageName) {
			continue
		}

		info := &VersionManagerInfo{
			Manager:     managerName,
			PackageName: packageName,
			Paths:       []string{},
			Commands:    []string{},
		}

		// Get paths and commands based on manager and package
		switch managerName {
		case "nvm":
			if packageName == "node" || packageName == "npm" {
				versionsPath := filepath.Join(home, ".nvm/versions/node")
				if matches, _ := filepath.Glob(filepath.Join(versionsPath, "*")); len(matches) > 0 {
					info.Paths = append(info.Paths, matches...)

					// Get current version for uninstall command
					if currentVersion := getSystemVersion(packageName); currentVersion != "" {
						info.Commands = append(info.Commands, fmt.Sprintf("nvm uninstall %s", currentVersion))
					}
				}
				info.Paths = append(info.Paths, manager.HomePath)
			}
		case "pyenv":
			if packageName == "python" {
				versionsPath := filepath.Join(home, ".pyenv/versions")
				if matches, _ := filepath.Glob(filepath.Join(versionsPath, "*")); len(matches) > 0 {
					info.Paths = append(info.Paths, matches...)

					// Get current version for uninstall command
					if currentVersion := getSystemVersion(packageName); currentVersion != "" {
						info.Commands = append(info.Commands, fmt.Sprintf("pyenv uninstall %s", currentVersion))
					}
				}
				info.Paths = append(info.Paths, manager.HomePath)
			}
		case "sdkman":
			if packageName == "java" {
				candidatesPath := filepath.Join(home, ".sdkman/candidates/java")
				if matches, _ := filepath.Glob(filepath.Join(candidatesPath, "*")); len(matches) > 0 {
					info.Paths = append(info.Paths, matches...)

					// Get current version for uninstall command
					if currentVersion := getSystemVersion(packageName); currentVersion != "" {
						info.Commands = append(info.Commands, fmt.Sprintf("sdk uninstall java %s", currentVersion))
					}
				}
				info.Paths = append(info.Paths, manager.HomePath)
			}
		}

		return info, nil
	}

	return nil, fmt.Errorf("package %s not found in any version manager", packageName)
}

// ExecuteVersionManagerCommand executes a version manager command
func ExecuteVersionManagerCommand(command string) error {
	args := strings.Fields(command)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(args[0], args[1:]...)
	return cmd.Run()
}

// getSystemVersion gets the system-installed version of a package
func getSystemVersion(packageName string) string {
	return GetSystemVersion(packageName)
}

// GetSystemVersion gets the system-installed version of a package
func GetSystemVersion(packageName string) string {
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
	case "java":
		if out, err := exec.Command("java", "-version").Output(); err == nil {
			output := string(out)
			// Parse Java version from output like "openjdk version "17.0.2" 2022-01-18"
			if strings.Contains(output, "version") {
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(line, "version") {
						parts := strings.Fields(line)
						for i, part := range parts {
							if part == "version" && i+1 < len(parts) {
								ver := strings.Trim(parts[i+1], "\"")
								verParts := strings.Split(ver, ".")
								if len(verParts) >= 1 {
									return verParts[0]
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// ListInstalledVersions returns all installed versions for a given package via its version manager
func ListInstalledVersions(packageName string) ([]string, error) {
	home := os.Getenv("HOME")
	var versions []string

	switch packageName {
	case "python":
		versionsPath := filepath.Join(home, ".pyenv/versions")
		matches, _ := filepath.Glob(filepath.Join(versionsPath, "*"))
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				versions = append(versions, filepath.Base(match))
			}
		}
	case "node":
		versionsPath := filepath.Join(home, ".nvm/versions/node")
		matches, _ := filepath.Glob(filepath.Join(versionsPath, "*"))
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				versions = append(versions, filepath.Base(match))
			}
		}
	case "java":
		candidatesPath := filepath.Join(home, ".sdkman/candidates/java")
		matches, _ := filepath.Glob(filepath.Join(candidatesPath, "*"))
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				versions = append(versions, filepath.Base(match))
			}
		}
	case "php":
		versionsPath := filepath.Join(home, ".phpenv/versions")
		matches, _ := filepath.Glob(filepath.Join(versionsPath, "*"))
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				versions = append(versions, filepath.Base(match))
			}
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %s", packageName)
	}
	return versions, nil
}

// CheckRequiredVersionManagers ensures all required version managers are installed
func CheckRequiredVersionManagers() error {
	required := []string{"pyenv", "nvm", "sdkman", "phpenv"}
	missing := []string{}
	for _, mgr := range required {
		if !IsVersionManagerInstalled(mgr) {
			missing = append(missing, mgr)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required version managers: %s", strings.Join(missing, ", "))
	}
	return nil
}

// EnsureVersionManagerInstalled checks if the required version manager for a package is installed, and installs it if missing
func EnsureVersionManagerInstalled(packageName string) error {
	// home := os.Getenv("HOME")
	var managerName, installCmd string

	switch packageName {
	case "python":
		managerName = "pyenv"
		installCmd = "curl https://pyenv.run | bash"
	case "node":
		managerName = "nvm"
		installCmd = "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash"
	case "java":
		managerName = "sdkman"
		installCmd = "curl -s 'https://get.sdkman.io' | bash"
	case "php":
		managerName = "phpenv"
		installCmd = "git clone https://github.com/phpenv/phpenv.git ~/.phpenv && git clone https://github.com/php-build/php-build.git ~/.phpenv/plugins/php-build"
	default:
		return nil // Not a version-managed package
	}

	if IsVersionManagerInstalled(managerName) {
		return nil
	}

	// Install the version manager
	fmt.Printf("Installing %s for %s...\n", managerName, packageName)
	cmd := exec.Command("bash", "-c", installCmd)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", managerName, err)
	}
	return nil
}

// SetActiveVersion sets the specified version as the active/default version using the version manager
func SetActiveVersion(packageName, version string) error {
	var setCmd *exec.Cmd

	switch packageName {
	case "python":
		home := os.Getenv("HOME")
		setCmd = exec.Command("pyenv", "global", version)
		setCmd.Env = append(os.Environ(), fmt.Sprintf("PYENV_ROOT=%s/.pyenv", home), fmt.Sprintf("PATH=%s/.pyenv/bin:%s", home, os.Getenv("PATH")))
	case "node":
		setCmd = exec.Command("bash", "-c", fmt.Sprintf("source $HOME/.nvm/nvm.sh && nvm alias default %s", version))
	case "java":
		setCmd = exec.Command("bash", "-c", fmt.Sprintf("source $HOME/.sdkman/bin/sdkman-init.sh && sdk default java %s", version))
	case "php":
		home := os.Getenv("HOME")
		setCmd = exec.Command("phpenv", "global", version)
		setCmd.Env = append(os.Environ(), fmt.Sprintf("PHPENV_ROOT=%s/.phpenv", home), fmt.Sprintf("PATH=%s/.phpenv/bin:%s", home, os.Getenv("PATH")))
	default:
		return fmt.Errorf("set-active not supported for %s", packageName)
	}

	setCmd.Stdout = os.Stdout
	setCmd.Stderr = os.Stderr
	if err := setCmd.Run(); err != nil {
		return fmt.Errorf("failed to set %s version %s as active: %w", packageName, version, err)
	}
	return nil
}
