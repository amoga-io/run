package pkg

import (
	"fmt"
	"os"
	"path/filepath"
)

// PackageConfig represents package configuration
type PackageConfig struct {
	Name              string   `yaml:"name"`
	Description       string   `yaml:"description"`
	ScriptPath        string   `yaml:"script_path"`
	Dependencies      []string `yaml:"dependencies"`
	Commands          []string `yaml:"commands"`
	Category          string   `yaml:"category"`
	VersionSupport    bool     `yaml:"version_support"`
	DefaultVersion    string   `yaml:"default_version"`
	SupportedVersions []string `yaml:"supported_versions"`
	AptPackageName    string   `yaml:"apt_package_name"`
}

// ConfigManager manages package configurations
type ConfigManager struct {
	configDir string
	packages  map[string]PackageConfig
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable is not set")
	}

	configDir := filepath.Join(homeDir, ".run", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	cm := &ConfigManager{
		configDir: configDir,
		packages:  make(map[string]PackageConfig),
	}

	// Load default configurations
	if err := cm.loadDefaultConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load default configurations: %w", err)
	}

	return cm, nil
}

// loadDefaultConfigs loads the default package configurations
func (cm *ConfigManager) loadDefaultConfigs() error {
	// Default package configurations
	defaultConfigs := map[string]PackageConfig{
		"node": {
			Name:              "node",
			Description:       "Node.js runtime with npm",
			ScriptPath:        "scripts/packages/node.sh",
			Dependencies:      []string{"curl", "build-essential"},
			Commands:          []string{"node", "npm"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "18",
			SupportedVersions: []string{"16", "18", "20", "21"},
			AptPackageName:    "nodejs",
		},
		"docker": {
			Name:              "docker",
			Description:       "Docker containerization platform",
			ScriptPath:        "scripts/packages/docker.sh",
			Dependencies:      []string{"ca-certificates", "curl", "gnupg"},
			Commands:          []string{"docker"},
			Category:          "devops",
			VersionSupport:    false,
			DefaultVersion:    "",
			SupportedVersions: []string{},
			AptPackageName:    "",
		},
		"nginx": {
			Name:              "nginx",
			Description:       "High-performance web server",
			ScriptPath:        "scripts/packages/nginx.sh",
			Dependencies:      []string{"curl", "gnupg"},
			Commands:          []string{"nginx"},
			Category:          "web",
			VersionSupport:    true,
			DefaultVersion:    "stable",
			SupportedVersions: []string{"stable", "mainline"},
			AptPackageName:    "nginx",
		},
		"postgres": {
			Name:              "postgres",
			Description:       "PostgreSQL 17 database server",
			ScriptPath:        "scripts/packages/postgres17.sh",
			Dependencies:      []string{"curl", "gnupg", "lsb-release"},
			Commands:          []string{"psql"},
			Category:          "database",
			VersionSupport:    true,
			DefaultVersion:    "17",
			SupportedVersions: []string{"15", "16", "17"},
			AptPackageName:    "postgresql",
		},
		"php": {
			Name:              "php",
			Description:       "PHP 8.3 programming language with FPM",
			ScriptPath:        "scripts/packages/php.sh",
			Dependencies:      []string{"software-properties-common"},
			Commands:          []string{"php"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "8.3",
			SupportedVersions: []string{"8.1", "8.2", "8.3"},
			AptPackageName:    "php",
		},
		"java": {
			Name:              "java",
			Description:       "OpenJDK Java Development Kit 17",
			ScriptPath:        "scripts/packages/java.sh",
			Dependencies:      []string{},
			Commands:          []string{"java", "javac"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "17",
			SupportedVersions: []string{"11", "17", "21"},
			AptPackageName:    "openjdk-17-jdk",
		},
		"pm2": {
			Name:              "pm2",
			Description:       "Process manager for Node.js applications",
			ScriptPath:        "scripts/packages/pm2.sh",
			Dependencies:      []string{"node"},
			Commands:          []string{"pm2"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "latest",
			SupportedVersions: []string{"latest", "5.3.0", "5.4.0", "5.5.0"},
			AptPackageName:    "pm2",
		},
		"essentials": {
			Name:              "essentials",
			Description:       "System essentials and development tools",
			ScriptPath:        "scripts/system/essentials.sh",
			Dependencies:      []string{"gcc", "make", "redis-server"},
			Commands:          []string{"gcc", "make", "redis-server"},
			Category:          "system",
			VersionSupport:    false,
			DefaultVersion:    "",
			SupportedVersions: []string{},
			AptPackageName:    "",
		},
	}

	// Load configurations
	for name, config := range defaultConfigs {
		cm.packages[name] = config
	}

	return nil
}

// GetPackageConfig returns a package configuration
func (cm *ConfigManager) GetPackageConfig(name string) (PackageConfig, bool) {
	config, exists := cm.packages[name]
	return config, exists
}

// ListPackageConfigs returns all package configurations
func (cm *ConfigManager) ListPackageConfigs() map[string]PackageConfig {
	configs := make(map[string]PackageConfig)
	for name, config := range cm.packages {
		configs[name] = config
	}
	return configs
}

// GetPackagesByCategory returns packages filtered by category
func (cm *ConfigManager) GetPackagesByCategory(category string) []PackageConfig {
	var configs []PackageConfig
	for _, config := range cm.packages {
		if config.Category == category {
			configs = append(configs, config)
		}
	}
	return configs
}

// BuildIntensivePackages returns the list of packages that benefit from build tools
func (cm *ConfigManager) BuildIntensivePackages() []string {
	return []string{"node", "php"}
}

// RelatedPackages returns related package suggestions
func (cm *ConfigManager) RelatedPackages() map[string][]string {
	return map[string][]string{
		"nginx":    {"php", "node"},
		"postgres": {"java"},
		"docker":   {"node"},
		"node":     {"pm2"},
	}
}

// ConvertToPackage converts PackageConfig to Package
func (pc PackageConfig) ConvertToPackage() Package {
	return Package{
		Name:              pc.Name,
		Description:       pc.Description,
		ScriptPath:        pc.ScriptPath,
		Dependencies:      pc.Dependencies,
		Commands:          pc.Commands,
		Category:          pc.Category,
		VersionSupport:    pc.VersionSupport,
		DefaultVersion:    pc.DefaultVersion,
		SupportedVersions: pc.SupportedVersions,
		AptPackageName:    pc.AptPackageName,
	}
}
