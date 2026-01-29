package parser

import (
	"testing"

	"f1c/ast"
	"f1c/lexer"
)

// =============================================================================
// Helper Functions
// =============================================================================

func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}

func testIntegerLiteral(t *testing.T, expr ast.Expression, expected int64) {
	t.Helper()
	intLit, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expr not *ast.IntegerLiteral. got=%T", expr)
	}
	if intLit.Value != expected {
		t.Errorf("intLit.Value = %d, want %d", intLit.Value, expected)
	}
}

func testIdentifier(t *testing.T, expr ast.Expression, expected string) {
	t.Helper()
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expr not *ast.Identifier. got=%T", expr)
	}
	if ident.Value != expected {
		t.Errorf("ident.Value = %q, want %q", ident.Value, expected)
	}
}

func testBooleanLiteral(t *testing.T, expr ast.Expression, expected bool) {
	t.Helper()
	boolLit, ok := expr.(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("expr not *ast.BooleanLiteral. got=%T", expr)
	}
	if boolLit.Value != expected {
		t.Errorf("boolLit.Value = %t, want %t", boolLit.Value, expected)
	}
}

func testLiteralExpression(t *testing.T, expr ast.Expression, expected interface{}) {
	t.Helper()
	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, expr, int64(v))
	case int64:
		testIntegerLiteral(t, expr, v)
	case string:
		testIdentifier(t, expr, v)
	case bool:
		testBooleanLiteral(t, expr, v)
	default:
		t.Fatalf("type of expected not handled. got=%T", expected)
	}
}

func testInfixExpression(t *testing.T, expr ast.Expression, left interface{}, operator string, right interface{}) {
	t.Helper()
	infixExpr, ok := expr.(*ast.InfixExpr)
	if !ok {
		t.Fatalf("expr not *ast.InfixExpr. got=%T", expr)
	}
	testLiteralExpression(t, infixExpr.Left, left)
	if infixExpr.Operator != operator {
		t.Errorf("infixExpr.Operator = %q, want %q", infixExpr.Operator, operator)
	}
	testLiteralExpression(t, infixExpr.Right, right)
}

// =============================================================================
// Literal Tests
// =============================================================================

func TestIntegerLiteralExpression(t *testing.T) {
	input := `5;`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
	}

	testIntegerLiteral(t, stmt.Expression, 5)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
	}

	strLit, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expr not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if strLit.Value != "hello world" {
		t.Errorf("strLit.Value = %q, want %q", strLit.Value, "hello world")
	}
}

func TestBooleanLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"greenlight;", true},
		{"redlight;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
			if !ok {
				t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
			}

			testBooleanLiteral(t, stmt.Expression, tt.expected)
		})
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `foobar;`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
	}

	testIdentifier(t, stmt.Expression, "foobar")
}

// =============================================================================
// Prefix Expression Tests
// =============================================================================

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"-5;", "-", 5},
		{"!greenlight;", "!", true},
		{"!redlight;", "!", false},
		{"-foobar;", "-", "foobar"},
		{"!foobar;", "!", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
			if !ok {
				t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
			}

			prefixExpr, ok := stmt.Expression.(*ast.PrefixExpr)
			if !ok {
				t.Fatalf("expr not *ast.PrefixExpr. got=%T", stmt.Expression)
			}

			if prefixExpr.Operator != tt.operator {
				t.Errorf("prefixExpr.Operator = %q, want %q", prefixExpr.Operator, tt.operator)
			}

			testLiteralExpression(t, prefixExpr.Right, tt.value)
		})
	}
}

// =============================================================================
// Infix Expression Tests
// =============================================================================

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 % 5;", 5, "%", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 <= 5;", 5, "<=", 5},
		{"greenlight == greenlight;", true, "==", true},
		{"greenlight != redlight;", true, "!=", false},
		{"greenlight && redlight;", true, "&&", false},
		{"greenlight || redlight;", true, "||", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
			if !ok {
				t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
			}

			testInfixExpression(t, stmt.Expression, tt.left, tt.operator, tt.right)
		})
	}
}

// =============================================================================
// Operator Precedence Tests
// =============================================================================

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b;", "((-a) * b)"},
		{"!-a;", "(!(-a))"},
		{"a + b + c;", "((a + b) + c)"},
		{"a + b - c;", "((a + b) - c)"},
		{"a * b * c;", "((a * b) * c)"},
		{"a * b / c;", "((a * b) / c)"},
		{"a + b / c;", "(a + (b / c))"},
		{"a + b * c + d / e - f;", "(((a + (b * c)) + (d / e)) - f)"},
		{"5 > 4 == 3 < 4;", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4;", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5;", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"greenlight;", "greenlight"},
		{"redlight;", "redlight"},
		{"3 > 5 == redlight;", "((3 > 5) == redlight)"},
		{"(5 + 5) * 2;", "((5 + 5) * 2)"},
		{"2 / (5 + 5);", "(2 / (5 + 5))"},
		{"-(5 + 5);", "(-(5 + 5))"},
		{"!(greenlight == greenlight);", "(!(greenlight == greenlight))"},
		{"a + add(b * c) + d;", "((a + add((b * c))) + d)"},
		{"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8));", "add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))"},
		{"a && b || c;", "((a && b) || c)"},
		{"a || b && c;", "(a || (b && c))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			actual := program.Statements[0].(*ast.ExpressionStmt).Expression.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

// =============================================================================
// Grouped Expression Tests
// =============================================================================

func TestGroupedExpression(t *testing.T) {
	input := `(5 + 5);`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
	}

	infixExpr, ok := stmt.Expression.(*ast.InfixExpr)
	if !ok {
		t.Fatalf("expr not *ast.InfixExpr. got=%T", stmt.Expression)
	}

	testInfixExpression(t, infixExpr, 5, "+", 5)
}

// =============================================================================
// Variable Declaration Tests
// =============================================================================

func TestVarDeclaration(t *testing.T) {
	tests := []struct {
		input         string
		expectedName  string
		expectedValue interface{}
	}{
		{"driver x = 5;", "x", 5},
		{"driver y = greenlight;", "y", true},
		{"driver foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.VarDecl)
			if !ok {
				t.Fatalf("stmt not *ast.VarDecl. got=%T", program.Statements[0])
			}

			if stmt.Name.Value != tt.expectedName {
				t.Errorf("stmt.Name.Value = %q, want %q", stmt.Name.Value, tt.expectedName)
			}

			testLiteralExpression(t, stmt.Value, tt.expectedValue)
		})
	}
}

// =============================================================================
// Assignment Tests
// =============================================================================

func TestAssignment(t *testing.T) {
	input := `x = 10;`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("stmt not *ast.Assignment. got=%T", program.Statements[0])
	}

	if stmt.Name.Value != "x" {
		t.Errorf("stmt.Name.Value = %q, want %q", stmt.Name.Value, "x")
	}

	testIntegerLiteral(t, stmt.Value, 10)
}

// =============================================================================
// Return Statement Tests
// =============================================================================

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"finish 5;", 5},
		{"finish greenlight;", true},
		{"finish foobar;", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ReturnStmt)
			if !ok {
				t.Fatalf("stmt not *ast.ReturnStmt. got=%T", program.Statements[0])
			}

			testLiteralExpression(t, stmt.Value, tt.expectedValue)
		})
	}
}

func TestReturnStatementVoid(t *testing.T) {
	input := `finish;`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ReturnStmt. got=%T", program.Statements[0])
	}

	if stmt.Value != nil {
		t.Errorf("stmt.Value should be nil for void return, got=%T", stmt.Value)
	}
}

// =============================================================================
// If Statement Tests
// =============================================================================

func TestIfStatement(t *testing.T) {
	input := `drs (x < y) { x; }`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("stmt not *ast.IfStmt. got=%T", program.Statements[0])
	}

	testInfixExpression(t, stmt.Condition, "x", "<", "y")

	if len(stmt.Consequence.Statements) != 1 {
		t.Fatalf("consequence has %d statements, want 1", len(stmt.Consequence.Statements))
	}

	consequence, ok := stmt.Consequence.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("consequence stmt not *ast.ExpressionStmt. got=%T", stmt.Consequence.Statements[0])
	}

	testIdentifier(t, consequence.Expression, "x")

	if stmt.Alternative != nil {
		t.Errorf("stmt.Alternative should be nil, got=%T", stmt.Alternative)
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `drs (x < y) { x; } defend { y; }`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("stmt not *ast.IfStmt. got=%T", program.Statements[0])
	}

	testInfixExpression(t, stmt.Condition, "x", "<", "y")

	if len(stmt.Consequence.Statements) != 1 {
		t.Fatalf("consequence has %d statements, want 1", len(stmt.Consequence.Statements))
	}

	alternative, ok := stmt.Alternative.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("alternative not *ast.BlockStmt. got=%T", stmt.Alternative)
	}

	if len(alternative.Statements) != 1 {
		t.Fatalf("alternative has %d statements, want 1", len(alternative.Statements))
	}

	altStmt, ok := alternative.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("alternative stmt not *ast.ExpressionStmt. got=%T", alternative.Statements[0])
	}

	testIdentifier(t, altStmt.Expression, "y")
}

func TestIfElseIfStatement(t *testing.T) {
	input := `drs (x < 0) { x; } defend drs (x == 0) { y; } defend { z; }`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("stmt not *ast.IfStmt. got=%T", program.Statements[0])
	}

	// Check the else-if chain
	elseIf, ok := stmt.Alternative.(*ast.IfStmt)
	if !ok {
		t.Fatalf("alternative not *ast.IfStmt. got=%T", stmt.Alternative)
	}

	testInfixExpression(t, elseIf.Condition, "x", "==", 0)

	// Check the final else
	finalElse, ok := elseIf.Alternative.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("final alternative not *ast.BlockStmt. got=%T", elseIf.Alternative)
	}

	if len(finalElse.Statements) != 1 {
		t.Fatalf("final else has %d statements, want 1", len(finalElse.Statements))
	}
}

// =============================================================================
// For Loop Tests
// =============================================================================

func TestForLoop(t *testing.T) {
	input := `lap (driver i = 0; i < 10; i = i + 1) { radio(i); }`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ForStmt. got=%T", program.Statements[0])
	}

	// Check init
	if stmt.Init.Name.Value != "i" {
		t.Errorf("init name = %q, want %q", stmt.Init.Name.Value, "i")
	}
	testIntegerLiteral(t, stmt.Init.Value, 0)

	// Check condition
	testInfixExpression(t, stmt.Condition, "i", "<", 10)

	// Check update
	if stmt.Update.Name.Value != "i" {
		t.Errorf("update name = %q, want %q", stmt.Update.Name.Value, "i")
	}

	// Check body
	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(stmt.Body.Statements))
	}
}

// =============================================================================
// Function Declaration Tests
// =============================================================================

func TestFunctionDeclaration(t *testing.T) {
	input := `pitstop add(driver a, driver b) { finish a + b; }`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("stmt not *ast.FunctionDecl. got=%T", program.Statements[0])
	}

	if stmt.Name.Value != "add" {
		t.Errorf("function name = %q, want %q", stmt.Name.Value, "add")
	}

	if len(stmt.Params) != 2 {
		t.Fatalf("function has %d params, want 2", len(stmt.Params))
	}

	if stmt.Params[0].Name.Value != "a" {
		t.Errorf("param[0] = %q, want %q", stmt.Params[0].Name.Value, "a")
	}

	if stmt.Params[1].Name.Value != "b" {
		t.Errorf("param[1] = %q, want %q", stmt.Params[1].Name.Value, "b")
	}

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(stmt.Body.Statements))
	}
}

func TestFunctionDeclarationNoParams(t *testing.T) {
	input := `pitstop noop() { }`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("stmt not *ast.FunctionDecl. got=%T", program.Statements[0])
	}

	if stmt.Name.Value != "noop" {
		t.Errorf("function name = %q, want %q", stmt.Name.Value, "noop")
	}

	if len(stmt.Params) != 0 {
		t.Fatalf("function has %d params, want 0", len(stmt.Params))
	}
}

// =============================================================================
// Function Call Tests
// =============================================================================

func TestCallExpression(t *testing.T) {
	input := `add(1, 2 * 3, 4 + 5);`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
	}

	callExpr, ok := stmt.Expression.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expr not *ast.CallExpr. got=%T", stmt.Expression)
	}

	testIdentifier(t, callExpr.Function, "add")

	if len(callExpr.Arguments) != 3 {
		t.Fatalf("wrong number of arguments. got=%d, want=3", len(callExpr.Arguments))
	}

	testIntegerLiteral(t, callExpr.Arguments[0], 1)
	testInfixExpression(t, callExpr.Arguments[1], 2, "*", 3)
	testInfixExpression(t, callExpr.Arguments[2], 4, "+", 5)
}

func TestCallExpressionNoArgs(t *testing.T) {
	input := `noop();`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt not *ast.ExpressionStmt. got=%T", program.Statements[0])
	}

	callExpr, ok := stmt.Expression.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expr not *ast.CallExpr. got=%T", stmt.Expression)
	}

	testIdentifier(t, callExpr.Function, "noop")

	if len(callExpr.Arguments) != 0 {
		t.Fatalf("wrong number of arguments. got=%d, want=0", len(callExpr.Arguments))
	}
}

// =============================================================================
// Lambda Tests
// =============================================================================

func TestLambdaExpression(t *testing.T) {
	input := `driver addOne = pitstop(driver x) { finish x + 1; };`

	program := parseProgram(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("stmt not *ast.VarDecl. got=%T", program.Statements[0])
	}

	lambda, ok := stmt.Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("value not *ast.LambdaExpr. got=%T", stmt.Value)
	}

	if len(lambda.Params) != 1 {
		t.Fatalf("lambda has %d params, want 1", len(lambda.Params))
	}

	if lambda.Params[0].Name.Value != "x" {
		t.Errorf("param[0] = %q, want %q", lambda.Params[0].Name.Value, "x")
	}

	if len(lambda.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(lambda.Body.Statements))
	}
}

// =============================================================================
// Complete Program Tests
// =============================================================================

func TestCompleteProgram(t *testing.T) {
	input := `
pitstop fib(driver n) {
	drs (n < 2) {
		finish n;
	}
	finish fib(n - 1) + fib(n - 2);
}

driver result = fib(10);
radio(result);
`

	program := parseProgram(t, input)

	if len(program.Statements) != 3 {
		t.Fatalf("program has %d statements, want 3", len(program.Statements))
	}

	// First statement: function declaration
	_, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("stmt[0] not *ast.FunctionDecl. got=%T", program.Statements[0])
	}

	// Second statement: variable declaration
	_, ok = program.Statements[1].(*ast.VarDecl)
	if !ok {
		t.Fatalf("stmt[1] not *ast.VarDecl. got=%T", program.Statements[1])
	}

	// Third statement: expression statement (function call)
	_, ok = program.Statements[2].(*ast.ExpressionStmt)
	if !ok {
		t.Fatalf("stmt[2] not *ast.ExpressionStmt. got=%T", program.Statements[2])
	}
}

// =============================================================================
// Error Tests
// =============================================================================

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing semicolon", "driver x = 5"},
		{"missing closing paren", "add(1, 2"},
		{"missing closing brace", "drs (x) {"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Errorf("expected parser errors, got none")
			}
		})
	}
}
