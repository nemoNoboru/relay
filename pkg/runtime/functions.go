package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// evaluateFuncCall evaluates function calls
func (e *Evaluator) evaluateFuncCall(expr *parser.FuncCallExpr, env *Environment) (*Value, error) {
	// Look up the function
	funcValue, exists := env.Get(expr.Name)
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", expr.Name)
	}

	if funcValue.Type != ValueTypeFunction {
		return nil, fmt.Errorf("'%s' is not a function", expr.Name)
	}

	// Evaluate arguments
	args := make([]*Value, 0, len(expr.Args))
	for _, arg := range expr.Args {
		value, err := e.EvaluateWithEnv(arg, env)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}

	// Call the function
	if funcValue.Function.IsBuiltin {
		return funcValue.Function.Builtin(args)
	}

	// Handle user-defined functions
	return e.CallUserFunction(funcValue.Function, args, env)
}

// CallUserFunction calls a user-defined function
func (e *Evaluator) CallUserFunction(fn *Function, args []*Value, parentEnv *Environment) (*Value, error) {
	// Check parameter count
	if len(args) != len(fn.Parameters) {
		return nil, fmt.Errorf("function '%s' expects %d arguments, got %d", fn.Name, len(fn.Parameters), len(args))
	}

	// Create new environment for function execution
	var funcEnv *Environment
	if fn.ClosureEnv != nil {
		// Use closure environment as parent (for closures)
		funcEnv = NewEnvironment(fn.ClosureEnv)
	} else {
		// Use parent environment (for regular functions)
		funcEnv = NewEnvironment(parentEnv)
	}

	// Bind parameters
	for i, param := range fn.Parameters {
		funcEnv.Define(param, args[i])
	}

	// Execute function body
	result, err := e.evaluateBlock(fn.Body, funcEnv)
	if err != nil {
		// Check if it's a return value
		if returnVal, ok := err.(*ReturnValue); ok {
			return returnVal.Value, nil
		}
		return nil, err
	}

	return result, nil
}

// evaluateFunctionExpr evaluates function definitions
func (e *Evaluator) evaluateFunctionExpr(expr *parser.FunctionExpr, env *Environment) (*Value, error) {
	// Extract parameter names
	paramNames := make([]string, len(expr.Parameters))
	for i, param := range expr.Parameters {
		paramNames[i] = param.Name
	}

	// Create function value with closure environment
	function := &Function{
		Name:       "anonymous", // Default name
		Parameters: paramNames,
		Body:       expr.Body,
		IsBuiltin:  false,
		ClosureEnv: env, // Capture the current environment for closures
	}

	// If the function has a name, store it in the environment
	if expr.Name != nil {
		function.Name = *expr.Name
		funcValue := &Value{
			Type:     ValueTypeFunction,
			Function: function,
		}
		env.Define(*expr.Name, funcValue)
		return funcValue, nil
	}

	// Return anonymous function
	return &Value{
		Type:     ValueTypeFunction,
		Function: function,
	}, nil
}
