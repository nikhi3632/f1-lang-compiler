package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("F1-Lang Compiler")
		fmt.Println("Usage: f1c <command> [file]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  run <file>    Compile and run")
		fmt.Println("  build <file>  Compile to binary")
		fmt.Println("  emit <file>   Show LLVM IR")
		fmt.Println("  lex <file>    Tokenize only")
		fmt.Println("  parse <file>  Parse only (show AST)")
		fmt.Println("  check <file>  Type check only")
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "lex":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "error: missing file argument")
			os.Exit(1)
		}
		lexFile(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q (not yet implemented)\n", cmd)
		os.Exit(1)
	}
}

func lexFile(path string) {
	// TODO: implement in Phase 8
	fmt.Printf("lex: %s (not yet implemented)\n", path)
}
