package run

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestHelloCmd_Success(t *testing.T) {
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	scriptPath := filepath.Join(scriptDir, "hello.sh")
	os.MkdirAll(scriptDir, 0755)
	scriptContent := "#!/bin/bash\necho 'Hello, Test!'\n"
	os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	defer os.RemoveAll(filepath.Join(home, ".devkit"))

	buf := new(bytes.Buffer)
	HelloCmd.SetOut(buf)
	HelloCmd.SetErr(buf)
	arg = ""
	HelloCmd.Run(HelloCmd, []string{})

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Hello, Test!")) {
		t.Errorf("expected output to contain 'Hello, Test!', got: %s", output)
	}
}

func TestHelloCmd_ScriptNotFound(t *testing.T) {
	home, _ := os.UserHomeDir()
	scriptDir := filepath.Join(home, ".devkit", "scripts")
	os.RemoveAll(scriptDir)

	buf := new(bytes.Buffer)
	HelloCmd.SetOut(buf)
	HelloCmd.SetErr(buf)
	arg = ""
	HelloCmd.Run(HelloCmd, []string{})

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Script not found")) {
		t.Errorf("expected error about script not found, got: %s", output)
	}
}
