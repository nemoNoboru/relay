# Relay Language Parser Technical Specification

**Version:** 0.3 "Cloudpunks Edition"  
**Parser Library:** Participle v2  
**Language Target:** Go

---

## 1. Overview

This document specifies the technical implementation of the Relay language parser using the Participle parser combinator library for Go. The parser transforms Relay source code into an Abstract Syntax Tree (AST) suitable for compilation or interpretation.

---

## 2. Lexer Specification

The Relay lexer is implemented using Participle's simple lexer with the following token rules:

```go
var relayLexer = lexer.MustSimple([]lexer.SimpleRule{
    // Comments
    {"Comment", `//[^\n]*`},
    {"BlockComment", `/\*([^*]|\*[^/])*\*/`},
    
    // Literals
    {"String", `"(\\"|[^"])*"`},
    {"Number", `[-+]?(\d*\.)?\d+`},
    {"Bool", `\b(true|false)\b`},
    {"DateTime", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`},
    
    // Identifiers and Keywords
    {"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
    
    // Operators and Punctuation
    {"Arrow", `->`},
    {"Assign", `=`},
    {"Eq", `==`},
    {"Ne", `!=`},
    {"Le", `<=`},
    {"Ge", `>=`},
    {"Lt", `<`},
    {"Gt", `>`},
    {"Plus", `\+`},
    {"Minus", `-`},
    {"Multiply", `\*`},
    {"Divide", `/`},
    {"And", `&&`},
    {"Or", `\|\|`},
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
    {"Newline", `\n`},
    {"Whitespace", `[ \t\r]+`},
})
```

### Reserved Keywords

The following keywords are reserved and handled specially during parsing:

- `struct`, `protocol`, `server`, `state`, `receive`, `send`, `template`, `dispatch`
- `config`, `set`, `fn`, `if`, `else`, `for`, `in`, `try`, `catch`, `return`, `throw`
- `implements`, `from`, `optional`, `string`, `number`, `bool`, `datetime`

---

## 3. AST Node Definitions

### 3.1 Root Program Structure

```go
type Program struct {
    Pos lexer.Position
    
    Statements []*Statement `@@*`
}

type Statement struct {
    Pos lexer.Position
    
    StructDef   *StructDef   `  @@`
    ProtocolDef *ProtocolDef `| @@`
    ServerDef   *ServerDef   `| @@`
    ConfigDef   *ConfigDef   `| @@`
    FunctionDef *FunctionDef `| @@`
    Template    *Template    `| @@`
    Expression  *Expression  `| @@`
}
```

### 3.2 Data Type Definitions

```go
type Type struct {
    Pos lexer.Position
    
    Name       string      `@Ident`
    Array      bool        `( "[" "]" @("")?`
    Optional   bool        `| "optional" "(" @("") `
    Object     *ObjectType `| @@`
    Validation *Validation `)?`
    Generic    *Type       `( @@ )?`
}

type ObjectType struct {
    Pos lexer.Position
    
    Fields []*ObjectField `"{" ( @@ ( "," @@ )* )? "}"`
}

type ObjectField struct {
    Pos lexer.Position
    
    Key   string `@Ident`
    Value *Type  `":" @@`
}

type Validation struct {
    Pos lexer.Position
    
    Constraints []*Constraint `( "." @@ )*`
}

type Constraint struct {
    Pos lexer.Position
    
    Name string         `@Ident`
    Args []*Expression  `( "(" ( @@ ( "," @@ )* )? ")" )?`
}
```

### 3.3 Struct Definitions

```go
type StructDef struct {
    Pos lexer.Position
    
    Name   string         `"struct" @Ident`
    Fields []*StructField `"{" ( @@ ( "," @@ )* )? "}"`
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
    Methods []*ProtocolMethod `"{" @@* "}"`
}

type ProtocolMethod struct {
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
    
    Name       string            `"server"`
    ServerName *string           `( @Ident )?`
    Implements *string           `( "implements" @Ident )?`
    Body       *ServerBody       `@@`
}

type ServerBody struct {
    Pos lexer.Position
    
    StateBlock   *StateBlock       `"{" ( @@`
    ReceiveBlocks []*ReceiveBlock  `@@*`
    OtherStatements []*Statement   `@@* ) "}"`
}

type StateBlock struct {
    Pos lexer.Position
    
    Variables []*StateVariable `"state" "{" @@* "}"`
}

type StateVariable struct {
    Pos lexer.Position
    
    Name         string      `@Ident`
    Type         *Type       `":" @@`
    DefaultValue *Expression `( "=" @@ )?`
}

type ReceiveBlock struct {
    Pos lexer.Position
    
    Name       string       `"receive" @Ident`
    Parameters []*Parameter `"{" ( @@ ( "," @@ )* )? "}"`
    ReturnType *Type        `( "->" @@ )?`
    Body       *BlockStmt   `@@`
}
```

### 3.6 Function Definitions

```go
type FunctionDef struct {
    Pos lexer.Position
    
    Name       string       `"fn" @Ident`
    Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
    ReturnType *Type        `( "->" @@ )?`
    Body       *BlockStmt   `@@`
}
```

### 3.7 Template Definitions

```go
type Template struct {
    Pos lexer.Position
    
    Path   string      `"template" @String`
    Method string      `"from" @Ident`
    Params *CallParams `@@`
}
```

### 3.8 Configuration

```go
type ConfigDef struct {
    Pos lexer.Position
    
    Properties []*ConfigProperty `"config" "{" @@* "}"`
}

type ConfigProperty struct {
    Pos lexer.Position
    
    Key   string      `@Ident`
    Value *Expression `":" @@`
}
```

### 3.9 Expressions

```go
type Expression struct {
    Pos lexer.Position
    
    Or []*AndExpr `@@ ( "||" @@ )*`
}

type AndExpr struct {
    Pos lexer.Position
    
    And []*EqualityExpr `@@ ( "&&" @@ )*`
}

type EqualityExpr struct {
    Pos lexer.Position
    
    Left  *ComparisonExpr `@@`
    Right []*OpComparison `@@*`
}

type OpComparison struct {
    Pos lexer.Position
    
    Operator string          `@( "==" | "!=" | "<=" | ">=" | "<" | ">" )`
    Right    *ComparisonExpr `@@`
}

type ComparisonExpr struct {
    Pos lexer.Position
    
    Left  *AdditiveExpr `@@`
    Right []*OpAdditive `@@*`
}

type OpAdditive struct {
    Pos lexer.Position
    
    Operator string        `@( "+" | "-" )`
    Right    *AdditiveExpr `@@`
}

type AdditiveExpr struct {
    Pos lexer.Position
    
    Left  *MultiplicativeExpr `@@`
    Right []*OpMultiplicative `@@*`
}

type OpMultiplicative struct {
    Pos lexer.Position
    
    Operator string               `@( "*" | "/" )`
    Right    *MultiplicativeExpr `@@`
}

type MultiplicativeExpr struct {
    Pos lexer.Position
    
    Left  *UnaryExpr `@@`
    Right []*OpUnary `@@*`
}

type OpUnary struct {
    Pos lexer.Position
    
    Operator string     `@( "!" | "-" )`
    Right    *UnaryExpr `@@`
}

type UnaryExpr struct {
    Pos lexer.Position
    
    Primary *PrimaryExpr `@@`
}

type PrimaryExpr struct {
    Pos lexer.Position
    
    Literal      *Literal      `  @@`
    Identifier   *string       `| @Ident`
    FunctionCall *FunctionCall `| @@`
    MethodCall   *MethodCall   `| @@`
    SendCall     *SendCall     `| @@`
    ArrayAccess  *ArrayAccess  `| @@`
    FieldAccess  *FieldAccess  `| @@`
    ObjectLiteral *ObjectLiteral `| @@`
    ArrayLiteral  *ArrayLiteral  `| @@`
    Lambda        *Lambda        `| @@`
    Grouped       *Expression    `| "(" @@ ")"`
}
```

### 3.10 Literals and Complex Expressions

```go
type Literal struct {
    Pos lexer.Position
    
    String   *string  `  @String`
    Number   *float64 `| @Number`
    Bool     *bool    `| @Bool`
    DateTime *string  `| @DateTime`
}

type FunctionCall struct {
    Pos lexer.Position
    
    Name   string      `@Ident`
    Params *CallParams `@@`
}

type CallParams struct {
    Pos lexer.Position
    
    Params []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type MethodCall struct {
    Pos lexer.Position
    
    Object *PrimaryExpr `@@`
    Method string       `"." @Ident`
    Params *CallParams  `@@`
}

type SendCall struct {
    Pos lexer.Position
    
    Target string      `"send" @Ident`
    Method string      `@Ident`
    Params *CallParams `@@`
}

type ArrayAccess struct {
    Pos lexer.Position
    
    Array *PrimaryExpr `@@`
    Index *Expression  `"[" @@ "]"`
}

type FieldAccess struct {
    Pos lexer.Position
    
    Object *PrimaryExpr `@@`
    Field  string       `"." @Ident`
}

type ObjectLiteral struct {
    Pos lexer.Position
    
    Fields []*ObjectLiteralField `"{" ( @@ ( "," @@ )* )? "}"`
}

type ObjectLiteralField struct {
    Pos lexer.Position
    
    Key   string      `@Ident`
    Value *Expression `":" @@`
}

type ArrayLiteral struct {
    Pos lexer.Position
    
    Elements []*Expression `"[" ( @@ ( "," @@ )* )? "]"`
}

type Lambda struct {
    Pos lexer.Position
    
    Parameters []*Parameter `( @Ident | "(" ( @@ ( "," @@ )* )? ")" )`
    Body       *Expression  `"=>" @@`
}
```

### 3.11 Statements

```go
type Statement struct {
    Pos lexer.Position
    
    VarDecl    *VarDecl    `  @@`
    Assignment *Assignment `| @@`
    IfStmt     *IfStmt     `| @@`
    ForStmt    *ForStmt    `| @@`
    TryStmt    *TryStmt    `| @@`
    ReturnStmt *ReturnStmt `| @@`
    ThrowStmt  *ThrowStmt  `| @@`
    ExprStmt   *Expression `| @@`
    BlockStmt  *BlockStmt  `| @@`
}

type VarDecl struct {
    Pos lexer.Position
    
    Name  string      `"set" @Ident`
    Type  *Type       `( ":" @@ )?`
    Value *Expression `"=" @@`
}

type Assignment struct {
    Pos lexer.Position
    
    Target *Expression `@@`
    Value  *Expression `"=" @@`
}

type BlockStmt struct {
    Pos lexer.Position
    
    Statements []*Statement `"{" @@* "}"`
}

type IfStmt struct {
    Pos lexer.Position
    
    Condition *Expression `"if" @@`
    ThenBlock *BlockStmt  `@@`
    ElseBlock *BlockStmt  `( "else" @@ )?`
}

type ForStmt struct {
    Pos lexer.Position
    
    Variable   string      `"for" @Ident`
    Collection *Expression `"in" @@`
    Body       *BlockStmt  `@@`
}

type TryStmt struct {
    Pos lexer.Position
    
    TryBlock   *BlockStmt `"try" @@`
    CatchVar   *string    `"catch" ( @Ident )?`
    CatchBlock *BlockStmt `@@`
}

type ReturnStmt struct {
    Pos lexer.Position
    
    Value *Expression `"return" ( @@ )?`
}

type ThrowStmt struct {
    Pos lexer.Position
    
    Value *Expression `"throw" @@`
}
```

### 3.12 Dispatch Statements

```go
type DispatchStmt struct {
    Pos lexer.Position
    
    Value *Expression     `"dispatch" @@`
    Cases []*DispatchCase `"{" @@* "}"`
}

type DispatchCase struct {
    Pos lexer.Position
    
    Pattern *Expression `@@`
    Body    *BlockStmt  `"->" @@`
}
```

---

## 4. Parser Configuration

The parser is configured with the following Participle options:

```go
var relayParser = participle.MustBuild[Program](
    participle.Lexer(relayLexer),
    participle.CaseInsensitive("Ident"),
    participle.Unquote("String"),
    participle.UseLookahead(3),
    participle.Elide("Whitespace", "Comment", "BlockComment"),
)
```

### Parser Options Explained

- **Lexer**: Uses the custom Relay lexer defined above
- **CaseInsensitive**: Makes identifier matching case-insensitive
- **Unquote**: Automatically unquotes string literals
- **UseLookahead(3)**: Enables 3-token lookahead for disambiguation
- **Elide**: Removes whitespace and comments from the parse tree

---

## 5. Error Handling

The parser provides detailed error information including:

- **Position information**: Line and column numbers for all syntax errors
- **Context-aware messages**: Specific error messages based on parsing context
- **Recovery strategies**: Partial parsing for better IDE integration

```go
type ParseError struct {
    Pos     lexer.Position
    Message string
    Context string
}
```

---

## 6. Integration Points

### 6.1 Entry Point Function

```go
func Parse(filename string, r io.Reader) (*Program, error) {
    program, err := relayParser.Parse(filename, r)
    if err != nil {
        return nil, &ParseError{
            Message: err.Error(),
            Context: filename,
        }
    }
    
    // Post-processing validation
    if err := validateAST(program); err != nil {
        return nil, err
    }
    
    return program, nil
}
```

### 6.2 AST Validation

Post-parsing validation includes:

- **Type checking**: Validate type annotations and references
- **Symbol resolution**: Ensure all identifiers are properly defined
- **Protocol compliance**: Verify server implementations match protocols
- **State management**: Validate state variable usage

### 6.3 Pretty Printing

The AST supports conversion back to source code for debugging and code generation:

```go
func (p *Program) String() string {
    // Implementation for converting AST back to source
}
```

---

## 7. Testing Strategy

### 7.1 Unit Tests

- **Token-level tests**: Verify lexer correctly tokenizes all language constructs
- **Grammar tests**: Test parsing of individual language constructs
- **Error tests**: Verify appropriate error messages for invalid syntax

### 7.2 Integration Tests

- **Complete programs**: Parse full Relay programs from the spec
- **Edge cases**: Test boundary conditions and complex nesting
- **Performance tests**: Ensure parser handles large programs efficiently

### 7.3 Golden File Tests

- **Round-trip tests**: Parse → Pretty-print → Parse should be identical
- **Regression tests**: Prevent breaking changes to parser output

---

## 8. Performance Considerations

- **Memory usage**: AST nodes include position information but minimize overhead
- **Parse speed**: Optimized grammar rules to minimize backtracking  
- **Error recovery**: Fast failure paths for common syntax errors
- **Incremental parsing**: Future support for IDE integration

---

## 9. Future Extensions

The parser is designed to accommodate future language features:

- **Macro system**: Reserved space in AST for macro definitions
- **Type inference**: Enhanced type annotations for gradual typing
- **Generics**: Type parameter support in protocols and structs
- **Pattern matching**: Extended dispatch statement capabilities

This specification provides the foundation for implementing a robust, efficient parser for the Relay language using the Participle library. 