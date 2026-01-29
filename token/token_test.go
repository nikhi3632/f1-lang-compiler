package token

import "testing"

func TestLookupIdent_Keywords(t *testing.T) {
	tests := []struct {
		ident    string
		expected TokenType
	}{
		{"driver", DRIVER},
		{"pitstop", PITSTOP},
		{"drs", DRS},
		{"defend", DEFEND},
		{"lap", LAP},
		{"finish", FINISH},
		{"greenlight", GREENLIGHT},
		{"redlight", REDLIGHT},
	}

	for _, tt := range tests {
		t.Run(tt.ident, func(t *testing.T) {
			got := LookupIdent(tt.ident)
			if got != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.ident, got, tt.expected)
			}
		})
	}
}

func TestLookupIdent_Identifiers(t *testing.T) {
	tests := []string{
		"x",
		"foobar",
		"myVar",
		"_private",
		"driver1",
		"Driver", // case-sensitive: not a keyword
		"DRIVER", // case-sensitive: not a keyword
	}

	for _, ident := range tests {
		t.Run(ident, func(t *testing.T) {
			got := LookupIdent(ident)
			if got != IDENT {
				t.Errorf("LookupIdent(%q) = %v, want IDENT", ident, got)
			}
		})
	}
}

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{ILLEGAL, "ILLEGAL"},
		{EOF, "EOF"},
		{INT, "INT"},
		{STRING, "STRING"},
		{IDENT, "IDENT"},
		{DRIVER, "DRIVER"},
		{PITSTOP, "PITSTOP"},
		{DRS, "DRS"},
		{DEFEND, "DEFEND"},
		{LAP, "LAP"},
		{FINISH, "FINISH"},
		{GREENLIGHT, "GREENLIGHT"},
		{REDLIGHT, "REDLIGHT"},
		{PLUS, "PLUS"},
		{MINUS, "MINUS"},
		{STAR, "STAR"},
		{SLASH, "SLASH"},
		{PERCENT, "PERCENT"},
		{EQ, "EQ"},
		{NEQ, "NEQ"},
		{LT, "LT"},
		{GT, "GT"},
		{LTE, "LTE"},
		{GTE, "GTE"},
		{AND, "AND"},
		{OR, "OR"},
		{NOT, "NOT"},
		{ASSIGN, "ASSIGN"},
		{LPAREN, "LPAREN"},
		{RPAREN, "RPAREN"},
		{LBRACE, "LBRACE"},
		{RBRACE, "RBRACE"},
		{SEMICOLON, "SEMICOLON"},
		{COMMA, "COMMA"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.tokenType.String()
			if got != tt.expected {
				t.Errorf("TokenType(%d).String() = %q, want %q", tt.tokenType, got, tt.expected)
			}
		})
	}
}

func TestTokenType_String_Unknown(t *testing.T) {
	// Test an unknown token type (beyond defined constants)
	unknown := TokenType(9999)
	got := unknown.String()
	if got != "UNKNOWN" {
		t.Errorf("TokenType(9999).String() = %q, want %q", got, "UNKNOWN")
	}
}

func TestToken_Struct(t *testing.T) {
	tok := Token{
		Type:    INT,
		Literal: "42",
		Line:    1,
		Column:  5,
	}

	if tok.Type != INT {
		t.Errorf("tok.Type = %v, want INT", tok.Type)
	}
	if tok.Literal != "42" {
		t.Errorf("tok.Literal = %q, want %q", tok.Literal, "42")
	}
	if tok.Line != 1 {
		t.Errorf("tok.Line = %d, want 1", tok.Line)
	}
	if tok.Column != 5 {
		t.Errorf("tok.Column = %d, want 5", tok.Column)
	}
}
