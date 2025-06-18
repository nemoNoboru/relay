package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// evaluateBinaryExpr evaluates binary expressions
func (e *Evaluator) evaluateBinaryExpr(expr *parser.BinaryExpr, env *Environment) (*Value, error) {
	// Evaluate the left operand
	left, err := e.evaluateUnaryExpr(expr.Left, env)
	if err != nil {
		return nil, err
	}

	// If there are no binary operations, just return the left operand
	if len(expr.Right) == 0 {
		return left, nil
	}

	// Process binary operations left to right
	result := left
	for _, op := range expr.Right {
		right, err := e.evaluateUnaryExpr(op.Right, env)
		if err != nil {
			return nil, err
		}

		result, err = e.applyBinaryOperation(result, op.Op, right)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evaluateUnaryExpr evaluates unary expressions
func (e *Evaluator) evaluateUnaryExpr(expr *parser.UnaryExpr, env *Environment) (*Value, error) {
	// Evaluate the primary expression first
	primary, err := e.evaluatePrimaryExpr(expr.Primary, env)
	if err != nil {
		return nil, err
	}

	// Apply unary operator if present
	if expr.Op != nil {
		switch *expr.Op {
		case "!":
			return NewBool(!primary.IsTruthy()), nil
		case "-":
			if primary.Type != ValueTypeNumber {
				return nil, fmt.Errorf("cannot negate non-number value")
			}
			return NewNumber(-primary.Number), nil
		default:
			return nil, fmt.Errorf("unsupported unary operator: %s", *expr.Op)
		}
	}

	return primary, nil
}

// evaluatePrimaryExpr evaluates primary expressions
func (e *Evaluator) evaluatePrimaryExpr(expr *parser.PrimaryExpr, env *Environment) (*Value, error) {
	// Evaluate the base expression
	base, err := e.evaluateBaseExpr(expr.Base, env)
	if err != nil {
		return nil, err
	}

	// Apply any method calls
	result := base
	for _, access := range expr.Access {
		if access.MethodCall != nil {
			result, err = e.evaluateMethodCall(result, access.MethodCall, env)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// evaluateBaseExpr evaluates base expressions
func (e *Evaluator) evaluateBaseExpr(expr *parser.BaseExpr, env *Environment) (*Value, error) {
	if expr.FuncCall != nil {
		return e.evaluateFuncCall(expr.FuncCall, env)
	}

	if expr.StructConstructor != nil {
		return e.evaluateStructConstructor(expr.StructConstructor, env)
	}

	if expr.ObjectLit != nil {
		return e.evaluateObjectLiteral(expr.ObjectLit, env)
	}

	if expr.SendExpr != nil {
		return e.evaluateSendExpr(expr.SendExpr, env)
	}

	if expr.Lambda != nil {
		return e.evaluateLambda(expr.Lambda, env)
	}

	if expr.Block != nil {
		return e.evaluateBlock(expr.Block, env)
	}

	if expr.Grouped != nil {
		return e.EvaluateWithEnv(expr.Grouped, env)
	}

	if expr.Literal != nil {
		return e.evaluateLiteral(expr.Literal, env)
	}

	if expr.Identifier != nil {
		value, exists := env.Get(*expr.Identifier)
		if !exists {
			return nil, fmt.Errorf("undefined variable: %s", *expr.Identifier)
		}
		return value, nil
	}

	return NewNil(), fmt.Errorf("unsupported base expression")
}

// evaluateObjectLiteral evaluates object literals
func (e *Evaluator) evaluateObjectLiteral(expr *parser.ObjectLit, env *Environment) (*Value, error) {
	obj := make(map[string]*Value)

	for _, field := range expr.Fields {
		value, err := e.EvaluateWithEnv(field.Value, env)
		if err != nil {
			return nil, err
		}

		// Convert symbol keys to strings (remove leading :)
		key := field.Key
		if len(key) > 0 && key[0] == ':' {
			key = key[1:]
		}

		obj[key] = value
	}

	return NewObject(obj), nil
}

// evaluateLambda evaluates lambda expressions
func (e *Evaluator) evaluateLambda(expr *parser.LambdaExpr, env *Environment) (*Value, error) {
	// Extract parameter names
	paramNames := make([]string, len(expr.Parameters))
	for i, param := range expr.Parameters {
		paramNames[i] = param.Name
	}

	// Create function value with closure environment
	function := &Function{
		Name:       "lambda",
		Parameters: paramNames,
		Body:       expr.Body,
		IsBuiltin:  false,
		ClosureEnv: env, // Capture the current environment for closures
	}

	return &Value{
		Type:     ValueTypeFunction,
		Function: function,
	}, nil
}

// evaluateSendExpr evaluates send expressions for server communication
func (e *Evaluator) evaluateSendExpr(expr *parser.SendExpr, env *Environment) (*Value, error) {
	// Find the server
	server, exists := e.GetServer(expr.Target)
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", expr.Target)
	}

	// Evaluate the arguments object
	args := make([]*Value, 0)
	if expr.Args != nil {
		// Extract values from the object literal as a list of arguments
		for _, field := range expr.Args.Fields {
			value, err := e.EvaluateWithEnv(field.Value, env)
			if err != nil {
				return nil, err
			}
			args = append(args, value)
		}
	}

	// Send message to server (synchronous call)
	result, err := server.Server.SendMessage(expr.Method, args, true)
	if err != nil {
		return nil, err
	}

	return result, nil
}
