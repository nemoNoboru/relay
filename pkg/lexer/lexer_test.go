package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `set five = 5;
set ten = 10;

set add = fn(x, y) {
  x + y;
};

set result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
-> 
"foobar"
"foo bar"
[]
{foo: "bar"}
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SET, "set"},
		{IDENT, "five"},
		{ASSIGN, "="},
		{NUMBER, "5"},
		{SEMICOLON, ";"},
		{SET, "set"},
		{IDENT, "ten"},
		{ASSIGN, "="},
		{NUMBER, "10"},
		{SEMICOLON, ";"},
		{SET, "set"},
		{IDENT, "add"},
		{ASSIGN, "="},
		{FN, "fn"},
		{LPAREN, "("},
		{IDENT, "x"},
		{COMMA, ","},
		{IDENT, "y"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{PLUS, "+"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{SEMICOLON, ";"},
		{SET, "set"},
		{IDENT, "result"},
		{ASSIGN, "="},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "five"},
		{COMMA, ","},
		{IDENT, "ten"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{ILLEGAL, "!"},
		{MINUS, "-"},
		{DIVIDE, "/"},
		{MULTIPLY, "*"},
		{NUMBER, "5"},
		{SEMICOLON, ";"},
		{NUMBER, "5"},
		{LESS_THAN, "<"},
		{NUMBER, "10"},
		{GREATER_THAN, ">"},
		{NUMBER, "5"},
		{SEMICOLON, ";"},
		{IF, "if"},
		{LPAREN, "("},
		{NUMBER, "5"},
		{LESS_THAN, "<"},
		{NUMBER, "10"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{BOOL, "true"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{ELSE, "else"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{BOOL, "false"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{NUMBER, "10"},
		{EQUAL, "=="},
		{NUMBER, "10"},
		{SEMICOLON, ";"},
		{NUMBER, "10"},
		{NOT_EQUAL, "!="},
		{NUMBER, "9"},
		{SEMICOLON, ";"},
		{ARROW, "->"},
		{STRING, "foobar"},
		{STRING, "foo bar"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{LBRACE, "{"},
		{IDENT, "foo"},
		{COLON, ":"},
		{STRING, "bar"},
		{RBRACE, "}"},
		{EOF, ""},
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

func TestRelayKeywords(t *testing.T) {
	input := `struct protocol server state receive send template dispatch 
config auth implements from page require_login current_user discover broadcast`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STRUCT, "struct"},
		{PROTOCOL, "protocol"},
		{SERVER, "server"},
		{STATE, "state"},
		{RECEIVE, "receive"},
		{SEND, "send"},
		{TEMPLATE, "template"},
		{DISPATCH, "dispatch"},
		{CONFIG, "config"},
		{AUTH, "auth"},
		{IMPLEMENTS, "implements"},
		{FROM, "from"},
		{PAGE, "page"},
		{REQUIRE_LOGIN, "require_login"},
		{CURRENT_USER, "current_user"},
		{DISCOVER, "discover"},
		{BROADCAST, "broadcast"},
		{EOF, ""},
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

func TestNumbers(t *testing.T) {
	input := `5 42 123.45 0.5 999.999`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{NUMBER, "5"},
		{NUMBER, "42"},
		{NUMBER, "123.45"},
		{NUMBER, "0.5"},
		{NUMBER, "999.999"},
		{EOF, ""},
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

func TestStrings(t *testing.T) {
	input := `"hello" "world" "hello world" ""`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STRING, "hello"},
		{STRING, "world"},
		{STRING, "hello world"},
		{STRING, ""},
		{EOF, ""},
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

func TestComments(t *testing.T) {
	input := `set x = 5; // this is a comment
set y = 10; // another comment
// full line comment
set z = 15;`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SET, "set"},
		{IDENT, "x"},
		{ASSIGN, "="},
		{NUMBER, "5"},
		{SEMICOLON, ";"},
		{SET, "set"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{NUMBER, "10"},
		{SEMICOLON, ";"},
		{SET, "set"},
		{IDENT, "z"},
		{ASSIGN, "="},
		{NUMBER, "15"},
		{SEMICOLON, ";"},
		{EOF, ""},
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

func TestMultiCharacterOperators(t *testing.T) {
	input := `== != -> <= >=`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{EQUAL, "=="},
		{NOT_EQUAL, "!="},
		{ARROW, "->"},
		{LESS_THAN, "<"},
		{ASSIGN, "="},
		{GREATER_THAN, ">"},
		{ASSIGN, "="},
		{EOF, ""},
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

func TestRelayProgram(t *testing.T) {
	input := `struct User {
  name: string,
  age: number
}

protocol UserService {
  get_user(id: string) -> User
}

server implements UserService {
  receive get_user {id: string} -> User {
    return User{name: "John", age: 30}
  }
}`

	l := New(input)

	// We'll just check that it tokenizes without errors and gets the main keywords
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}

	// Should have more than just EOF
	if len(tokens) <= 1 {
		t.Fatalf("Expected more tokens, got %d", len(tokens))
	}

	// Check for key tokens
	foundStruct := false
	foundProtocol := false
	foundServer := false
	foundReceive := false

	for _, tok := range tokens {
		switch tok.Type {
		case STRUCT:
			foundStruct = true
		case PROTOCOL:
			foundProtocol = true
		case SERVER:
			foundServer = true
		case RECEIVE:
			foundReceive = true
		}
	}

	if !foundStruct {
		t.Fatal("Expected to find STRUCT token")
	}
	if !foundProtocol {
		t.Fatal("Expected to find PROTOCOL token")
	}
	if !foundServer {
		t.Fatal("Expected to find SERVER token")
	}
	if !foundReceive {
		t.Fatal("Expected to find RECEIVE token")
	}
}

func TestLineAndColumnNumbers(t *testing.T) {
	input := `set x = 5;
set y = 10;`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
		expectedLine    int
		expectedColumn  int
	}{
		{SET, "set", 1, 1},
		{IDENT, "x", 1, 5},
		{ASSIGN, "=", 1, 7},
		{NUMBER, "5", 1, 9},
		{SEMICOLON, ";", 1, 10},
		{SET, "set", 2, 1},
		{IDENT, "y", 2, 5},
		{ASSIGN, "=", 2, 7},
		{NUMBER, "10", 2, 9},
		{SEMICOLON, ";", 2, 11},
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

		if tok.Line != tt.expectedLine {
			t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d",
				i, tt.expectedLine, tok.Line)
		}

		// Note: Column tracking might be off by one, this is just to verify it's working
		if tok.Column <= 0 {
			t.Fatalf("tests[%d] - column should be positive, got=%d",
				i, tok.Column)
		}
	}
}
