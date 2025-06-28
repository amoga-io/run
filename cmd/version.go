package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// Version information - set during build time via ldflags
var (
	Version   = "dev"             // Default fallback version
	GitCommit = "unknown"         // Git commit hash
	BuildDate = "unknown"         // Build timestamp
	GoVersion = runtime.Version() // Go version used to build
)

// validateVersion ensures version is properly formatted
func validateVersion(version string) string {
	if version == "" || version == "dev" {
		return "dev"
	}

	// Ensure version starts with 'v' if it's a proper version
	if !strings.HasPrefix(version, "v") && !strings.Contains(version, "dev") {
		return "v" + version
	}

	return version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display version, build information, and system details for the run CLI",
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	validatedVersion := validateVersion(Version)

	fmt.Printf("run CLI %s\n", validatedVersion)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Built: %s\n", BuildDate)
	fmt.Printf("Go version: %s\n", GoVersion)
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Show additional info for dev builds
	if strings.Contains(validatedVersion, "dev") {
		fmt.Println("\n⚠️  This is a development build")
		fmt.Println("For production use, install from a tagged release")
	}
}
