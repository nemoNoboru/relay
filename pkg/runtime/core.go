// Package runtime provides the core evaluation engine for the Relay programming language.
// This file contains the unified expression evaluation system that handles all language constructs
// including functions, structs, servers, arithmetic, logic, and method calls.
//
// The evaluation follows a hierarchical pattern:
// Expression -> Binary -> Unary -> Primary -> Base -> Literals/Identifiers
//
// Key features:
// - Immutable-by-default semantics with method-based operations
// - First-class functions with closure support
// - Struct definitions and instantiation
// - Server state management with automatic persistence
// - Short-circuit evaluation for logical operators
// - Comprehensive error handling and type checking
package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// EvaluateExpression is the unified entry point for all expression evaluation.
// It dispatches to appropriate handlers based on the expression type.
// This replaces the previous fragmented approach across multiple files.
func (e *Evaluator) EvaluateExpression(expr *parser.Expression, env *Environment) (*Value, error) {
	if expr == nil {
		return NewNil(), nil
	}

	// Handle top-level expression types
	switch {
	case expr.StructExpr != nil:
		return e.evaluateStructExpr(expr.StructExpr, env)
	case expr.ServerExpr != nil:
		return e.evaluateServerExpr(expr.ServerExpr, env)
	case expr.FunctionExpr != nil:
		return e.evaluateFunction(expr.FunctionExpr, env)
	case expr.SetExpr != nil:
		return e.evaluateSet(expr.SetExpr, env)
	case expr.ReturnExpr != nil:
		return e.evaluateReturn(expr.ReturnExpr, env)
	case expr.IfExpr != nil:
		return e.evaluateIf(expr.IfExpr, env)
	case expr.Binary != nil:
		return e.evaluateBinary(expr.Binary, env)
	default:
		return NewNil(), fmt.Errorf("unsupported expression type")
	}
}

// evaluateBinary handles all binary operations including arithmetic (+, -, *, /),
// logical (&&, ||), comparisons (==, !=, <, >, <=, >=), and null coalescing (??).
// Implements short-circuit evaluation for logical operators for efficiency.
func (e *Evaluator) evaluateBinary(expr *parser.BinaryExpr, env *Environment) (*Value, error) {
	left, err := e.evaluateUnary(expr.Left, env)
	if err != nil {
		return nil, err
	}

	// Handle short-circuit evaluation for logical operators
	for _, op := range expr.Right {
		switch op.Op {
		case "&&":
			if !left.IsTruthy() {
				return NewBool(false), nil
			}
			right, err := e.evaluateUnary(op.Right, env)
			if err != nil {
				return nil, err
			}
			left = NewBool(right.IsTruthy())
		case "||":
			if left.IsTruthy() {
				return NewBool(true), nil
			}
			right, err := e.evaluateUnary(op.Right, env)
			if err != nil {
				return nil, err
			}
			left = NewBool(right.IsTruthy())
		case "??":
			if left.Type != ValueTypeNil {
				return left, nil
			}
			right, err := e.evaluateUnary(op.Right, env)
			if err != nil {
				return nil, err
			}
			left = right
		default:
			// Regular binary operations
			right, err := e.evaluateUnary(op.Right, env)
			if err != nil {
				return nil, err
			}
			left, err = e.applyBinaryOperation(left, op.Op, right)
			if err != nil {
				return nil, err
			}
		}
	}

	return left, nil
}

// evaluateUnary handles unary operations: logical NOT (!) and numeric negation (-).
// Validates type compatibility for operations (e.g., negation only works on numbers).
func (e *Evaluator) evaluateUnary(expr *parser.UnaryExpr, env *Environment) (*Value, error) {
	primary, err := e.evaluatePrimary(expr.Primary, env)
	if err != nil {
		return nil, err
	}

	if expr.Op == nil || *expr.Op == "" {
		return primary, nil
	}

	switch *expr.Op {
	case "!":
		return NewBool(!primary.IsTruthy()), nil
	case "-":
		if primary.Type != ValueTypeNumber {
			return nil, fmt.Errorf("unary minus requires number, got %s", primary.Type)
		}
		return NewNumber(-primary.Number), nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", *expr.Op)
	}
}

// evaluatePrimary handles primary expressions and chains of access operations.
// Supports method calls (.method()), field access (.field), and function calls.
// Enables method chaining like obj.set("key", value).get("other").method().
func (e *Evaluator) evaluatePrimary(expr *parser.PrimaryExpr, env *Environment) (*Value, error) {
	base, err := e.evaluateBase(expr.Base, env)
	if err != nil {
		return nil, err
	}

	// Apply access operations (method calls, field access)
	for _, access := range expr.Access {
		switch {
		case access.MethodCall != nil:
			args, err := e.evaluateArguments(access.MethodCall.Args, env)
			if err != nil {
				return nil, err
			}
			base, err = e.methodDispatcher.CallMethod(base, access.MethodCall.Method, args)
			if err != nil {
				return nil, err
			}
		case access.FieldAccess != nil:
			base, err = e.evaluateFieldAccess(base, access.FieldAccess, env)
			if err != nil {
				return nil, err
			}
		case access.FuncCall != nil:
			// This is a function call on the base value
			args, err := e.evaluateArguments(access.FuncCall.Args, env)
			if err != nil {
				return nil, err
			}
			base, err = e.callFunction(base, args, env)
			if err != nil {
				return nil, err
			}
		}
	}

	return base, nil
}

// evaluateBase handles the fundamental building blocks of expressions:
// literals (numbers, strings, booleans), identifiers (variables), function calls,
// struct constructors, object literals, server send operations, and grouped expressions.
func (e *Evaluator) evaluateBase(expr *parser.BaseExpr, env *Environment) (*Value, error) {
	switch {
	case expr.Literal != nil:
		return e.evaluateLiteral(expr.Literal, env)
	case expr.Identifier != nil:
		return e.evaluateIdentifier(*expr.Identifier, env)
	case expr.FuncCall != nil:
		return e.evaluateFunctionCall(expr.FuncCall, env)
	case expr.StructConstructor != nil:
		return e.evaluateStructConstructor(expr.StructConstructor, env)
	case expr.ObjectLit != nil:
		return e.evaluateObjectLiteral(expr.ObjectLit, env)
	case expr.SendExpr != nil:
		return e.evaluateSendExpr(expr.SendExpr, env)
	case expr.Grouped != nil:
		return e.EvaluateExpression(expr.Grouped, env)
	default:
		return NewNil(), fmt.Errorf("unsupported base expression")
	}
}

// evaluateLiteral handles all literal values: numbers, strings, booleans, symbols,
// arrays, and embedded function calls. Converts parser literals to runtime values.
func (e *Evaluator) evaluateLiteral(lit *parser.Literal, env *Environment) (*Value, error) {
	switch {
	case lit.Number != nil:
		return NewNumber(*lit.Number), nil
	case lit.String != nil:
		return NewString(*lit.String), nil
	case lit.Bool != nil:
		boolVal := lit.GetBoolValue()
		if boolVal != nil {
			return NewBool(*boolVal), nil
		}
		return NewBool(false), nil
	case lit.Symbol != nil:
		return NewString(*lit.Symbol), nil
	case lit.Array != nil:
		return e.evaluateArrayLiteral(lit.Array.Elements, env)
	case lit.FuncCall != nil:
		// Convert FuncCall to FuncCallExpr
		funcCallExpr := &parser.FuncCallExpr{
			Name: lit.FuncCall.Name,
			Args: lit.FuncCall.Args,
		}
		return e.evaluateFunctionCall(funcCallExpr, env)
	default:
		return NewNil(), nil
	}
}

// evaluateArrayLiteral creates an array from literal elements.
// Evaluates each element expression and collects them into a runtime array value.
func (e *Evaluator) evaluateArrayLiteral(elements []*parser.Expression, env *Environment) (*Value, error) {
	values := make([]*Value, len(elements))
	for i, element := range elements {
		value, err := e.EvaluateExpression(element, env)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return NewArray(values), nil
}

// evaluateObjectLiteral creates an object from key-value pairs.
// Supports both string keys and symbol keys (:symbol syntax).
func (e *Evaluator) evaluateObjectLiteral(obj *parser.ObjectLit, env *Environment) (*Value, error) {
	fields := make(map[string]*Value)
	for _, field := range obj.Fields {
		value, err := e.EvaluateExpression(field.Value, env)
		if err != nil {
			return nil, err
		}
		fields[field.Key] = value
	}
	return NewObject(fields), nil
}

// evaluateIdentifier looks up a variable in the environment chain.
// Returns an error if the variable is not defined in any scope.
func (e *Evaluator) evaluateIdentifier(name string, env *Environment) (*Value, error) {
	value, exists := env.Get(name)
	if !exists {
		return nil, fmt.Errorf("undefined variable: %s", name)
	}
	return value, nil
}

// evaluateFunctionCall handles function calls by looking up the function,
// evaluating arguments, and delegating to callFunction. Provides specific
// error messages for undefined functions vs undefined variables.
func (e *Evaluator) evaluateFunctionCall(call *parser.FuncCallExpr, env *Environment) (*Value, error) {
	// Get the function
	function, err := e.evaluateIdentifier(call.Name, env)
	if err != nil {
		// Provide more specific error message for function calls
		return nil, fmt.Errorf("undefined function: %s", call.Name)
	}

	// Evaluate arguments
	args, err := e.evaluateArguments(call.Args, env)
	if err != nil {
		return nil, err
	}

	// Call the function
	return e.callFunction(function, args, env)
}

// evaluateArguments evaluates a list of argument expressions in order.
// Used for function calls, method calls, and other operations requiring arguments.
func (e *Evaluator) evaluateArguments(args []*parser.Expression, env *Environment) ([]*Value, error) {
	values := make([]*Value, len(args))
	for i, arg := range args {
		value, err := e.EvaluateExpression(arg, env)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// callFunction handles both built-in and user-defined function calls.
// Supports closures by preserving captured environments and manages parameter binding.
// Validates argument count and creates appropriate execution environments.
func (e *Evaluator) callFunction(function *Value, args []*Value, env *Environment) (*Value, error) {
	if function.Type != ValueTypeFunction {
		return nil, fmt.Errorf("cannot call non-function value of type %s", function.Type)
	}

	fn := function.Function

	// Handle built-in functions
	if fn.IsBuiltin {
		return fn.Builtin(args)
	}

	// Check parameter count
	if len(args) != len(fn.Parameters) {
		return nil, fmt.Errorf("function '%s' expects %d arguments, got %d", fn.Name, len(fn.Parameters), len(args))
	}

	// Create function environment
	var funcEnv *Environment
	if fn.ClosureEnv != nil {
		// For closures, use captured environment as parent
		funcEnv = NewEnvironment(fn.ClosureEnv)
	} else {
		// For regular functions, use calling environment as parent
		funcEnv = NewEnvironment(env)
	}

	// Bind parameters
	for i, param := range fn.Parameters {
		funcEnv.Define(param, args[i])
	}

	// Execute function body
	return e.evaluateBlock(fn.Body, funcEnv)
}

// evaluateBlock executes a block of expressions sequentially.
// Returns the value of the last expression. Handles early returns via error mechanism.
func (e *Evaluator) evaluateBlock(block *parser.Block, env *Environment) (*Value, error) {
	var result *Value = NewNil()

	for _, expr := range block.Expressions {
		value, err := e.EvaluateExpression(expr, env)
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

// evaluateSet handles variable assignment (set variable = value).
// Evaluates the value expression and defines it in the current environment.
func (e *Evaluator) evaluateSet(expr *parser.SetExpr, env *Environment) (*Value, error) {
	value, err := e.EvaluateExpression(expr.Value, env)
	if err != nil {
		return nil, err
	}
	env.Define(expr.Variable, value)
	return value, nil
}

// evaluateReturn handles return statements by evaluating the return value
// and wrapping it in a special error type for early return handling.
func (e *Evaluator) evaluateReturn(expr *parser.ReturnExpr, env *Environment) (*Value, error) {
	value, err := e.EvaluateExpression(expr.Value, env)
	if err != nil {
		return nil, err
	}
	return nil, NewReturn(value)
}

// evaluateIf handles if-else expressions by evaluating the condition
// and executing the appropriate branch. Returns nil if no else branch and condition is false.
func (e *Evaluator) evaluateIf(expr *parser.IfExpr, env *Environment) (*Value, error) {
	condition, err := e.EvaluateExpression(expr.Condition, env)
	if err != nil {
		return nil, err
	}

	if condition.IsTruthy() {
		return e.evaluateBlock(expr.ThenBlock, env)
	} else if expr.ElseBlock != nil {
		return e.evaluateBlock(expr.ElseBlock, env)
	}

	return NewNil(), nil
}

// evaluateFunction handles function definitions, creating closures that capture
// the current environment. Supports both named and anonymous functions.
// Named functions are automatically defined in the current environment.
func (e *Evaluator) evaluateFunction(expr *parser.FunctionExpr, env *Environment) (*Value, error) {
	// Extract parameter names
	params := make([]string, len(expr.Parameters))
	for i, param := range expr.Parameters {
		params[i] = param.Name
	}

	// Get function name (could be nil for anonymous functions)
	var name string
	if expr.Name != nil {
		name = *expr.Name
	}

	// Create function value with closure environment
	function := &Function{
		Name:       name,
		Parameters: params,
		Body:       expr.Body,
		IsBuiltin:  false,
		ClosureEnv: env, // Capture current environment for closures
	}

	value := &Value{
		Type:     ValueTypeFunction,
		Function: function,
	}

	// If it's a named function, define it in the environment
	if expr.Name != nil && *expr.Name != "" {
		env.Define(*expr.Name, value)
	}

	return value, nil
}

// applyBinaryOperation applies a binary operator to two values.
// Dispatches to specific operation handlers and provides unified error handling.
func (e *Evaluator) applyBinaryOperation(left *Value, op string, right *Value) (*Value, error) {
	switch op {
	case "+":
		return e.applyAddition(left, right)
	case "-":
		return e.applySubtraction(left, right)
	case "*":
		return e.applyMultiplication(left, right)
	case "/":
		return e.applyDivision(left, right)
	case "==":
		return NewBool(left.IsEqual(right)), nil
	case "!=":
		return NewBool(!left.IsEqual(right)), nil
	case "<":
		return e.applyComparison(left, right, "<")
	case "<=":
		return e.applyComparison(left, right, "<=")
	case ">":
		return e.applyComparison(left, right, ">")
	case ">=":
		return e.applyComparison(left, right, ">=")
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", op)
	}
}

// applyAddition handles addition and string concatenation.
// Delegates to operations.go for the actual implementation.
func (e *Evaluator) applyAddition(left, right *Value) (*Value, error) {
	return e.add(left, right)
}

// applySubtraction handles numeric subtraction.
// Delegates to operations.go for the actual implementation.
func (e *Evaluator) applySubtraction(left, right *Value) (*Value, error) {
	return e.subtract(left, right)
}

// applyMultiplication handles numeric multiplication.
// Delegates to operations.go for the actual implementation.
func (e *Evaluator) applyMultiplication(left, right *Value) (*Value, error) {
	return e.multiply(left, right)
}

// applyDivision handles numeric division with zero-checking.
// Delegates to operations.go for the actual implementation.
func (e *Evaluator) applyDivision(left, right *Value) (*Value, error) {
	return e.divide(left, right)
}

// applyComparison handles numeric comparison operations (<, <=, >, >=).
// Delegates to operations.go for the actual implementation.
func (e *Evaluator) applyComparison(left, right *Value, op string) (*Value, error) {
	switch op {
	case "<":
		return e.less(left, right)
	case "<=":
		return e.lessEqual(left, right)
	case ">":
		return e.greater(left, right)
	case ">=":
		return e.greaterEqual(left, right)
	default:
		return nil, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// evaluateFieldAccess evaluates direct field access (obj.field).
// Supports both object and struct field access with appropriate error handling.
func (e *Evaluator) evaluateFieldAccess(object *Value, access *parser.FieldAccess, env *Environment) (*Value, error) {
	switch object.Type {
	case ValueTypeObject:
		// Direct field access on objects
		if value, exists := object.Object[access.Field]; exists {
			return value, nil
		}
		return NewNil(), nil // Return nil for non-existent fields

	case ValueTypeStruct:
		// Direct field access on structs
		if value, exists := object.Struct.Fields[access.Field]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("field '%s' not found in struct '%s'", access.Field, object.Struct.Name)

	default:
		return nil, fmt.Errorf("cannot access field '%s' on %s", access.Field, object.Type)
	}
}

func FromNative(v interface{}) (*Value, error) {
	switch val := v.(type) {
	case nil:
		return NewNil(), nil
	case bool:
		return NewBool(val), nil
	case int:
		return NewNumber(float64(val)), nil
	case int8:
		return NewNumber(float64(val)), nil
	case int16:
		return NewNumber(float64(val)), nil
	case int32:
		return NewNumber(float64(val)), nil
	case int64:
		return NewNumber(float64(val)), nil
	case uint:
		return NewNumber(float64(val)), nil
	case uint8:
		return NewNumber(float64(val)), nil
	case uint16:
		return NewNumber(float64(val)), nil
	case uint32:
		return NewNumber(float64(val)), nil
	case uint64:
		return NewNumber(float64(val)), nil
	case float32:
		return NewNumber(float64(val)), nil
	case float64:
		return NewNumber(val), nil
	case string:
		return NewString(val), nil
	default:
		return nil, fmt.Errorf("unsupported native type: %T", v)
	}
}
