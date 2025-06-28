/*
Copyright © 2025 Amoga
*/
package main

import (
	"os"

	"github.com/amoga-io/run/cmd"
	"github.com/amoga-io/run/internal/logger"
)

func main() {
	// Initialize logger
	if err := logger.InitLogger(logger.INFO); err != nil {
		// If logging fails, just continue without it
		os.Stderr.WriteString("Warning: Failed to initialize logging\n")
	}
	defer logger.GetLogger().Close()

	// Log application start
	logger.Info("run CLI starting")

	cmd.Execute()
}
