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

// NewReturnValue creates a new return value
func NewReturn(value *Value) *ReturnValue {
	return &ReturnValue{Value: value}
}

// Evaluator executes parsed AST nodes
type Evaluator struct {
	globalEnv  *Environment
	structDefs map[string]*StructDefinition // Store struct definitions
}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	global := NewEnvironment(nil)

	// Add some built-in functions
	eval := &Evaluator{
		globalEnv:  global,
		structDefs: make(map[string]*StructDefinition),
	}
	eval.defineBuiltins()

	return eval
}

// defineBuiltins adds built-in functions to the global environment
func (e *Evaluator) defineBuiltins() {
	// Add print function
	printFunc := &Value{
		Type: ValueTypeFunction,
		Function: &Function{
			Name:      "print",
			IsBuiltin: true,
			Builtin: func(args []*Value) (*Value, error) {
				for i, arg := range args {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(arg.String())
				}
				fmt.Println()
				return NewNil(), nil
			},
		},
	}
	e.globalEnv.Define("print", printFunc)
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

	// Handle method calls/property access
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

// evaluateBaseExpr evaluates base expressions (literals, identifiers, etc.)
func (e *Evaluator) evaluateBaseExpr(expr *parser.BaseExpr, env *Environment) (*Value, error) {
	if expr.Literal != nil {
		return e.evaluateLiteral(expr.Literal, env)
	}

	if expr.Identifier != nil {
		// Look up variable
		value, exists := env.Get(*expr.Identifier)
		if !exists {
			return nil, fmt.Errorf("undefined variable: %s", *expr.Identifier)
		}
		return value, nil
	}

	if expr.Grouped != nil {
		return e.EvaluateWithEnv(expr.Grouped, env)
	}

	if expr.FuncCall != nil {
		return e.evaluateFuncCall(expr.FuncCall, env)
	}

	if expr.StructConstructor != nil {
		return e.evaluateStructConstructor(expr.StructConstructor, env)
	}

	return NewNil(), fmt.Errorf("unsupported base expression")
}

// evaluateLiteral evaluates literal values
func (e *Evaluator) evaluateLiteral(literal *parser.Literal, env *Environment) (*Value, error) {
	if literal.Number != nil {
		return NewNumber(*literal.Number), nil
	}

	if literal.String != nil {
		return NewString(*literal.String), nil
	}

	if literal.Bool != nil {
		return NewBool(*literal.Bool == "true"), nil
	}

	if literal.Array != nil {
		elements := make([]*Value, 0, len(literal.Array.Elements))
		for _, elem := range literal.Array.Elements {
			value, err := e.evaluateLiteral(elem, env)
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

	// Evaluate arguments (these are literals, so use evaluateLiteral)
	args := make([]*Value, 0, len(funcCall.Args))
	for _, arg := range funcCall.Args {
		value, err := e.evaluateLiteral(arg, env)
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
	return e.callUserFunction(funcValue.Function, args, env)
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
	return e.callUserFunction(funcValue.Function, args, env)
}

// callUserFunction calls a user-defined function
func (e *Evaluator) callUserFunction(fn *Function, args []*Value, parentEnv *Environment) (*Value, error) {
	// Check parameter count
	if len(args) != len(fn.Parameters) {
		return nil, fmt.Errorf("function '%s' expects %d arguments, got %d",
			fn.Name, len(fn.Parameters), len(args))
	}

	// Create new environment for function scope
	// Use closure environment as parent if available, otherwise use calling environment
	var envParent *Environment
	if fn.ClosureEnv != nil {
		envParent = fn.ClosureEnv
	} else {
		envParent = parentEnv
	}
	funcEnv := NewEnvironment(envParent)

	// Bind parameters to arguments
	for i, paramName := range fn.Parameters {
		funcEnv.Define(paramName, args[i])
	}

	// Execute function body
	result, err := e.evaluateBlock(fn.Body, funcEnv)
	if err != nil {
		// Check if it's a return value (should be handled by evaluateBlock)
		if returnVal, ok := err.(*ReturnValue); ok {
			return returnVal.Value, nil
		}
		return nil, err
	}

	return result, nil
}

// evaluateMethodCall evaluates method calls
func (e *Evaluator) evaluateMethodCall(object *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	// Handle built-in methods based on object type
	switch object.Type {
	case ValueTypeArray:
		return e.evaluateArrayMethod(object, call, env)
	case ValueTypeObject:
		return e.evaluateObjectMethod(object, call, env)
	case ValueTypeString:
		return e.evaluateStringMethod(object, call, env)
	case ValueTypeStruct:
		return e.evaluateStructMethod(object, call, env)
	default:
		return nil, fmt.Errorf("method '%s' not supported for %v", call.Method, object.Type)
	}
}

// evaluateArrayMethod evaluates array methods
func (e *Evaluator) evaluateArrayMethod(array *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "length":
		return NewNumber(float64(len(array.Array))), nil
	default:
		return nil, fmt.Errorf("unknown array method: %s", call.Method)
	}
}

// evaluateObjectMethod evaluates object methods
func (e *Evaluator) evaluateObjectMethod(object *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "get":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("get method expects 1 argument")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("object key must be a string")
		}

		if value, exists := object.Object[key.Str]; exists {
			return value, nil
		}
		return NewNil(), nil

	default:
		return nil, fmt.Errorf("unknown object method: %s", call.Method)
	}
}

// evaluateStringMethod evaluates string methods
func (e *Evaluator) evaluateStringMethod(str *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "length":
		return NewNumber(float64(len(str.Str))), nil
	default:
		return nil, fmt.Errorf("unknown string method: %s", call.Method)
	}
}

// evaluateStructMethod evaluates struct methods
func (e *Evaluator) evaluateStructMethod(structVal *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "get":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("get method expects 1 argument")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("struct field name must be a string")
		}

		if value, exists := structVal.Struct.Fields[key.Str]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("struct %s has no field '%s'", structVal.Struct.Name, key.Str)

	default:
		return nil, fmt.Errorf("unknown struct method: %s", call.Method)
	}
}

// applyBinaryOperation applies a binary operation to two values
func (e *Evaluator) applyBinaryOperation(left *Value, op string, right *Value) (*Value, error) {
	switch op {
	case "+":
		return e.add(left, right)
	case "-":
		return e.subtract(left, right)
	case "*":
		return e.multiply(left, right)
	case "/":
		return e.divide(left, right)
	case "==":
		return NewBool(left.IsEqual(right)), nil
	case "!=":
		return NewBool(!left.IsEqual(right)), nil
	case "<":
		return e.less(left, right)
	case "<=":
		return e.lessEqual(left, right)
	case ">":
		return e.greater(left, right)
	case ">=":
		return e.greaterEqual(left, right)
	case "&&":
		return NewBool(left.IsTruthy() && right.IsTruthy()), nil
	case "||":
		return NewBool(left.IsTruthy() || right.IsTruthy()), nil
	case "??":
		if left.Type == ValueTypeNil {
			return right, nil
		}
		return left, nil
	default:
		return nil, fmt.Errorf("unsupported binary operator: %s", op)
	}
}

// Arithmetic operations
func (e *Evaluator) add(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewNumber(left.Number + right.Number), nil
	}
	if left.Type == ValueTypeString && right.Type == ValueTypeString {
		return NewString(left.Str + right.Str), nil
	}
	return nil, fmt.Errorf("invalid operands for addition")
}

func (e *Evaluator) subtract(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewNumber(left.Number - right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for subtraction")
}

func (e *Evaluator) multiply(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewNumber(left.Number * right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for multiplication")
}

func (e *Evaluator) divide(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		if right.Number == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return NewNumber(left.Number / right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for division")
}

// Comparison operations
func (e *Evaluator) less(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number < right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

func (e *Evaluator) lessEqual(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number <= right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

func (e *Evaluator) greater(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number > right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

func (e *Evaluator) greaterEqual(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number >= right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

// evaluateFunctionExpr handles function definitions
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

// evaluateStructExpr handles struct definitions
func (e *Evaluator) evaluateStructExpr(expr *parser.StructExpr, env *Environment) (*Value, error) {
	// Create struct definition
	fieldTypes := make(map[string]string)
	for _, field := range expr.Fields {
		// For now, we'll store the type name as a string
		// In a more complete implementation, we'd have a proper type system
		fieldTypes[field.Name] = e.getTypeName(field.Type)
	}

	structDef := &StructDefinition{
		Name:   expr.Name,
		Fields: fieldTypes,
	}

	// Store the struct definition
	e.structDefs[expr.Name] = structDef

	// Struct definitions don't return a value, they register a type
	return NewNil(), nil
}

// evaluateStructConstructor handles struct instantiation (User{name: "John"})
func (e *Evaluator) evaluateStructConstructor(expr *parser.StructConstructor, env *Environment) (*Value, error) {
	// Check if the struct type is defined
	structDef, exists := e.structDefs[expr.Name]
	if !exists {
		return nil, fmt.Errorf("undefined struct type: %s", expr.Name)
	}

	// Evaluate field values
	fields := make(map[string]*Value)
	for _, field := range expr.Fields {
		value, err := e.EvaluateWithEnv(field.Value, env)
		if err != nil {
			return nil, err
		}
		fields[field.Key] = value
	}

	// Validate that all required fields are provided
	for fieldName := range structDef.Fields {
		if _, provided := fields[fieldName]; !provided {
			return nil, fmt.Errorf("missing required field '%s' for struct %s", fieldName, expr.Name)
		}
	}

	// Create struct instance
	return NewStruct(expr.Name, fields), nil
}

// getTypeName extracts a type name from a TypeRef (simplified)
func (e *Evaluator) getTypeName(typeRef *parser.TypeRef) string {
	if typeRef == nil {
		return "unknown"
	}

	if typeRef.Name != "" {
		return typeRef.Name
	}

	if typeRef.Array != nil {
		return "[]" + e.getTypeName(typeRef.Array)
	}

	if typeRef.Function != nil {
		return "function"
	}

	if typeRef.Parameterized != nil {
		return typeRef.Parameterized.Name
	}

	return "unknown"
}
