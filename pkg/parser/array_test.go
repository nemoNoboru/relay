package parser

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ARRAY TYPE PARSING TESTS
// =============================================================================

// Test array type lexer tokenization
func TestLexer_ArrayTypeTokens(t *testing.T) {
	input := `[string]`

	lex, err := relayLexer.Lex("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	// Get symbol mappings
	symbols := relayLexer.Symbols()
	whitespaceType := symbols["Whitespace"]

	// Collect all tokens
	var filteredTokens []lexer.Token
	for {
		token, err := lex.Next()
		if err != nil {
			break
		}
		if token.Type != whitespaceType {
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
		{"LBracket", "["},
		{"Ident", "string"},
		{"RBracket", "]"},
		{"EOF", ""},
	}

	require.Len(t, filteredTokens, len(expectedTokens), "Token count mismatch")

	for i, expected := range expectedTokens {
		expectedType := symbols[expected.TypeName]
		assert.Equal(t, expectedType, filteredTokens[i].Type, "Token type mismatch at position %d", i)
		assert.Equal(t, expected.Value, filteredTokens[i].Value, "Token value mismatch at position %d", i)
	}
}

// Test basic array type parsing
func TestParser_ArrayTypeBasic(t *testing.T) {
	input := `struct Container {
  items: [string]
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 1)

	field := structDef.Fields[0]
	assert.Equal(t, "items", field.Name)
	require.NotNil(t, field.Type.Array)
	assert.Equal(t, "string", field.Type.Array.Name)
}

// Test nested array types
func TestParser_ArrayTypeNested(t *testing.T) {
	input := `struct NestedContainer {
  matrix: [[number]],
  grid: [[[string]]]
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 2)

	// Test matrix: [[number]]
	matrixField := structDef.Fields[0]
	assert.Equal(t, "matrix", matrixField.Name)
	require.NotNil(t, matrixField.Type.Array)       // Outer array
	require.NotNil(t, matrixField.Type.Array.Array) // Inner array
	assert.Equal(t, "number", matrixField.Type.Array.Array.Name)

	// Test grid: [[[string]]]
	gridField := structDef.Fields[1]
	assert.Equal(t, "grid", gridField.Name)
	require.NotNil(t, gridField.Type.Array)             // Level 1 array
	require.NotNil(t, gridField.Type.Array.Array)       // Level 2 array
	require.NotNil(t, gridField.Type.Array.Array.Array) // Level 3 array
	assert.Equal(t, "string", gridField.Type.Array.Array.Array.Name)
}

// Test array types with custom types
func TestParser_ArrayTypeCustom(t *testing.T) {
	input := `struct Blog {
  posts: [Post],
  authors: [User],
  categories: [Category]
}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 3)

	// Test posts: [Post]
	postsField := structDef.Fields[0]
	assert.Equal(t, "posts", postsField.Name)
	require.NotNil(t, postsField.Type.Array)
	assert.Equal(t, "Post", postsField.Type.Array.Name)

	// Test authors: [User]
	authorsField := structDef.Fields[1]
	assert.Equal(t, "authors", authorsField.Name)
	require.NotNil(t, authorsField.Type.Array)
	assert.Equal(t, "User", authorsField.Type.Array.Name)

	// Test categories: [Category]
	categoriesField := structDef.Fields[2]
	assert.Equal(t, "categories", categoriesField.Name)
	require.NotNil(t, categoriesField.Type.Array)
	assert.Equal(t, "Category", categoriesField.Type.Array.Name)
}

// Benchmark array type parsing performance
func BenchmarkParser_ArrayTypeParsing(b *testing.B) {
	input := `struct Container {
  simple: [string],
  nested: [[number]],
  custom: [User]
}`

	for i := 0; i < b.N; i++ {
		_, err := relayParser.Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// MIXED PARSING TESTS (Struct + Protocol + Array Types)
// =============================================================================

// Test mixed struct and protocol definitions with array types
func TestParser_MixedStructProtocolArrays(t *testing.T) {
	input := `struct User {
		name: string,
		email: string,
		tags: [string]
	}
	
	protocol UserService {
		create_user(user: User) -> User
		get_users() -> [User]
		get_user_tags(id: string) -> [string]
		bulk_update(users: [User]) -> [User]
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 2)

	// First should be struct
	structStmt := program.Statements[0]
	require.NotNil(t, structStmt.StructDef)
	structDef := structStmt.StructDef
	assert.Equal(t, "User", structDef.Name)
	require.Len(t, structDef.Fields, 3)

	// Check array field in struct
	tagsField := structDef.Fields[2]
	assert.Equal(t, "tags", tagsField.Name)
	require.NotNil(t, tagsField.Type.Array)
	assert.Equal(t, "string", tagsField.Type.Array.Name)

	// Second should be protocol
	protocolStmt := program.Statements[1]
	require.NotNil(t, protocolStmt.ProtocolDef)
	protocolDef := protocolStmt.ProtocolDef
	assert.Equal(t, "UserService", protocolDef.Name)
	require.Len(t, protocolDef.Methods, 4)

	// Check array return types and parameters in protocol
	getUsersMethod := protocolDef.Methods[1]
	assert.Equal(t, "get_users", getUsersMethod.Name)
	require.NotNil(t, getUsersMethod.ReturnType.Array)
	assert.Equal(t, "User", getUsersMethod.ReturnType.Array.Name)

	bulkUpdateMethod := protocolDef.Methods[3]
	assert.Equal(t, "bulk_update", bulkUpdateMethod.Name)
	require.Len(t, bulkUpdateMethod.Parameters, 1)
	param := bulkUpdateMethod.Parameters[0]
	require.NotNil(t, param.Type.Array)
	assert.Equal(t, "User", param.Type.Array.Name)
	require.NotNil(t, bulkUpdateMethod.ReturnType.Array)
	assert.Equal(t, "User", bulkUpdateMethod.ReturnType.Array.Name)
}
