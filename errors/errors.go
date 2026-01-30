package errors

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrorKind represents the phase where an error occurred.
type ErrorKind string

const (
	LexerError ErrorKind = "lexer error"
	ParseError ErrorKind = "parse error"
	TypeError  ErrorKind = "type error"
)

// Error represents a compiler error with source location.
type Error struct {
	Kind    ErrorKind
	Message string
	Line    int
	Column  int
	Length  int // length of the offending token for underlining
}

// Reporter collects and formats compiler errors.
type Reporter struct {
	filename string
	source   string
	lines    []string
	errors   []Error
}

// NewReporter creates a new error reporter for the given source file.
func NewReporter(filename, source string) *Reporter {
	return &Reporter{
		filename: filename,
		source:   source,
		lines:    strings.Split(source, "\n"),
		errors:   []Error{},
	}
}

// Add records a new error with the given location and message.
func (r *Reporter) Add(kind ErrorKind, line, col, length int, format string, args ...any) {
	r.errors = append(r.errors, Error{
		Kind:    kind,
		Message: fmt.Sprintf(format, args...),
		Line:    line,
		Column:  col,
		Length:  length,
	})
}

// HasErrors returns true if any errors have been recorded.
func (r *Reporter) HasErrors() bool {
	return len(r.errors) > 0
}

// Errors returns the list of recorded errors.
func (r *Reporter) Errors() []Error {
	return r.errors
}

// Format returns a formatted string of all errors with source context.
func (r *Reporter) Format() string {
	var sb strings.Builder
	r.formatTo(&sb)
	return sb.String()
}

// Print writes formatted errors to stderr.
func (r *Reporter) Print() {
	r.formatTo(os.Stderr)
}

// formatTo writes formatted errors to the given writer.
func (r *Reporter) formatTo(w io.Writer) {
	// Calculate max line number width for padding
	maxLine := 0
	for _, err := range r.errors {
		if err.Line > maxLine {
			maxLine = err.Line
		}
	}
	lineWidth := len(fmt.Sprintf("%d", maxLine))
	if lineWidth < 1 {
		lineWidth = 1
	}

	for i, err := range r.errors {
		if i > 0 {
			fmt.Fprintln(w)
		}
		r.formatError(w, err, lineWidth)
	}
}

// formatError formats a single error with source context.
func (r *Reporter) formatError(w io.Writer, err Error, lineWidth int) {
	// Line 1: error kind and message
	fmt.Fprintf(w, "%s: %s\n", err.Kind, err.Message)

	// Line 2: file location
	fmt.Fprintf(w, "  --> %s:%d:%d\n", r.filename, err.Line, err.Column)

	// Line 3: separator
	fmt.Fprintf(w, "%s |\n", strings.Repeat(" ", lineWidth+1))

	// Line 4: source line (if available)
	if err.Line >= 1 && err.Line <= len(r.lines) {
		sourceLine := r.lines[err.Line-1]
		// Convert tabs to spaces for consistent alignment
		sourceLine = strings.ReplaceAll(sourceLine, "\t", "    ")
		fmt.Fprintf(w, "%*d | %s\n", lineWidth, err.Line, sourceLine)

		// Line 5: carets
		caretCount := err.Length
		if caretCount < 1 {
			caretCount = 1
		}

		// Calculate caret position (accounting for tab expansion)
		caretPos := err.Column - 1
		if caretPos < 0 {
			caretPos = 0
		}

		// Account for tabs before the error position in the original line
		originalLine := r.lines[err.Line-1]
		tabsBefore := 0
		for i := 0; i < err.Column-1 && i < len(originalLine); i++ {
			if originalLine[i] == '\t' {
				tabsBefore++
			}
		}
		// Each tab becomes 4 spaces, so add 3 extra spaces per tab
		caretPos += tabsBefore * 3

		carets := strings.Repeat("^", caretCount)
		fmt.Fprintf(w, "%s | %s%s\n", strings.Repeat(" ", lineWidth), strings.Repeat(" ", caretPos), carets)
	} else {
		// Out of bounds - just show empty context
		fmt.Fprintf(w, "%s | <source not available>\n", strings.Repeat(" ", lineWidth))
		fmt.Fprintf(w, "%s |\n", strings.Repeat(" ", lineWidth))
	}
}
