package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// ReturnValue represents a return statement (using error mechanism for control flow)
type ReturnValue struct {
	Value *Value
}

func (r ReturnValue) Error() string {
	return "return"
}

// NewReturn creates a new return value
func NewReturn(value *Value) *ReturnValue {
	return &ReturnValue{Value: value}
}

// Evaluator executes parsed AST nodes
type Evaluator struct {
	globalEnv  *Environment
	structDefs map[string]*StructDefinition // Store struct definitions
	servers    map[string]*Value            // Store running servers
}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	global := NewEnvironment(nil)

	eval := &Evaluator{
		globalEnv:  global,
		structDefs: make(map[string]*StructDefinition),
		servers:    make(map[string]*Value),
	}
	eval.defineBuiltins()

	return eval
}

// Evaluate evaluates an expression
func (e *Evaluator) Evaluate(expr *parser.Expression) (*Value, error) {
	return e.EvaluateWithEnv(expr, e.globalEnv)
}

// EvaluateWithEnv evaluates an expression with a specific environment
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error) {
	if expr == nil {
		return NewNil(), nil
	}

	// Handle different expression types
	if expr.StructExpr != nil {
		return e.evaluateStructExpr(expr.StructExpr, env)
	}

	if expr.ServerExpr != nil {
		return e.evaluateServerExpr(expr.ServerExpr, env)
	}

	if expr.Binary != nil {
		return e.evaluateBinaryExpr(expr.Binary, env)
	}

	if expr.SetExpr != nil {
		return e.evaluateSetExpr(expr.SetExpr, env)
	}

	if expr.FunctionExpr != nil {
		return e.evaluateFunctionExpr(expr.FunctionExpr, env)
	}

	if expr.ReturnExpr != nil {
		return e.evaluateReturnExpr(expr.ReturnExpr, env)
	}

	return NewNil(), fmt.Errorf("unsupported expression type")
}

// evaluateReturnExpr handles return statements
func (e *Evaluator) evaluateReturnExpr(expr *parser.ReturnExpr, env *Environment) (*Value, error) {
	value, err := e.EvaluateWithEnv(expr.Value, env)
	if err != nil {
		return nil, err
	}

	// Use the error mechanism to implement early return
	return nil, NewReturn(value)
}

// evaluateBlock evaluates a block of expressions
func (e *Evaluator) evaluateBlock(block *parser.Block, env *Environment) (*Value, error) {
	var result *Value = NewNil()

	for _, expr := range block.Expressions {
		value, err := e.EvaluateWithEnv(expr, env)
		if err != nil {
			// Check if it's a return value
			if returnVal, ok := err.(*ReturnValue); ok {
				return returnVal.Value, nil
			}
			return nil, err
		}
		result = value
	}

	return result, nil
}

// evaluateSetExpr evaluates set expressions (variable assignment)
func (e *Evaluator) evaluateSetExpr(expr *parser.SetExpr, env *Environment) (*Value, error) {
	value, err := e.EvaluateWithEnv(expr.Value, env)
	if err != nil {
		return nil, err
	}

	env.Define(expr.Variable, value)
	return value, nil
}
