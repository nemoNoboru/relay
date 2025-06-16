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
}

type StructDef struct {
	Name   string   `"struct" @Ident`
	Fields []*Field `"{" @@* "}"`
}

type ProtocolDef struct {
	Name    string             `"protocol" @Ident`
	Methods []*MethodSignature `"{" @@* "}"`
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

// Parser configuration
var relayParser = participle.MustBuild[Program](
	participle.Lexer(relayLexer),
	participle.CaseInsensitive("Ident"),
	participle.Unquote("String"),
	participle.UseLookahead(2),
	participle.Elide("Whitespace", "Newline", "Comment", "BlockComment"),
)

// Parse function to parse Relay source code
func Parse(filename string, r io.Reader) (*Program, error) {
	program, err := relayParser.Parse(filename, r)
	if err != nil {
		return nil, err
	}
	return program, nil
}
