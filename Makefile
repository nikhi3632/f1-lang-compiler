.PHONY: build test cover clean run fmt lint lex parse check

# Build the compiler with race detection
build:
	@mkdir -p build
	go build -race -o build/f1c ./cmd/f1c

# Run all tests with race detection
test:
	go test -race ./...

# Run tests with verbose output
test-v:
	go test -race ./... -v

# Run tests with coverage report
cover:
	go test -race ./... -cover

# Run tests with HTML coverage report
cover-html:
	go test -race ./... -coverprofile=build/coverage.out
	go tool cover -html=build/coverage.out -o build/coverage.html
	@echo "Coverage report: build/coverage.html"

# Clean build artifacts
clean:
	rm -rf build/*

# Format all Go code
fmt:
	gofmt -w .

# Check formatting (CI-friendly)
fmt-check:
	@gofmt -l . | grep -q . && { echo "Files need formatting:"; gofmt -l .; exit 1; } || echo "All files formatted"

# Run the compiler (usage: make run F=examples/hello.f1)
run: build
	./build/f1c run $(F)

# Emit LLVM IR (usage: make emit F=examples/hello.f1)
emit: build
	./build/f1c emit $(F)

# Tokenize only (usage: make lex F=examples/test.f1)
lex: build
	./build/f1c lex $(F)

# Parse only (usage: make parse F=examples/test.f1)
parse: build
	./build/f1c parse $(F)

# Type check only (usage: make check F=examples/test.f1)
check: build
	./build/f1c check $(F)

# Default target
all: fmt test build

# Help
help:
	@echo "F1-Lang Compiler Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build compiler with race detection"
	@echo "  make test        Run all tests with race detection"
	@echo "  make test-v      Run tests with verbose output"
	@echo "  make cover       Run tests with coverage"
	@echo "  make cover-html  Generate HTML coverage report"
	@echo "  make clean       Remove build artifacts"
	@echo "  make fmt         Format all Go code"
	@echo "  make fmt-check   Check if code is formatted"
	@echo ""
	@echo "  make run F=file.f1    Compile and run a file"
	@echo "  make emit F=file.f1   Show LLVM IR"
	@echo "  make lex F=file.f1    Tokenize only"
	@echo "  make parse F=file.f1  Parse only (show AST)"
	@echo "  make check F=file.f1  Type check only"
