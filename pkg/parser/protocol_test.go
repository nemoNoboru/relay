package parser

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// PROTOCOL PARSING TESTS
// =============================================================================

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
func TestParser_ProtocolDefinition(t *testing.T) {
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
func TestParser_ProtocolEmpty(t *testing.T) {
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
func TestParser_ProtocolSingleMethod(t *testing.T) {
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
func TestParser_ProtocolNoReturnType(t *testing.T) {
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
func TestParser_ProtocolBasicTypes(t *testing.T) {
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

// Test protocol with array types
func TestParser_ProtocolArrayTypes(t *testing.T) {
	input := `protocol DataService {
		get_items() -> [Item]
		get_tags() -> [string]
		get_matrix() -> [[number]]
		process_list(items: [string]) -> [Item]
		bulk_create(users: [User], posts: [Post]) -> [Result]
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	protocolDef := program.Statements[0].ProtocolDef
	require.NotNil(t, protocolDef)
	require.Len(t, protocolDef.Methods, 5)

	// Test get_items() -> [Item]
	method1 := protocolDef.Methods[0]
	assert.Equal(t, "get_items", method1.Name)
	assert.Len(t, method1.Parameters, 0)
	require.NotNil(t, method1.ReturnType.Array)
	assert.Equal(t, "Item", method1.ReturnType.Array.Name)

	// Test get_tags() -> [string]
	method2 := protocolDef.Methods[1]
	assert.Equal(t, "get_tags", method2.Name)
	require.NotNil(t, method2.ReturnType.Array)
	assert.Equal(t, "string", method2.ReturnType.Array.Name)

	// Test get_matrix() -> [[number]] (nested arrays)
	method3 := protocolDef.Methods[2]
	assert.Equal(t, "get_matrix", method3.Name)
	require.NotNil(t, method3.ReturnType.Array)       // Outer array
	require.NotNil(t, method3.ReturnType.Array.Array) // Inner array
	assert.Equal(t, "number", method3.ReturnType.Array.Array.Name)

	// Test process_list(items: [string]) -> [Item] (array parameters)
	method4 := protocolDef.Methods[3]
	assert.Equal(t, "process_list", method4.Name)
	require.Len(t, method4.Parameters, 1)
	param := method4.Parameters[0]
	assert.Equal(t, "items", param.Name)
	require.NotNil(t, param.Type.Array)
	assert.Equal(t, "string", param.Type.Array.Name)
	require.NotNil(t, method4.ReturnType.Array)
	assert.Equal(t, "Item", method4.ReturnType.Array.Name)

	// Test bulk_create(users: [User], posts: [Post]) -> [Result] (multiple array parameters)
	method5 := protocolDef.Methods[4]
	assert.Equal(t, "bulk_create", method5.Name)
	require.Len(t, method5.Parameters, 2)

	param1 := method5.Parameters[0]
	assert.Equal(t, "users", param1.Name)
	require.NotNil(t, param1.Type.Array)
	assert.Equal(t, "User", param1.Type.Array.Name)

	param2 := method5.Parameters[1]
	assert.Equal(t, "posts", param2.Name)
	require.NotNil(t, param2.Type.Array)
	assert.Equal(t, "Post", param2.Type.Array.Name)

	require.NotNil(t, method5.ReturnType.Array)
	assert.Equal(t, "Result", method5.ReturnType.Array.Name)
}

// Test multiple protocols in one file
func TestParser_ProtocolMultiple(t *testing.T) {
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
