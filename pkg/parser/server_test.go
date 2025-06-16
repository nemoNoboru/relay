package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SERVER PARSING TESTS
// =============================================================================

// Test parsing basic server definition with protocol implementation
func TestParser_ServerDefinitionWithProtocol(t *testing.T) {
	input := `server blog_service implements BlogService {
		state {
			posts: [Post] = []
		}
		
		receive get_posts {} -> [Post] {
			return state.posts
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	// Verify we have exactly one statement
	require.Len(t, program.Statements, 1)

	// Verify it's a server definition
	stmt := program.Statements[0]
	require.NotNil(t, stmt.ServerDef)

	serverDef := stmt.ServerDef

	// Verify server name
	assert.Equal(t, "blog_service", serverDef.Name)

	// Verify protocol implementation
	require.NotNil(t, serverDef.Protocol)
	assert.Equal(t, "BlogService", *serverDef.Protocol)

	// Verify body contains state and receive
	require.NotNil(t, serverDef.Body)
	require.Len(t, serverDef.Body.Elements, 2)

	// Verify state element
	stateElement := serverDef.Body.Elements[0]
	require.NotNil(t, stateElement.State)
	require.Len(t, stateElement.State.Fields, 1)
	assert.Equal(t, "posts", stateElement.State.Fields[0].Name)

	// Verify receive element
	receiveElement := serverDef.Body.Elements[1]
	require.NotNil(t, receiveElement.Receive)
	assert.Equal(t, "get_posts", receiveElement.Receive.Name)
}

// Test parsing server definition without protocol implementation
func TestParser_ServerDefinitionNoProtocol(t *testing.T) {
	input := `server simple_service {
		receive ping {} -> bool {
			return true
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, program)

	serverDef := program.Statements[0].ServerDef
	require.NotNil(t, serverDef)

	// Verify server name
	assert.Equal(t, "simple_service", serverDef.Name)

	// Verify no protocol implementation
	assert.Nil(t, serverDef.Protocol)

	// Verify body contains receive
	require.NotNil(t, serverDef.Body)
	require.Len(t, serverDef.Body.Elements, 1)

	receiveElement := serverDef.Body.Elements[0]
	require.NotNil(t, receiveElement.Receive)
	assert.Equal(t, "ping", receiveElement.Receive.Name)
}

// Test parsing empty server definition
func TestParser_ServerDefinitionEmpty(t *testing.T) {
	input := `server empty_service {}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	serverDef := program.Statements[0].ServerDef
	require.NotNil(t, serverDef)

	assert.Equal(t, "empty_service", serverDef.Name)
	assert.Nil(t, serverDef.Protocol)
	require.NotNil(t, serverDef.Body)
	assert.Len(t, serverDef.Body.Elements, 0)
}

// Test parsing server with only state
func TestParser_ServerDefinitionStateOnly(t *testing.T) {
	input := `server stateful_service {
		state {
			counter: number = 0,
			active: bool = true
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	serverDef := program.Statements[0].ServerDef
	require.NotNil(t, serverDef)

	assert.Equal(t, "stateful_service", serverDef.Name)
	require.NotNil(t, serverDef.Body)
	require.Len(t, serverDef.Body.Elements, 1)

	stateElement := serverDef.Body.Elements[0]
	require.NotNil(t, stateElement.State)
	require.Len(t, stateElement.State.Fields, 2)

	// Verify state fields
	field1 := stateElement.State.Fields[0]
	assert.Equal(t, "counter", field1.Name)
	assert.Equal(t, "number", field1.Type.Name)

	field2 := stateElement.State.Fields[1]
	assert.Equal(t, "active", field2.Name)
	assert.Equal(t, "bool", field2.Type.Name)
}

// Test parsing server with multiple receive handlers
func TestParser_ServerDefinitionMultipleReceive(t *testing.T) {
	input := `server multi_service implements TestService {
		receive get_data {} -> [string] {
			return []
		}
		
		receive set_data {value: string} {
			// implementation
		}
		
		receive process {input: object} -> object {
			return input
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	serverDef := program.Statements[0].ServerDef
	require.NotNil(t, serverDef)

	assert.Equal(t, "multi_service", serverDef.Name)
	require.NotNil(t, serverDef.Protocol)
	assert.Equal(t, "TestService", *serverDef.Protocol)

	require.NotNil(t, serverDef.Body)
	require.Len(t, serverDef.Body.Elements, 3)

	// Verify all three receive handlers
	for i, expectedName := range []string{"get_data", "set_data", "process"} {
		receiveElement := serverDef.Body.Elements[i]
		require.NotNil(t, receiveElement.Receive)
		assert.Equal(t, expectedName, receiveElement.Receive.Name)
	}
}

// Test parsing server with state and multiple receive handlers
func TestParser_ServerDefinitionComplete(t *testing.T) {
	input := `server complete_service implements CompleteService {
		state {
			items: [string] = [],
			count: number = 0,
			metadata: string = ""
		}
		
		receive get_items {} -> [string] {
			return state.items
		}
		
		receive add_item {item: string} -> bool {
			state.count = state.count + 1
			return true
		}
		
		receive get_count {} -> number {
			return state.count
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	serverDef := program.Statements[0].ServerDef
	require.NotNil(t, serverDef)

	// Verify server properties
	assert.Equal(t, "complete_service", serverDef.Name)
	require.NotNil(t, serverDef.Protocol)
	assert.Equal(t, "CompleteService", *serverDef.Protocol)

	// Verify body structure
	require.NotNil(t, serverDef.Body)
	require.Len(t, serverDef.Body.Elements, 4) // 1 state + 3 receive

	// Verify state element
	stateElement := serverDef.Body.Elements[0]
	require.NotNil(t, stateElement.State)
	require.Len(t, stateElement.State.Fields, 3)

	stateFields := stateElement.State.Fields
	assert.Equal(t, "items", stateFields[0].Name)
	assert.Equal(t, "count", stateFields[1].Name)
	assert.Equal(t, "metadata", stateFields[2].Name)

	// Verify receive elements
	expectedReceiveNames := []string{"get_items", "add_item", "get_count"}
	for i, expectedName := range expectedReceiveNames {
		receiveElement := serverDef.Body.Elements[i+1]
		require.NotNil(t, receiveElement.Receive)
		assert.Equal(t, expectedName, receiveElement.Receive.Name)
	}
}

// Test parsing multiple server definitions in one file
func TestParser_ServerDefinitionMultiple(t *testing.T) {
	input := `server user_service implements UserService {
		receive get_user {id: string} -> User {
			return {id: id}
		}
	}
	
	server blog_service implements BlogService {
		state {
			posts: [Post] = []
		}
		
		receive get_posts {} -> [Post] {
			return state.posts
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 2)

	// Verify first server
	server1 := program.Statements[0].ServerDef
	require.NotNil(t, server1)
	assert.Equal(t, "user_service", server1.Name)
	assert.Equal(t, "UserService", *server1.Protocol)

	// Verify second server
	server2 := program.Statements[1].ServerDef
	require.NotNil(t, server2)
	assert.Equal(t, "blog_service", server2.Name)
	assert.Equal(t, "BlogService", *server2.Protocol)
}

// Test mixed server and other constructs
func TestParser_ServerDefinitionMixed(t *testing.T) {
	input := `struct Post {
		title: string,
		content: string
	}
	
	protocol BlogService {
		get_posts() -> [Post]
	}
	
	server blog_service implements BlogService {
		state {
			posts: [Post] = []
		}
		
		receive get_posts {} -> [Post] {
			return state.posts
		}
	}
	
	state {
		global_counter: number = 0
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, program.Statements, 4)

	// Verify struct
	assert.NotNil(t, program.Statements[0].StructDef)
	assert.Equal(t, "Post", program.Statements[0].StructDef.Name)

	// Verify protocol
	assert.NotNil(t, program.Statements[1].ProtocolDef)
	assert.Equal(t, "BlogService", program.Statements[1].ProtocolDef.Name)

	// Verify server
	assert.NotNil(t, program.Statements[2].ServerDef)
	assert.Equal(t, "blog_service", program.Statements[2].ServerDef.Name)

	// Verify global state
	assert.NotNil(t, program.Statements[3].StateDef)
}

// Test server parsing error cases
func TestParser_ServerDefinitionErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing_server_name", `server {}`},
		{"missing_opening_brace", `server test_service`},
		{"missing_closing_brace", `server test_service {`},
		{"invalid_server_name", `server 123_service {}`},
		{"invalid_protocol_name", `server test implements 123Protocol {}`},
		{"missing_implements_protocol", `server test implements {}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tt.input))
			assert.Error(t, err, "Expected parsing error for %s", tt.name)
		})
	}
}

// Test server case sensitivity
func TestParser_ServerDefinitionCaseSensitivity(t *testing.T) {
	input := `server BlogService implements IBlogService {
		receive GetPosts {} -> [Post] {
			return []
		}
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	require.NoError(t, err)

	serverDef := program.Statements[0].ServerDef
	require.NotNil(t, serverDef)

	// Server names and protocol names should preserve case
	assert.Equal(t, "BlogService", serverDef.Name)
	assert.Equal(t, "IBlogService", *serverDef.Protocol)

	receiveElement := serverDef.Body.Elements[0]
	assert.Equal(t, "GetPosts", receiveElement.Receive.Name)
}

// Benchmark server definition parsing
func BenchmarkParser_ServerParsing(b *testing.B) {
	input := `server benchmark_service implements BenchmarkService {
		state {
			items: [string] = [],
			count: number = 0,
			active: bool = true
		}
		
		receive get_items {} -> [string] {
			return state.items
		}
		
		receive add_item {item: string} -> bool {
			state.items = state.items.add(item)
			state.count = state.count + 1
			return true
		}
		
		receive get_stats {} -> object {
			return {
				count: state.count,
				active: state.active
			}
		}
	}`

	for i := 0; i < b.N; i++ {
		_, err := Parse("test.relay", strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
