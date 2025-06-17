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

// AST Node Definitions - MINIMAL VERSION

// Program represents the root of a Relay program
type Program struct {
	Pos lexer.Position

	Expressions []*Expression `@@*`
}

type StructExpr struct {
	Name   string   `"struct" @Ident`
	Fields []*Field `"{" @@* "}"`
}

type ProtocolExpr struct {
	Name    string             `"protocol" @Ident`
	Methods []*MethodSignature `"{" @@* "}"`
}

type ServerExpr struct {
	Name     string      `"server" @Ident`
	Protocol *string     `( "implements" @Ident )?`
	Body     *ServerBody `@@`
}

type ServerBody struct {
	Elements []*ServerElement `"{" @@* "}"`
}

type ServerElement struct {
	State   *StateExpr   `@@`
	Receive *ReceiveExpr `| @@`
}

type StateExpr struct {
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
	Symbol   *string       `| @Symbol`
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
	Type *TypeRef `( ":" @@ )?`
}

type Field struct {
	Name string `@Ident`
	// Arguments  []*Argument `( "(" ( @@ ( "," @@ )* )? ")" )?`
	Type  *TypeRef `":" @@`
	Comma string   `(",")?`
}

type TypeRef struct {
	Array         *TypeRef           `(   "[" @@ "]"`
	Function      *FunctionType      `| @@`
	Parameterized *ParameterizedType `| @@`
	Name          string             `| @Ident )`
}

type ParameterizedType struct {
	Name string     `@Ident`
	Args []*TypeRef `"(" ( @@ ( "," @@ )* )? ")"`
}

type FunctionType struct {
	Parameters []*Parameter `"fn" "(" ( @@ ( "," @@ )* )? ")"`
	ReturnType *TypeRef     `( "->" @@ )?`
}

type ReceiveExpr struct {
	Name       string       `"receive" "fn" @Ident`
	Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
	ReturnType *TypeRef     `( "->" @@ )?`
	Body       *Block       `@@`
}

type FunctionExpr struct {
	Name       *string      `"fn" ( @Ident )?`
	Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
	ReturnType *TypeRef     `( "->" @@ )?`
	Body       *Block       `@@`
}

type TemplateExpr struct {
	Path     string       `"template" @String`
	FromFunc *FuncRefExpr `"from" @@`
}

type FuncRefExpr struct {
	Name       string       `@Ident`
	Parameters []*Parameter `"(" ( @@ ( "," @@ )* )? ")"`
}

type ConfigExpr struct {
	Fields []*ObjectField `"config" "{" ( @@ ( "," @@ )* )? "}"`
}

type LambdaExpr struct {
	Parameters []*Parameter `"fn" "(" ( @@ ( "," @@ )* )? ")"`
	ReturnType *TypeRef     `( "->" @@ )?`
	Body       *Block       `@@`
}

type Block struct {
	Expressions []*Expression `"{" @@* "}"`
}

type SetExpr struct {
	Variable string      `"set" @Ident`
	Value    *Expression `"=" @@`
}

type ReturnExpr struct {
	Value *Expression `"return" @@`
}

type IfExpr struct {
	Condition *Expression `"if" @@`
	ThenBlock *Block      `@@`
	ElseBlock *Block      `( "else" @@ )?`
}

type ThrowExpr struct {
	Value *Expression `"throw" @@`
}

type ForExpr struct {
	Variable   string      `"for" @Ident`
	Collection *Expression `"in" @@`
	Body       *Block      `@@`
}

type TryExpr struct {
	TryBlock   *Block  `"try" @@`
	CatchVar   *string `"catch" ( @Ident )?`
	CatchBlock *Block  `@@`
}

type DispatchExpr struct {
	Value *Expression     `"dispatch" @@`
	Cases []*DispatchCase `"{" ( @@ ( "," @@ )* )? "}"`
}

type DispatchCase struct {
	Pattern *Literal    `@@`
	Body    *Expression `":" @@`
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
	Logical      *LogicalExpr  `| @@`
}

type LogicalExpr struct {
	Left  *NullCoalesceExpr `@@`
	Right []*LogicalOp      `@@*`
}

type LogicalOp struct {
	Op    string            `@( "&&" | "||" )`
	Right *NullCoalesceExpr `@@`
}

type NullCoalesceExpr struct {
	Left  *EqualityExpr     `@@`
	Right []*NullCoalesceOp `@@*`
}

type NullCoalesceOp struct {
	Op    string        `@"??"`
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
	Literal           *Literal           `@@`
	Identifier        *string            `| @( Ident | "state" | "send" | "receive" | "protocol" | "struct" | "for" | "in" | "try" | "catch" | "dispatch" | "set" | "return" | "if" | "else" | "throw" | "server" | "implements" )`
	StructConstructor *StructConstructor `| @@`
	ObjectLit         *ObjectLit         `| @@`
	SendExpr          *SendExpr          `| @@`
	Lambda            *LambdaExpr        `| @@`
	FuncCall          *FuncCallExpr      `| @@`
	Block             *Block             `| @@`
	Grouped           *Expression        `| "(" @@ ")"`
}

type StructConstructor struct {
	Name   string         `@Ident`
	Fields []*ObjectField `"{" ( @@ ( "," @@ )* )? "}"`
}

type AccessExpr struct {
	MethodCall *MethodCall `@@`
}

type MethodCall struct {
	Method string        `"." @( Ident | "set" | "get" | "add" | "remove" | "filter" | "map" | "find" | "sort_by" | "reduce" )`
	Args   []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type ObjectLit struct {
	Fields []*ObjectField `"{" ( @@ ( "," @@ )* )? "}"`
}

type ObjectField struct {
	Key   string      `@( Ident | Symbol )`
	Value *Expression `":" @@`
}

// Parser configuration
var relayParser = participle.MustBuild[Program](
	participle.Lexer(relayLexer),
	participle.CaseInsensitive("Ident"),
	participle.Unquote("String"),
	participle.UseLookahead(4),
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
