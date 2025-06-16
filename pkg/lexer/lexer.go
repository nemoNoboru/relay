package lexer

import (
	"fmt"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF

	// Identifiers and literals
	IDENT  // variable names, function names, etc.
	STRING // "hello"
	NUMBER // 123, 123.45
	BOOL   // true, false

	// Operators
	ASSIGN       // =
	PLUS         // +
	MINUS        // -
	MULTIPLY     // *
	DIVIDE       // /
	EQUAL        // ==
	NOT_EQUAL    // !=
	LESS_THAN    // <
	GREATER_THAN // >
	ARROW        // ->

	// Delimiters
	COMMA     // ,
	SEMICOLON // ;
	COLON     // :
	DOT       // .

	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]

	// Keywords
	STRUCT
	PROTOCOL
	SERVER
	STATE
	RECEIVE
	SEND
	TEMPLATE
	DISPATCH
	CONFIG
	AUTH
	SET
	FN
	IF
	ELSE
	FOR
	IN
	TRY
	CATCH
	RETURN
	THROW
	IMPLEMENTS
	FROM
	PAGE
	REQUIRE_LOGIN
	CURRENT_USER
	DISCOVER
	BROADCAST
)

var keywords = map[string]TokenType{
	"struct":        STRUCT,
	"protocol":      PROTOCOL,
	"server":        SERVER,
	"state":         STATE,
	"receive":       RECEIVE,
	"send":          SEND,
	"template":      TEMPLATE,
	"dispatch":      DISPATCH,
	"config":        CONFIG,
	"auth":          AUTH,
	"set":           SET,
	"fn":            FN,
	"if":            IF,
	"else":          ELSE,
	"for":           FOR,
	"in":            IN,
	"try":           TRY,
	"catch":         CATCH,
	"return":        RETURN,
	"throw":         THROW,
	"implements":    IMPLEMENTS,
	"from":          FROM,
	"page":          PAGE,
	"require_login": REQUIRE_LOGIN,
	"current_user":  CURRENT_USER,
	"discover":      DISCOVER,
	"broadcast":     BROADCAST,
	"true":          BOOL,
	"false":         BOOL,
}

// Token represents a single token
type Token struct {
	Type     TokenType
	Literal  string
	Line     int
	Column   int
	Position int
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %s, Line: %d, Column: %d}",
		t.Type.String(), t.Literal, t.Line, t.Column)
}

// Lexer tokenizes Relay source code
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new lexer instance
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar gives us the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL represents EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken scans the input and returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQUAL, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = newToken(ASSIGN, l.ch, l.line, l.column)
		}
	case '+':
		tok = newToken(PLUS, l.ch, l.line, l.column)
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = newToken(MINUS, l.ch, l.line, l.column)
		}
	case '*':
		tok = newToken(MULTIPLY, l.ch, l.line, l.column)
	case '/':
		if l.peekChar() == '/' {
			// Single line comment
			l.skipComment()
			return l.NextToken()
		} else {
			tok = newToken(DIVIDE, l.ch, l.line, l.column)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQUAL, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = newToken(ILLEGAL, l.ch, l.line, l.column)
		}
	case '<':
		tok = newToken(LESS_THAN, l.ch, l.line, l.column)
	case '>':
		tok = newToken(GREATER_THAN, l.ch, l.line, l.column)
	case ',':
		tok = newToken(COMMA, l.ch, l.line, l.column)
	case ';':
		tok = newToken(SEMICOLON, l.ch, l.line, l.column)
	case ':':
		tok = newToken(COLON, l.ch, l.line, l.column)
	case '.':
		tok = newToken(DOT, l.ch, l.line, l.column)
	case '(':
		tok = newToken(LPAREN, l.ch, l.line, l.column)
	case ')':
		tok = newToken(RPAREN, l.ch, l.line, l.column)
	case '{':
		tok = newToken(LBRACE, l.ch, l.line, l.column)
	case '}':
		tok = newToken(RBRACE, l.ch, l.line, l.column)
	case '[':
		tok = newToken(LBRACKET, l.ch, l.line, l.column)
	case ']':
		tok = newToken(RBRACKET, l.ch, l.line, l.column)
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		tok.Column = l.column
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			tok.Line = l.line
			tok.Column = l.column
			return tok // early return to avoid readChar() call
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			tok.Line = l.line
			tok.Column = l.column
			return tok // early return to avoid readChar() call
		} else {
			tok = newToken(ILLEGAL, l.ch, l.line, l.column)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType TokenType, ch byte, line, column int) Token {
	return Token{Type: tokenType, Literal: string(ch), Line: line, Column: column}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	// Handle decimal numbers
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// String returns a string representation of the TokenType
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case BOOL:
		return "BOOL"
	case ASSIGN:
		return "ASSIGN"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case MULTIPLY:
		return "MULTIPLY"
	case DIVIDE:
		return "DIVIDE"
	case EQUAL:
		return "EQUAL"
	case NOT_EQUAL:
		return "NOT_EQUAL"
	case LESS_THAN:
		return "LESS_THAN"
	case GREATER_THAN:
		return "GREATER_THAN"
	case ARROW:
		return "ARROW"
	case COMMA:
		return "COMMA"
	case SEMICOLON:
		return "SEMICOLON"
	case COLON:
		return "COLON"
	case DOT:
		return "DOT"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
	case STRUCT:
		return "STRUCT"
	case PROTOCOL:
		return "PROTOCOL"
	case SERVER:
		return "SERVER"
	case STATE:
		return "STATE"
	case RECEIVE:
		return "RECEIVE"
	case SEND:
		return "SEND"
	case TEMPLATE:
		return "TEMPLATE"
	case DISPATCH:
		return "DISPATCH"
	case CONFIG:
		return "CONFIG"
	case AUTH:
		return "AUTH"
	case SET:
		return "SET"
	case FN:
		return "FN"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case FOR:
		return "FOR"
	case IN:
		return "IN"
	case TRY:
		return "TRY"
	case CATCH:
		return "CATCH"
	case RETURN:
		return "RETURN"
	case THROW:
		return "THROW"
	case IMPLEMENTS:
		return "IMPLEMENTS"
	case FROM:
		return "FROM"
	case PAGE:
		return "PAGE"
	case REQUIRE_LOGIN:
		return "REQUIRE_LOGIN"
	case CURRENT_USER:
		return "CURRENT_USER"
	case DISCOVER:
		return "DISCOVER"
	case BROADCAST:
		return "BROADCAST"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(t))
	}
}
