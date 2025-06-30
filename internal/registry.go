package internal

var InstallPackageRegistry = map[string]string{
	"docker":   "docker.sh",
	"java":     "java.sh",
	"nginx":    "nginx.sh",
	"node":     "node.sh",
	"php":      "php.sh",
	"pm2":      "pm2.sh",
	"postgres": "postgres17.sh",
}

var RemovePackageRegistry = map[string]string{
	"nginx":    "remove-nginx.sh",
	"node":     "remove-node.sh",
	"postgres": "remove-postgres.sh",
}
