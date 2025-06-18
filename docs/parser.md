# Relay Language Parser Technical Specification

**Version:** 0.3 "Cloudpunks Edition"  
**Parser Library:** Participle v2  
**Language Target:** Go  
**Status:** Simplified Grammar Implementation

---

## 1. Overview

This document specifies the complete technical implementation of the simplified Relay language parser using the Participle parser combinator library for Go. The parser transforms the streamlined Relay source code into an Abstract Syntax Tree (AST) suitable for compilation or interpretation.

**Key Simplifications Implemented:**
- Consistent `fn` syntax for all functions and lambdas
- Unified field access via `.get("field")` method calls only
- Symbol literals `:identifier` as string shorthand
- Expression-based design with blocks as expressions
- Method-based state updates via `.set("field", value)`
- `receive fn` pattern for server methods
- Strict keyword reservation eliminates ambiguity

---

## 2. Complete Lexer Specification

The simplified Relay lexer uses comprehensive token rules with no context sensitivity:

```go
var relayLexer = lexer.MustSimple([]lexer.SimpleRule{
    // Comments
    {"Comment", `//[^\n]*`},
    {"BlockComment", `/\*([^*]|\*[^/])*\*/`},
    
    // Literals
    {"String", `"(\\"|[^"])*"`},
    {"Symbol", `:[a-zA-Z_][a-zA-Z0-9_]*`},
    {"Number", `[-+]?(\d*\.)?\d+([eE][-+]?\d+)?`},
    {"Bool", `\b(true|false)\b`},
    {"DateTime", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`},
    
    // Strictly Reserved Keywords (MUST come before Ident)
    {"Struct", `\bstruct\b`},
    {"Protocol", `\bprotocol\b`},
    {"Server", `\bserver\b`},
    {"State", `\bstate\b`},
    {"Receive", `\breceive\b`},
    {"Send", `\bsend\b`},
    {"Template", `\btemplate\b`},
    {"Dispatch", `\bdispatch\b`},
    {"Config", `\bconfig\b`},
    {"Fn", `\bfn\b`},
    {"Set", `\bset\b`},
    {"If", `\bif\b`},
    {"Else", `\belse\b`},
    {"For", `\bfor\b`},
    {"In", `\bin\b`},

    {"Return", `\breturn\b`},
    {"Throw", `\bthrow\b`},
    {"From", `\bfrom\b`},
    {"Optional", `\boptional\b`},
    
    // Type Keywords
    {"String_Type", `\bstring\b`},
    {"Number_Type", `\bnumber\b`},
    {"Bool_Type", `\bbool\b`},
    {"DateTime_Type", `\bdatetime\b`},
    {"Object_Type", `\bobject\b`},
    
    // Identifiers (after all keywords)
    {"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
    
    // Operators (order matters - longest first)
    {"Arrow", `->`},
    {"NullCoalesce", `\?\?`},
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
    
    // Whitespace
    {"Whitespace", `[ \t\r\n]+`},
})
```

---

## 3. Simplified AST Node Definitions

### 3.1 Root Program Structure

```go
type Program struct {
    Pos lexer.Position
    
    Statements []*TopLevelStatement `@@*`
}

type TopLevelStatement struct {
    Pos lexer.Position
    
    StructDef   *StructDef   `  @@`
    ProtocolDef *ProtocolDef `| @@`
    ServerDef   *ServerDef   `| @@`
    ConfigDef   *ConfigDef   `| @@`
    FunctionDef *FunctionDef `| @@`
    Template    *Template    `| @@`
}
```

### 3.2 Simplified Type System

```go
type Type struct {
    Pos lexer.Position
    
    BasicType    *BasicType    `  @@`
    ArrayType    *ArrayType    `| @@`
    ObjectType   *ObjectType   `| @@`
    OptionalType *OptionalType `| @@`
    FunctionType *FunctionType `| @@`
}

type BasicType struct {
    Pos lexer.Position
    
    Name string `@( "string" | "number" | "bool" | "datetime" | "object" | Ident )`
}

type ArrayType struct {
    Pos lexer.Position
    
    ElementType *Type `"[" @@ "]"`
}

type ObjectType struct {
    Pos lexer.Position
    
    KeyType   *Type `"{" @@`
    ValueType *Type `":" @@ "}"`
}

type OptionalType struct {
    Pos lexer.Position
    
    InnerType *Type `"optional" "(" @@ ")"`
}

type FunctionType struct {
    Pos lexer.Position
    
    Parameters []*Type `"fn" "(" ( @@ ( "," @@ )* )? ")"`
    ReturnType *Type   `( "->" @@ )?`
}
```

### 3.3 Struct Definitions

```go
type StructDef struct {
    Pos lexer.Position
    
    Name   string         `"struct" @Ident`
    Fields []*StructField `"{" ( @@ ( "," @@ )* ","? )? "}"`
}

type StructField struct {
    Pos lexer.Position
    
    Name string `@Ident`
    Type *Type  `":" @@`
}
```

### 3.4 Protocol Definitions

```go
type ProtocolDef struct {
    Pos lexer.Position
    
    Name    string           `"protocol" @Ident`
    Methods []*MethodSignature `"{" @@* "}"`
}

type MethodSignature struct {
    Pos lexer.Position
    
    Name       string       `@Ident`
    Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
    ReturnType *Type        `( "->" @@ )?`
}

type Parameter struct {
    Pos lexer.Position
    
    Name string `@Ident`
    Type *Type  `":" @@`
}
```

### 3.5 Server Definitions

```go
type ServerDef struct {
    Pos lexer.Position
    
    Name string      `"server" @Ident`
    Body *ServerBody `@@`
}

type ServerBody struct {
    Pos lexer.Position
    
    StateBlock    *StateBlock     `"{" ( @@`
    ReceiveBlocks []*ReceiveBlock `@@* ) "}"`
}

type StateBlock struct {
    Pos lexer.Position
    
    Variables []*StateVariable `"state" "{" ( @@ ( "," @@ )* ","? )? "}"`
}

type StateVariable struct {
    Pos lexer.Position
    
    Name         string      `@Ident`
    Type         *Type       `":" @@`
    DefaultValue *Expression `( "=" @@ )?`
}

type ReceiveBlock struct {
    Pos lexer.Position
    
    Function *FunctionDef `"receive" @@`
}
```

### 3.6 Function Definitions

```go
type FunctionDef struct {
    Pos lexer.Position
    
    Name       string       `"fn" @Ident`
    Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
    ReturnType *Type        `( "->" @@ )?`
    Body       *BlockExpr   `@@`
}

type Lambda struct {
    Pos lexer.Position
    
    Parameters []*Parameter `"fn" "(" ( @@ ( "," @@ )* )? ")"`
    Body       *BlockExpr   `@@`
}
```

### 3.7 Template Definitions

```go
type Template struct {
    Pos lexer.Position
    
    Path     string            `"template" @String`
    Function *TemplateFuncCall `"from" @@`
}

type TemplateFuncCall struct {
    Pos lexer.Position
    
    Name       string       `@Ident`
    Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
}
```

### 3.8 Configuration Definitions

```go
type ConfigDef struct {
    Pos lexer.Position
    
    Properties []*ConfigProperty `"config" "{" ( @@ ( "," @@ )* ","? )? "}"`
}

type ConfigProperty struct {
    Pos lexer.Position
    
    Key   string      `@Ident`
    Value *Expression `":" @@`
}
```

### 3.9 Expression System

```go
type Expression struct {
    Pos lexer.Position
    
    Logical *LogicalOrExpr `@@`
}

type LogicalOrExpr struct {
    Pos lexer.Position
    
    Left  *LogicalAndExpr `@@`
    Right []*LogicalOrOp  `@@*`
}

type LogicalOrOp struct {
    Pos lexer.Position
    
    Right *LogicalAndExpr `"||" @@`
}

type LogicalAndExpr struct {
    Pos lexer.Position
    
    Left  *EqualityExpr   `@@`
    Right []*LogicalAndOp `@@*`
}

type LogicalAndOp struct {
    Pos lexer.Position
    
    Right *EqualityExpr `"&&" @@`
}

type EqualityExpr struct {
    Pos lexer.Position
    
    Left  *RelationalExpr `@@`
    Right []*EqualityOp   `@@*`
}

type EqualityOp struct {
    Pos lexer.Position
    
    Operator string          `@( "==" | "!=" )`
    Right    *RelationalExpr `@@`
}

type RelationalExpr struct {
    Pos lexer.Position
    
    Left  *AdditiveExpr   `@@`
    Right []*RelationalOp `@@*`
}

type RelationalOp struct {
    Pos lexer.Position
    
    Operator string        `@( "<=" | ">=" | "<" | ">" )`
    Right    *AdditiveExpr `@@`
}

type AdditiveExpr struct {
    Pos lexer.Position
    
    Left  *MultiplicativeExpr `@@`
    Right []*AdditiveOp       `@@*`
}

type AdditiveOp struct {
    Pos lexer.Position
    
    Operator string              `@( "+" | "-" )`
    Right    *MultiplicativeExpr `@@`
}

type MultiplicativeExpr struct {
    Pos lexer.Position
    
    Left  *UnaryExpr          `@@`
    Right []*MultiplicativeOp `@@*`
}

type MultiplicativeOp struct {
    Pos lexer.Position
    
    Operator string     `@( "*" | "/" )`
    Right    *UnaryExpr `@@`
}

type UnaryExpr struct {
    Pos lexer.Position
    
    Operator *string      `@( "!" | "-" )?`
    Primary  *PrimaryExpr `@@`
}

type PrimaryExpr struct {
    Pos lexer.Position
    
    Base     *BaseExpr     `@@`
    Accesses []*AccessExpr `@@*`
}

type BaseExpr struct {
    Pos lexer.Position
    
    Literal       *Literal       `  @@`
    Identifier    *string        `| @Ident`
    SendExpr      *SendExpr      `| @@`
    DispatchExpr  *DispatchExpr  `| @@`
    ObjectLiteral *ObjectLiteral `| @@`
    ArrayLiteral  *ArrayLiteral  `| @@`
    Lambda        *Lambda        `| @@`
    BlockExpr     *BlockExpr     `| @@`
    Grouped       *Expression    `| "(" @@ ")"`
}

type AccessExpr struct {
    Pos lexer.Position
    
    MethodCall  *MethodCall  `@@`
    ArrayAccess *ArrayAccess `| "[" @@ "]"`
}

type MethodCall struct {
    Pos lexer.Position
    
    Method string        `"." @Ident`
    Args   []*Expression `( "(" ( @@ ( "," @@ )* )? ")" )`
}
```

### 3.10 Literals and Complex Expressions

```go
type Literal struct {
    Pos lexer.Position
    
    String   *string  `  @String`
    Symbol   *string  `| @Symbol`
    Number   *float64 `| @Number`
    Bool     *bool    `| @Bool`
    DateTime *string  `| @DateTime`
}

type SendExpr struct {
    Pos lexer.Position
    
    Target string          `"send" @String`
    Method string          `@Ident`
    Args   *ObjectLiteral  `@@`
}

type DispatchExpr struct {
    Pos lexer.Position
    
    Value *Expression     `"dispatch" @@`
    Cases []*DispatchCase `"{" ( @@ ( "," @@ )* ","? )? "}"`
}

type DispatchCase struct {
    Pos lexer.Position
    
    Pattern *Literal `@@`
    Handler *Lambda  `":" @@`
}

type ObjectLiteral struct {
    Pos lexer.Position
    
    Fields []*ObjectField `"{" ( @@ ( "," @@ )* ","? )? "}"`
}

type ObjectField struct {
    Pos lexer.Position
    
    Key   string      `@( Ident | Symbol )`
    Value *Expression `":" @@`
}

type ArrayLiteral struct {
    Pos lexer.Position
    
    Elements []*Expression `"[" ( @@ ( "," @@ )* ","? )? "]"`
}
```

### 3.11 Block Expressions and Statements

```go
type BlockExpr struct {
    Pos lexer.Position
    
    Statements []*Statement   `"{" @@*`
    Result     *Expression    `( @@ )? "}"`
}

type Statement struct {
    Pos lexer.Position
    
    SetStatement    *SetStatement    `  @@`
    IfStatement     *IfStatement     `| @@`
    ForStatement    *ForStatement    `| @@`
    TryStatement    *TryStatement    `| @@`
    ReturnStatement *ReturnStatement `| @@`
    ThrowStatement  *ThrowStatement  `| @@`
    ExpressionStmt  *Expression      `| @@`
}

type SetStatement struct {
    Pos lexer.Position
    
    Name  string      `"set" @Ident`
    Type  *Type       `( ":" @@ )?`
    Value *Expression `"=" @@`
}

type IfStatement struct {
    Pos lexer.Position
    
    Condition *Expression `"if" @@`
    ThenBlock *BlockExpr  `@@`
    ElseBlock *BlockExpr  `( "else" @@ )?`
}

type ForStatement struct {
    Pos lexer.Position
    
    Variable   string      `"for" @Ident`
    Collection *Expression `"in" @@`
    Body       *BlockExpr  `@@`
}



type ReturnStatement struct {
    Pos lexer.Position
    
    Value *Expression `"return" @@`
}

type ThrowStatement struct {
    Pos lexer.Position
    
    Value *Expression `"throw" @@`
}
```

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