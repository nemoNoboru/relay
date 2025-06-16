package parser

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Lexer definition
var relayLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Comments
	{"Comment", `//[^\n]*`},
	{"BlockComment", `/\*([^*]|\*[^/])*\*/`},

	// Literals
	{"String", `"(\\"|[^"])*"`},
	{"Number", `[-+]?(\d*\.)?\d+`},
	{"Bool", `\b(true|false)\b`},
	{"DateTime", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`},

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
	{"State", `\bstate\b`},
	{"Receive", `\breceive\b`},

	// Identifiers and Keywords
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},

	// Operators and Punctuation (compound operators first)
	{"Arrow", `->`},
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

// AST Node Definitions - MINIMAL VERSION

// Program represents the root of a Relay program
type Program struct {
	Pos lexer.Position

	Statements []*Statement `@@*`
}

type Statement struct {
	Pos lexer.Position

	StructDef   *StructDef   `@@`
	ProtocolDef *ProtocolDef `| @@`
	StateDef    *StateDef    `| @@`
	ReceiveDef  *ReceiveDef  `| @@`
}

type StructDef struct {
	Name   string   `"struct" @Ident`
	Fields []*Field `"{" @@* "}"`
}

type ProtocolDef struct {
	Name    string             `"protocol" @Ident`
	Methods []*MethodSignature `"{" @@* "}"`
}

type StateDef struct {
	Fields []*StateField `"state" "{" @@* "}"`
}

type StateField struct {
	Name         string   `@Ident`
	Type         *TypeRef `":" @@`
	DefaultValue *Literal `( "=" @@ )?`
	Comma        string   `(",")?`
}

type Literal struct {
	String   *string       `@String`
	Number   *float64      `| @Number`
	Bool     *string       `| @Bool`
	Array    *ArrayLiteral `| @@`
	FuncCall *FuncCall     `| @@`
}

// GetBoolValue returns the boolean value from a Bool string
func (l *Literal) GetBoolValue() *bool {
	if l.Bool == nil {
		return nil
	}
	val := *l.Bool == "true"
	return &val
}

type ArrayLiteral struct {
	Elements []*Literal `"[" ( @@ ( "," @@ )* )? "]"`
}

type FuncCall struct {
	Name string     `@Ident`
	Args []*Literal `"(" ( @@ ( "," @@ )* )? ")"`
}

type MethodSignature struct {
	Name       string       `@Ident`
	Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
	ReturnType *TypeRef     `( "->" @@ )?`
}

type Parameter struct {
	Name string   `@Ident`
	Type *TypeRef `":" @@`
}

type Field struct {
	Name string `@Ident`
	// Arguments  []*Argument `( "(" ( @@ ( "," @@ )* )? ")" )?`
	Type  *TypeRef `":" @@`
	Comma string   `(",")?`
}

type TypeRef struct {
	Array *TypeRef `(   "[" @@ "]"`
	Name  string   `  | @Ident )`
}

type ReceiveDef struct {
	Name       string       `"receive" @Ident`
	Parameters []*Parameter `"{" ( @@ ( "," @@ )* )? "}"`
	ReturnType *TypeRef     `( "->" @@ )?`
	Body       *Block       `@@`
}

type Block struct {
	Statements []*BlockStatement `"{" @@* "}"`
}

type BlockStatement struct {
	Pos lexer.Position

	ForStatement      *ForStatement      `@@`
	TryStatement      *TryStatement      `| @@`
	DispatchStatement *DispatchStatement `| @@`
	SetStatement      *SetStatement      `| @@`
	ReturnStatement   *ReturnStatement   `| @@`
	IfStatement       *IfStatement       `| @@`
	ThrowStatement    *ThrowStatement    `| @@`
	Assignment        *Assignment        `| @@`
	ExprStatement     *ExprStatement     `| @@`
}

type SetStatement struct {
	Variable string      `"set" @Ident`
	Value    *Expression `"=" @@`
}

type ReturnStatement struct {
	Value *Expression `"return" @@`
}

type IfStatement struct {
	Condition *Expression `"if" @@`
	ThenBlock *Block      `@@`
	ElseBlock *Block      `( "else" @@ )?`
}

type ThrowStatement struct {
	Value *Expression `"throw" @@`
}

type Assignment struct {
	Target *Expression `@@`
	Op     string      `@( "=" | "+=" | "-=" | "*=" | "/=" )`
	Value  *Expression `@@`
}

type ForStatement struct {
	Variable   string      `"for" @Ident`
	Collection *Expression `"in" @@`
	Body       *Block      `@@`
}

type TryStatement struct {
	TryBlock   *Block  `"try" @@`
	CatchVar   *string `"catch" ( @Ident )?`
	CatchBlock *Block  `@@`
}

type DispatchStatement struct {
	Value *Expression     `"dispatch" @@`
	Cases []*DispatchCase `"{" @@* "}"`
}

type DispatchCase struct {
	Pattern *Literal `@@`
	Body    *Block   `"->" @@`
}

type SendExpr struct {
	Target string     `"send" @String`
	Method string     `@Ident`
	Args   *ObjectLit `@@`
}

type FuncCallExpr struct {
	Name string        `@Ident`
	Args []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type ExprStatement struct {
	Expression *Expression `@@`
}

type Expression struct {
	Logical *LogicalExpr `@@`
}

type LogicalExpr struct {
	Left  *EqualityExpr `@@`
	Right []*LogicalOp  `@@*`
}

type LogicalOp struct {
	Op    string        `@( "&&" | "||" )`
	Right *EqualityExpr `@@`
}

type EqualityExpr struct {
	Left  *RelationalExpr `@@`
	Right []*EqualityOp   `@@*`
}

type EqualityOp struct {
	Op    string          `@( "==" | "!=" )`
	Right *RelationalExpr `@@`
}

type RelationalExpr struct {
	Left  *AdditiveExpr   `@@`
	Right []*RelationalOp `@@*`
}

type RelationalOp struct {
	Op    string        `@( "<=" | ">=" | "<" | ">" )`
	Right *AdditiveExpr `@@`
}

type AdditiveExpr struct {
	Left  *MultiplicativeExpr `@@`
	Right []*AdditiveOp       `@@*`
}

type AdditiveOp struct {
	Op    string              `@( "+" | "-" )`
	Right *MultiplicativeExpr `@@`
}

type MultiplicativeExpr struct {
	Left  *UnaryExpr          `@@`
	Right []*MultiplicativeOp `@@*`
}

type MultiplicativeOp struct {
	Op    string     `@( "*" | "/" )`
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

type BaseExpr struct {
	Literal    *Literal      `@@`
	Identifier *string       `| @Ident`
	ObjectLit  *ObjectLit    `| @@`
	SendExpr   *SendExpr     `| @@`
	FuncCall   *FuncCallExpr `| @@`
	Grouped    *Expression   `| "(" @@ ")"`
}

type AccessExpr struct {
	FieldAccess *string     `"." @Ident`
	MethodCall  *MethodCall `| @@`
}

type MethodCall struct {
	Method string        `"." @Ident`
	Args   []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type ObjectLit struct {
	Fields []*ObjectField `"{" ( @@ ( "," @@ )* )? "}"`
}

type ObjectField struct {
	Key   string      `@Ident`
	Value *Expression `":" @@`
}

// Parser configuration
var relayParser = participle.MustBuild[Program](
	participle.Lexer(relayLexer),
	participle.CaseInsensitive("Ident"),
	participle.Unquote("String"),
	participle.UseLookahead(2),
	participle.Elide("whitespace", "Comment", "BlockComment"),
)

// Parse function to parse Relay source code
func Parse(filename string, r io.Reader) (*Program, error) {
	program, err := relayParser.Parse(filename, r)
	if err != nil {
		return nil, err
	}
	return program, nil
}
