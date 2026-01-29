package irgen

import (
	"strings"
	"testing"

	"f1c/lexer"
	"f1c/parser"
	"f1c/typechecker"
)

func generateIR(t *testing.T, input string) string {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := typechecker.New()
	c.Check(program)

	if len(c.Errors()) > 0 {
		t.Fatalf("type errors: %v", c.Errors())
	}

	gen := New()
	mod := gen.Generate(program)
	return mod.String()
}

func TestIRGen_IntegerLiteral(t *testing.T) {
	input := `driver x = 42;`
	ir := generateIR(t, input)

	// Should have alloca and store
	if !strings.Contains(ir, "alloca int") {
		t.Errorf("expected alloca int, got:\n%s", ir)
	}
	if !strings.Contains(ir, "store int 42") {
		t.Errorf("expected store int 42, got:\n%s", ir)
	}
}

func TestIRGen_BooleanLiteral(t *testing.T) {
	input := `driver b = greenlight;`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "alloca bool") {
		t.Errorf("expected alloca bool, got:\n%s", ir)
	}
	if !strings.Contains(ir, "store bool true") {
		t.Errorf("expected store bool true, got:\n%s", ir)
	}
}

func TestIRGen_BinaryExpression(t *testing.T) {
	input := `driver x = 1 + 2;`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "add int") {
		t.Errorf("expected add int, got:\n%s", ir)
	}
}

func TestIRGen_MultipleOperations(t *testing.T) {
	input := `
driver a = 10;
driver b = 20;
driver c = a + b;
`
	ir := generateIR(t, input)

	// Should have loads for a and b
	if !strings.Contains(ir, "load int") {
		t.Errorf("expected load int, got:\n%s", ir)
	}
	if !strings.Contains(ir, "add int") {
		t.Errorf("expected add int, got:\n%s", ir)
	}
}

func TestIRGen_ComparisonExpression(t *testing.T) {
	input := `driver x = 5 < 10;`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "lt int") {
		t.Errorf("expected lt int, got:\n%s", ir)
	}
	if !strings.Contains(ir, "alloca bool") {
		t.Errorf("expected alloca bool for result, got:\n%s", ir)
	}
}

func TestIRGen_FunctionDeclaration(t *testing.T) {
	input := `
pitstop add(driver a, driver b) {
    finish a + b;
}
`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "define int @add(int %a, int %b)") {
		t.Errorf("expected function definition, got:\n%s", ir)
	}
	if !strings.Contains(ir, "ret int") {
		t.Errorf("expected ret int, got:\n%s", ir)
	}
}

func TestIRGen_VoidFunction(t *testing.T) {
	input := `
pitstop noop() {
    driver x = 1;
}
`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "define void @noop()") {
		t.Errorf("expected void function, got:\n%s", ir)
	}
	if !strings.Contains(ir, "ret void") {
		t.Errorf("expected ret void, got:\n%s", ir)
	}
}

func TestIRGen_FunctionCall(t *testing.T) {
	input := `
pitstop double(driver x) {
    finish x + x;
}
driver result = double(21);
`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "call int @double") {
		t.Errorf("expected call to double, got:\n%s", ir)
	}
}

func TestIRGen_IfStatement(t *testing.T) {
	input := `
driver x = 5;
drs (x > 0) {
    driver y = 1;
}
`
	ir := generateIR(t, input)

	// Should have conditional branch
	if !strings.Contains(ir, "br %") {
		t.Errorf("expected conditional branch, got:\n%s", ir)
	}
	// Should have then and merge blocks
	if !strings.Contains(ir, "then") || !strings.Contains(ir, "merge") {
		t.Errorf("expected then and merge blocks, got:\n%s", ir)
	}
}

func TestIRGen_IfElseStatement(t *testing.T) {
	input := `
driver x = 5;
drs (x > 0) {
    driver y = 1;
} defend {
    driver y = 0;
}
`
	ir := generateIR(t, input)

	// Should have then, else, and merge blocks
	if !strings.Contains(ir, "then") {
		t.Errorf("expected then block, got:\n%s", ir)
	}
	if !strings.Contains(ir, "else") {
		t.Errorf("expected else block, got:\n%s", ir)
	}
}

func TestIRGen_ForLoop(t *testing.T) {
	input := `
lap (driver i = 0; i < 10; i = i + 1) {
    driver x = i;
}
`
	ir := generateIR(t, input)

	// Should have loop structure
	if !strings.Contains(ir, "loop") || !strings.Contains(ir, "cond") {
		t.Errorf("expected loop blocks, got:\n%s", ir)
	}
	// Should have conditional branch back
	if !strings.Contains(ir, "br %") {
		t.Errorf("expected conditional branch, got:\n%s", ir)
	}
}

func TestIRGen_UnaryNegation(t *testing.T) {
	input := `driver x = -5;`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "neg int") {
		t.Errorf("expected neg int, got:\n%s", ir)
	}
}

func TestIRGen_LogicalNot(t *testing.T) {
	input := `driver x = !greenlight;`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "not bool") {
		t.Errorf("expected not bool, got:\n%s", ir)
	}
}

func TestIRGen_LogicalAnd(t *testing.T) {
	input := `driver x = greenlight && redlight;`
	ir := generateIR(t, input)

	if !strings.Contains(ir, "and bool") {
		t.Errorf("expected and bool, got:\n%s", ir)
	}
}

func TestIRGen_MainFunction(t *testing.T) {
	input := `driver x = 42;`
	ir := generateIR(t, input)

	// Top-level code should be wrapped in @main
	if !strings.Contains(ir, "@main") {
		t.Errorf("expected @main function, got:\n%s", ir)
	}
}

func TestIRGen_Assignment(t *testing.T) {
	input := `
driver x = 5;
x = 10;
`
	ir := generateIR(t, input)

	// Should have two stores to x
	count := strings.Count(ir, "store int")
	if count < 2 {
		t.Errorf("expected at least 2 stores, got %d in:\n%s", count, ir)
	}
}
