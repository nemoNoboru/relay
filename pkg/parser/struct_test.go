package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// STRUCT PARSING TESTS
// =============================================================================

// Test parsing basic struct definition
func TestParser_StructDefinition(t *testing.T) {
	input := `struct User {
		username: string,
		email: string,
		age: number,
		active: bool
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
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
	require.Len(t, structDef.Fields, 4)

	// Test fields
	field1 := structDef.Fields[0]
	assert.Equal(t, "username", field1.Name)
	assert.Equal(t, "string", field1.Type.Name)

	field2 := structDef.Fields[1]
	assert.Equal(t, "email", field2.Name)
	assert.Equal(t, "string", field2.Type.Name)

	field3 := structDef.Fields[2]
	assert.Equal(t, "age", field3.Name)
	assert.Equal(t, "number", field3.Type.Name)

	field4 := structDef.Fields[3]
	assert.Equal(t, "active", field4.Name)
	assert.Equal(t, "bool", field4.Type.Name)
}

// Test struct with no fields (empty struct)
func TestParser_StructEmpty(t *testing.T) {
	input := `struct Empty {}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	require.Len(t, program.Statements, 1)
	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)

	assert.Equal(t, "Empty", structDef.Name)
	assert.Len(t, structDef.Fields, 0)
}

// Test struct with single field
func TestParser_StructSingleField(t *testing.T) {
	input := `struct Simple {
  id: string
}`

	program, err := Parse("test.relay", strings.NewReader(input))
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
func TestParser_StructBasicTypes(t *testing.T) {
	input := `struct Complex {
  name: string,
  age: number,
  active: bool,
  created_at: datetime
}`

	program, err := Parse("test.relay", strings.NewReader(input))
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

// Test struct with array types
func TestParser_StructArrayTypes(t *testing.T) {
	input := `struct Post {
  title: string,
  tags: [string],
  comments: [Comment],
  scores: [number]
}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 4)

	// Test title: string (basic type)
	titleField := structDef.Fields[0]
	assert.Equal(t, "title", titleField.Name)
	assert.Equal(t, "string", titleField.Type.Name)
	assert.Nil(t, titleField.Type.Array) // Not an array

	// Test tags: [string] (array of strings)
	tagsField := structDef.Fields[1]
	assert.Equal(t, "tags", tagsField.Name)
	require.NotNil(t, tagsField.Type.Array)
	assert.Equal(t, "string", tagsField.Type.Array.Name)

	// Test comments: [Comment] (array of custom types)
	commentsField := structDef.Fields[2]
	assert.Equal(t, "comments", commentsField.Name)
	require.NotNil(t, commentsField.Type.Array)
	assert.Equal(t, "Comment", commentsField.Type.Array.Name)

	// Test scores: [number] (array of numbers)
	scoresField := structDef.Fields[3]
	assert.Equal(t, "scores", scoresField.Name)
	require.NotNil(t, scoresField.Type.Array)
	assert.Equal(t, "number", scoresField.Type.Array.Name)
}

// Test struct with nested array types
func TestParser_StructNestedArrayTypes(t *testing.T) {
	input := `struct Matrix {
  data: [[number]],
  labels: [[string]]
}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 2)

	// Test data: [[number]] (array of array of numbers)
	dataField := structDef.Fields[0]
	assert.Equal(t, "data", dataField.Name)
	require.NotNil(t, dataField.Type.Array)       // Outer array
	require.NotNil(t, dataField.Type.Array.Array) // Inner array
	assert.Equal(t, "number", dataField.Type.Array.Array.Name)

	// Test labels: [[string]] (array of array of strings)
	labelsField := structDef.Fields[1]
	assert.Equal(t, "labels", labelsField.Name)
	require.NotNil(t, labelsField.Type.Array)       // Outer array
	require.NotNil(t, labelsField.Type.Array.Array) // Inner array
	assert.Equal(t, "string", labelsField.Type.Array.Array.Name)
}

// Test struct with optional types - DISABLED (optional not implemented yet)
func TestParser_StructOptionalTypes(t *testing.T) {
	t.Skip("Optional types not implemented yet")
}

// Test struct with validation constraints - DISABLED (validation not implemented yet)
func TestParser_StructValidation(t *testing.T) {
	t.Skip("Validation constraints not implemented yet")
}

// Test struct with object types - DISABLED (object types not implemented yet)
func TestParser_StructObjectTypes(t *testing.T) {
	t.Skip("Object types not fully implemented yet")
}

// Test struct parsing error cases
func TestParser_StructErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing_struct_name", `struct {}`},
		{"missing_opening_brace", `struct User name: string }`},
		{"missing_closing_brace", `struct User { name: string`},
		{"missing_field_type", `struct User { name: }`},
		{"missing_colon", `struct User { name string }`},
		{"invalid_field_name", `struct User { 123: string }`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tt.input))
			assert.Error(t, err, "Expected parsing error for %s", tt.name)
		})
	}
}

// Test multiple structs in one file
func TestParser_StructMultiple(t *testing.T) {
	input := `struct User {
  name: string,
  email: string
}

struct Post {
  title: string,
  content: string
}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	require.Len(t, program.Statements, 2)

	// First struct
	struct1 := program.Statements[0].StructDef
	require.NotNil(t, struct1)
	assert.Equal(t, "User", struct1.Name)
	require.Len(t, struct1.Fields, 2)
	assert.Equal(t, "name", struct1.Fields[0].Name)
	assert.Equal(t, "email", struct1.Fields[1].Name)

	// Second struct
	struct2 := program.Statements[1].StructDef
	require.NotNil(t, struct2)
	assert.Equal(t, "Post", struct2.Name)
	require.Len(t, struct2.Fields, 2)
	assert.Equal(t, "title", struct2.Fields[0].Name)
	assert.Equal(t, "content", struct2.Fields[1].Name)
}

// Test struct with trailing comma - DISABLED (trailing comma not supported yet)
func TestParser_StructTrailingComma(t *testing.T) {
	t.Skip("Trailing comma not supported yet")
}

// Test struct case sensitivity
func TestParser_StructCaseSensitivity(t *testing.T) {
	input := `struct UserProfile {
  UserName: string,
  emailAddress: string
}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	structDef := program.Statements[0].StructDef
	require.NotNil(t, structDef)
	require.Len(t, structDef.Fields, 2)

	// Fields should maintain their original case
	field1 := structDef.Fields[0]
	assert.Equal(t, "UserName", field1.Name)

	field2 := structDef.Fields[1]
	assert.Equal(t, "emailAddress", field2.Name)
}

// Benchmark struct parsing performance
func BenchmarkParser_StructParsing(b *testing.B) {
	input := `struct User {
  name: string,
  email: string,
  age: number
}`

	for i := 0; i < b.N; i++ {
		_, err := Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
