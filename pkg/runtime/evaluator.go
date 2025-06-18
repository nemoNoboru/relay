package runtime

import (
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
	return e.EvaluateExpression(expr, env)
}

// Note: These methods have been moved to core.go for consolidation
