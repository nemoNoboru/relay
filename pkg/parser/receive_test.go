package parser

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RECEIVE PARSING TESTS
// =============================================================================

// Test the lexer tokenization for receive definition
func TestLexer_ReceiveTokens(t *testing.T) {
	input := `receive get_posts {} -> [Post] {}`

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
		{"Ident", "receive"},
		{"Ident", "get_posts"},
		{"LBrace", "{"},
		{"RBrace", "}"},
		{"Arrow", "->"},
		{"LBracket", "["},
		{"Ident", "Post"},
		{"RBracket", "]"},
		{"LBrace", "{"},
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

// Test parsing basic receive definition without parameters
func TestParser_ReceiveDefinitionNoParams(t *testing.T) {
	input := `receive get_posts {} -> [Post] {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	// Verify we have exactly one statement
	require.Len(t, program.Statements, 1)

	// Verify it's a receive definition
	stmt := program.Statements[0]
	require.NotNil(t, stmt.ReceiveDef)

	receiveDef := stmt.ReceiveDef

	// Verify basic properties
	assert.Equal(t, "get_posts", receiveDef.Name)
	assert.Len(t, receiveDef.Parameters, 0)

	// Verify return type
	require.NotNil(t, receiveDef.ReturnType)
	require.NotNil(t, receiveDef.ReturnType.Array)
	assert.Equal(t, "Post", receiveDef.ReturnType.Array.Name)

	// Verify body
	require.NotNil(t, receiveDef.Body)
}

// Test parsing receive definition with parameters
func TestParser_ReceiveDefinitionWithParams(t *testing.T) {
	input := `receive create_post {title: string, content: string} -> Post {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	receiveDef := program.Statements[0].ReceiveDef
	require.NotNil(t, receiveDef)

	// Verify name
	assert.Equal(t, "create_post", receiveDef.Name)

	// Verify parameters
	require.Len(t, receiveDef.Parameters, 2)

	// First parameter: title: string
	param1 := receiveDef.Parameters[0]
	assert.Equal(t, "title", param1.Name)
	require.NotNil(t, param1.Type)
	assert.Equal(t, "string", param1.Type.Name)

	// Second parameter: content: string
	param2 := receiveDef.Parameters[1]
	assert.Equal(t, "content", param2.Name)
	require.NotNil(t, param2.Type)
	assert.Equal(t, "string", param2.Type.Name)

	// Verify return type
	require.NotNil(t, receiveDef.ReturnType)
	assert.Equal(t, "Post", receiveDef.ReturnType.Name)

	// Verify body
	require.NotNil(t, receiveDef.Body)
}

// Test receive definition without return type
func TestParser_ReceiveDefinitionNoReturnType(t *testing.T) {
	input := `receive ping {} {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	receiveDef := program.Statements[0].ReceiveDef
	require.NotNil(t, receiveDef)

	assert.Equal(t, "ping", receiveDef.Name)
	assert.Len(t, receiveDef.Parameters, 0)
	assert.Nil(t, receiveDef.ReturnType)
	require.NotNil(t, receiveDef.Body)
}

// Test receive definition with single parameter
func TestParser_ReceiveDefinitionSingleParam(t *testing.T) {
	input := `receive get_post {id: string} -> Post {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	receiveDef := program.Statements[0].ReceiveDef
	require.NotNil(t, receiveDef)

	assert.Equal(t, "get_post", receiveDef.Name)
	require.Len(t, receiveDef.Parameters, 1)

	param := receiveDef.Parameters[0]
	assert.Equal(t, "id", param.Name)
	assert.Equal(t, "string", param.Type.Name)

	require.NotNil(t, receiveDef.ReturnType)
	assert.Equal(t, "Post", receiveDef.ReturnType.Name)
}

// Test receive definition with various parameter types
func TestParser_ReceiveDefinitionBasicTypes(t *testing.T) {
	input := `receive test_method {
		name: string,
		count: number,
		active: bool,
		created: datetime
	} -> object {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	receiveDef := program.Statements[0].ReceiveDef
	require.NotNil(t, receiveDef)

	assert.Equal(t, "test_method", receiveDef.Name)
	require.Len(t, receiveDef.Parameters, 4)

	expectedParams := []struct {
		name     string
		typeName string
	}{
		{"name", "string"},
		{"count", "number"},
		{"active", "bool"},
		{"created", "datetime"},
	}

	for i, expected := range expectedParams {
		param := receiveDef.Parameters[i]
		assert.Equal(t, expected.name, param.Name)
		assert.Equal(t, expected.typeName, param.Type.Name)
	}

	require.NotNil(t, receiveDef.ReturnType)
	assert.Equal(t, "object", receiveDef.ReturnType.Name)
}

// Test receive definition with array parameter types
func TestParser_ReceiveDefinitionArrayTypes(t *testing.T) {
	input := `receive process_data {
		items: [string],
		scores: [number],
		users: [User]
	} -> [Result] {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	receiveDef := program.Statements[0].ReceiveDef
	require.NotNil(t, receiveDef)

	assert.Equal(t, "process_data", receiveDef.Name)
	require.Len(t, receiveDef.Parameters, 3)

	// Test items: [string]
	param1 := receiveDef.Parameters[0]
	assert.Equal(t, "items", param1.Name)
	require.NotNil(t, param1.Type.Array)
	assert.Equal(t, "string", param1.Type.Array.Name)

	// Test scores: [number]
	param2 := receiveDef.Parameters[1]
	assert.Equal(t, "scores", param2.Name)
	require.NotNil(t, param2.Type.Array)
	assert.Equal(t, "number", param2.Type.Array.Name)

	// Test users: [User]
	param3 := receiveDef.Parameters[2]
	assert.Equal(t, "users", param3.Name)
	require.NotNil(t, param3.Type.Array)
	assert.Equal(t, "User", param3.Type.Array.Name)

	// Test return type: [Result]
	require.NotNil(t, receiveDef.ReturnType)
	require.NotNil(t, receiveDef.ReturnType.Array)
	assert.Equal(t, "Result", receiveDef.ReturnType.Array.Name)
}

// Test receive definition with nested array types
func TestParser_ReceiveDefinitionNestedArrayTypes(t *testing.T) {
	input := `receive process_matrix {
		matrix: [[number]],
		grid: [[[string]]]
	} -> [[Result]] {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	receiveDef := program.Statements[0].ReceiveDef
	require.NotNil(t, receiveDef)

	require.Len(t, receiveDef.Parameters, 2)

	// Test matrix: [[number]]
	param1 := receiveDef.Parameters[0]
	assert.Equal(t, "matrix", param1.Name)
	require.NotNil(t, param1.Type.Array)       // Outer array
	require.NotNil(t, param1.Type.Array.Array) // Inner array
	assert.Equal(t, "number", param1.Type.Array.Array.Name)

	// Test grid: [[[string]]]
	param2 := receiveDef.Parameters[1]
	assert.Equal(t, "grid", param2.Name)
	require.NotNil(t, param2.Type.Array)             // Outer array
	require.NotNil(t, param2.Type.Array.Array)       // Middle array
	require.NotNil(t, param2.Type.Array.Array.Array) // Inner array
	assert.Equal(t, "string", param2.Type.Array.Array.Array.Name)

	// Test return type: [[Result]]
	require.NotNil(t, receiveDef.ReturnType)
	require.NotNil(t, receiveDef.ReturnType.Array)       // Outer array
	require.NotNil(t, receiveDef.ReturnType.Array.Array) // Inner array
	assert.Equal(t, "Result", receiveDef.ReturnType.Array.Array.Name)
}

// Test multiple receive definitions
func TestParser_ReceiveDefinitionMultiple(t *testing.T) {
	input := `receive get_posts {} -> [Post] {}
	
	receive create_post {title: string, content: string} -> Post {}
	
	receive delete_post {id: string} -> bool {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 3)

	// First receive: get_posts
	receive1 := program.Statements[0].ReceiveDef
	require.NotNil(t, receive1)
	assert.Equal(t, "get_posts", receive1.Name)
	assert.Len(t, receive1.Parameters, 0)
	require.NotNil(t, receive1.ReturnType)
	assert.Equal(t, "Post", receive1.ReturnType.Array.Name)

	// Second receive: create_post
	receive2 := program.Statements[1].ReceiveDef
	require.NotNil(t, receive2)
	assert.Equal(t, "create_post", receive2.Name)
	require.Len(t, receive2.Parameters, 2)
	assert.Equal(t, "title", receive2.Parameters[0].Name)
	assert.Equal(t, "content", receive2.Parameters[1].Name)

	// Third receive: delete_post
	receive3 := program.Statements[2].ReceiveDef
	require.NotNil(t, receive3)
	assert.Equal(t, "delete_post", receive3.Name)
	require.Len(t, receive3.Parameters, 1)
	assert.Equal(t, "id", receive3.Parameters[0].Name)
	assert.Equal(t, "bool", receive3.ReturnType.Name)
}

// Test receive case sensitivity
func TestParser_ReceiveDefinitionCaseSensitivity(t *testing.T) {
	input := `receive GetPosts {} -> [Post] {}
	
	receive get_posts {} -> [post] {}
	
	receive GET_POSTS {} -> [POST] {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 3)

	// Methods should maintain their original case
	assert.Equal(t, "GetPosts", program.Statements[0].ReceiveDef.Name)
	assert.Equal(t, "get_posts", program.Statements[1].ReceiveDef.Name)
	assert.Equal(t, "GET_POSTS", program.Statements[2].ReceiveDef.Name)

	// Return types should also maintain case
	assert.Equal(t, "Post", program.Statements[0].ReceiveDef.ReturnType.Array.Name)
	assert.Equal(t, "post", program.Statements[1].ReceiveDef.ReturnType.Array.Name)
	assert.Equal(t, "POST", program.Statements[2].ReceiveDef.ReturnType.Array.Name)
}

// Test receive parsing error cases
func TestParser_ReceiveDefinitionErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing_method_name", `receive {} -> Post {}`},
		{"missing_parameter_braces", `receive test_method -> Post {}`},
		{"missing_body_braces", `receive test_method {} -> Post {`},
		{"missing_parameter_type", `receive test_method {id} -> Post {}`},
		{"missing_colon", `receive test_method {id string} -> Post {}`},
		{"invalid_method_name", `receive 123invalid {} -> Post {}`},
		{"unclosed_parameter_braces", `receive test_method {id: string -> Post {}`},
		{"unclosed_body_braces", `receive test_method {} -> Post {`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := relayParser.Parse("test.relay", strings.NewReader(tt.input))
			assert.Error(t, err, "Expected parsing error for %s", tt.name)
		})
	}
}

// Test mixed receive with other constructs
func TestParser_MixedReceiveWithOtherConstructs(t *testing.T) {
	input := `struct Post {
		id: string,
		title: string
	}
	
	protocol BlogService {
		get_posts() -> [Post]
	}
	
	state {
		posts: [Post] = []
	}
	
	receive get_posts {} -> [Post] {}
	
	receive create_post {title: string} -> Post {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 5)

	// First should be struct
	structStmt := program.Statements[0]
	require.NotNil(t, structStmt.StructDef)
	assert.Equal(t, "Post", structStmt.StructDef.Name)

	// Second should be protocol
	protocolStmt := program.Statements[1]
	require.NotNil(t, protocolStmt.ProtocolDef)
	assert.Equal(t, "BlogService", protocolStmt.ProtocolDef.Name)

	// Third should be state
	stateStmt := program.Statements[2]
	require.NotNil(t, stateStmt.StateDef)
	require.Len(t, stateStmt.StateDef.Fields, 1)

	// Fourth should be receive: get_posts
	receiveStmt1 := program.Statements[3]
	require.NotNil(t, receiveStmt1.ReceiveDef)
	assert.Equal(t, "get_posts", receiveStmt1.ReceiveDef.Name)
	assert.Len(t, receiveStmt1.ReceiveDef.Parameters, 0)

	// Fifth should be receive: create_post
	receiveStmt2 := program.Statements[4]
	require.NotNil(t, receiveStmt2.ReceiveDef)
	assert.Equal(t, "create_post", receiveStmt2.ReceiveDef.Name)
	require.Len(t, receiveStmt2.ReceiveDef.Parameters, 1)
	assert.Equal(t, "title", receiveStmt2.ReceiveDef.Parameters[0].Name)
}

// Benchmark receive parsing performance
func BenchmarkParser_ReceiveParsing(b *testing.B) {
	input := `receive create_post {
		title: string,
		content: string,
		author: string,
		tags: [string]
	} -> Post {}`

	for i := 0; i < b.N; i++ {
		_, err := relayParser.Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
