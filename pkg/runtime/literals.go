package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// Note: evaluateLiteral moved to core.go

// evaluateLiteralFuncCall evaluates function calls from literals
func (e *Evaluator) evaluateLiteralFuncCall(funcCall *parser.FuncCall, env *Environment) (*Value, error) {
	// Look up the function
	funcValue, exists := env.Get(funcCall.Name)
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", funcCall.Name)
	}

	if funcValue.Type != ValueTypeFunction {
		return nil, fmt.Errorf("'%s' is not a function", funcCall.Name)
	}

	// Evaluate arguments (these are now expressions, so use EvaluateWithEnv)
	args := make([]*Value, 0, len(funcCall.Args))
	for _, arg := range funcCall.Args {
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
