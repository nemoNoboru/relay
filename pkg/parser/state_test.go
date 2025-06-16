package parser

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// STATE PARSING TESTS
// =============================================================================

// Test the lexer tokenization for state definition
func TestLexer_StateTokens(t *testing.T) {
	input := `state {
		count: number = 0,
		active: bool = true,
		name: string = "default"
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
		{"Ident", "state"},
		{"LBrace", "{"},
		{"Ident", "count"},
		{"Colon", ":"},
		{"Ident", "number"},
		{"Assign", "="},
		{"Number", "0"},
		{"Comma", ","},
		{"Ident", "active"},
		{"Colon", ":"},
		{"Ident", "bool"},
		{"Assign", "="},
		{"Bool", "true"},
		{"Comma", ","},
		{"Ident", "name"},
		{"Colon", ":"},
		{"Ident", "string"},
		{"Assign", "="},
		{"String", `"default"`},
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

// Test parsing basic state definition
func TestParser_StateDefinition(t *testing.T) {
	input := `state {
		count: number = 0,
		active: bool = true,
		name: string = "default"
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	// Verify we have exactly one statement
	require.Len(t, program.Statements, 1)

	// Verify it's a state definition
	stmt := program.Statements[0]
	require.NotNil(t, stmt.StateDef)

	stateDef := stmt.StateDef

	// Verify fields
	require.Len(t, stateDef.Fields, 3)

	// Test first field: count: number = 0
	field1 := stateDef.Fields[0]
	assert.Equal(t, "count", field1.Name)
	require.NotNil(t, field1.Type)
	assert.Equal(t, "number", field1.Type.Name)
	require.NotNil(t, field1.DefaultValue)
	require.NotNil(t, field1.DefaultValue.Number)
	assert.Equal(t, float64(0), *field1.DefaultValue.Number)

	// Test second field: active: bool = true
	field2 := stateDef.Fields[1]
	assert.Equal(t, "active", field2.Name)
	require.NotNil(t, field2.Type)
	assert.Equal(t, "bool", field2.Type.Name)
	require.NotNil(t, field2.DefaultValue)
	require.NotNil(t, field2.DefaultValue.Bool)
	assert.True(t, *field2.DefaultValue.GetBoolValue())

	// Test third field: name: string = "default"
	field3 := stateDef.Fields[2]
	assert.Equal(t, "name", field3.Name)
	require.NotNil(t, field3.Type)
	assert.Equal(t, "string", field3.Type.Name)
	require.NotNil(t, field3.DefaultValue)
	require.NotNil(t, field3.DefaultValue.String)
	assert.Equal(t, "default", *field3.DefaultValue.String)
}

// Test state with no fields (empty state)
func TestParser_StateEmpty(t *testing.T) {
	input := `state {}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	require.Len(t, program.Statements, 1)
	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)

	assert.Len(t, stateDef.Fields, 0)
}

// Test state with single field
func TestParser_StateSingleField(t *testing.T) {
	input := `state {
		counter: number = 42
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)

	require.Len(t, stateDef.Fields, 1)

	field := stateDef.Fields[0]
	assert.Equal(t, "counter", field.Name)
	assert.Equal(t, "number", field.Type.Name)
	require.NotNil(t, field.DefaultValue.Number)
	assert.Equal(t, float64(42), *field.DefaultValue.Number)
}

// Test state with fields without default values
func TestParser_StateFieldsNoDefaults(t *testing.T) {
	input := `state {
		count: number,
		name: string,
		active: bool
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)
	require.Len(t, stateDef.Fields, 3)

	// All fields should have no default values
	for i, expectedName := range []string{"count", "name", "active"} {
		field := stateDef.Fields[i]
		assert.Equal(t, expectedName, field.Name)
		assert.Nil(t, field.DefaultValue)
	}
}

// Test state with array types and array literals
func TestParser_StateArrayTypes(t *testing.T) {
	input := `state {
		items: [string] = ["a", "b", "c"],
		scores: [number] = [1, 2, 3],
		flags: [bool] = [true, false],
		empty_list: [object] = []
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)
	require.Len(t, stateDef.Fields, 4)

	// Test items: [string] = ["a", "b", "c"]
	itemsField := stateDef.Fields[0]
	assert.Equal(t, "items", itemsField.Name)
	require.NotNil(t, itemsField.Type.Array)
	assert.Equal(t, "string", itemsField.Type.Array.Name)
	require.NotNil(t, itemsField.DefaultValue.Array)
	require.Len(t, itemsField.DefaultValue.Array.Elements, 3)
	assert.Equal(t, "a", *itemsField.DefaultValue.Array.Elements[0].String)
	assert.Equal(t, "b", *itemsField.DefaultValue.Array.Elements[1].String)
	assert.Equal(t, "c", *itemsField.DefaultValue.Array.Elements[2].String)

	// Test scores: [number] = [1, 2, 3]
	scoresField := stateDef.Fields[1]
	assert.Equal(t, "scores", scoresField.Name)
	require.NotNil(t, scoresField.Type.Array)
	assert.Equal(t, "number", scoresField.Type.Array.Name)
	require.NotNil(t, scoresField.DefaultValue.Array)
	require.Len(t, scoresField.DefaultValue.Array.Elements, 3)
	assert.Equal(t, float64(1), *scoresField.DefaultValue.Array.Elements[0].Number)
	assert.Equal(t, float64(2), *scoresField.DefaultValue.Array.Elements[1].Number)
	assert.Equal(t, float64(3), *scoresField.DefaultValue.Array.Elements[2].Number)

	// Test flags: [bool] = [true, false]
	flagsField := stateDef.Fields[2]
	assert.Equal(t, "flags", flagsField.Name)
	require.NotNil(t, flagsField.Type.Array)
	assert.Equal(t, "bool", flagsField.Type.Array.Name)
	require.NotNil(t, flagsField.DefaultValue.Array)
	require.Len(t, flagsField.DefaultValue.Array.Elements, 2)
	assert.True(t, *flagsField.DefaultValue.Array.Elements[0].GetBoolValue())
	assert.False(t, *flagsField.DefaultValue.Array.Elements[1].GetBoolValue())

	// Test empty_list: [object] = []
	emptyField := stateDef.Fields[3]
	assert.Equal(t, "empty_list", emptyField.Name)
	require.NotNil(t, emptyField.Type.Array)
	assert.Equal(t, "object", emptyField.Type.Array.Name)
	require.NotNil(t, emptyField.DefaultValue.Array)
	assert.Len(t, emptyField.DefaultValue.Array.Elements, 0)
}

// Test state with function call defaults
func TestParser_StateFunctionCallDefaults(t *testing.T) {
	input := `state {
		created_at: datetime = now(),
		id: string = uuid(),
		timestamp: number = time()
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)
	require.Len(t, stateDef.Fields, 3)

	// Test created_at: datetime = now()
	field1 := stateDef.Fields[0]
	assert.Equal(t, "created_at", field1.Name)
	assert.Equal(t, "datetime", field1.Type.Name)
	require.NotNil(t, field1.DefaultValue.FuncCall)
	assert.Equal(t, "now", field1.DefaultValue.FuncCall.Name)
	assert.Len(t, field1.DefaultValue.FuncCall.Args, 0)

	// Test id: string = uuid()
	field2 := stateDef.Fields[1]
	assert.Equal(t, "id", field2.Name)
	assert.Equal(t, "string", field2.Type.Name)
	require.NotNil(t, field2.DefaultValue.FuncCall)
	assert.Equal(t, "uuid", field2.DefaultValue.FuncCall.Name)
	assert.Len(t, field2.DefaultValue.FuncCall.Args, 0)

	// Test timestamp: number = time()
	field3 := stateDef.Fields[2]
	assert.Equal(t, "timestamp", field3.Name)
	assert.Equal(t, "number", field3.Type.Name)
	require.NotNil(t, field3.DefaultValue.FuncCall)
	assert.Equal(t, "time", field3.DefaultValue.FuncCall.Name)
	assert.Len(t, field3.DefaultValue.FuncCall.Args, 0)
}

// Test state with various data types
func TestParser_StateBasicTypes(t *testing.T) {
	input := `state {
		name: string = "test",
		count: number = 123,
		pi: number = 3.14,
		active: bool = true,
		inactive: bool = false,
		created: datetime = now()
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)
	require.Len(t, stateDef.Fields, 6)

	expectedFields := []struct {
		name       string
		typeName   string
		hasDefault bool
	}{
		{"name", "string", true},
		{"count", "number", true},
		{"pi", "number", true},
		{"active", "bool", true},
		{"inactive", "bool", true},
		{"created", "datetime", true},
	}

	for i, expected := range expectedFields {
		field := stateDef.Fields[i]
		assert.Equal(t, expected.name, field.Name)
		assert.Equal(t, expected.typeName, field.Type.Name)
		if expected.hasDefault {
			assert.NotNil(t, field.DefaultValue, "Field %s should have a default value", expected.name)
		} else {
			assert.Nil(t, field.DefaultValue, "Field %s should not have a default value", expected.name)
		}
	}

	// Verify specific default values
	assert.Equal(t, "test", *stateDef.Fields[0].DefaultValue.String)
	assert.Equal(t, float64(123), *stateDef.Fields[1].DefaultValue.Number)
	assert.Equal(t, 3.14, *stateDef.Fields[2].DefaultValue.Number)
	assert.True(t, *stateDef.Fields[3].DefaultValue.GetBoolValue())
	assert.False(t, *stateDef.Fields[4].DefaultValue.GetBoolValue())
	assert.Equal(t, "now", stateDef.Fields[5].DefaultValue.FuncCall.Name)
}

// Test state parsing error cases
func TestParser_StateErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing_opening_brace", `state name: string = "test" }`},
		{"missing_closing_brace", `state { name: string = "test"`},
		{"missing_field_type", `state { name = "test" }`},
		{"missing_colon", `state { name string = "test" }`},
		{"invalid_field_name", `state { 123: string = "test" }`},
		{"missing_assignment_value", `state { name: string = }`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := relayParser.Parse("test.relay", strings.NewReader(tt.input))
			assert.Error(t, err, "Expected parsing error for %s", tt.name)
		})
	}
}

// Test state with nested array types
func TestParser_StateNestedArrayTypes(t *testing.T) {
	input := `state {
		matrix: [[number]] = [[1, 2], [3, 4]],
		grid: [[[string]]] = [[["a", "b"]], [["c", "d"]]]
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)
	require.Len(t, stateDef.Fields, 2)

	// Test matrix: [[number]] = [[1, 2], [3, 4]]
	matrixField := stateDef.Fields[0]
	assert.Equal(t, "matrix", matrixField.Name)
	require.NotNil(t, matrixField.Type.Array)       // Outer array
	require.NotNil(t, matrixField.Type.Array.Array) // Inner array
	assert.Equal(t, "number", matrixField.Type.Array.Array.Name)

	// Check default value structure (nested arrays)
	require.NotNil(t, matrixField.DefaultValue.Array)
	require.Len(t, matrixField.DefaultValue.Array.Elements, 2)

	// First sub-array [1, 2]
	firstSubArray := matrixField.DefaultValue.Array.Elements[0].Array
	require.NotNil(t, firstSubArray)
	require.Len(t, firstSubArray.Elements, 2)
	assert.Equal(t, float64(1), *firstSubArray.Elements[0].Number)
	assert.Equal(t, float64(2), *firstSubArray.Elements[1].Number)
}

// Test state case sensitivity
func TestParser_StateCaseSensitivity(t *testing.T) {
	input := `state {
		UserCount: number = 0,
		userCount: number = 1,
		USERCOUNT: number = 2
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	stateDef := program.Statements[0].StateDef
	require.NotNil(t, stateDef)
	require.Len(t, stateDef.Fields, 3)

	// Fields should maintain their original case
	assert.Equal(t, "UserCount", stateDef.Fields[0].Name)
	assert.Equal(t, "userCount", stateDef.Fields[1].Name)
	assert.Equal(t, "USERCOUNT", stateDef.Fields[2].Name)

	// Values should be different
	assert.Equal(t, float64(0), *stateDef.Fields[0].DefaultValue.Number)
	assert.Equal(t, float64(1), *stateDef.Fields[1].DefaultValue.Number)
	assert.Equal(t, float64(2), *stateDef.Fields[2].DefaultValue.Number)
}

// Test state with trailing comma - DISABLED (trailing comma not supported yet)
func TestParser_StateTrailingComma(t *testing.T) {
	t.Skip("Trailing comma not supported yet")
}

// Test mixed state with other constructs
func TestParser_MixedStateWithOtherConstructs(t *testing.T) {
	input := `struct User {
		name: string,
		email: string
	}
	
	state {
		users: [User] = [],
		count: number = 0
	}
	
	protocol UserService {
		get_users() -> [User]
	}`

	program, err := relayParser.Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 3)

	// First should be struct
	structStmt := program.Statements[0]
	require.NotNil(t, structStmt.StructDef)
	assert.Equal(t, "User", structStmt.StructDef.Name)

	// Second should be state
	stateStmt := program.Statements[1]
	require.NotNil(t, stateStmt.StateDef)
	stateDef := stateStmt.StateDef
	require.Len(t, stateDef.Fields, 2)

	// Check state fields
	usersField := stateDef.Fields[0]
	assert.Equal(t, "users", usersField.Name)
	require.NotNil(t, usersField.Type.Array)
	assert.Equal(t, "User", usersField.Type.Array.Name)
	require.NotNil(t, usersField.DefaultValue.Array)
	assert.Len(t, usersField.DefaultValue.Array.Elements, 0) // empty array

	countField := stateDef.Fields[1]
	assert.Equal(t, "count", countField.Name)
	assert.Equal(t, "number", countField.Type.Name)
	assert.Equal(t, float64(0), *countField.DefaultValue.Number)

	// Third should be protocol
	protocolStmt := program.Statements[2]
	require.NotNil(t, protocolStmt.ProtocolDef)
	assert.Equal(t, "UserService", protocolStmt.ProtocolDef.Name)
}

// Benchmark state parsing performance
func BenchmarkParser_StateParsing(b *testing.B) {
	input := `state {
		count: number = 0,
		active: bool = true,
		name: string = "default",
		items: [string] = ["a", "b", "c"],
		created_at: datetime = now()
	}`

	for i := 0; i < b.N; i++ {
		_, err := relayParser.Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
