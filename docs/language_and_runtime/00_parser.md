# Relay Language Parser Technical Specification

**Version:** 0.3 "Cloudpunks Edition"  
**Parser Library:** Participle v2  
**Language Target:** Go  
**Status:** Implemented

---

## 1. Overview

This document specifies the complete technical implementation of the Relay language parser using the Participle parser combinator library for Go. The parser transforms Relay source code into an Abstract Syntax Tree (AST) suitable for compilation or interpretation.

---

## 2. Complete Lexer Specification

The Relay lexer definition:

```go
var relayLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Comments
	{"Comment", `//[^\n]*`},
	{"BlockComment", `/\*([^*]|\*[^/])*\*/`},

	// Literals
	{"String", `"(\\"|[^"])*"`},
	{"Number", `(\d*\.)?\d+`},
	{"Bool", `\b(true|false)\b`},
	{"Nil", `\bnil\b`},
	{"DateTime", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`},
	{"Symbol", `:[a-zA-Z_][a-zA-Z0-9_]*`},

	// Keywords (must come before Ident)
	{"For", `\bfor\b`},
	{"In", `\bin\b`},
	{"Try", `\btry\b`},
	{"Catch", `\bcatch\b`},
	{"Dispatch", `\bdispatch\b`},
	{"Send", `\bsend\b`},
	{"Set", `\bset\b`},
	{"Return", `\breturn\b`},
	{"If", `\bif\b`},
	{"Else", `\belse\b`},
	{"Throw", `\bthrow\b`},
	{"Struct", `\bstruct\b`},
	{"Protocol", `\bprotocol\b`},
	{"Server", `\bserver\b`},
	{"Implements", `\bimplements\b`},
	{"Fn", `\bfn\b`},
	{"Template", `\btemplate\b`},
	{"Config", `\bconfig\b`},
	{"From", `\bfrom\b`},

	{"State", `\bstate\b`},
	{"Receive", `\breceive\b`},

	// Identifiers and Keywords
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},

	// Operators and Punctuation (compound operators first)
	{"Arrow", `->`},
	{"NullCoalesce", `\?\?`},
	{"PlusAssign", `\+=`},
	{"MinusAssign", `-=`},
	{"MultiplyAssign", `\*=`},
	{"DivideAssign", `/=`},
	{"Eq", `==`},
	{"Ne", `!=`},
	{"Le", `<=`},
	{"Ge", `>=`},
	{"And", `&&`},
	{"Or", `\|\|`},
	{"Assign", `=`},
	{"Lt", `<`},
	{"Gt", `>`},
	{"Plus", `\+`},
	{"Minus", `-`},
	{"Multiply", `\*`},
	{"Divide", `/`},
	{"Not", `!`},
	{"Dot", `\.`},
	{"Comma", `,`},
	{"Semicolon", `;`},
	{"Colon", `:`},
	{"LParen", `\(`},
	{"RParen", `\)`},
	{"LBrace", `\{`},
	{"RBrace", `\}`},
	{"LBracket", `\[`},
	{"RBracket", `\]`},
	{"Question", `\?`},
	{"Pipe", `\|`},

	// Whitespace
	{"whitespace", `[ \t\r\n]+`},
})
```

---

## 3. AST Node Definitions

### 3.1 Root Program Structure

```go
// Program represents the root of a Relay program
type Program struct {
	Pos lexer.Position

	Expressions []*Expression `@@*`
}
```

An `Expression` is the top-level node for all language constructs.

```go
type Expression struct {
	StructExpr   *StructExpr   `@@`
	ProtocolExpr *ProtocolExpr `| @@`
	ServerExpr   *ServerExpr   `| @@`
	FunctionExpr *FunctionExpr `| @@`
	TemplateExpr *TemplateExpr `| @@`
	ConfigExpr   *ConfigExpr   `| @@`
	ForExpr      *ForExpr      `| @@`
	TryExpr      *TryExpr      `| @@`
	DispatchExpr *DispatchExpr `| @@`
	SetExpr      *SetExpr      `| @@`
	ReturnExpr   *ReturnExpr   `| @@`
	IfExpr       *IfExpr       `| @@`
	ThrowExpr    *ThrowExpr    `| @@`
	Binary       *BinaryExpr   `| @@`
}
```
This reflects a grammar where any of these constructs can appear at the top level of a file.

### 3.2 Expression Hierarchy

The parser uses a standard recursive descent structure for expressions.

```go
type BinaryExpr struct {
	Left  *UnaryExpr  `@@`
	Right []*BinaryOp `@@*`
}

type BinaryOp struct {
	Op    string     `@( "&&" | "||" | "??" | "==" | "!=" | "<=" | ">=" | "<" | ">" | "+" | "-" | "*" | "/" )`
	Right *UnaryExpr `@@`
}

type UnaryExpr struct {
	Op      *string      `@( "!" | "-" )?`
	Primary *PrimaryExpr `@@`
}

type PrimaryExpr struct {
	Base   *BaseExpr     `@@`
	Access []*AccessExpr `@@*`
}
```

### 3.3 Base Expressions and Literals

The `BaseExpr` is the foundation of an expression, representing a literal, identifier, or grouped expression.

```go
type BaseExpr struct {
	Literal           *Literal           `@@`
	StructConstructor *StructConstructor `| @@`
	SendExpr          *SendExpr          `| @@`
	Lambda            *LambdaExpr        `| @@`
	FuncCall          *FuncCallExpr      `| @@`
	ObjectLit         *ObjectLit         `| @@`
	Block             *Block             `| @@`
	Grouped           *Expression        `| "(" @@ ")"`
	Identifier        *string            `| @( Ident | "state" | "send" | "receive" | "protocol" | "struct" | "for" | "in" | "try" | "catch" | "dispatch" | "set" | "return" | "if" | "else" | "throw" | "server" | "implements" )`
}

type Literal struct {
	String   *string       `@String`
	Number   *float64      `| @Number`
	Bool     *string       `| @Bool`
	Nil      *string       `| @Nil`
	Symbol   *string       `| @Symbol`
	Array    *ArrayLiteral `| @@`
	FuncCall *FuncCall     `| @@`
}
```

### 3.4 Field and Method Access

The parser supports both direct field access and method calls.

```go
type AccessExpr struct {
	MethodCall  *MethodCall     `@@`
	FieldAccess *FieldAccess    `| @@`
	FuncCall    *FuncCallAccess `| @@`
}

type MethodCall struct {
	Method string        `"." @( Ident | "set" | "get" | "add" | "remove" | "filter" | "map" | "find" | "sort_by" | "reduce" )`
	Args   []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type FieldAccess struct {
	Field string `"." @Ident`
}
```

The rest of the AST nodes for structs, servers, functions, etc., are detailed in the `pkg/parser/parser.go` file and follow a similar structure.

---

## 4. Parser Configuration

```go
var relayParser = participle.MustBuild[Program](
	participle.Lexer(relayLexer),
	participle.Unquote("String"),
	participle.UseLookahead(2), // Reduced lookahead due to simplified grammar
	participle.Elide("Whitespace", "Comment", "BlockComment"),
	
	// Custom token transformations
	participle.Map(func(token lexer.Token) (lexer.Token, error) {
		// Transform symbol tokens: :hello -> "hello"
		if token.Type == "Symbol" {
			return lexer.Token{
				Type:  "String",
				Value: `"` + token.Value[1:] + `"`, // Remove : and add quotes
				Pos:   token.Pos,
			}, nil
		}
		return token, nil
	}),
)

func Parse(filename string, r io.Reader) (*Program, error) {
	program, err := relayParser.Parse(filename, r)
	if err != nil {
		return nil, fmt.Errorf("parse error in %s: %w", filename, err)
	}
	
	// Post-parsing validation
	if err := validateAST(program); err != nil {
		return nil, fmt.Errorf("validation error in %s: %w", filename, err)
	}
	
	return program, nil
}

func validateAST(program *Program) error {
	validator := &ASTValidator{}
	return validator.Validate(program)
}
```

---

## 5. Key Grammar Simplifications

### 5.1 Eliminated Ambiguities

**No More Context Sensitivity:**
```go
// All keywords are strictly reserved
state.get("field")    // 'state' must be identifier
send "service" method // 'send' must be keyword
```

**Unified Function Syntax:**
```go
// All functions follow same pattern
fn named_function(param: type) { body }
fn (param: type) { body }  // Lambda
receive fn method_name(param: type) { body }
```

**Clear Field Access:**
```go
// Always method call - no ambiguity
object.get("field")      // Field access
object.method(args)      // Method call
```

### 5.2 Expression-First Design

**Blocks as Expressions:**
```go
type BlockExpr struct {
	Statements []*Statement   // Intermediate statements
	Result     *Expression    // Final expression (optional)
}
```

**Everything Returns Values:**
```go
set result = if condition { "true" } else { "false" }
set value = { set x = 10; x * 2 }  // Block evaluates to 20
```

### 5.3 Consistent Operator Precedence

```
Precedence levels (highest to lowest):
1. Primary: literals, identifiers, grouping
2. Postfix: method calls, array access
3. Unary: !, -
4. Multiplicative: *, /
5. Additive: +, -
6. Relational: <, >, <=, >=
7. Equality: ==, !=
8. Logical AND: &&
9. Logical OR: ||
```

---

## 6. Error Handling and Validation

### 6.1 Syntax Error Reporting

```go
type ParseError struct {
	Pos     lexer.Position
	Message string
	Context string
	
	Expected []string
	Found    string
	Line     string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s\nExpected: %v\nFound: %s", 
		e.Context, e.Pos.Line, e.Pos.Column, e.Message,
		e.Expected, e.Found)
}
```

### 6.2 AST Validation

```go
type ASTValidator struct {
	errors []error
}

func (v *ASTValidator) Validate(program *Program) error {
	v.validateStructs(program)
	v.validateProtocols(program)
	v.validateServers(program)
	v.validateFunctions(program)
	
	if len(v.errors) > 0 {
		return fmt.Errorf("validation failed: %v", v.errors)
	}
	return nil
}
```

---

## 7. Testing Strategy

### 7.1 Grammar Test Coverage

```go
func TestSimplifiedGrammar(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		// Function definitions
		{
			name:  "named function",
			input: `fn add(x: number, y: number) -> number { x + y }`,
			valid: true,
		},
		{
			name:  "lambda function",
			input: `set double = fn (x: number) { x * 2 }`,
			valid: true,
		},
		
		// Field access
		{
			name:  "field access via get",
			input: `user.get("name")`,
			valid: true,
		},
		{
			name:  "method chaining",
			input: `users.filter(fn (u) { u.get("active") }).map(fn (u) { u.get("name") })`,
			valid: true,
		},
		
		// Symbol literals
		{
			name:  "symbol in dispatch",
			input: `dispatch action { :create: fn (data) { "created" } }`,
			valid: true,
		},
		
		// State updates
		{
			name:  "state update via method",
			input: `state.set("count", state.get("count") + 1)`,
			valid: true,
		},
		
		// Block expressions
		{
			name:  "block as expression",
			input: `set result = { set x = 10; x * 2 }`,
			valid: true,
		},
		
		// Server definitions
		{
			name: "complete server",
			input: `
				server test_service {
					state { count: number = 0 }
					receive fn increment() -> number {
						state.set("count", state.get("count") + 1)
						state.get("count")
					}
				}`,
			valid: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tc.input))
			if tc.valid && err != nil {
				t.Errorf("Expected valid input, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected invalid input, got no error")
			}
		})
	}
}
```

---

## 8. Performance Optimizations

### 8.1 Reduced Lookahead

- Simplified grammar requires only 2-token lookahead
- No context-sensitive keyword handling
- Clear delimiters eliminate ambiguity

### 8.2 Efficient Token Recognition

- Symbol literals handled at lexer level
- Keyword reservation eliminates backtracking
- Consistent syntax patterns reduce parser complexity

### 8.3 Memory Efficiency

- Streamlined AST nodes
- Minimal overhead for position tracking
- Efficient handling of expression trees

---

## 9. Integration with Compiler Pipeline

### 9.1 AST to IR Conversion

```go
func (p *Program) ToIR() (*ir.Module, error) {
	converter := &IRConverter{}
	return converter.Convert(p)
}
```

### 9.2 Type Checking

```go
func (p *Program) TypeCheck() error {
	checker := &TypeChecker{}
	return checker.Check(p)
}
```

### 9.3 Code Generation

```go
func (p *Program) Generate(target string) ([]byte, error) {
	generator := &CodeGenerator{Target: target}
	return generator.Generate(p)
}
```

---

This simplified parser specification eliminates all the ambiguities and complexities of the original grammar while maintaining full language expressiveness. The consistent syntax patterns make the parser much easier to implement and maintain, while the expression-based design provides powerful composability. 