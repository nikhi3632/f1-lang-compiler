package main

import (
	"fmt"
	"io"
	"os"

	"f1c/lexer"
	"f1c/parser"
	"f1c/typechecker"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

// run executes the CLI and returns an exit code.
// This is extracted from main() to make it testable.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)
		return 0
	}

	cmd := args[1]

	switch cmd {
	case "lex":
		if len(args) < 3 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return lexFile(args[2], stdout, stderr)
	case "parse":
		if len(args) < 3 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return parseFile(args[2], stdout, stderr)
	case "check":
		if len(args) < 3 {
			fmt.Fprintln(stderr, "error: missing file argument")
			return 1
		}
		return checkFile(args[2], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "error: unknown command %q (not yet implemented)\n", cmd)
		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "F1-Lang Compiler")
	fmt.Fprintln(w, "Usage: f1c <command> [file]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  run <file>    Compile and run")
	fmt.Fprintln(w, "  build <file>  Compile to binary")
	fmt.Fprintln(w, "  emit <file>   Show LLVM IR")
	fmt.Fprintln(w, "  lex <file>    Tokenize only")
	fmt.Fprintln(w, "  parse <file>  Parse only (show AST)")
	fmt.Fprintln(w, "  check <file>  Type check only")
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
