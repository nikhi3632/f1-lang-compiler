package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompile_SimpleProgram(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.f1")
	outFile := filepath.Join(tmpDir, "test")

	// Simple program that just assigns a variable
	err := os.WriteFile(srcFile, []byte("driver x = 42;"), 0644)
	if err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	c := New()
	err = c.Compile(srcFile, outFile)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	// Check that output file exists
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Errorf("output file should exist at %s", outFile)
	}
}

func TestCompile_WithFunction(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.f1")
	outFile := filepath.Join(tmpDir, "test")

	code := `
pitstop add(driver a, driver b) {
    finish a + b;
}
driver result = add(1, 2);
`
	err := os.WriteFile(srcFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	c := New()
	err = c.Compile(srcFile, outFile)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Errorf("output file should exist at %s", outFile)
	}
}

func TestCompile_ParseError(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.f1")
	outFile := filepath.Join(tmpDir, "test")

	// Invalid syntax
	err := os.WriteFile(srcFile, []byte("driver x ="), 0644)
	if err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	c := New()
	err = c.Compile(srcFile, outFile)
	if err == nil {
		t.Errorf("compile should fail with parse error")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention parse, got: %v", err)
	}
}

func TestCompile_TypeError(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.f1")
	outFile := filepath.Join(tmpDir, "test")

	// Type error: adding int and string
	err := os.WriteFile(srcFile, []byte(`driver x = 5 + "hello";`), 0644)
	if err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	c := New()
	err = c.Compile(srcFile, outFile)
	if err == nil {
		t.Errorf("compile should fail with type error")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("error should mention type, got: %v", err)
	}
}

func TestCompile_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test")

	c := New()
	err := c.Compile("/nonexistent/file.f1", outFile)
	if err == nil {
		t.Errorf("compile should fail for non-existent file")
	}
}

func TestGenerateLLVM_SimpleProgram(t *testing.T) {
	c := New()
	llvm, err := c.GenerateLLVM("driver x = 42;")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if !strings.Contains(llvm, "define") {
		t.Errorf("LLVM IR should contain 'define', got: %s", llvm)
	}
	if !strings.Contains(llvm, "@main") {
		t.Errorf("LLVM IR should contain '@main', got: %s", llvm)
	}
}

func TestRuntimePath(t *testing.T) {
	c := New()
	path := c.RuntimePath()

	// Should return a path ending in runtime/runtime.c
	if !strings.HasSuffix(path, "runtime/runtime.c") {
		t.Errorf("runtime path should end with runtime/runtime.c, got: %s", path)
	}
}
