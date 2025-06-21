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
	dispatcher.RegisterHandler(ValueTypeServerState, NewServerStateMethodHandler())

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

// Note: The core evaluation logic has been consolidated in core.go for better organization.
// This file now focuses on evaluator initialization and high-level coordination.
