package ast

import (
	"testing"

	"f1c/token"
)

func TestProgram_String(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarDecl{
				Token: token.Token{Type: token.DRIVER, Literal: "driver"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
				Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
			},
		},
	}

	expected := "driver x = 5;"
	if program.String() != expected {
		t.Errorf("program.String() = %q, want %q", program.String(), expected)
	}
}

func TestProgram_TokenLiteral(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarDecl{
				Token: token.Token{Type: token.DRIVER, Literal: "driver"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
				Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
			},
		},
	}

	if program.TokenLiteral() != "driver" {
		t.Errorf("program.TokenLiteral() = %q, want %q", program.TokenLiteral(), "driver")
	}

	// Empty program
	emptyProgram := &Program{Statements: []Statement{}}
	if emptyProgram.TokenLiteral() != "" {
		t.Errorf("emptyProgram.TokenLiteral() = %q, want %q", emptyProgram.TokenLiteral(), "")
	}
}

func TestIdentifier_String(t *testing.T) {
	ident := &Identifier{
		Token: token.Token{Type: token.IDENT, Literal: "foobar"},
		Value: "foobar",
	}

	if ident.String() != "foobar" {
		t.Errorf("ident.String() = %q, want %q", ident.String(), "foobar")
	}
}

func TestIntegerLiteral_String(t *testing.T) {
	intLit := &IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: "42"},
		Value: 42,
	}

	if intLit.String() != "42" {
		t.Errorf("intLit.String() = %q, want %q", intLit.String(), "42")
	}
}

func TestStringLiteral_String(t *testing.T) {
	strLit := &StringLiteral{
		Token: token.Token{Type: token.STRING, Literal: "hello"},
		Value: "hello",
	}

	if strLit.String() != `"hello"` {
		t.Errorf("strLit.String() = %q, want %q", strLit.String(), `"hello"`)
	}
}

func TestBooleanLiteral_String(t *testing.T) {
	tests := []struct {
		value    bool
		literal  string
		expected string
	}{
		{true, "greenlight", "greenlight"},
		{false, "redlight", "redlight"},
	}

	for _, tt := range tests {
		t.Run(tt.literal, func(t *testing.T) {
			boolLit := &BooleanLiteral{
				Token: token.Token{Literal: tt.literal},
				Value: tt.value,
			}
			if boolLit.String() != tt.expected {
				t.Errorf("boolLit.String() = %q, want %q", boolLit.String(), tt.expected)
			}
		})
	}
}

func TestPrefixExpr_String(t *testing.T) {
	expr := &PrefixExpr{
		Token:    token.Token{Type: token.MINUS, Literal: "-"},
		Operator: "-",
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
	}

	expected := "(-5)"
	if expr.String() != expected {
		t.Errorf("expr.String() = %q, want %q", expr.String(), expected)
	}
}

func TestInfixExpr_String(t *testing.T) {
	expr := &InfixExpr{
		Token:    token.Token{Type: token.PLUS, Literal: "+"},
		Left:     &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
		Operator: "+",
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 10},
	}

	expected := "(5 + 10)"
	if expr.String() != expected {
		t.Errorf("expr.String() = %q, want %q", expr.String(), expected)
	}
}

func TestCallExpr_String(t *testing.T) {
	expr := &CallExpr{
		Token:    token.Token{Type: token.LPAREN, Literal: "("},
		Function: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "add"}, Value: "add"},
		Arguments: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
		},
	}

	expected := "add(1, 2)"
	if expr.String() != expected {
		t.Errorf("expr.String() = %q, want %q", expr.String(), expected)
	}
}

func TestCallExpr_String_NoArgs(t *testing.T) {
	expr := &CallExpr{
		Token:     token.Token{Type: token.LPAREN, Literal: "("},
		Function:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "noop"}, Value: "noop"},
		Arguments: []Expression{},
	}

	expected := "noop()"
	if expr.String() != expected {
		t.Errorf("expr.String() = %q, want %q", expr.String(), expected)
	}
}

func TestVarDecl_String(t *testing.T) {
	stmt := &VarDecl{
		Token: token.Token{Type: token.DRIVER, Literal: "driver"},
		Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
	}

	expected := "driver x = 5;"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestAssignment_String(t *testing.T) {
	stmt := &Assignment{
		Token: token.Token{Type: token.IDENT, Literal: "x"},
		Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 10},
	}

	expected := "x = 10;"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestReturnStmt_String(t *testing.T) {
	stmt := &ReturnStmt{
		Token: token.Token{Type: token.FINISH, Literal: "finish"},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "42"}, Value: 42},
	}

	expected := "finish 42;"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestReturnStmt_String_Void(t *testing.T) {
	stmt := &ReturnStmt{
		Token: token.Token{Type: token.FINISH, Literal: "finish"},
		Value: nil,
	}

	expected := "finish;"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestExpressionStmt_String(t *testing.T) {
	stmt := &ExpressionStmt{
		Token:      token.Token{Type: token.INT, Literal: "5"},
		Expression: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
	}

	expected := "5"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestExpressionStmt_String_Nil(t *testing.T) {
	stmt := &ExpressionStmt{
		Token:      token.Token{},
		Expression: nil,
	}

	expected := ""
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestBlockStmt_String(t *testing.T) {
	block := &BlockStmt{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&ExpressionStmt{
				Token:      token.Token{Type: token.IDENT, Literal: "x"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			},
		},
	}

	expected := "{ x }"
	if block.String() != expected {
		t.Errorf("block.String() = %q, want %q", block.String(), expected)
	}
}

func TestIfStmt_String(t *testing.T) {
	stmt := &IfStmt{
		Token: token.Token{Type: token.DRS, Literal: "drs"},
		Condition: &InfixExpr{
			Token:    token.Token{Type: token.LT, Literal: "<"},
			Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			Operator: "<",
			Right:    &Identifier{Token: token.Token{Type: token.IDENT, Literal: "y"}, Value: "y"},
		},
		Consequence: &BlockStmt{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStmt{
					Token:      token.Token{Type: token.IDENT, Literal: "x"},
					Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
				},
			},
		},
		Alternative: nil,
	}

	expected := "drs ((x < y)) { x }"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestIfStmt_String_WithElse(t *testing.T) {
	stmt := &IfStmt{
		Token: token.Token{Type: token.DRS, Literal: "drs"},
		Condition: &BooleanLiteral{
			Token: token.Token{Type: token.GREENLIGHT, Literal: "greenlight"},
			Value: true,
		},
		Consequence: &BlockStmt{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
		Alternative: &BlockStmt{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}

	expected := "drs (greenlight) {  } defend {  }"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestFunctionDecl_String(t *testing.T) {
	stmt := &FunctionDecl{
		Token: token.Token{Type: token.PITSTOP, Literal: "pitstop"},
		Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "add"}, Value: "add"},
		Params: []*Parameter{
			{
				Token: token.Token{Type: token.DRIVER, Literal: "driver"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			},
			{
				Token: token.Token{Type: token.DRIVER, Literal: "driver"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
			},
		},
		Body: &BlockStmt{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}

	expected := "pitstop add(driver a, driver b) {  }"
	if stmt.String() != expected {
		t.Errorf("stmt.String() = %q, want %q", stmt.String(), expected)
	}
}

func TestLambdaExpr_String(t *testing.T) {
	expr := &LambdaExpr{
		Token: token.Token{Type: token.PITSTOP, Literal: "pitstop"},
		Params: []*Parameter{
			{
				Token: token.Token{Type: token.DRIVER, Literal: "driver"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			},
		},
		Body: &BlockStmt{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}

	expected := "pitstop(driver x) {  }"
	if expr.String() != expected {
		t.Errorf("expr.String() = %q, want %q", expr.String(), expected)
	}
}

func TestParameter_String(t *testing.T) {
	param := &Parameter{
		Token: token.Token{Type: token.DRIVER, Literal: "driver"},
		Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}

	expected := "driver x"
	if param.String() != expected {
		t.Errorf("param.String() = %q, want %q", param.String(), expected)
	}
}

func TestGroupedExpr_String(t *testing.T) {
	expr := &GroupedExpr{
		Token: token.Token{Type: token.LPAREN, Literal: "("},
		Expression: &InfixExpr{
			Token:    token.Token{Type: token.PLUS, Literal: "+"},
			Left:     &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			Operator: "+",
			Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
		},
	}

	expected := "((1 + 2))"
	if expr.String() != expected {
		t.Errorf("expr.String() = %q, want %q", expr.String(), expected)
	}
}

// Test TokenLiteral methods
func TestTokenLiteral_Methods(t *testing.T) {
	tests := []struct {
		name     string
		node     Node
		expected string
	}{
		{
			"VarDecl",
			&VarDecl{Token: token.Token{Literal: "driver"}},
			"driver",
		},
		{
			"Assignment",
			&Assignment{Token: token.Token{Literal: "x"}},
			"x",
		},
		{
			"ExpressionStmt",
			&ExpressionStmt{Token: token.Token{Literal: "5"}},
			"5",
		},
		{
			"BlockStmt",
			&BlockStmt{Token: token.Token{Literal: "{"}},
			"{",
		},
		{
			"IfStmt",
			&IfStmt{Token: token.Token{Literal: "drs"}},
			"drs",
		},
		{
			"ForStmt",
			&ForStmt{Token: token.Token{Literal: "lap"}},
			"lap",
		},
		{
			"ReturnStmt",
			&ReturnStmt{Token: token.Token{Literal: "finish"}},
			"finish",
		},
		{
			"FunctionDecl",
			&FunctionDecl{Token: token.Token{Literal: "pitstop"}},
			"pitstop",
		},
		{
			"Parameter",
			&Parameter{Token: token.Token{Literal: "driver"}},
			"driver",
		},
		{
			"Identifier",
			&Identifier{Token: token.Token{Literal: "foo"}},
			"foo",
		},
		{
			"IntegerLiteral",
			&IntegerLiteral{Token: token.Token{Literal: "42"}},
			"42",
		},
		{
			"StringLiteral",
			&StringLiteral{Token: token.Token{Literal: "hello"}},
			"hello",
		},
		{
			"BooleanLiteral",
			&BooleanLiteral{Token: token.Token{Literal: "greenlight"}},
			"greenlight",
		},
		{
			"PrefixExpr",
			&PrefixExpr{Token: token.Token{Literal: "-"}},
			"-",
		},
		{
			"InfixExpr",
			&InfixExpr{Token: token.Token{Literal: "+"}},
			"+",
		},
		{
			"CallExpr",
			&CallExpr{Token: token.Token{Literal: "("}},
			"(",
		},
		{
			"GroupedExpr",
			&GroupedExpr{Token: token.Token{Literal: "("}},
			"(",
		},
		{
			"LambdaExpr",
			&LambdaExpr{Token: token.Token{Literal: "pitstop"}},
			"pitstop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.node.TokenLiteral() != tt.expected {
				t.Errorf("%s.TokenLiteral() = %q, want %q", tt.name, tt.node.TokenLiteral(), tt.expected)
			}
		})
	}
}
