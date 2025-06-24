package cmd

import (
	"fmt"
	"os"
)

func GetScriptPath(scriptName string) string {
	return fmt.Sprintf("%s/.gocli/scripts/%s", os.Getenv("HOME"), scriptName)
}
