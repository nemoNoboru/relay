package parser

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the lexer tokenization for struct definition
func TestLexer_StructTokens(t *testing.T) {
	input := `struct User {
 		name: string,
  		email: string,
  		age: number
	}`

	lex, err := relayLexer.Lex("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	if err != nil {
		println(err.Error())
	}

	// Get symbol mappings
	symbols := relayLexer.Symbols()
	whitespaceType := symbols["Whitespace"]
	newlineType := symbols["Newline"]

	// Collect all tokens
	var filteredTokens []lexer.Token
	for {
		token, err := lex.Next()
		if err != nil {
			break
		}
		if token.Type != whitespaceType && token.Type != newlineType {
			filteredTokens = append(filteredTokens, token)
		}
		if token.EOF() {
			break
		}
	}

	expectedTokens := []struct {
		TypeName string
		Value    string
	}{
		{"Ident", "struct"},
		{"Ident", "User"},
		{"LBrace", "{"},
		{"Ident", "name"},
		{"Colon", ":"},
		{"Ident", "string"},
		{"Comma", ","},
		{"Ident", "email"},
		{"Colon", ":"},
		{"Ident", "string"},
		{"Comma", ","},
		{"Ident", "age"},
		{"Colon", ":"},
		{"Ident", "number"},
		{"RBrace", "}"},
		{"EOF", ""},
	}

	require.Len(t, filteredTokens, len(expectedTokens), "Token count mismatch")

	for i, expected := range expectedTokens {
		expectedType := symbols[expected.TypeName]
		assert.Equal(t, expectedType, filteredTokens[i].Type, "Token type mismatch at position %d", i)
		assert.Equal(t, expected.Value, filteredTokens[i].Value, "Token value mismatch at position %d", i)
	}
}

// Test parsing the complete struct definition
func TestParser_StructDefinition(t *testing.T) {
	input := `struct User {
  name: string,
  email: string,
  age: number
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	// Verify we have exactly one statement
	require.Len(t, program.Statements, 1)

	// Verify it's a struct definition
	stmt := program.Statements[0]
	require.NotNil(t, stmt.StructDef)

	structDef := stmt.StructDef

	// Verify struct name
	assert.Equal(t, "User", structDef.Name)

	// Verify fields
	require.Len(t, structDef.Fields, 3)

	// Test first field: name: string
	field1 := structDef.Fields[0]
	assert.Equal(t, "name", field1.Name)
	require.NotNil(t, field1.Type)
	assert.Equal(t, "string", field1.Type.Name)

	// Test second field: email: string
	field2 := structDef.Fields[1]
	assert.Equal(t, "email", field2.Name)
	require.NotNil(t, field2.Type)
	assert.Equal(t, "string", field2.Type.Name)

	// Test third field: age: number
	field3 := structDef.Fields[2]
	assert.Equal(t, "age", field3.Name)
	require.NotNil(t, field3.Type)
	assert.Equal(t, "number", field3.Type.Name)
}

// Test struct with no fields (empty struct)
func TestParser_EmptyStruct(t *testing.T) {
	input := `struct Empty {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	require.Len(t, program.Statements, 1)
	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)

	assert.Equal(t, "Empty", structDef.Name)
	assert.Len(t, structDef.Fields, 0)
}

// Test struct with single field
func TestParser_SingleFieldStruct(t *testing.T) {
	input := `struct Simple {
  id: string
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)

	assert.Equal(t, "Simple", structDef.Name)
	require.Len(t, structDef.Fields, 1)

	field := structDef.Fields[0]
	assert.Equal(t, "id", field.Name)
	assert.Equal(t, "string", field.Type.Name)
}

// Test struct with various data types
func TestParser_StructWithDifferentTypes(t *testing.T) {
	input := `struct Complex {
  name: string,
  age: number,
  active: bool,
  created_at: datetime
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 4)

	expectedFields := []struct {
		name     string
		typeName string
	}{
		{"name", "string"},
		{"age", "number"},
		{"active", "bool"},
		{"created_at", "datetime"},
	}

	for i, expected := range expectedFields {
		field := structDef.Fields[i]
		assert.Equal(t, expected.name, field.Name)
		assert.Equal(t, expected.typeName, field.Type.Name)
	}
}

// Test struct with array types - DISABLED (arrays not implemented yet)
func TestParser_StructWithArrayTypes(t *testing.T) {
	t.Skip("Array types not implemented yet")
}

// Test struct with optional types - DISABLED (optional not implemented yet)
func TestParser_StructWithOptionalTypes(t *testing.T) {
	t.Skip("Optional types not implemented yet")
}

// Test struct with validation constraints - DISABLED (validation not implemented yet)
func TestParser_StructWithValidation(t *testing.T) {
	t.Skip("Validation constraints not implemented yet")
}

// Test struct with object types - DISABLED (object types not fully implemented yet)
func TestParser_StructWithObjectTypes(t *testing.T) {
	t.Skip("Object types not fully implemented yet")
}

// Test error cases
func TestParser_StructErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing struct name",
			input: `struct {}`,
		},
		{
			name:  "missing opening brace",
			input: `struct User name: string }`,
		},
		{
			name:  "missing closing brace",
			input: `struct User { name: string`,
		},
		{
			name:  "missing field type",
			input: `struct User { name: }`,
		},
		{
			name:  "missing colon",
			input: `struct User { name string }`,
		},
		{
			name:  "invalid field name",
			input: `struct User { 123: string }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := relayParser.Parse("test.relay", strings.NewReader(tt.input))
			assert.Error(t, err, "Expected parsing error for: %s", tt.input)
		})
	}
}

// Test multiple structs in one program
func TestParser_MultipleStructs(t *testing.T) {
	input := `struct User {
  name: string,
  email: string,
}

struct Post {
  title: string,
  content: string,
  author: string
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	// Should have two statements
	require.Len(t, program.Statements, 2)

	// First struct: User
	struct1 := program.Statements[0].StructDef
	require.NotNil(t, struct1)
	assert.Equal(t, "User", struct1.Name)
	require.Len(t, struct1.Fields, 2)
	assert.Equal(t, "name", struct1.Fields[0].Name)
	assert.Equal(t, "email", struct1.Fields[1].Name)

	// Second struct: Post
	struct2 := program.Statements[1].StructDef
	require.NotNil(t, struct2)
	assert.Equal(t, "Post", struct2.Name)
	require.Len(t, struct2.Fields, 3)
	assert.Equal(t, "title", struct2.Fields[0].Name)
	assert.Equal(t, "content", struct2.Fields[1].Name)
	assert.Equal(t, "author", struct2.Fields[2].Name)
}

// Test struct with trailing comma - DISABLED (trailing comma not supported yet)
func TestParser_StructWithTrailingComma(t *testing.T) {
	t.Skip("Trailing comma not supported yet")
}

// Test case sensitivity
func TestParser_StructCaseSensitivity(t *testing.T) {
	input := `struct User {
  Name: String,
  EMAIL: STRING
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)

	// Field names should preserve case
	assert.Equal(t, "Name", structDef.Fields[0].Name)
	assert.Equal(t, "EMAIL", structDef.Fields[1].Name)

	// Type names should be case-insensitive (based on parser config)
	// But stored as they appear in source
	assert.Equal(t, "String", structDef.Fields[0].Type.Name)
	assert.Equal(t, "STRING", structDef.Fields[1].Type.Name)
}

// Benchmark test for parsing performance
func BenchmarkParser_StructParsing(b *testing.B) {
	input := `struct User {
  name: string,
  email: string,
  age: number
}`

	for i := 0; i < b.N; i++ {
		_, err := relayParser.Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ===============================
// Protocol Parsing Tests
// ===============================

// Test the lexer tokenization for protocol definition
func TestLexer_ProtocolTokens(t *testing.T) {
	input := `protocol BlogService {
		get_posts() -> [Post]
		create_post(title: string, content: string) -> Post
	}`

	lex, err := relayLexer.Lex("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	// Get symbol mappings
	symbols := relayLexer.Symbols()
	whitespaceType := symbols["Whitespace"]
	newlineType := symbols["Newline"]

	// Collect all tokens
	var filteredTokens []lexer.Token
	for {
		token, err := lex.Next()
		if err != nil {
			break
		}
		if token.Type != whitespaceType && token.Type != newlineType {
			filteredTokens = append(filteredTokens, token)
		}
		if token.EOF() {
			break
		}
	}

	expectedTokens := []struct {
		TypeName string
		Value    string
	}{
		{"Ident", "protocol"},
		{"Ident", "BlogService"},
		{"LBrace", "{"},
		{"Ident", "get_posts"},
		{"LParen", "("},
		{"RParen", ")"},
		{"Arrow", "->"},
		{"LBracket", "["},
		{"Ident", "Post"},
		{"RBracket", "]"},
		{"Ident", "create_post"},
		{"LParen", "("},
		{"Ident", "title"},
		{"Colon", ":"},
		{"Ident", "string"},
		{"Comma", ","},
		{"Ident", "content"},
		{"Colon", ":"},
		{"Ident", "string"},
		{"RParen", ")"},
		{"Arrow", "->"},
		{"Ident", "Post"},
		{"RBrace", "}"},
		{"EOF", ""},
	}

	require.Len(t, filteredTokens, len(expectedTokens), "Token count mismatch")

	for i, expected := range expectedTokens {
		expectedType := symbols[expected.TypeName]
		assert.Equal(t, expectedType, filteredTokens[i].Type, "Token type mismatch at position %d", i)
		assert.Equal(t, expected.Value, filteredTokens[i].Value, "Token value mismatch at position %d", i)
	}
}

// Test parsing basic protocol definition
func TestParser_BasicProtocolDefinition(t *testing.T) {
	input := `protocol BlogService {
		get_posts() -> [Post]
		create_post(title: string, content: string) -> Post
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	// Verify we have exactly one statement
	require.Len(t, program.Statements, 1)

	// Verify it's a protocol definition
	stmt := program.Statements[0]
	require.NotNil(t, stmt.ProtocolDef)

	protocolDef := stmt.ProtocolDef

	// Verify protocol name
	assert.Equal(t, "BlogService", protocolDef.Name)

	// Verify methods
	require.Len(t, protocolDef.Methods, 2)

	// Test first method: get_posts() -> [Post]
	method1 := protocolDef.Methods[0]
	assert.Equal(t, "get_posts", method1.Name)
	assert.Len(t, method1.Parameters, 0) // No parameters
	require.NotNil(t, method1.ReturnType)
	require.NotNil(t, method1.ReturnType.Array)
	assert.Equal(t, "Post", method1.ReturnType.Array.Name)

	// Test second method: create_post(title: string, content: string) -> Post
	method2 := protocolDef.Methods[1]
	assert.Equal(t, "create_post", method2.Name)
	require.Len(t, method2.Parameters, 2)

	// Check parameters
	param1 := method2.Parameters[0]
	assert.Equal(t, "title", param1.Name)
	assert.Equal(t, "string", param1.Type.Name)

	param2 := method2.Parameters[1]
	assert.Equal(t, "content", param2.Name)
	assert.Equal(t, "string", param2.Type.Name)

	// Check return type
	require.NotNil(t, method2.ReturnType)
	assert.Equal(t, "Post", method2.ReturnType.Name)
}

// Test protocol with no methods (empty protocol)
func TestParser_EmptyProtocol(t *testing.T) {
	input := `protocol EmptyService {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	require.Len(t, program.Statements, 1)
	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)

	assert.Equal(t, "EmptyService", protocolDef.Name)
	assert.Len(t, protocolDef.Methods, 0)
}

// Test protocol with single method
func TestParser_SingleMethodProtocol(t *testing.T) {
	input := `protocol SimpleService {
		ping() -> bool
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)

	assert.Equal(t, "SimpleService", protocolDef.Name)
	require.Len(t, protocolDef.Methods, 1)

	method := protocolDef.Methods[0]
	assert.Equal(t, "ping", method.Name)
	assert.Len(t, method.Parameters, 0)
	require.NotNil(t, method.ReturnType)
	assert.Equal(t, "bool", method.ReturnType.Name)
}

// Test protocol with method that has no return type
func TestParser_MethodWithoutReturnType(t *testing.T) {
	input := `protocol ActionService {
		do_something(input: string)
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)
	require.Len(t, protocolDef.Methods, 1)

	method := protocolDef.Methods[0]
	assert.Equal(t, "do_something", method.Name)
	require.Len(t, method.Parameters, 1)

	param := method.Parameters[0]
	assert.Equal(t, "input", param.Name)
	assert.Equal(t, "string", param.Type.Name)

	assert.Nil(t, method.ReturnType) // No return type
}

// Test protocol with various parameter types
func TestParser_ProtocolWithDifferentTypes(t *testing.T) {
	input := `protocol ComplexService {
		get_user(id: string) -> User
		update_stats(count: number, active: bool) -> Stats
		process_data(timestamp: datetime) -> bool
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)
	require.Len(t, protocolDef.Methods, 3)

	// Test get_user method
	method1 := protocolDef.Methods[0]
	assert.Equal(t, "get_user", method1.Name)
	require.Len(t, method1.Parameters, 1)
	assert.Equal(t, "id", method1.Parameters[0].Name)
	assert.Equal(t, "string", method1.Parameters[0].Type.Name)
	assert.Equal(t, "User", method1.ReturnType.Name)

	// Test update_stats method
	method2 := protocolDef.Methods[1]
	assert.Equal(t, "update_stats", method2.Name)
	require.Len(t, method2.Parameters, 2)
	assert.Equal(t, "count", method2.Parameters[0].Name)
	assert.Equal(t, "number", method2.Parameters[0].Type.Name)
	assert.Equal(t, "active", method2.Parameters[1].Name)
	assert.Equal(t, "bool", method2.Parameters[1].Type.Name)
	assert.Equal(t, "Stats", method2.ReturnType.Name)

	// Test process_data method
	method3 := protocolDef.Methods[2]
	assert.Equal(t, "process_data", method3.Name)
	require.Len(t, method3.Parameters, 1)
	assert.Equal(t, "timestamp", method3.Parameters[0].Name)
	assert.Equal(t, "datetime", method3.Parameters[0].Type.Name)
	assert.Equal(t, "bool", method3.ReturnType.Name)
}

// Test protocol with array return types - DISABLED (arrays not fully implemented yet)
func TestParser_ProtocolWithArrayTypes(t *testing.T) {
	t.Skip("Array types not fully implemented yet")

	input := `protocol DataService {
		get_items() -> [Item]
		get_tags() -> [string]
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)
	require.Len(t, protocolDef.Methods, 2)

	// Test array return types
	method1 := protocolDef.Methods[0]
	assert.Equal(t, "get_items", method1.Name)
	require.NotNil(t, method1.ReturnType.Array)
	assert.Equal(t, "Item", method1.ReturnType.Array.Name)

	method2 := protocolDef.Methods[1]
	assert.Equal(t, "get_tags", method2.Name)
	require.NotNil(t, method2.ReturnType.Array)
	assert.Equal(t, "string", method2.ReturnType.Array.Name)
}

// Test multiple protocols in one file
func TestParser_MultipleProtocols(t *testing.T) {
	input := `protocol UserService {
		get_user(id: string) -> User
	}
	
	protocol BlogService {
		get_posts() -> [Post]
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	require.Len(t, program.Statements, 2)

	// First protocol
	protocol1 := program.Statements[0].ProtocolDef
	require.NotNil(t, protocol1)
	assert.Equal(t, "UserService", protocol1.Name)
	require.Len(t, protocol1.Methods, 1)
	assert.Equal(t, "get_user", protocol1.Methods[0].Name)

	// Second protocol
	protocol2 := program.Statements[1].ProtocolDef
	require.NotNil(t, protocol2)
	assert.Equal(t, "BlogService", protocol2.Name)
	require.Len(t, protocol2.Methods, 1)
	assert.Equal(t, "get_posts", protocol2.Methods[0].Name)
}

// Test protocol case sensitivity
func TestParser_ProtocolCaseSensitivity(t *testing.T) {
	input := `protocol UserService {
		GetUser(ID: string) -> User
		getUser(id: string) -> User
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)
	require.Len(t, protocolDef.Methods, 2)

	// Methods should maintain their original case
	method1 := protocolDef.Methods[0]
	assert.Equal(t, "GetUser", method1.Name)
	assert.Equal(t, "ID", method1.Parameters[0].Name)

	method2 := protocolDef.Methods[1]
	assert.Equal(t, "getUser", method2.Name)
	assert.Equal(t, "id", method2.Parameters[0].Name)
}

// Test mixed struct and protocol definitions
func TestParser_MixedStructAndProtocol(t *testing.T) {
	input := `struct User {
		name: string,
		email: string
	}
	
	protocol UserService {
		create_user(user: User) -> User
		get_user(id: string) -> User
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 2)

	// First should be struct
	structStmt := program.Statements[0]
	require.NotNil(t, structStmt.StructDef)
	assert.Equal(t, "User", structStmt.StructDef.Name)

	// Second should be protocol
	protocolStmt := program.Statements[1]
	require.NotNil(t, protocolStmt.ProtocolDef)
	assert.Equal(t, "UserService", protocolStmt.ProtocolDef.Name)
}

// Benchmark protocol parsing performance
func BenchmarkParser_ProtocolParsing(b *testing.B) {
	input := `protocol BlogService {
		get_posts() -> [Post]
		create_post(title: string, content: string) -> Post
		get_post(id: string) -> Post
		delete_post(id: string) -> bool
	}`

	for i := 0; i < b.N; i++ {
		_, err := relayParser.Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
