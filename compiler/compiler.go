package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"f1c/codegen"
	"f1c/errors"
	"f1c/irgen"
	"f1c/lexer"
	"f1c/optimizer"
	"f1c/parser"
	"f1c/typechecker"
)

// Compiler orchestrates the full compilation pipeline.
type Compiler struct {
	// Verbose enables debug output
	Verbose bool
	// Optimize enables optimization passes
	Optimize bool
}

// New creates a new compiler instance with default options.
func New() *Compiler {
	return &Compiler{
		Optimize: true, // Default: optimization enabled
	}
}

// NewWithOptions creates a compiler with specific options.
func NewWithOptions(optimize, verbose bool) *Compiler {
	return &Compiler{
		Optimize: optimize,
		Verbose:  verbose,
	}
}

// Compile compiles a source file to an executable.
func (c *Compiler) Compile(srcPath, outPath string) error {
	// Read source file
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading source: %w", err)
	}

	// Generate LLVM IR
	llvmIR, err := c.GenerateLLVMWithFilename(srcPath, string(content))
	if err != nil {
		return err
	}

	// Create temp directory for intermediate files
	tmpDir, err := os.MkdirTemp("", "f1c-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write LLVM IR to temp file
	llFile := filepath.Join(tmpDir, "program.ll")
	if err := os.WriteFile(llFile, []byte(llvmIR), 0644); err != nil {
		return fmt.Errorf("writing LLVM IR: %w", err)
	}

	// Find clang
	clang, err := c.findClang()
	if err != nil {
		return err
	}

	// Compile with clang
	// Link with runtime if it exists
	args := []string{"-o", outPath, llFile}

	// Add SDK path for macOS if needed
	if sdkPath := c.findMacOSSDK(); sdkPath != "" {
		args = append([]string{"-isysroot", sdkPath}, args...)
	}

	runtimePath := c.RuntimePath()
	if _, err := os.Stat(runtimePath); err == nil {
		args = append(args, runtimePath)
	}

	if c.Verbose {
		fmt.Printf("Running: %s %s\n", clang, strings.Join(args, " "))
	}

	cmd := exec.Command(clang, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clang failed: %v\n%s", err, string(output))
	}

	return nil
}

// Run compiles and executes a source file.
func (c *Compiler) Run(srcPath string) (string, error) {
	// Create temp file for executable
	tmpDir, err := os.MkdirTemp("", "f1c-run-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "program")

	// Compile
	if err := c.Compile(srcPath, outPath); err != nil {
		return "", err
	}

	// Execute
	cmd := exec.Command(outPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's just a non-zero exit code (which is fine for our programs)
		if _, ok := err.(*exec.ExitError); ok {
			return string(output), nil
		}
		return "", fmt.Errorf("execution failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}

// GenerateLLVM generates LLVM IR from source code.
func (c *Compiler) GenerateLLVM(source string) (string, error) {
	return c.GenerateLLVMWithFilename("<input>", source)
}

// GenerateLLVMWithFilename generates LLVM IR from source code with filename for error reporting.
func (c *Compiler) GenerateLLVMWithFilename(filename, source string) (string, error) {
	reporter := errors.NewReporter(filename, source)

	// Lex
	l := lexer.NewWithReporter(source, reporter)

	// Parse
	p := parser.NewWithReporter(l, reporter)
	program := p.ParseProgram()

	if reporter.HasErrors() {
		return "", fmt.Errorf("%s", reporter.Format())
	}

	// Type check
	tc := typechecker.NewWithReporter(reporter)
	tc.Check(program)

	if reporter.HasErrors() {
		return "", fmt.Errorf("%s", reporter.Format())
	}

	// Generate F1-IR
	gen := irgen.New()
	mod := gen.Generate(program)

	// Optimize (if enabled)
	if c.Optimize {
		if c.Verbose {
			fmt.Println("[verbose] Running optimizations")
		}
		opt := optimizer.New()
		opt.Run(mod)
	} else if c.Verbose {
		fmt.Println("[verbose] Optimizations disabled")
	}

	// Generate LLVM IR
	llvmGen := codegen.New()
	llvmIR := llvmGen.Generate(mod)

	return llvmIR, nil
}

// RuntimePath returns the path to the runtime.c file.
func (c *Compiler) RuntimePath() string {
	// Get the directory of this source file
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "runtime/runtime.c"
	}

	// Go up one level from compiler/ to project root
	projectRoot := filepath.Dir(filepath.Dir(thisFile))
	return filepath.Join(projectRoot, "runtime", "runtime.c")
}

// findClang finds the clang compiler.
func (c *Compiler) findClang() (string, error) {
	// Prefer system clang (from Xcode) on macOS for better compatibility
	candidates := []string{
		"/usr/bin/clang", // Xcode/CommandLineTools clang - best for macOS
		"clang",          // System PATH
	}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("clang not found; please install Xcode Command Line Tools")
}

// findMacOSSDK finds the macOS SDK path for compilation.
func (c *Compiler) findMacOSSDK() string {
	// Only needed on macOS
	if goos := os.Getenv("GOOS"); goos != "" && goos != "darwin" {
		return ""
	}

	// Try xcrun to find SDK
	cmd := exec.Command("xcrun", "--show-sdk-path")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
