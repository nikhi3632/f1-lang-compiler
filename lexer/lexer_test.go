package lexer

import (
	"testing"

	"f1c/token"
)

func TestNextToken_SingleCharTokens(t *testing.T) {
	input := `+-*/%(){}=;,<>!`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.STAR, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.ASSIGN, "="},
		{token.SEMICOLON, ";"},
		{token.COMMA, ","},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.NOT, "!"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_TwoCharTokens(t *testing.T) {
	input := `== != <= >= && ||`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.EQ, "=="},
		{token.NEQ, "!="},
		{token.LTE, "<="},
		{token.GTE, ">="},
		{token.AND, "&&"},
		{token.OR, "||"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	input := `driver pitstop drs defend lap finish greenlight redlight`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DRIVER, "driver"},
		{token.PITSTOP, "pitstop"},
		{token.DRS, "drs"},
		{token.DEFEND, "defend"},
		{token.LAP, "lap"},
		{token.FINISH, "finish"},
		{token.GREENLIGHT, "greenlight"},
		{token.REDLIGHT, "redlight"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Identifiers(t *testing.T) {
	input := `x points max_speed driver1 _private`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.IDENT, "points"},
		{token.IDENT, "max_speed"},
		{token.IDENT, "driver1"},
		{token.IDENT, "_private"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Integers(t *testing.T) {
	input := `0 42 1000000 9223372036854775807`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "0"},
		{token.INT, "42"},
		{token.INT, "1000000"},
		{token.INT, "9223372036854775807"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{
			name:            "simple string",
			input:           `"Hello, World!"`,
			expectedType:    token.STRING,
			expectedLiteral: "Hello, World!",
		},
		{
			name:            "empty string",
			input:           `""`,
			expectedType:    token.STRING,
			expectedLiteral: "",
		},
		{
			name:            "string with newline escape",
			input:           `"Line 1\nLine 2"`,
			expectedType:    token.STRING,
			expectedLiteral: "Line 1\nLine 2",
		},
		{
			name:            "string with tab escape",
			input:           `"col1\tcol2"`,
			expectedType:    token.STRING,
			expectedLiteral: "col1\tcol2",
		},
		{
			name:            "string with backslash escape",
			input:           `"path\\file"`,
			expectedType:    token.STRING,
			expectedLiteral: "path\\file",
		},
		{
			name:            "string with quote escape",
			input:           `"He said \"hello\""`,
			expectedType:    token.STRING,
			expectedLiteral: `He said "hello"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Errorf("tokentype wrong. expected=%q, got=%q",
					tt.expectedType, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `// This is a comment
driver x = 5; // inline comment
// Another comment
radio(x);`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DRIVER, "driver"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "radio"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Whitespace(t *testing.T) {
	input := `
	driver   x   =   5   ;
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DRIVER, "driver"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_CompleteProgram(t *testing.T) {
	input := `pitstop add(driver a, driver b) {
	finish a + b;
}

driver result = add(3, 4);
radio(result);`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PITSTOP, "pitstop"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.DRIVER, "driver"},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.DRIVER, "driver"},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.FINISH, "finish"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.DRIVER, "driver"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.INT, "3"},
		{token.COMMA, ","},
		{token.INT, "4"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "radio"},
		{token.LPAREN, "("},
		{token.IDENT, "result"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_LineAndColumn(t *testing.T) {
	input := `driver x = 5;
driver y = 10;`

	l := New(input)

	// First token: "driver" at line 1, column 1
	tok := l.NextToken()
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("expected line=1, column=1, got line=%d, column=%d", tok.Line, tok.Column)
	}

	// Skip to "5"
	l.NextToken() // x
	l.NextToken() // =
	tok = l.NextToken()
	if tok.Type != token.INT || tok.Literal != "5" {
		t.Fatalf("expected INT '5', got %s '%s'", tok.Type, tok.Literal)
	}
	if tok.Line != 1 || tok.Column != 12 {
		t.Errorf("expected line=1, column=12, got line=%d, column=%d", tok.Line, tok.Column)
	}

	// Skip to second line "driver"
	l.NextToken() // ;
	tok = l.NextToken()
	if tok.Type != token.DRIVER {
		t.Fatalf("expected DRIVER, got %s", tok.Type)
	}
	if tok.Line != 2 {
		t.Errorf("expected line=2, got line=%d", tok.Line)
	}
}

func TestNextToken_IllegalCharacter(t *testing.T) {
	input := `driver x$y = 5;`

	l := New(input)

	l.NextToken() // driver
	l.NextToken() // x
	tok := l.NextToken()

	if tok.Type != token.ILLEGAL {
		t.Errorf("expected ILLEGAL token, got %s", tok.Type)
	}
	if tok.Literal != "$" {
		t.Errorf("expected literal '$', got %q", tok.Literal)
	}
}

func TestNextToken_InvalidEscapeSequence(t *testing.T) {
	input := `"hello\q"`

	l := New(input)
	tok := l.NextToken()

	if tok.Type != token.ILLEGAL {
		t.Errorf("expected ILLEGAL token for invalid escape, got %s", tok.Type)
	}
}

func TestNextToken_UnterminatedString(t *testing.T) {
	input := `"hello`

	l := New(input)
	tok := l.NextToken()

	if tok.Type != token.ILLEGAL {
		t.Errorf("expected ILLEGAL token for unterminated string, got %s", tok.Type)
	}
}

func TestNextToken_IntegerOverflow(t *testing.T) {
	// This number exceeds int64 max (9223372036854775807)
	input := `99999999999999999999999`

	l := New(input)
	tok := l.NextToken()

	if tok.Type != token.ILLEGAL {
		t.Errorf("expected ILLEGAL token for overflow, got %s", tok.Type)
	}
}

func TestNextToken_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{
			name:            "single equals vs double equals",
			input:           `=`,
			expectedType:    token.ASSIGN,
			expectedLiteral: "=",
		},
		{
			name:            "less than vs less than or equal",
			input:           `<`,
			expectedType:    token.LT,
			expectedLiteral: "<",
		},
		{
			name:            "exclamation vs not equal",
			input:           `!`,
			expectedType:    token.NOT,
			expectedLiteral: "!",
		},
		{
			name:            "ampersand alone is illegal",
			input:           `&`,
			expectedType:    token.ILLEGAL,
			expectedLiteral: "&",
		},
		{
			name:            "pipe alone is illegal",
			input:           `|`,
			expectedType:    token.ILLEGAL,
			expectedLiteral: "|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Errorf("tokentype wrong. expected=%q, got=%q",
					tt.expectedType, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}
