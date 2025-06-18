package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// evaluateLiteral evaluates literal values
func (e *Evaluator) evaluateLiteral(literal *parser.Literal, env *Environment) (*Value, error) {
	if literal.Number != nil {
		return NewNumber(*literal.Number), nil
	}

	if literal.String != nil {
		return NewString(*literal.String), nil
	}

	if literal.Symbol != nil {
		// Convert symbol :hello to string "hello" (remove the leading :)
		symbolValue := *literal.Symbol
		if len(symbolValue) > 0 && symbolValue[0] == ':' {
			symbolValue = symbolValue[1:]
		}
		return NewString(symbolValue), nil
	}

	if literal.Bool != nil {
		return NewBool(*literal.Bool == "true"), nil
	}

	if literal.Nil != nil {
		return NewNil(), nil
	}

	if literal.Array != nil {
		elements := make([]*Value, 0, len(literal.Array.Elements))
		for _, elem := range literal.Array.Elements {
			value, err := e.Evaluate(elem)
			if err != nil {
				return nil, err
			}
			elements = append(elements, value)
		}
		return NewArray(elements), nil
	}

	if literal.FuncCall != nil {
		return e.evaluateLiteralFuncCall(literal.FuncCall, env)
	}

	return NewNil(), nil
}

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
