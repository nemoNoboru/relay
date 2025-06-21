// Package runtime/evaluator provides the main evaluation engine for Relay.
// This file contains the Evaluator type and its initialization, while the actual
// evaluation logic has been consolidated in core.go for better organization.
package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// ReturnValue represents a return statement using Go's error mechanism for control flow.
// This allows early returns from functions and blocks without complex control structures.
type ReturnValue struct {
	Value *Value // The value being returned
}

func (r ReturnValue) Error() string {
	return "return"
}

// NewReturn creates a new return value wrapper for early return handling.
func NewReturn(value *Value) *ReturnValue {
	return &ReturnValue{Value: value}
}

// Evaluator is the main execution engine for Relay programs.
// It maintains global state including environments, struct definitions, and running servers.
// The actual evaluation logic is implemented in core.go for better organization.
type Evaluator struct {
	globalEnv        *Environment                 // Global variable environment
	structDefs       map[string]*StructDefinition // Registered struct types
	servers          map[string]*Value            // Running server instances
	methodDispatcher MethodDispatcher             // Method dispatch system
}

// NewEvaluator creates a new evaluator with built-in functions and empty state.
// Initializes the global environment and registers all built-in functions.
func NewEvaluator() *Evaluator {
	global := NewEnvironment(nil)

	eval := &Evaluator{
		globalEnv:  global,
		structDefs: make(map[string]*StructDefinition),
		servers:    make(map[string]*Value),
	}

	// Initialize method dispatcher with all type handlers
	eval.methodDispatcher = eval.newMethodDispatcherWithAllHandlers()

	eval.defineBuiltins()

	return eval
}

// ExecuteFunction implements FunctionExecutor interface for higher-order array methods
func (e *Evaluator) ExecuteFunction(fn *Value, args []*Value) (*Value, error) {
	if fn.Type != ValueTypeFunction {
		return nil, fmt.Errorf("cannot call non-function value")
	}

	// Use the evaluator's function calling mechanism
	return e.callFunction(fn, args, e.globalEnv)
}

// newMethodDispatcherWithAllHandlers creates a method dispatcher with all built-in handlers
func (e *Evaluator) newMethodDispatcherWithAllHandlers() MethodDispatcher {
	dispatcher := NewMethodDispatcher()

	// Create array handler and set function executor
	arrayHandler := NewArrayMethodHandler()
	arrayHandler.SetFunctionExecutor(e) // evaluator implements FunctionExecutor

	// Register all built-in type handlers
	dispatcher.RegisterHandler(ValueTypeArray, arrayHandler)
	dispatcher.RegisterHandler(ValueTypeObject, NewObjectMethodHandler())
	dispatcher.RegisterHandler(ValueTypeStruct, NewStructMethodHandler())
	dispatcher.RegisterHandler(ValueTypeString, NewStringMethodHandler())

	return dispatcher
}

// Evaluate evaluates an expression in the global environment.
// This is the main entry point for evaluating top-level expressions.
func (e *Evaluator) Evaluate(expr *parser.Expression) (*Value, error) {
	return e.EvaluateWithEnv(expr, e.globalEnv)
}

// EvaluateWithEnv evaluates an expression with a specific environment.
// Used for function calls, block scoping, and other context-sensitive evaluations.
// Delegates to the unified evaluation engine in core.go.
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error) {
	return e.EvaluateExpression(expr, env)
}

// GetAllServers returns all registered servers (thread-safe copy)
func (e *Evaluator) GetAllServers() map[string]*Value {
	result := make(map[string]*Value)
	for name, server := range e.servers {
		result[name] = server
	}
	return result
}

// GetServer returns a registered server by name.
func (e *Evaluator) GetServer(name string) (*Value, bool) {
	server, exists := e.servers[name]
	return server, exists
}

// StopAllServers stops all running servers managed by the evaluator.
func (e *Evaluator) StopAllServers() {
	for _, serverVal := range e.servers {
		if serverVal.Type == ValueTypeServer && serverVal.Server.Running {
			serverVal.Server.Stop()
		}
	}
}

// RegisterServer adds a server to the evaluator's registry.
// This is useful for testing or for manually adding servers.
func (e *Evaluator) RegisterServer(name string, server *Value) {
	e.servers[name] = server
}

// Note: The core evaluation logic has been consolidated in core.go for better organization.
// This file now focuses on evaluator initialization and high-level coordination.

func (e *Evaluator) evaluateServerExpr(expr *parser.ServerExpr, env *Environment) (*Value, error) {
	// Extract state and receive functions from the server body
	state := make(map[string]*Value)
	receives := make(map[string]*Function)
	serverEnv := NewEnvironment(env)

	if expr.Body != nil {
		for _, element := range expr.Body.Elements {
			if element.State != nil {
				for _, field := range element.State.Fields {
					var val *Value
					if field.DefaultValue != nil {
						// For now, we'll just handle simple literals.
						// A full implementation would need to evaluate expressions.
						if field.DefaultValue.Number != nil {
							val = NewNumber(*field.DefaultValue.Number)
						} else if field.DefaultValue.String != nil {
							val = NewString(*field.DefaultValue.String)
						} else {
							val = NewNil()
						}
					} else {
						val = NewNil()
					}
					state[field.Name] = val
				}
			}
			if element.Receive != nil {
				params := make([]string, len(element.Receive.Parameters))
				for i, p := range element.Receive.Parameters {
					params[i] = p.Name
				}
				fn := &Function{
					Name:       element.Receive.Name,
					Parameters: params,
					Body:       element.Receive.Body,
					ClosureEnv: serverEnv,
				}
				receives[element.Receive.Name] = fn
			}
		}
	}

	// Create a new server instance
	server := NewServer(expr.Name, state, receives, serverEnv)

	// Register the server with the evaluator
	e.servers[expr.Name] = server

	// Server definitions don't return a value, they register a server
	return NewNil(), nil
}
