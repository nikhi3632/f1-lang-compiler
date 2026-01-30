package lexer

import (
	"strconv"

	"f1c/errors"
	"f1c/token"
)

// Lexer tokenizes F1-Lang source code.
type Lexer struct {
	input    string
	pos      int  // current position in input (points to current char)
	readPos  int  // reading position (after current char)
	ch       byte // current char under examination
	line     int
	column   int
	reporter *errors.Reporter
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

// NewWithReporter creates a new Lexer with an error reporter.
func NewWithReporter(input string, reporter *errors.Reporter) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0, reporter: reporter}
	l.readChar()
	return l
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	l.skipComments()

	// Record token position before reading
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '+':
		tok = l.newToken(token.PLUS, l.ch)
	case '-':
		tok = l.newToken(token.MINUS, l.ch)
	case '*':
		tok = l.newToken(token.STAR, l.ch)
	case '/':
		tok = l.newToken(token.SLASH, l.ch)
	case '%':
		tok = l.newToken(token.PERCENT, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = l.newToken(token.COMMA, l.ch)

	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}

	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = l.newToken(token.NOT, l.ch)
		}

	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = l.newToken(token.LT, l.ch)
		}

	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = l.newToken(token.GT, l.ch)
		}

	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			l.reportError(l.line, l.column, 1, "unexpected character '%c', did you mean '&&'?", l.ch)
			tok = l.newToken(token.ILLEGAL, l.ch)
		}

	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			l.reportError(l.line, l.column, 1, "unexpected character '%c', did you mean '||'?", l.ch)
			tok = l.newToken(token.ILLEGAL, l.ch)
		}

	case '"':
		tok.Line = l.line
		tok.Column = l.column
		str, ok := l.readString()
		if ok {
			tok.Type = token.STRING
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = str
		}
		return tok

	case 0:
		tok.Literal = ""
		tok.Type = token.EOF

	default:
		if isLetter(l.ch) {
			tok.Line = l.line
			tok.Column = l.column
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Line = l.line
			tok.Column = l.column
			literal, ok := l.readNumber()
			if ok {
				tok.Type = token.INT
				tok.Literal = literal
			} else {
				l.reportError(tok.Line, tok.Column, len(literal), "integer literal overflow: value exceeds 64-bit signed integer range")
				tok.Type = token.ILLEGAL
				tok.Literal = literal
			}
			return tok
		} else {
			l.reportError(l.line, l.column, 1, "unexpected character '%c'", l.ch)
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    l.line,
		Column:  l.column,
	}
}

// reportError reports an error to the reporter if one is configured.
func (l *Lexer) reportError(line, col, length int, format string, args ...any) {
	if l.reporter != nil {
		l.reporter.Add(errors.LexerError, line, col, length, format, args...)
	}
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComments() {
	for l.ch == '/' && l.peekChar() == '/' {
		// Skip until end of line
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		l.skipWhitespace()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.pos]
}

func (l *Lexer) readNumber() (string, bool) {
	position := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	literal := l.input[position:l.pos]

	// Check for overflow
	_, err := strconv.ParseInt(literal, 10, 64)
	if err != nil {
		// Check if it's an overflow error specifically
		if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
			return literal, false
		}
		// For numbers larger than int64 max, ParseInt returns error
		// Check if number is too large
		if len(literal) > 19 || (len(literal) == 19 && literal > "9223372036854775807") {
			return literal, false
		}
	}

	return literal, true
}

func (l *Lexer) readString() (string, bool) {
	var result []byte
	startLine := l.line
	startCol := l.column

	l.readChar() // skip opening quote

	for {
		if l.ch == '"' {
			l.readChar() // skip closing quote
			return string(result), true
		}

		if l.ch == 0 || l.ch == '\n' {
			// Unterminated string
			l.reportError(startLine, startCol, 1, "unterminated string literal")
			return string(result), false
		}

		if l.ch == '\\' {
			escLine := l.line
			escCol := l.column
			l.readChar() // read escape character
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			default:
				// Invalid escape sequence
				l.reportError(escLine, escCol, 2, "invalid escape sequence '\\%c'", l.ch)
				return string(result), false
			}
		} else {
			result = append(result, l.ch)
		}

		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
