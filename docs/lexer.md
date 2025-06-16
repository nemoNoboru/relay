# Relay Lexer Technical Documentation

## Overview

The Relay lexer is responsible for the first phase of the compilation process: **tokenization**. It takes raw Relay source code as input and breaks it down into a stream of tokens that represent the fundamental building blocks of the language.

## Architecture

### Core Components

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Source Code   │───▶│    Lexer     │───▶│   Token Stream  │
│  (.relay file)  │    │              │    │                 │
└─────────────────┘    └──────────────┘    └─────────────────┘
```

**Location**: `pkg/lexer/`

**Key Files**:
- `lexer.go` - Main lexer implementation
- `lexer_test.go` - Comprehensive unit tests
- `integration_test.go` - Integration tests with real Relay code

## Token Structure

### TokenType Enum

The lexer recognizes the following categories of tokens:

#### Special Tokens
```go
ILLEGAL   // Invalid character
EOF       // End of file
```

#### Identifiers & Literals
```go
IDENT     // variable names, function names (e.g., "username", "get_posts")
STRING    // string literals (e.g., "hello world")
NUMBER    // numeric literals (e.g., 42, 123.45)
BOOL      // boolean literals (true, false)
```

#### Operators
```go
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
```

#### Delimiters
```go
COMMA        // ,
SEMICOLON    // ;
COLON        // :
DOT          // .
LPAREN       // (
RPAREN       // )
LBRACE       // {
RBRACE       // }
LBRACKET     // [
RBRACKET     // ]
```

#### Relay Keywords
```go
STRUCT       // struct
PROTOCOL     // protocol
SERVER       // server
STATE        // state
RECEIVE      // receive
SEND         // send
TEMPLATE     // template
DISPATCH     // dispatch
CONFIG       // config
AUTH         // auth
SET          // set
FN           // fn
IF           // if
ELSE         // else
FOR          // for
IN           // in
TRY          // try
CATCH        // catch
RETURN       // return
THROW        // throw
IMPLEMENTS   // implements
FROM         // from
PAGE         // page
REQUIRE_LOGIN// require_login
CURRENT_USER // current_user
DISCOVER     // discover
BROADCAST    // broadcast
```

### Token Structure

```go
type Token struct {
    Type     TokenType  // The type of token
    Literal  string     // The actual text from source
    Line     int        // Line number (1-based)
    Column   int        // Column number (1-based)
    Position int        // Absolute position in source
}
```

**Example Token**:
```go
Token{
    Type: STRUCT, 
    Literal: "struct", 
    Line: 1, 
    Column: 1
}
```

## Lexer Implementation

### Core Algorithm

The lexer uses a **single-pass scanning algorithm** with **lookahead**:

1. **Character Reading**: Reads one character at a time from input
2. **Pattern Matching**: Matches character patterns to token types
3. **Token Generation**: Creates tokens with position information
4. **State Management**: Tracks line/column numbers for error reporting

### Key Methods

#### `New(input string) *Lexer`
Creates a new lexer instance for the given input string.

```go
l := lexer.New(sourceCode)
```

#### `NextToken() Token`
Advances the lexer and returns the next token in the stream.

```go
for {
    tok := l.NextToken()
    if tok.Type == lexer.EOF {
        break
    }
    // Process token...
}
```

### Character Processing

#### Multi-Character Operators
The lexer uses **lookahead** to handle multi-character operators:

```go
case '=':
    if l.peekChar() == '=' {
        // Found "=="
        ch := l.ch
        l.readChar()
        tok = Token{Type: EQUAL, Literal: string(ch) + string(l.ch)}
    } else {
        // Found "="
        tok = newToken(ASSIGN, l.ch)
    }
```

#### String Literals
Strings are delimited by double quotes and can contain spaces:

```go
"hello world"     // → Token{Type: STRING, Literal: "hello world"}
""                // → Token{Type: STRING, Literal: ""}
```

#### Numbers
Supports both integers and floating-point numbers:

```go
42               // → Token{Type: NUMBER, Literal: "42"}
123.45          // → Token{Type: NUMBER, Literal: "123.45"}
0.5             // → Token{Type: NUMBER, Literal: "0.5"}
```

#### Comments
Single-line comments starting with `//` are automatically skipped:

```go
set x = 5; // this is a comment
// This entire line is ignored
set y = 10;
```

## Usage Examples

### Basic Tokenization

```go
package main

import (
    "fmt"
    "github.com/cloudpunks/relay/pkg/lexer"
)

func main() {
    input := `struct User {
        name: string,
        age: number
    }`
    
    l := lexer.New(input)
    
    for {
        tok := l.NextToken()
        fmt.Printf("Token: %s\n", tok.String())
        
        if tok.Type == lexer.EOF {
            break
        }
    }
}
```

### Working with Token Stream

```go
func parseTokens(input string) []lexer.Token {
    l := lexer.New(input)
    var tokens []lexer.Token
    
    for {
        tok := l.NextToken()
        tokens = append(tokens, tok)
        
        if tok.Type == lexer.EOF {
            break
        }
    }
    
    return tokens
}
```

## Testing

### Test Coverage

The lexer has **comprehensive test coverage** with 8+ test functions:

#### Unit Tests (`lexer_test.go`)

1. **`TestNextToken`** - Tests all basic token types and operators
2. **`TestRelayKeywords`** - Verifies all Relay keywords are recognized
3. **`TestNumbers`** - Tests integer and decimal parsing
4. **`TestStrings`** - Tests string literal parsing
5. **`TestComments`** - Verifies comment handling
6. **`TestMultiCharacterOperators`** - Tests `==`, `!=`, `->` 
7. **`TestRelayProgram`** - Tests complete program tokenization
8. **`TestLineAndColumnNumbers`** - Tests position tracking

#### Integration Tests (`integration_test.go`)

1. **`TestRelayFileTokenization`** - Tests real `.relay` files
2. **`TestTokenizeAndPrint`** - Debug output for development

### Running Tests

```bash
# Run all lexer tests
go test ./pkg/lexer/ -v

# Run specific test
go test ./pkg/lexer/ -v -run TestRelayKeywords

# Run with coverage
go test ./pkg/lexer/ -cover
```

### Test Example

```go
func TestSimpleStruct(t *testing.T) {
    input := `struct User { name: string }`
    
    tests := []struct {
        expectedType    TokenType
        expectedLiteral string
    }{
        {STRUCT, "struct"},
        {IDENT, "User"},
        {LBRACE, "{"},
        {IDENT, "name"},
        {COLON, ":"},
        {IDENT, "string"},
        {RBRACE, "}"},
        {EOF, ""},
    }
    
    l := New(input)
    
    for i, tt := range tests {
        tok := l.NextToken()
        
        if tok.Type != tt.expectedType {
            t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
                i, tt.expectedType, tok.Type)
        }
        
        if tok.Literal != tt.expectedLiteral {
            t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
                i, tt.expectedLiteral, tok.Literal)
        }
    }
}
```

## Interactive Testing

### test-lexer Utility

We provide a command-line utility for interactive testing:

```bash
# Build the utility
go build -o test-lexer cmd/test-lexer/main.go

# Test a .relay file
./test-lexer examples/simple_blog.relay

# Test inline code
echo 'struct User { name: string }' | ./test-lexer -
```

**Output Example**:
```
Tokenizing:
struct User { name: string }

==================================================
  0: Token{Type: STRUCT, Literal: struct, Line: 1, Column: 7}
  1: Token{Type: IDENT, Literal: User, Line: 1, Column: 12}
  2: Token{Type: LBRACE, Literal: {, Line: 1, Column: 13}
  3: Token{Type: IDENT, Literal: name, Line: 1, Column: 19}
  4: Token{Type: COLON, Literal: :, Line: 1, Column: 19}
  5: Token{Type: IDENT, Literal: string, Line: 1, Column: 27}
  6: Token{Type: RBRACE, Literal: }, Line: 1, Column: 28}
  7: Token{Type: EOF, Literal: , Line: 2, Column: 1}

Total tokens: 8
```

## Performance Characteristics

### Time Complexity
- **O(n)** where n is the length of input
- **Single-pass** algorithm with minimal backtracking
- **Constant memory** per token

### Memory Usage
- **Minimal allocation** - reuses character buffer
- **Position tracking** adds ~12 bytes per token
- **Token storage** depends on consumer (parser)

### Benchmarks

```bash
# Run performance benchmarks (when available)
go test ./pkg/lexer/ -bench=.
```

## Error Handling

### ILLEGAL Tokens
When the lexer encounters an invalid character, it generates an `ILLEGAL` token:

```go
set x = 5@  // '@' is not valid in Relay
// Results in: [..., Token{Type: ILLEGAL, Literal: "@"}, ...]
```

### Position Information
All tokens include precise position information for error reporting:

```go
if tok.Type == lexer.ILLEGAL {
    fmt.Printf("Illegal character '%s' at line %d, column %d\n", 
        tok.Literal, tok.Line, tok.Column)
}
```

## Extending the Lexer

### Adding New Keywords

1. **Add to TokenType enum**:
```go
const (
    // ... existing tokens ...
    ASYNC     // async
)
```

2. **Add to keywords map**:
```go
var keywords = map[string]TokenType{
    // ... existing keywords ...
    "async": ASYNC,
}
```

3. **Add to String() method**:
```go
func (t TokenType) String() string {
    switch t {
    // ... existing cases ...
    case ASYNC:
        return "ASYNC"
    }
}
```

4. **Add tests**:
```go
func TestAsyncKeyword(t *testing.T) {
    input := "async"
    l := New(input)
    
    tok := l.NextToken()
    if tok.Type != ASYNC {
        t.Fatalf("Expected ASYNC, got %v", tok.Type)
    }
}
```

### Adding New Operators

For single-character operators, add to `NextToken()` switch:

```go
case '%':
    tok = newToken(MODULO, l.ch)
```

For multi-character operators, add lookahead logic:

```go
case '<':
    if l.peekChar() == '=' {
        ch := l.ch
        l.readChar()
        tok = Token{Type: LESS_EQUAL, Literal: string(ch) + string(l.ch)}
    } else {
        tok = newToken(LESS_THAN, l.ch)
    }
```

## Common Issues & Solutions

### Issue: Unexpected ILLEGAL tokens
**Cause**: Invalid characters in source code  
**Solution**: Check for unsupported characters, ensure proper encoding

### Issue: Wrong line/column numbers  
**Cause**: Position tracking bug  
**Solution**: Verify `readChar()` properly updates line/column counters

### Issue: Keywords not recognized
**Cause**: Missing from keywords map or typo  
**Solution**: Check keywords map and ensure exact spelling

### Issue: String literals not closed
**Cause**: Missing closing quote  
**Solution**: Lexer will read until EOF - parser should catch this

## Integration with Parser

The lexer is designed to work seamlessly with the parser:

```go
// Parser consumes tokens from lexer
type Parser struct {
    lexer *lexer.Lexer
    curToken Token
    peekToken Token
}

func (p *Parser) nextToken() {
    p.curToken = p.peekToken
    p.peekToken = p.lexer.NextToken()
}
```

## Future Enhancements

### Planned Features
- [ ] **Template string literals** with `${expression}` interpolation
- [ ] **Multiline comments** with `/* ... */`
- [ ] **Raw string literals** with backticks
- [ ] **Binary/hex number literals** (0b1010, 0xFF)
- [ ] **Unicode identifier support**
- [ ] **Better error recovery**

### Performance Optimizations
- [ ] **String interning** for keywords/identifiers
- [ ] **Memory pooling** for token allocation
- [ ] **SIMD optimizations** for character scanning

## References

- [Relay Language Specification](../spec.md)
- [AST Documentation](./ast.md) *(coming next)*
- [Parser Documentation](./parser.md) *(coming next)*

---

*Last updated: [Current Date]*  
*Maintainers: Relay Core Team* 