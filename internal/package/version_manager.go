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
