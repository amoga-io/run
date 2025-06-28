package output

import (
	"fmt"
	"strings"
)

// Output interface for standardized user output
type Output interface {
	Info(message string, args ...interface{})
	Success(message string, args ...interface{})
	Error(message string, args ...interface{})
	Warning(message string, args ...interface{})
	Progress(message string, args ...interface{})
	Section(title string)
	Summary(title string, successful []string, failed []string, total int, retryCommand string)
}

// ConsoleOutput implements Output for console display
type ConsoleOutput struct{}

// NewConsoleOutput creates a new console output
func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{}
}

// Info displays an informational message
func (co *ConsoleOutput) Info(message string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf(message+"\n", args...)
	} else {
		fmt.Println(message)
	}
}

// Success displays a success message
func (co *ConsoleOutput) Success(message string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("✓ "+message+"\n", args...)
	} else {
		fmt.Println("✓ " + message)
	}
}

// Error displays an error message
func (co *ConsoleOutput) Error(message string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("✗ "+message+"\n", args...)
	} else {
		fmt.Println("✗ " + message)
	}
}

// Warning displays a warning message
func (co *ConsoleOutput) Warning(message string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("⚠ "+message+"\n", args...)
	} else {
		fmt.Println("⚠ " + message)
	}
}

// Progress displays a progress message
func (co *ConsoleOutput) Progress(message string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("→ "+message+"\n", args...)
	} else {
		fmt.Println("→ " + message)
	}
}

// Section displays a section header
func (co *ConsoleOutput) Section(title string) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println(strings.ToUpper(title))
	fmt.Println(strings.Repeat("=", 50))
}

// Summary displays an operation summary
func (co *ConsoleOutput) Summary(title string, successful []string, failed []string, total int, retryCommand string) {
	co.Section(title + " SUMMARY")

	if len(successful) > 0 {
		co.Success("Successfully %s (%d): %s", title, len(successful), strings.Join(successful, ", "))
	}

	if len(failed) > 0 {
		co.Error("Failed to %s (%d): %s", title, len(failed), strings.Join(failed, ", "))
		fmt.Println("\nFailed packages details:")
		for _, pkg := range failed {
			fmt.Printf("  • %s\n", pkg)
		}
	}

	if total > 0 {
		fmt.Printf("\nTotal: %d packages processed\n", total)
		successRate := float64(len(successful)) / float64(total) * 100
		fmt.Printf("Success rate: %.1f%% (%d/%d)\n", successRate, len(successful), total)

		if len(failed) > 0 && retryCommand != "" {
			fmt.Printf("\nTo retry failed packages: %s %s\n", retryCommand, strings.Join(failed, " "))
		}
	}
}

// Global output instance
var globalOutput Output = NewConsoleOutput()

// SetOutput sets the global output instance
func SetOutput(output Output) {
	globalOutput = output
}

// GetOutput returns the global output instance
func GetOutput() Output {
	return globalOutput
}

// Convenience functions for global output
func Info(message string, args ...interface{}) {
	globalOutput.Info(message, args...)
}

func Success(message string, args ...interface{}) {
	globalOutput.Success(message, args...)
}

func Error(message string, args ...interface{}) {
	globalOutput.Error(message, args...)
}

func Warning(message string, args ...interface{}) {
	globalOutput.Warning(message, args...)
}

func Progress(message string, args ...interface{}) {
	globalOutput.Progress(message, args...)
}

func Section(title string) {
	globalOutput.Section(title)
}

func Summary(title string, successful []string, failed []string, total int, retryCommand string) {
	globalOutput.Summary(title, successful, failed, total, retryCommand)
}
