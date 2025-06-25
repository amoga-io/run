package utils

import (
	"os"
	"path/filepath"
)

func GetScriptPath(scriptName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".devkit", "scripts", scriptName)
}
