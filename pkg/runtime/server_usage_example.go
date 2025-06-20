package runtime

import (
	"fmt"
)

// Example showing how to use the new server abstractions

// MigrateEvaluatorToAbstractedServers shows how to migrate from the old tightly-coupled
// server implementation to the new abstracted version
func MigrateEvaluatorToAbstractedServers(evaluator *Evaluator) {
	// Create the abstraction layer components
	engine := NewEvaluatorExecutionEngine(evaluator)
	registry := NewEvaluatorServerRegistry(evaluator)
	factory := NewEvaluatorServerFactory(evaluator, registry, engine)

	// Now you can create servers using the abstracted interface
	// instead of calling evaluateServerExpr directly

	// Example: Create a simple counter server
	counterDefinition := ServerDefinition{
		Name: "counter",
		State: map[string]StateField{
			"count": {
				Name:         "count",
				DefaultValue: nil, // Will default to nil
			},
		},
		Receivers: map[string]FunctionHandler{
			"increment": &SimpleIncrementHandler{},
			"get_count": &SimpleGetCountHandler{},
		},
	}

	env := NewEnvironment(nil)
	serverValue, err := factory.CreateServer("counter", counterDefinition, env)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		return
	}

	// Server is now running and can receive messages
	fmt.Printf("Created server: %s\n", serverValue.String())
}

// Example FunctionHandler implementations that are decoupled from parser AST

type SimpleIncrementHandler struct{}

func (h *SimpleIncrementHandler) GetName() string {
	return "increment"
}

func (h *SimpleIncrementHandler) GetParameters() []string {
	return []string{} // No parameters
}

func (h *SimpleIncrementHandler) Execute(engine ExecutionEngine, args []*Value, env *Environment) (*Value, error) {
	// Get current count from state
	state, exists := env.Get("state")
	if !exists {
		return NewNil(), fmt.Errorf("state not available")
	}

	// This is server state, so we can use .get/.set methods
	currentCount, err := callMethod(state, "get", []*Value{NewString("count")})
	if err != nil {
		return NewNil(), err
	}

	// Increment
	newCount := NewNumber(currentCount.Number + 1)

	// Set new value
	_, err = callMethod(state, "set", []*Value{NewString("count"), newCount})
	if err != nil {
		return NewNil(), err
	}

	return newCount, nil
}

type SimpleGetCountHandler struct{}

func (h *SimpleGetCountHandler) GetName() string {
	return "get_count"
}

func (h *SimpleGetCountHandler) GetParameters() []string {
	return []string{} // No parameters
}

func (h *SimpleGetCountHandler) Execute(engine ExecutionEngine, args []*Value, env *Environment) (*Value, error) {
	// Get current count from state
	state, exists := env.Get("state")
	if !exists {
		return NewNil(), fmt.Errorf("state not available")
	}

	// Get the count value
	count, err := callMethod(state, "get", []*Value{NewString("count")})
	if err != nil {
		return NewNil(), err
	}

	return count, nil
}

// Helper function to call methods (you'd need to implement this based on your method system)
func callMethod(obj *Value, method string, args []*Value) (*Value, error) {
	// This would call your existing method dispatch system
	// For now, just a stub
	return NewNil(), fmt.Errorf("method calling not implemented in example")
}

// NewServerSystem creates a complete server system with all abstractions in place
func NewServerSystem() (*EvaluatorExecutionEngine, ServerRegistry, ServerFactory) {
	evaluator := NewEvaluator()
	engine := NewEvaluatorExecutionEngine(evaluator)
	registry := NewEvaluatorServerRegistry(evaluator)
	factory := NewEvaluatorServerFactory(evaluator, registry, engine)

	return engine, registry, factory
}

// Example showing how the new system allows you to work on server internals
// without worrying about parser/evaluator details
func ExampleServerInternalsWork() {
	_, _, factory := NewServerSystem()

	// Create a server with custom behavior
	definition := ServerDefinition{
		Name:      "custom_server",
		State:     map[string]StateField{},
		Receivers: map[string]FunctionHandler{},
	}

	env := NewEnvironment(nil)
	server, _ := factory.CreateServer("custom", definition, env)

	// Now you can work with server internals without parser dependencies
	fmt.Printf("Server created: %s\n", server.String())

	// Example of using a custom execution engine for specialized server behavior
	customEngine := &CustomExecutionEngine{}
	fmt.Printf("Custom engine available: %v\n", customEngine != nil)
}

// CustomExecutionEngine shows how you can create alternative execution engines
type CustomExecutionEngine struct{}

func (e *CustomExecutionEngine) ExecuteFunction(handler FunctionHandler, args []*Value, env *Environment) (*Value, error) {
	// Custom execution logic - could be for HTTP servers, Go actors, etc.
	fmt.Printf("Custom execution of: %s\n", handler.GetName())
	return NewString("custom_result"), nil
}

func (e *CustomExecutionEngine) CreateEnvironment(parent *Environment) *Environment {
	return NewEnvironment(parent)
}

func (e *CustomExecutionEngine) EvaluateDefaultValue(expr interface{}, env *Environment) (*Value, error) {
	// Custom default value evaluation
	return NewNil(), nil
}
