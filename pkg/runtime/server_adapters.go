package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// EvaluatorExecutionEngine implements ExecutionEngine using the current evaluator
type EvaluatorExecutionEngine struct {
	evaluator *Evaluator
}

// NewEvaluatorExecutionEngine creates a new execution engine adapter
func NewEvaluatorExecutionEngine(evaluator *Evaluator) *EvaluatorExecutionEngine {
	return &EvaluatorExecutionEngine{
		evaluator: evaluator,
	}
}

// ExecuteFunction executes a function using the evaluator
func (e *EvaluatorExecutionEngine) ExecuteFunction(handler FunctionHandler, args []*Value, env *Environment) (*Value, error) {
	return handler.Execute(e, args, env)
}

// CreateEnvironment creates a new environment
func (e *EvaluatorExecutionEngine) CreateEnvironment(parent *Environment) *Environment {
	return NewEnvironment(parent)
}

// EvaluateDefaultValue evaluates a default value using the evaluator
func (e *EvaluatorExecutionEngine) EvaluateDefaultValue(expr interface{}, env *Environment) (*Value, error) {
	if literal, ok := expr.(*parser.Literal); ok {
		return e.evaluator.evaluateLiteral(literal, env)
	}
	return NewNil(), nil
}

// FunctionHandlerAdapter adapts the current Function type to FunctionHandler interface
type FunctionHandlerAdapter struct {
	function *Function
}

// NewFunctionHandlerAdapter creates a new function handler adapter
func NewFunctionHandlerAdapter(function *Function) *FunctionHandlerAdapter {
	return &FunctionHandlerAdapter{
		function: function,
	}
}

// GetName returns the function name
func (f *FunctionHandlerAdapter) GetName() string {
	return f.function.Name
}

// GetParameters returns the parameter names
func (f *FunctionHandlerAdapter) GetParameters() []string {
	return f.function.Parameters
}

// Execute executes the function using the execution engine
func (f *FunctionHandlerAdapter) Execute(engine ExecutionEngine, args []*Value, env *Environment) (*Value, error) {
	// For current implementation, we need to cast back to the evaluator
	if evalEngine, ok := engine.(*EvaluatorExecutionEngine); ok {
		// Create a modified function with the correct environment for state access
		modifiedFunction := &Function{
			Name:       f.function.Name,
			Parameters: f.function.Parameters,
			Body:       f.function.Body,
			IsBuiltin:  f.function.IsBuiltin,
			ClosureEnv: env,
		}

		return evalEngine.evaluator.CallUserFunction(modifiedFunction, args, env)
	}

	return NewNil(), fmt.Errorf("unsupported execution engine")
}

// EvaluatorServerRegistry implements ServerRegistry using the evaluator's server map
type EvaluatorServerRegistry struct {
	evaluator *Evaluator
}

// NewEvaluatorServerRegistry creates a new server registry adapter
func NewEvaluatorServerRegistry(evaluator *Evaluator) *EvaluatorServerRegistry {
	return &EvaluatorServerRegistry{
		evaluator: evaluator,
	}
}

// RegisterServer registers a server instance
func (r *EvaluatorServerRegistry) RegisterServer(name string, server *Value) {
	r.evaluator.servers[name] = server
}

// GetServer retrieves a server by name
func (r *EvaluatorServerRegistry) GetServer(name string) (*Value, bool) {
	server, exists := r.evaluator.servers[name]
	return server, exists
}

// StopAllServers stops all running servers
func (r *EvaluatorServerRegistry) StopAllServers() {
	for _, server := range r.evaluator.servers {
		if server.Server.Running {
			server.Server.Stop()
		}
	}
}

// EvaluatorServerFactory implements ServerFactory using the current evaluator
type EvaluatorServerFactory struct {
	evaluator *Evaluator
	registry  ServerRegistry
	engine    ExecutionEngine
}

// NewEvaluatorServerFactory creates a new server factory adapter
func NewEvaluatorServerFactory(evaluator *Evaluator, registry ServerRegistry, engine ExecutionEngine) *EvaluatorServerFactory {
	return &EvaluatorServerFactory{
		evaluator: evaluator,
		registry:  registry,
		engine:    engine,
	}
}

// CreateServer creates a server from a server definition
func (f *EvaluatorServerFactory) CreateServer(name string, definition ServerDefinition, env *Environment) (*Value, error) {
	// Initialize server state
	state := make(map[string]*Value)
	for fieldName, field := range definition.State {
		if field.DefaultValue != nil {
			val, err := f.engine.EvaluateDefaultValue(field.DefaultValue, env)
			if err != nil {
				return nil, fmt.Errorf("error evaluating default value for state field %s: %v", fieldName, err)
			}
			state[fieldName] = val
		} else {
			state[fieldName] = NewNil()
		}
	}

	// Create server core
	serverCore := NewServerCore(name, state, definition.Receivers, env, f.engine)

	// Wrap in Value type for compatibility
	serverValue := &Value{
		Type: ValueTypeServer,
		Server: &Server{
			Name:        serverCore.Name,
			State:       serverCore.State,
			Receivers:   convertHandlersToFunctions(definition.Receivers),
			MessageChan: serverCore.MessageChan,
			Running:     serverCore.Running,
			Environment: serverCore.Environment,
		},
	}

	// Register and start the server
	f.registry.RegisterServer(name, serverValue)
	serverCore.Start()

	return serverValue, nil
}

// Helper function to convert FunctionHandlers back to Functions for compatibility
func convertHandlersToFunctions(handlers map[string]FunctionHandler) map[string]*Function {
	functions := make(map[string]*Function)
	for name, handler := range handlers {
		if adapter, ok := handler.(*FunctionHandlerAdapter); ok {
			functions[name] = adapter.function
		}
	}
	return functions
}
