package errors

import (
	"strings"
	"testing"
)

func TestNewReporter(t *testing.T) {
	source := "driver x = 5;\nradio(x);"
	r := NewReporter("test.f1", source)

	if r == nil {
		t.Fatal("NewReporter returned nil")
	}
	if r.HasErrors() {
		t.Error("new reporter should have no errors")
	}
	if len(r.Errors()) != 0 {
		t.Errorf("expected 0 errors, got %d", len(r.Errors()))
	}
}

func TestReporter_AddError(t *testing.T) {
	source := "driver x = 5;"
	r := NewReporter("test.f1", source)

	r.Add(TypeError, 1, 8, 1, "undefined variable '%s'", "x")

	if !r.HasErrors() {
		t.Error("reporter should have errors after Add")
	}
	if len(r.Errors()) != 1 {
		t.Errorf("expected 1 error, got %d", len(r.Errors()))
	}

	err := r.Errors()[0]
	if err.Kind != TypeError {
		t.Errorf("expected TypeError, got %v", err.Kind)
	}
	if err.Line != 1 {
		t.Errorf("expected line 1, got %d", err.Line)
	}
	if err.Column != 8 {
		t.Errorf("expected column 8, got %d", err.Column)
	}
	if err.Length != 1 {
		t.Errorf("expected length 1, got %d", err.Length)
	}
	if err.Message != "undefined variable 'x'" {
		t.Errorf("unexpected message: %s", err.Message)
	}
}

func TestReporter_MultipleErrors(t *testing.T) {
	source := "driver x = 5;\nradio(y);\nradio(z);"
	r := NewReporter("test.f1", source)

	r.Add(TypeError, 2, 7, 1, "undefined variable 'y'")
	r.Add(TypeError, 3, 7, 1, "undefined variable 'z'")

	if len(r.Errors()) != 2 {
		t.Errorf("expected 2 errors, got %d", len(r.Errors()))
	}
}

func TestReporter_Format_SingleError(t *testing.T) {
	source := "radio(foo);"
	r := NewReporter("test.f1", source)
	r.Add(TypeError, 1, 7, 3, "undefined variable 'foo'")

	output := r.Format()

	// Check key components are present
	if !strings.Contains(output, "type error:") {
		t.Error("output should contain 'type error:'")
	}
	if !strings.Contains(output, "undefined variable 'foo'") {
		t.Error("output should contain error message")
	}
	if !strings.Contains(output, "--> test.f1:1:7") {
		t.Error("output should contain file location")
	}
	if !strings.Contains(output, "radio(foo);") {
		t.Error("output should contain source line")
	}
	if !strings.Contains(output, "^^^") {
		t.Error("output should contain carets for length 3")
	}
}

func TestReporter_Format_MultipleErrors(t *testing.T) {
	source := "driver x = 5\nradio(y);"
	r := NewReporter("test.f1", source)
	r.Add(ParseError, 1, 13, 1, "expected ';' after expression")
	r.Add(TypeError, 2, 7, 1, "undefined variable 'y'")

	output := r.Format()

	if !strings.Contains(output, "parse error:") {
		t.Error("output should contain 'parse error:'")
	}
	if !strings.Contains(output, "type error:") {
		t.Error("output should contain 'type error:'")
	}
}

func TestReporter_Format_AllErrorKinds(t *testing.T) {
	tests := []struct {
		kind     ErrorKind
		expected string
	}{
		{LexerError, "lexer error:"},
		{ParseError, "parse error:"},
		{TypeError, "type error:"},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			r := NewReporter("test.f1", "test")
			r.Add(tt.kind, 1, 1, 1, "test message")

			output := r.Format()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got: %s", tt.expected, output)
			}
		})
	}
}

func TestReporter_Format_LineNumberPadding(t *testing.T) {
	// Create source with 10+ lines
	lines := make([]string, 12)
	for i := range lines {
		lines[i] = "driver x = 1;"
	}
	source := strings.Join(lines, "\n")

	r := NewReporter("test.f1", source)
	r.Add(TypeError, 1, 1, 6, "error on line 1")
	r.Add(TypeError, 10, 1, 6, "error on line 10")

	output := r.Format()

	// Line 1 should be padded to match line 10's width
	// Looking for consistent alignment
	if !strings.Contains(output, " 1 |") || !strings.Contains(output, "10 |") {
		t.Errorf("line numbers should be padded consistently, got:\n%s", output)
	}
}

func TestReporter_Format_TabHandling(t *testing.T) {
	source := "\tradio(foo);"
	r := NewReporter("test.f1", source)
	r.Add(TypeError, 1, 8, 3, "undefined variable 'foo'")

	output := r.Format()

	// Tabs should be converted to spaces in output
	// The caret position should still align correctly
	if strings.Contains(output, "\t") {
		t.Error("tabs should be converted to spaces in formatted output")
	}
}

func TestReporter_Format_EmptyLength(t *testing.T) {
	source := "radio(x);"
	r := NewReporter("test.f1", source)
	r.Add(TypeError, 1, 7, 0, "some error")

	output := r.Format()

	// Length 0 should show at least one caret
	if !strings.Contains(output, "^") {
		t.Error("output should contain at least one caret even with length 0")
	}
}

func TestReporter_Format_OutOfBoundsLine(t *testing.T) {
	source := "driver x = 5;"
	r := NewReporter("test.f1", source)
	r.Add(TypeError, 99, 1, 1, "error on non-existent line")

	// Should not panic
	output := r.Format()

	// Should still show the error message
	if !strings.Contains(output, "error on non-existent line") {
		t.Error("should still show error message for out-of-bounds line")
	}
}

func TestErrorKind_String(t *testing.T) {
	tests := []struct {
		kind     ErrorKind
		expected string
	}{
		{LexerError, "lexer error"},
		{ParseError, "parse error"},
		{TypeError, "type error"},
	}

	for _, tt := range tests {
		if string(tt.kind) != tt.expected {
			t.Errorf("ErrorKind %v should be %q", tt.kind, tt.expected)
		}
	}
}
