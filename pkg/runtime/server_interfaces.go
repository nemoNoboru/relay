package runtime

// ExecutionEngine defines the interface for executing functions within the server context.
// This abstracts away the specific implementation details of the evaluator and parser.
type ExecutionEngine interface {
	// ExecuteFunction executes a function with given arguments in the provided environment
	ExecuteFunction(handler FunctionHandler, args []*Value, env *Environment) (*Value, error)

	// CreateEnvironment creates a new environment for server execution
	CreateEnvironment(parent *Environment) *Environment

	// EvaluateDefaultValue evaluates a default value expression
	EvaluateDefaultValue(expr interface{}, env *Environment) (*Value, error)
}

// FunctionHandler represents an abstract function that can be executed by the server.
// This replaces the direct dependency on *Function with parser.Block
type FunctionHandler interface {
	// GetName returns the function name
	GetName() string

	// GetParameters returns the parameter names
	GetParameters() []string

	// Execute executes the function with the given execution engine
	Execute(engine ExecutionEngine, args []*Value, env *Environment) (*Value, error)
}

// ServerFactory creates servers without requiring direct access to parser types
type ServerFactory interface {
	// CreateServer creates a server from a server definition
	CreateServer(name string, definition ServerDefinition, env *Environment) (*Value, error)
}

// ServerDefinition represents a server definition in an abstract way,
// without depending on specific parser types
type ServerDefinition struct {
	Name      string
	State     map[string]StateField
	Receivers map[string]FunctionHandler
}

// StateField represents a state field definition
type StateField struct {
	Name         string
	DefaultValue interface{} // Abstract representation of default value
}

// ServerRegistry manages server instances independently of the evaluator
type ServerRegistry interface {
	// RegisterServer registers a server instance
	RegisterServer(name string, server *Value)

	// GetServer retrieves a server by name
	GetServer(name string) (*Value, bool)

	// StopAllServers stops all running servers
	StopAllServers()
}
