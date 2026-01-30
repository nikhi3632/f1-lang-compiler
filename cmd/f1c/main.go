package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"f1c/codegen"
	"f1c/compiler"
	"f1c/irgen"
	"f1c/lexer"
	"f1c/optimizer"
	"f1c/parser"
	"f1c/typechecker"
)

// Options holds CLI flags
type Options struct {
	Optimize  bool   // -O1 (default: true)
	Verbose   bool   // -v, --verbose
	EmitLLVM  bool   // --emit-llvm
	OutputFile string // -o
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

// parseFlags extracts options from args, returns remaining args and options
func parseFlags(args []string) ([]string, Options) {
	opts := Options{Optimize: true} // Default: optimization enabled
	remaining := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-O0":
			opts.Optimize = false
		case "-O1":
			opts.Optimize = true
		case "-v", "--verbose":
			opts.Verbose = true
		case "--emit-llvm":
			opts.EmitLLVM = true
		case "-o":
			if i+1 < len(args) {
				opts.OutputFile = args[i+1]
				i++ // Skip the output file argument
			}
		default:
			remaining = append(remaining, arg)
		}
	}

	return remaining, opts
}

// run executes the CLI and returns an exit code.
// This is extracted from main() to make it testable.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)
		return 0
	}

	cmd := args[1]

	// Handle --help and -h flags
	if cmd == "--help" || cmd == "-h" || cmd == "help" {
		printUsage(stdout)
		return 0
	}

	// Parse flags from remaining args
	remaining, opts := parseFlags(args[2:])

	switch cmd {
	case "run":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return runFileWithOpts(remaining[0], opts, stdout, stderr)
	case "build":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		// Default output: same name as input but without .f1 extension
		outPath := opts.OutputFile
		if outPath == "" {
			outPath = strings.TrimSuffix(remaining[0], ".f1")
			if outPath == remaining[0] {
				outPath = remaining[0] + ".out"
			}
		}
		return buildFileWithOpts(remaining[0], outPath, opts, stdout, stderr)
	case "lex":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return lexFile(remaining[0], stdout, stderr)
	case "parse":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return parseFile(remaining[0], stdout, stderr)
	case "check":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return checkFile(remaining[0], stdout, stderr)
	case "emit-ir":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return emitIR(remaining[0], false, stdout, stderr)
	case "opt-ir":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return emitIR(remaining[0], true, stdout, stderr)
	case "emit":
		if len(remaining) < 1 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return emitLLVMWithOpts(remaining[0], opts, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "error: unknown command %q (not yet implemented)\n", cmd)
		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "F1-Lang Compiler")
	fmt.Fprintln(w, "Usage: f1c <command> [options] <file>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  run <file>      Compile and run")
	fmt.Fprintln(w, "  build <file>    Compile to binary")
	fmt.Fprintln(w, "  emit <file>     Show LLVM IR")
	fmt.Fprintln(w, "  emit-ir <file>  Show F1-IR")
	fmt.Fprintln(w, "  opt-ir <file>   Show optimized F1-IR")
	fmt.Fprintln(w, "  lex <file>      Tokenize only")
	fmt.Fprintln(w, "  parse <file>    Parse only (show AST)")
	fmt.Fprintln(w, "  check <file>    Type check only")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  -O0             Disable optimizations")
	fmt.Fprintln(w, "  -O1             Enable optimizations (default)")
	fmt.Fprintln(w, "  -v, --verbose   Verbose output")
	fmt.Fprintln(w, "  --emit-llvm     Also emit LLVM IR file (build command)")
	fmt.Fprintln(w, "  -o <file>       Output file path")
}

func lexFile(path string, stdout, stderr io.Writer) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	l := lexer.New(string(content))
	for {
		tok := l.NextToken()
		fmt.Fprintf(stdout, "%s\t%q\t(line %d, col %d)\n",
			tok.Type, tok.Literal, tok.Line, tok.Column)
		if tok.Type.String() == "EOF" {
			break
		}
	}
	return 0
}

func parseFile(path string, stdout, stderr io.Writer) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if errors := p.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "parse error: %s\n", e)
		}
		return 1
	}

	fmt.Fprintln(stdout, program.String())
	return 0
}

func checkFile(path string, stdout, stderr io.Writer) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if errors := p.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "parse error: %s\n", e)
		}
		return 1
	}

	c := typechecker.New()
	c.Check(program)

	if errors := c.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "type error: %s\n", e)
		}
		return 1
	}

	fmt.Fprintln(stdout, "type check passed")
	return 0
}

func emitIR(path string, optimize bool, stdout, stderr io.Writer) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if errors := p.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "parse error: %s\n", e)
		}
		return 1
	}

	c := typechecker.New()
	c.Check(program)

	if errors := c.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "type error: %s\n", e)
		}
		return 1
	}

	gen := irgen.New()
	mod := gen.Generate(program)

	if optimize {
		opt := optimizer.New()
		opt.Run(mod)
	}

	fmt.Fprint(stdout, mod.String())
	return 0
}

func emitLLVMWithOpts(path string, opts Options, stdout, stderr io.Writer) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if opts.Verbose {
		fmt.Fprintf(stderr, "[verbose] Reading file: %s\n", path)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if errors := p.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "parse error: %s\n", e)
		}
		return 1
	}

	if opts.Verbose {
		fmt.Fprintf(stderr, "[verbose] Parsing complete\n")
	}

	c := typechecker.New()
	c.Check(program)

	if errors := c.Errors(); len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "type error: %s\n", e)
		}
		return 1
	}

	if opts.Verbose {
		fmt.Fprintf(stderr, "[verbose] Type checking complete\n")
	}

	gen := irgen.New()
	mod := gen.Generate(program)

	// Optimize if enabled
	if opts.Optimize {
		if opts.Verbose {
			fmt.Fprintf(stderr, "[verbose] Running optimizations (-O1)\n")
		}
		opt := optimizer.New()
		opt.Run(mod)
	} else if opts.Verbose {
		fmt.Fprintf(stderr, "[verbose] Optimizations disabled (-O0)\n")
	}

	// Generate LLVM IR
	llvmGen := codegen.New()
	llvmIR := llvmGen.Generate(mod)

	fmt.Fprint(stdout, llvmIR)
	return 0
}

func runFileWithOpts(path string, opts Options, stdout, stderr io.Writer) int {
	if opts.Verbose {
		fmt.Fprintf(stderr, "[verbose] Compiling and running: %s\n", path)
	}

	c := compiler.NewWithOptions(opts.Optimize, opts.Verbose)
	output, err := c.Run(path)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprint(stdout, output)
	return 0
}

func buildFileWithOpts(srcPath, outPath string, opts Options, stdout, stderr io.Writer) int {
	// Ensure output directory exists
	outDir := filepath.Dir(outPath)
	if outDir != "" && outDir != "." {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Fprintf(stderr, "error: creating output directory: %v\n", err)
			return 1
		}
	}

	if opts.Verbose {
		fmt.Fprintf(stderr, "[verbose] Compiling %s -> %s\n", srcPath, outPath)
	}

	// Handle --emit-llvm flag
	if opts.EmitLLVM {
		llvmPath := outPath + ".ll"
		if opts.Verbose {
			fmt.Fprintf(stderr, "[verbose] Emitting LLVM IR to: %s\n", llvmPath)
		}
		// Generate LLVM IR and write to file
		content, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if errors := p.Errors(); len(errors) > 0 {
			for _, e := range errors {
				fmt.Fprintf(stderr, "parse error: %s\n", e)
			}
			return 1
		}

		tc := typechecker.New()
		tc.Check(program)

		if errors := tc.Errors(); len(errors) > 0 {
			for _, e := range errors {
				fmt.Fprintf(stderr, "type error: %s\n", e)
			}
			return 1
		}

		gen := irgen.New()
		mod := gen.Generate(program)

		if opts.Optimize {
			opt := optimizer.New()
			opt.Run(mod)
		}

		llvmGen := codegen.New()
		llvmIR := llvmGen.Generate(mod)

		if err := os.WriteFile(llvmPath, []byte(llvmIR), 0644); err != nil {
			fmt.Fprintf(stderr, "error: writing LLVM IR: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "emitted: %s\n", llvmPath)
	}

	c := compiler.NewWithOptions(opts.Optimize, opts.Verbose)
	if err := c.Compile(srcPath, outPath); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "compiled: %s\n", outPath)
	return 0
}
