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

// Statement represents a top-level statement
type Statement struct {
	Pos lexer.Position

	StructDef *StructDef `@@`
}

// Type represents a data type - simplified to avoid recursion
type Type struct {
	Pos lexer.Position

	// Basic type: just the name for now
	Name string `@Ident`
}

// ObjectType represents an object type like {key: string}
type ObjectType struct {
	Pos lexer.Position

	Fields []*ObjectField `"{" ( @@ ( "," @@ )* )? "}"`
}

// ObjectField represents a field in an object type
type ObjectField struct {
	Pos lexer.Position

	Key   string `@Ident`
	Value *Type  `":" @@`
}

// Validation represents type validation constraints
type Validation struct {
	Pos lexer.Position

	Constraints []*Constraint `( "." @@ )*`
}

// Constraint represents a single validation constraint
type Constraint struct {
	Pos lexer.Position

	Name string     `@Ident`
	Args []*Literal `( "(" ( @@ ( "," @@ )* )? ")" )?`
}

// StructDef represents a struct definition
type StructDef struct {
	Pos lexer.Position

	Name   string         `"struct" @Ident`
	Fields []*StructField `"{" ( @@ ( "," @@ )* )? "}"`
}

// StructField represents a field in a struct
type StructField struct {
	Pos lexer.Position

	Name string `@Ident`
	Type *Type  `":" @@`
}

// Literal represents a literal value - simple and non-recursive
type Literal struct {
	Pos lexer.Position

	String   *string  `  @String`
	Number   *float64 `| @Number`
	Bool     *bool    `| @Bool`
	DateTime *string  `| @DateTime`
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
