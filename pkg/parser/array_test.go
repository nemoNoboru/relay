package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ARRAY TYPE PARSING TESTS
// =============================================================================

// Test parsing basic array type [string]
func TestParser_ArrayTypeBasic(t *testing.T) {
	input := `struct User {
		tags: [string]
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 1)

	field := structDef.Fields[0]
	assert.Equal(t, "tags", field.Name)
	require.NotNil(t, field.Type.Array)
	assert.Equal(t, "string", field.Type.Array.Name)
}

// Test nested array types [[string]]
func TestParser_ArrayTypeNested(t *testing.T) {
	input := `struct Container {
		matrix: [[string]]
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 1)

	field := structDef.Fields[0]
	assert.Equal(t, "matrix", field.Name)
	require.NotNil(t, field.Type.Array)
	require.NotNil(t, field.Type.Array.Array)
	assert.Equal(t, "string", field.Type.Array.Array.Name)
}

// Test array of custom types
func TestParser_ArrayTypeCustom(t *testing.T) {
	input := `struct Blog {
		posts: [Post],
		authors: [User]
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 2)

	// Test posts: [Post]
	field1 := structDef.Fields[0]
	assert.Equal(t, "posts", field1.Name)
	require.NotNil(t, field1.Type.Array)
	assert.Equal(t, "Post", field1.Type.Array.Name)

	// Test authors: [User]
	field2 := structDef.Fields[1]
	assert.Equal(t, "authors", field2.Name)
	require.NotNil(t, field2.Type.Array)
	assert.Equal(t, "User", field2.Type.Array.Name)
}

// Benchmark array type parsing performance
func BenchmarkParser_ArrayTypeParsing(b *testing.B) {
	input := `struct Container {
  simple: [string],
  nested: [[number]],
  custom: [User]
}`

	for i := 0; i < b.N; i++ {
		_, err := Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// MIXED PARSING TESTS (Struct + Protocol + Array Types)
// =============================================================================

// Test mixed struct and protocol with array types
func TestParser_MixedStructProtocolArrays(t *testing.T) {
	input := `struct Post {
		tags: [string]
	}
	
	protocol BlogService {
		get_posts() -> [Post]
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 2)

	// Check struct
	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	assert.Equal(t, "Post", structDef.Name)

	// Check protocol
	protocolDef := program.Statements[1].ProtocolDef
	require.NotNil(t, protocolDef)
	assert.Equal(t, "BlogService", protocolDef.Name)
	require.Len(t, protocolDef.Methods, 1)

	method := protocolDef.Methods[0]
	assert.Equal(t, "get_posts", method.Name)
	require.NotNil(t, method.ReturnType.Array)
	assert.Equal(t, "Post", method.ReturnType.Array.Name)
}
