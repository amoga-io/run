package pkg

type Package struct {
	Name         string
	Description  string
	ScriptPath   string
	Dependencies []string // Required packages/commands before installation
	Commands     []string // Commands to check if package is installed
	Category     string
}

var PackageRegistry = map[string]Package{
	"python": {
		Name:         "python",
		Description:  "Python programming language with pip and venv",
		ScriptPath:   "scripts/packages/python.sh",
		Dependencies: []string{"build-essential", "curl"},
		Commands:     []string{"python3", "pip3"},
		Category:     "development",
	},
	"node": {
		Name:         "node",
		Description:  "Node.js runtime with npm",
		ScriptPath:   "scripts/packages/node.sh",
		Dependencies: []string{"curl", "build-essential"},
		Commands:     []string{"node", "npm"},
		Category:     "development",
	},
	"docker": {
		Name:         "docker",
		Description:  "Docker containerization platform",
		ScriptPath:   "scripts/packages/docker.sh",
		Dependencies: []string{"ca-certificates", "curl", "gnupg"},
		Commands:     []string{"docker"},
		Category:     "devops",
	},
	"nginx": {
		Name:         "nginx",
		Description:  "High-performance web server",
		ScriptPath:   "scripts/packages/nginx.sh",
		Dependencies: []string{"curl", "gnupg"},
		Commands:     []string{"nginx"},
		Category:     "web",
	},
	"postgres": {
		Name:         "postgres",
		Description:  "PostgreSQL 17 database server",
		ScriptPath:   "scripts/packages/postgres17.sh",
		Dependencies: []string{"curl", "gnupg", "lsb-release"},
		Commands:     []string{"psql"},
		Category:     "database",
	},
	"php": {
		Name:         "php",
		Description:  "PHP 8.3 programming language with FPM",
		ScriptPath:   "scripts/packages/php.sh",
		Dependencies: []string{"software-properties-common"},
		Commands:     []string{"php"},
		Category:     "development",
	},
	"java": {
		Name:         "java",
		Description:  "OpenJDK Java Development Kit 17",
		ScriptPath:   "scripts/packages/java.sh",
		Dependencies: []string{},
		Commands:     []string{"java", "javac"},
		Category:     "development",
	},
	"pm2": {
		Name:         "pm2",
		Description:  "Process manager for Node.js applications",
		ScriptPath:   "scripts/packages/pm2.sh",
		Dependencies: []string{"node"}, // Requires Node.js to be installed first
		Commands:     []string{"pm2"},
		Category:     "development",
	},
	"essentials": {
		Name:         "essentials",
		Description:  "System essentials and development tools",
		ScriptPath:   "scripts/system/essentials.sh",
		Dependencies: []string{},
		Commands:     []string{"gcc", "make", "redis-server"},
		Category:     "system",
	},
}

// GetPackage returns a package by name
func GetPackage(name string) (Package, bool) {
	pkg, exists := PackageRegistry[name]
	return pkg, exists
}

// ListPackages returns all available packages
func ListPackages() map[string]Package {
	return PackageRegistry
}

// GetPackagesByCategory returns packages filtered by category
func GetPackagesByCategory(category string) []Package {
	var packages []Package
	for _, pkg := range PackageRegistry {
		if pkg.Category == category {
			packages = append(packages, pkg)
		}
	}
	return packages
}
