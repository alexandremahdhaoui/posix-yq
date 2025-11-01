package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ShellFunctionTester helps execute and test individual shell functions
type ShellFunctionTester struct {
	t         *testing.T
	shellCode string
	tmpDir    string
}

// NewShellFunctionTester creates a new tester with the given shell code
func NewShellFunctionTester(t *testing.T, shellCode string) *ShellFunctionTester {
	tmpDir, err := os.MkdirTemp("", "posix-yq-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	return &ShellFunctionTester{
		t:         t,
		shellCode: shellCode,
		tmpDir:    tmpDir,
	}
}

// NewShellFunctionTesterWithDeps creates a tester with multiple shell code sections
// This is useful when functions depend on other functions
func NewShellFunctionTesterWithDeps(t *testing.T, codeSections ...string) *ShellFunctionTester {
	tmpDir, err := os.MkdirTemp("", "posix-yq-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Concatenate all code sections
	shellCode := ""
	for _, section := range codeSections {
		shellCode += section + "\n"
	}

	return &ShellFunctionTester{
		t:         t,
		shellCode: shellCode,
		tmpDir:    tmpDir,
	}
}

// Cleanup removes the temporary directory
func (sft *ShellFunctionTester) Cleanup() {
	os.RemoveAll(sft.tmpDir)
}

// WriteFile writes content to a temporary file and returns its path
func (sft *ShellFunctionTester) WriteFile(name string, content string) string {
	path := filepath.Join(sft.tmpDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		sft.t.Fatalf("Failed to write test file: %v", err)
	}
	return path
}

// ExecuteFunction executes a shell function with the given arguments and returns output
func (sft *ShellFunctionTester) ExecuteFunction(funcName string, args ...string) (string, error) {
	// Build a shell script that sources the function definition and calls it
	shellScript := sft.shellCode + "\n" + funcName
	for _, arg := range args {
		// Quote arguments properly for shell
		shellScript += fmt.Sprintf(" %q", arg)
	}
	shellScript += "\n"

	// Execute the shell script
	cmd := exec.Command("sh", "-c", shellScript)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExecuteFunctionExpect executes a function and expects specific output
func (sft *ShellFunctionTester) ExecuteFunctionExpect(expected string, funcName string, args ...string) {
	output, err := sft.ExecuteFunction(funcName, args...)

	// Remove trailing newline for comparison if present
	actualOutput := output
	if len(actualOutput) > 0 && actualOutput[len(actualOutput)-1] == '\n' {
		actualOutput = actualOutput[:len(actualOutput)-1]
	}

	if actualOutput != expected {
		sft.t.Errorf("Function %s with args %v: expected %q, got %q (error: %v)",
			funcName, args, expected, actualOutput, err)
	}
}

// ExecuteFunctionExpectError executes a function and expects it to fail
func (sft *ShellFunctionTester) ExecuteFunctionExpectError(funcName string, args ...string) {
	_, err := sft.ExecuteFunction(funcName, args...)
	if err == nil {
		sft.t.Errorf("Function %s with args %v: expected error but succeeded", funcName, args)
	}
}
