package typechecker

import (
	"strings"
	"testing"

	"f1c/lexer"
	"f1c/parser"
	"f1c/types"
)

func parseProgram(t *testing.T, input string) *parser.Parser {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return p
}

func checkProgram(t *testing.T, input string) (*Checker, []string) {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	c := New()
	c.Check(program)
	return c, c.Errors()
}

// --- Type Inference Tests ---

func TestChecker_IntegerLiteral(t *testing.T) {
	input := `driver x = 42;`
	c, errors := checkProgram(t, input)

	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	typ := c.TypeOf("x")
	if !types.Equal(typ, types.Int) {
		t.Errorf("type of x = %v, want int", typ)
	}
}

func TestChecker_StringLiteral(t *testing.T) {
	input := `driver s = "hello";`
	c, errors := checkProgram(t, input)

	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	typ := c.TypeOf("s")
	if !types.Equal(typ, types.String) {
		t.Errorf("type of s = %v, want string", typ)
	}
}

func TestChecker_BooleanLiteral(t *testing.T) {
	tests := []struct {
		input string
		name  string
	}{
		{`driver b = greenlight;`, "b"},
		{`driver f = redlight;`, "f"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, errors := checkProgram(t, tt.input)
			if len(errors) > 0 {
				t.Fatalf("unexpected errors: %v", errors)
			}
			typ := c.TypeOf(tt.name)
			if !types.Equal(typ, types.Bool) {
				t.Errorf("type of %s = %v, want bool", tt.name, typ)
			}
		})
	}
}

// --- Binary Operations ---

func TestChecker_ArithmeticOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Type
	}{
		{`driver x = 1 + 2;`, types.Int},
		{`driver x = 5 - 3;`, types.Int},
		{`driver x = 2 * 3;`, types.Int},
		{`driver x = 10 / 2;`, types.Int},
		{`driver x = 10 % 3;`, types.Int},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, errors := checkProgram(t, tt.input)
			if len(errors) > 0 {
				t.Fatalf("unexpected errors: %v", errors)
			}
			typ := c.TypeOf("x")
			if !types.Equal(typ, tt.expected) {
				t.Errorf("type = %v, want %v", typ, tt.expected)
			}
		})
	}
}

func TestChecker_ComparisonOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Type
	}{
		{`driver x = 1 < 2;`, types.Bool},
		{`driver x = 1 > 2;`, types.Bool},
		{`driver x = 1 <= 2;`, types.Bool},
		{`driver x = 1 >= 2;`, types.Bool},
		{`driver x = 1 == 2;`, types.Bool},
		{`driver x = 1 != 2;`, types.Bool},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, errors := checkProgram(t, tt.input)
			if len(errors) > 0 {
				t.Fatalf("unexpected errors: %v", errors)
			}
			typ := c.TypeOf("x")
			if !types.Equal(typ, tt.expected) {
				t.Errorf("type = %v, want %v", typ, tt.expected)
			}
		})
	}
}

func TestChecker_LogicalOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Type
	}{
		{`driver x = greenlight && redlight;`, types.Bool},
		{`driver x = greenlight || redlight;`, types.Bool},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, errors := checkProgram(t, tt.input)
			if len(errors) > 0 {
				t.Fatalf("unexpected errors: %v", errors)
			}
			typ := c.TypeOf("x")
			if !types.Equal(typ, tt.expected) {
				t.Errorf("type = %v, want %v", typ, tt.expected)
			}
		})
	}
}

func TestChecker_PrefixOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Type
	}{
		{`driver x = -5;`, types.Int},
		{`driver x = !greenlight;`, types.Bool},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, errors := checkProgram(t, tt.input)
			if len(errors) > 0 {
				t.Fatalf("unexpected errors: %v", errors)
			}
			typ := c.TypeOf("x")
			if !types.Equal(typ, tt.expected) {
				t.Errorf("type = %v, want %v", typ, tt.expected)
			}
		})
	}
}

// --- Function Type Checking ---

func TestChecker_FunctionDeclaration(t *testing.T) {
	input := `
pitstop add(driver a, driver b) {
    finish a + b;
}
`
	c, errors := checkProgram(t, input)
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	typ := c.TypeOf("add")
	fn, ok := typ.(*types.FunctionType)
	if !ok {
		t.Fatalf("add is not a function type, got %T", typ)
	}
	if len(fn.Params) != 2 {
		t.Errorf("param count = %d, want 2", len(fn.Params))
	}
	if !types.Equal(fn.Return, types.Int) {
		t.Errorf("return type = %v, want int", fn.Return)
	}
}

func TestChecker_VoidFunction(t *testing.T) {
	input := `
pitstop greet() {
    driver x = 1;
}
`
	c, errors := checkProgram(t, input)
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	typ := c.TypeOf("greet")
	fn, ok := typ.(*types.FunctionType)
	if !ok {
		t.Fatalf("greet is not a function type, got %T", typ)
	}
	if !types.Equal(fn.Return, types.Void) {
		t.Errorf("return type = %v, want void", fn.Return)
	}
}

func TestChecker_FunctionCall(t *testing.T) {
	input := `
pitstop add(driver a, driver b) {
    finish a + b;
}
driver result = add(1, 2);
`
	c, errors := checkProgram(t, input)
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	typ := c.TypeOf("result")
	if !types.Equal(typ, types.Int) {
		t.Errorf("type of result = %v, want int", typ)
	}
}

// --- Error Cases ---

func TestChecker_UndefinedVariable(t *testing.T) {
	input := `driver x = undefined_var;`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for undefined variable")
	}
	if !strings.Contains(errors[0], "undefined") {
		t.Errorf("error = %q, should mention 'undefined'", errors[0])
	}
}

func TestChecker_UndefinedFunction(t *testing.T) {
	input := `driver x = unknown_func();`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for undefined function")
	}
	if !strings.Contains(errors[0], "undefined") {
		t.Errorf("error = %q, should mention 'undefined'", errors[0])
	}
}

func TestChecker_TypeMismatch_BinaryOp(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"int + string", `driver x = 5 + "hello";`},
		{"string - int", `driver x = "hello" - 5;`},
		{"bool * int", `driver x = greenlight * 5;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := checkProgram(t, tt.input)
			if len(errors) == 0 {
				t.Fatal("expected type mismatch error")
			}
		})
	}
}

func TestChecker_WrongArgumentCount(t *testing.T) {
	input := `
pitstop add(driver a, driver b) {
    finish a + b;
}
driver x = add(1);
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for wrong argument count")
	}
	if !strings.Contains(errors[0], "argument") {
		t.Errorf("error = %q, should mention 'argument'", errors[0])
	}
}

func TestChecker_WrongArgumentType(t *testing.T) {
	input := `
pitstop add(driver a, driver b) {
    finish a + b;
}
driver x = add("hello", "world");
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for wrong argument type")
	}
}

func TestChecker_NonBooleanCondition(t *testing.T) {
	input := `
drs (42) {
    driver x = 1;
}
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for non-boolean condition")
	}
	if !strings.Contains(errors[0], "bool") {
		t.Errorf("error = %q, should mention 'bool'", errors[0])
	}
}

func TestChecker_RedeclarationInSameScope(t *testing.T) {
	input := `
driver x = 1;
driver x = 2;
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for redeclaration")
	}
	if !strings.Contains(errors[0], "already declared") {
		t.Errorf("error = %q, should mention 'already declared'", errors[0])
	}
}

func TestChecker_AssignmentTypeMismatch(t *testing.T) {
	input := `
driver x = 5;
x = "hello";
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for assignment type mismatch")
	}
}

func TestChecker_NonCallableType(t *testing.T) {
	input := `
driver x = 5;
driver y = x();
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for calling non-function")
	}
	if !strings.Contains(errors[0], "call") || !strings.Contains(errors[0], "function") {
		t.Errorf("error = %q, should mention cannot call non-function", errors[0])
	}
}

func TestChecker_InconsistentReturnTypes(t *testing.T) {
	input := `
pitstop bad(driver x) {
    drs (x > 0) {
        finish x;
    }
    finish "error";
}
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for inconsistent return types")
	}
}

// --- Scoping Tests ---

func TestChecker_VariableShadowing(t *testing.T) {
	input := `
driver x = 5;
drs (greenlight) {
    driver x = 10;
}
`
	_, errors := checkProgram(t, input)

	// Shadowing in different scope is allowed
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
}

func TestChecker_FunctionScope(t *testing.T) {
	input := `
pitstop test(driver x) {
    driver y = x + 1;
    finish y;
}
`
	_, errors := checkProgram(t, input)

	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
}

// --- Lambda Tests ---

func TestChecker_LambdaExpression(t *testing.T) {
	input := `
driver addOne = pitstop(driver x) { finish x + 1; };
`
	c, errors := checkProgram(t, input)

	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	typ := c.TypeOf("addOne")
	fn, ok := typ.(*types.FunctionType)
	if !ok {
		t.Fatalf("addOne is not a function type, got %T", typ)
	}
	if len(fn.Params) != 1 {
		t.Errorf("param count = %d, want 1", len(fn.Params))
	}
	if !types.Equal(fn.Return, types.Int) {
		t.Errorf("return type = %v, want int", fn.Return)
	}
}

func TestChecker_LambdaCapture(t *testing.T) {
	input := `
driver multiplier = 10;
driver scale = pitstop(driver x) { finish x * multiplier; };
`
	_, errors := checkProgram(t, input)

	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
}

// --- For Loop Tests ---

func TestChecker_ForLoop(t *testing.T) {
	input := `
lap (driver i = 0; i < 10; i = i + 1) {
    driver x = i;
}
`
	_, errors := checkProgram(t, input)

	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
}

func TestChecker_ForLoopConditionMustBeBool(t *testing.T) {
	input := `
lap (driver i = 0; i; i = i + 1) {
    driver x = i;
}
`
	_, errors := checkProgram(t, input)

	if len(errors) == 0 {
		t.Fatal("expected error for non-boolean loop condition")
	}
}

// --- Reserved Identifier Tests ---

func TestChecker_ReservedIdentifier_Variable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		reserved string
	}{
		{"radio", "driver radio = 1;", "radio"},
		{"bono", "driver bono = 1;", "bono"},
		{"canvas", "driver canvas = 1;", "canvas"},
		{"pixel", "driver pixel = 1;", "pixel"},
		{"render", "driver render = 1;", "render"},
		{"snapshot", "driver snapshot = 1;", "snapshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := checkProgram(t, tt.input)

			if len(errors) == 0 {
				t.Fatalf("expected error for reserved identifier '%s'", tt.reserved)
			}
			if !strings.Contains(errors[0], "reserved identifier") {
				t.Errorf("error should mention reserved identifier, got: %s", errors[0])
			}
		})
	}
}

func TestChecker_ReservedIdentifier_Function(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		reserved string
	}{
		{"radio", "pitstop radio() { finish 0; }", "radio"},
		{"bono", "pitstop bono() { finish 0; }", "bono"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := checkProgram(t, tt.input)

			if len(errors) == 0 {
				t.Fatalf("expected error for reserved identifier '%s'", tt.reserved)
			}
			if !strings.Contains(errors[0], "reserved identifier") {
				t.Errorf("error should mention reserved identifier, got: %s", errors[0])
			}
		})
	}
}
