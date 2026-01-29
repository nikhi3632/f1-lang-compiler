package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}

	if !strings.Contains(stdout.String(), "F1-Lang Compiler") {
		t.Errorf("stdout should contain 'F1-Lang Compiler', got %q", stdout.String())
	}

	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("stdout should contain 'Usage:', got %q", stdout.String())
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "unknown"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "unknown command") {
		t.Errorf("stderr should contain 'unknown command', got %q", stderr.String())
	}
}

func TestRun_LexMissingFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "lex"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "missing file argument") {
		t.Errorf("stderr should contain 'missing file argument', got %q", stderr.String())
	}
}

func TestRun_LexNonExistentFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "lex", "/nonexistent/file.f1"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("stderr should contain 'error:', got %q", stderr.String())
	}
}

func TestRun_LexValidFile(t *testing.T) {
	// Create a temporary file with F1 code
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.f1")
	err := os.WriteFile(tmpFile, []byte("driver x = 5;"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "lex", tmpFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0, stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "DRIVER") {
		t.Errorf("output should contain 'DRIVER', got %q", output)
	}
	if !strings.Contains(output, "IDENT") {
		t.Errorf("output should contain 'IDENT', got %q", output)
	}
	if !strings.Contains(output, "INT") {
		t.Errorf("output should contain 'INT', got %q", output)
	}
	if !strings.Contains(output, "EOF") {
		t.Errorf("output should contain 'EOF', got %q", output)
	}
}

func TestRun_ParseMissingFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "parse"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "missing file argument") {
		t.Errorf("stderr should contain 'missing file argument', got %q", stderr.String())
	}
}

func TestRun_ParseValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.f1")
	err := os.WriteFile(tmpFile, []byte("driver x = 5;"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "parse", tmpFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0, stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "driver") {
		t.Errorf("output should contain 'driver', got %q", output)
	}
	if !strings.Contains(output, "x") {
		t.Errorf("output should contain 'x', got %q", output)
	}
}

func TestRun_ParseNonExistentFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "parse", "/nonexistent/file.f1"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("stderr should contain 'error:', got %q", stderr.String())
	}
}

func TestRun_ParseInvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.f1")
	err := os.WriteFile(tmpFile, []byte("driver x ="), 0644) // missing value
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"f1c", "parse", tmpFile}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}

	if !strings.Contains(stderr.String(), "parse error") {
		t.Errorf("stderr should contain 'parse error', got %q", stderr.String())
	}
}

func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)

	output := buf.String()
	expectedContents := []string{
		"F1-Lang Compiler",
		"Usage: f1c <command> [file]",
		"run <file>",
		"build <file>",
		"emit <file>",
		"lex <file>",
		"parse <file>",
		"check <file>",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("output should contain %q, got %q", expected, output)
		}
	}
}

func TestLexFile_TokenOutput(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.f1")
	err := os.WriteFile(tmpFile, []byte("5 + 10"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := lexFile(tmpFile, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}

	output := stdout.String()
	// Check that tokens are properly formatted
	if !strings.Contains(output, "INT") {
		t.Errorf("output should contain INT token")
	}
	if !strings.Contains(output, "PLUS") {
		t.Errorf("output should contain PLUS token")
	}
	if !strings.Contains(output, "line") && !strings.Contains(output, "col") {
		t.Errorf("output should contain line and column info")
	}
}
