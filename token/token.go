package token

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special
	ILLEGAL TokenType = iota
	EOF

	// Literals
	INT    // 42
	STRING // "hello"
	IDENT  // myVar

	// Keywords
	DRIVER     // driver
	PITSTOP    // pitstop
	DRS        // drs
	DEFEND     // defend
	LAP        // lap
	FINISH     // finish
	GREENLIGHT // greenlight (true)
	REDLIGHT   // redlight (false)

	// Operators
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %

	EQ  // ==
	NEQ // !=
	LT  // <
	GT  // >
	LTE // <=
	GTE // >=

	AND // &&
	OR  // ||
	NOT // !

	ASSIGN // =

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	SEMICOLON // ;
	COMMA     // ,
)

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"driver":     DRIVER,
	"pitstop":    PITSTOP,
	"drs":        DRS,
	"defend":     DEFEND,
	"lap":        LAP,
	"finish":     FINISH,
	"greenlight": GREENLIGHT,
	"redlight":   REDLIGHT,
}

// LookupIdent checks if an identifier is a keyword.
// Returns the keyword token type if it is, IDENT otherwise.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// String returns a human-readable name for the token type.
func (t TokenType) String() string {
	names := map[TokenType]string{
		ILLEGAL:    "ILLEGAL",
		EOF:        "EOF",
		INT:        "INT",
		STRING:     "STRING",
		IDENT:      "IDENT",
		DRIVER:     "DRIVER",
		PITSTOP:    "PITSTOP",
		DRS:        "DRS",
		DEFEND:     "DEFEND",
		LAP:        "LAP",
		FINISH:     "FINISH",
		GREENLIGHT: "GREENLIGHT",
		REDLIGHT:   "REDLIGHT",
		PLUS:       "PLUS",
		MINUS:      "MINUS",
		STAR:       "STAR",
		SLASH:      "SLASH",
		PERCENT:    "PERCENT",
		EQ:         "EQ",
		NEQ:        "NEQ",
		LT:         "LT",
		GT:         "GT",
		LTE:        "LTE",
		GTE:        "GTE",
		AND:        "AND",
		OR:         "OR",
		NOT:        "NOT",
		ASSIGN:     "ASSIGN",
		LPAREN:     "LPAREN",
		RPAREN:     "RPAREN",
		LBRACE:     "LBRACE",
		RBRACE:     "RBRACE",
		SEMICOLON:  "SEMICOLON",
		COMMA:      "COMMA",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "UNKNOWN"
}
