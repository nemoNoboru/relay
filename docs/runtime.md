# Relay Runtime System Documentation

The Relay runtime system is responsible for executing parsed Abstract Syntax Trees (ASTs) and managing program state during execution. This document provides a comprehensive guide to understanding, using, and extending the runtime.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Value System](#value-system)
4. [Environment and Scoping](#environment-and-scoping)
5. [Expression Evaluation](#expression-evaluation)
6. [Function System](#function-system)
7. [Error Handling](#error-handling)
8. [Built-in Functions](#built-in-functions)
9. [Extending the Runtime](#extending-the-runtime)
10. [Examples](#examples)
11. [Best Practices](#best-practices)

## Overview

🎉 **MAJOR MILESTONE ACHIEVED**: Full closure support and first-class functions implemented!

The Relay runtime takes parsed AST nodes from the parser and executes them, maintaining program state and handling variable scoping, function calls, and control flow. The runtime is designed to be:

- **Type-safe**: Strong runtime type checking with clear error messages
- **Closure-enabled**: ✨ **NEW**: Full lexical scoping with environment capture for closures
- **Functional**: ✨ **NEW**: First-class functions supporting higher-order programming patterns
- **Scoped**: Proper lexical scoping with environment chains
- **Extensible**: Easy to add new value types, operators, and built-in functions
- **Error-resilient**: Comprehensive error handling with meaningful messages

### Key Components

- **Value System**: Represents all runtime values (numbers, strings, functions, structs, etc.)
- **Environment System**: Manages variable scoping and lookup with closure support
- **Evaluator**: Core engine that executes AST nodes
- **Function System**: ✨ **ENHANCED**: Full closure and first-class function support with lexical scoping
- **Struct System**: ✨ **NEW**: Complete struct definition, instantiation, and field access support

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Parser      │───▶│    Evaluator    │───▶│   Environment   │
│   (AST Nodes)   │    │                 │    │   (Variables)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               ▼
                       ┌─────────────────┐
                       │  Value System   │
                       │ (Runtime Values)│
                       └─────────────────┘
```

### Design Principles

1. **Separation of Concerns**: Parsing and execution are completely separate
2. **Immutable Values**: Runtime values are immutable for safety
3. **Environment Chaining**: Scoping implemented via parent environment references
4. **Error as Control Flow**: Return statements use Go's error mechanism
5. **Type Safety**: All operations validate types before execution

## Value System

The runtime uses a unified `Value` type to represent all data types. This provides type safety and consistent behavior across the system.

### Value Types

```go
type ValueType int

const (
    ValueTypeNil ValueType = iota
    ValueTypeNumber
    ValueTypeString
    ValueTypeBool
    ValueTypeArray
    ValueTypeObject
    ValueTypeFunction
    ValueTypeStruct
)
```

### Value Structure

```go
type Value struct {
    Type     ValueType
    Number   float64
    Str      string
    Bool     bool
    Array    []*Value
    Object   map[string]*Value
    Function *Function
    Struct   *Struct
}
```

### Creating Values

```go
// Numbers
value := runtime.NewNumber(42.0)

// Strings
value := runtime.NewString("hello")

// Booleans
value := runtime.NewBool(true)

// Arrays
elements := []*runtime.Value{
    runtime.NewNumber(1),
    runtime.NewNumber(2),
}
value := runtime.NewArray(elements)

// Objects
fields := map[string]*runtime.Value{
    "name": runtime.NewString("Alice"),
    "age":  runtime.NewNumber(30),
}
value := runtime.NewObject(fields)
```

### Value Methods

#### String Representation
```go
value.String() // Returns human-readable string representation
```

#### Truthiness
```go
value.IsTruthy() // Returns boolean truthiness
```

#### Equality
```go
value1.IsEqual(value2) // Deep equality comparison
```

### Truthiness Rules

- **Nil**: Always false
- **Numbers**: False if 0, true otherwise
- **Strings**: False if empty, true otherwise
- **Booleans**: Their boolean value
- **Arrays**: False if empty, true otherwise
- **Objects**: False if empty, true otherwise
- **Functions**: Always true

## Environment and Scoping

The environment system manages variable storage and lookup with proper lexical scoping.

### Environment Structure

```go
type Environment struct {
    variables map[string]*Value
    parent    *Environment
}
```

### Scoping Rules

1. **Variable Lookup**: Searches current environment, then parent chain
2. **Variable Definition**: Always defines in current environment
3. **Function Scoping**: Each function call creates a new environment
4. **Parameter Binding**: Function parameters are defined in function environment

### Environment Operations

```go
env := runtime.NewEnvironment(parentEnv)

// Define variable in current scope
env.Define("x", runtime.NewNumber(42))

// Look up variable (searches parent chain)
value, exists := env.Get("x")

// Set existing variable (updates in-place)
env.Set("x", runtime.NewNumber(100))
```

### Scoping Example

```relay
set global_var = 10

fn test_function(param) {
    set local_var = 20
    param + local_var + global_var  // Can access all three
}

test_function(5)  // Returns 35 (5 + 20 + 10)
```

## Expression Evaluation

The evaluator traverses AST nodes and executes them recursively.

### Evaluation Flow

```
Expression → Binary/Set/Function/Return
    ↓
UnaryExpr → Primary → Base
    ↓
Literal/Identifier/FuncCall/etc.
```

### Core Evaluation Methods

#### Main Entry Point
```go
func (e *Evaluator) Evaluate(expr *parser.Expression) (*Value, error)
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error)
```

#### Expression Types
```go
// Binary expressions (arithmetic, comparisons, logical)
func (e *Evaluator) evaluateBinaryExpr(expr *parser.BinaryExpr, env *Environment) (*Value, error)

// Variable assignments
func (e *Evaluator) evaluateSetExpr(expr *parser.SetExpr, env *Environment) (*Value, error)

// Function definitions
func (e *Evaluator) evaluateFunctionExpr(expr *parser.FunctionExpr, env *Environment) (*Value, error)

// Return statements
func (e *Evaluator) evaluateReturnExpr(expr *parser.ReturnExpr, env *Environment) (*Value, error)
```

### Binary Operations

The runtime supports comprehensive binary operations with type checking:

#### Arithmetic Operations
- **Addition** (`+`): Numbers and string concatenation
- **Subtraction** (`-`): Numbers only
- **Multiplication** (`*`): Numbers only
- **Division** (`/`): Numbers only (with division by zero check)

#### Comparison Operations
- **Equal** (`==`): Deep equality for all types
- **Not Equal** (`!=`): Negated deep equality
- **Less Than** (`<`): Numbers only
- **Less Equal** (`<=`): Numbers only
- **Greater Than** (`>`): Numbers only
- **Greater Equal** (`>=`): Numbers only

#### Logical Operations
- **And** (`&&`): Boolean truthiness of both operands
- **Or** (`||`): Boolean truthiness of either operand
- **Null Coalesce** (`??`): Returns right operand if left is nil

### Type Coercion Rules

The runtime is strictly typed with minimal coercion:

1. **No Automatic Coercion**: Operations require matching types
2. **String Addition**: Only strings can be concatenated with `+`
3. **Truthiness**: All types have defined truthiness for logical operations
4. **Error on Mismatch**: Clear error messages for type mismatches

## Function System

🎉 **MAJOR MILESTONE**: The function system now supports **closures** and **first-class functions** with full lexical scoping and environment capture!

Functions are first-class values with complete closure support, proper scoping, and parameter binding.

### Function Definition

Functions are first-class values stored in the environment with closure support:

```go
type Function struct {
    Name       string
    Parameters []string
    Body       *parser.Block
    IsBuiltin  bool
    Builtin    func(args []*Value) (*Value, error)
    ClosureEnv *Environment  // ✨ NEW: Captures lexical environment for closures
}
```

### Closure Support (NEW!)

**Functions are true closures** that capture their lexical environment:

```relay
set captured = 42

fn getCaptured() {
    captured  // ✅ Accesses variable from outer scope
}

getCaptured()  // Returns 42

fn makeCounter(start) {
    set count = start
    fn() { 
        set count = count + 1
        count 
    }
}

set counter = makeCounter(5)
counter()  // Returns 6
counter()  // Returns 7 - state is preserved!
```

**How Closures Work**:
1. When a function is defined, it captures the current environment as `ClosureEnv`
2. When called, the captured environment becomes the parent of the function's execution environment
3. Functions can access variables from their defining scope, even after that scope has exited

### First-Class Functions (NEW!)

**Functions are fully first-class citizens**:

#### Functions as Variables
```relay
set myFunc = double
myFunc(5)  // Returns 10
```

#### Functions as Arguments (Higher-Order Functions)
```relay
fn applyTwice(operation, value) {
    operation(operation(value))
}

fn increment(x) { x + 1 }
applyTwice(increment, 5)  // Returns 7

// With anonymous functions
applyTwice(fn(x) { x * 3 }, 2)  // Returns 18
```

#### Functions as Return Values
```relay
fn makeMultiplier(factor) {
    fn(x) { x * factor }  // Returns a function
}

set times3 = makeMultiplier(3)
times3(4)  // Returns 12
```

#### Complex Higher-Order Patterns
```relay
fn compose(f, g) {
    fn(x) { f(g(x)) }  // Function composition
}

fn square(x) { x * x }
fn double(x) { x * 2 }
set squareAndDouble = compose(double, square)
squareAndDouble(3)  // Returns 18 (3² = 9, 9 * 2 = 18)
```

### Function Creation

#### User-Defined Functions
```relay
fn add(x, y) {
    x + y
}

fn greet(name) {
    return "Hello, " + name + "!"
}
```

#### Anonymous Functions
```relay
fn (x) { x * 2 }  // Anonymous function
```

### Function Calls

Function calls are resolved through the environment chain:

1. **Lookup**: Find function in environment
2. **Type Check**: Verify it's a function value
3. **Argument Evaluation**: Evaluate all arguments
4. **Parameter Binding**: Create new environment with parameters
5. **Body Execution**: Execute function body in new environment
6. **Return Handling**: Handle explicit returns or implicit last expression

### Parameter Binding

```go
func (e *Evaluator) callUserFunction(fn *Function, args []*Value, parentEnv *Environment) (*Value, error) {
    // Create new environment for function scope
    funcEnv := NewEnvironment(parentEnv)
    
    // Bind parameters to arguments
    for i, paramName := range fn.Parameters {
        funcEnv.Define(paramName, args[i])
    }
    
    // Execute function body
    return e.evaluateBlock(fn.Body, funcEnv)
}
```

### Return Statements

Return statements are implemented using Go's error mechanism for control flow:

```go
type ReturnValue struct {
    Value *Value
}

func (r ReturnValue) Error() string {
    return "return"
}
```

This allows early returns from anywhere in the function body.

## Error Handling

The runtime provides comprehensive error handling with clear, actionable messages.

### Error Types

1. **Parse Errors**: Handled by parser (not runtime)
2. **Runtime Errors**: Type mismatches, undefined variables, etc.
3. **Return Values**: Special error type for function returns
4. **Built-in Errors**: Errors from built-in functions

### Error Examples

```go
// Undefined variable
return nil, fmt.Errorf("undefined variable: %s", name)

// Type mismatch
return nil, fmt.Errorf("invalid operands for addition")

// Parameter count mismatch
return nil, fmt.Errorf("function '%s' expects %d arguments, got %d", 
    fn.Name, len(fn.Parameters), len(args))

// Division by zero
return nil, fmt.Errorf("division by zero")
```

### Error Recovery

The REPL catches and displays runtime errors without terminating:

```relay
relay> undefined_variable
Runtime error: undefined variable: undefined_variable

relay> 5 / 0
Runtime error: division by zero

relay> add(5)  // Function expects 2 parameters
Runtime error: function 'add' expects 2 arguments, got 1
```

## Struct System

✨ **NEW FEATURE**: The runtime now supports complete struct definitions, instantiation, and field access!

### Struct Definition

Structs are user-defined types that group related data fields together:

```relay
struct User {
    name: string,
    age: number,
    active: bool
}

struct Post {
    title: string,
    author: string,
    views: number,
    tags: [string]
}
```

### Struct Storage

Struct definitions are stored in the evaluator and registered as types:

```go
type StructDefinition struct {
    Name   string            // Struct name
    Fields map[string]string // Field name -> type name mapping
}

type Struct struct {
    Name   string            // Struct type name (e.g., "User")
    Fields map[string]*Value // Field values
}
```

### Struct Instantiation

Create struct instances using constructor syntax:

```relay
// Create a User instance
set user = User{
    name: "John Doe",
    age: 30,
    active: true
}

// Field order doesn't matter
set user2 = User{
    active: false,
    name: "Jane Smith",
    age: 25
}

// Use expressions in field values
set post = Post{
    title: "Hello World",
    author: user.get("name"),
    views: 100,
    tags: ["programming", "tutorial"]
}
```

### Field Access

Access struct fields using the `.get()` method:

```relay
set name = user.get("name")        // Returns "John Doe"
set age = user.get("age")          // Returns 30
set isActive = user.get("active")  // Returns true
```

### Struct Operations

#### Equality Comparison
```relay
set user1 = User{ name: "John", age: 30, active: true }
set user2 = User{ name: "John", age: 30, active: true }
set user3 = User{ age: 30, name: "John", active: true }

user1 == user2  // Returns true (same values)
user1 == user3  // Returns true (field order doesn't matter)
```

#### Using Struct Fields in Expressions
```relay
// String concatenation
set greeting = "Hello, " + user.get("name")

// Arithmetic operations
set nextYear = user.get("age") + 1

// Conditional logic
if user.get("active") {
    print("User is active")
}
```

#### Complex Struct Usage
```relay
// Define multiple related structs
struct Address {
    street: string,
    city: string,
    zipcode: string
}

struct Employee {
    name: string,
    department: string,
    salary: number
}

// Create instances
set address = Address{
    street: "123 Main St",
    city: "Anytown",
    zipcode: "12345"
}

set employee = Employee{
    name: "Alice Johnson",
    department: "Engineering",
    salary: 75000
}

// Use in expressions
set summary = employee.get("name") + " works in " + employee.get("department")
```

### Error Handling

The struct system provides comprehensive error checking:

```relay
// Error: Undefined struct type
UnknownStruct{ field: "value" }  // Runtime error: undefined struct type: UnknownStruct

// Error: Missing required field
User{ name: "John" }  // Runtime error: missing required field 'age' for struct User

// Error: Accessing nonexistent field
user.get("email")  // Runtime error: struct User has no field 'email'
```

### Struct System Implementation

The struct system is implemented with:

1. **Type Registration**: Struct definitions stored in evaluator's `structDefs` map
2. **Constructor Validation**: All required fields must be provided during instantiation
3. **Field Access**: Safe field access with proper error handling
4. **Value Integration**: Structs are first-class values supporting equality, truthiness, and string representation

## Built-in Functions

The runtime comes with essential built-in functions that can be extended.

### Current Built-ins

#### print(...args)
Prints all arguments separated by spaces:

```relay
print("Hello", "World")  // Output: Hello World
print(42, true, "test")  // Output: 42 true test
```

### Built-in Implementation

```go
func (e *Evaluator) defineBuiltins() {
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
```

## Extending the Runtime

The runtime is designed for easy extension. Here are the main extension points:

### 1. Adding New Value Types

To add a new value type (e.g., `ValueTypeDate`):

1. **Add to ValueType enum**:
```go
const (
    ValueTypeNil ValueType = iota
    ValueTypeNumber
    ValueTypeString
    ValueTypeBool
    ValueTypeArray
    ValueTypeObject
    ValueTypeFunction
    ValueTypeDate  // New type
)
```

2. **Add field to Value struct**:
```go
type Value struct {
    Type     ValueType
    Number   float64
    Str      string
    Bool     bool
    Array    []*Value
    Object   map[string]*Value
    Function *Function
    Date     time.Time  // New field
}
```

3. **Add constructor**:
```go
func NewDate(t time.Time) *Value {
    return &Value{
        Type: ValueTypeDate,
        Date: t,
    }
}
```

4. **Update String() method**:
```go
case ValueTypeDate:
    return v.Date.Format("2006-01-02 15:04:05")
```

5. **Update IsTruthy() method**:
```go
case ValueTypeDate:
    return !v.Date.IsZero()
```

6. **Update IsEqual() method**:
```go
case ValueTypeDate:
    return v.Date.Equal(other.Date)
```

### 2. Adding New Binary Operators

To add a new operator (e.g., `%` for modulo):

1. **Add to parser** (if not already supported)
2. **Add to applyBinaryOperation**:
```go
func (e *Evaluator) applyBinaryOperation(left *Value, op string, right *Value) (*Value, error) {
    switch op {
    case "+":
        return e.add(left, right)
    // ... other cases
    case "%":
        return e.modulo(left, right)
    default:
        return nil, fmt.Errorf("unsupported binary operator: %s", op)
    }
}
```

3. **Implement the operation**:
```go
func (e *Evaluator) modulo(left, right *Value) (*Value, error) {
    if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
        if right.Number == 0 {
            return nil, fmt.Errorf("modulo by zero")
        }
        return NewNumber(math.Mod(left.Number, right.Number)), nil
    }
    return nil, fmt.Errorf("invalid operands for modulo")
}
```

### 3. Adding Built-in Functions

To add a new built-in function:

1. **Add to defineBuiltins()**:
```go
func (e *Evaluator) defineBuiltins() {
    // ... existing built-ins
    
    // Add new built-in
    mathPiFunc := &Value{
        Type: ValueTypeFunction,
        Function: &Function{
            Name:      "pi",
            IsBuiltin: true,
            Builtin: func(args []*Value) (*Value, error) {
                if len(args) != 0 {
                    return nil, fmt.Errorf("pi function takes no arguments")
                }
                return NewNumber(math.Pi), nil
            },
        },
    }
    e.globalEnv.Define("pi", mathPiFunc)
}
```

### 4. Adding Method Calls

To add methods to existing types:

1. **Add to evaluateMethodCall**:
```go
func (e *Evaluator) evaluateMethodCall(object *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
    switch object.Type {
    case ValueTypeArray:
        return e.evaluateArrayMethod(object, call, env)
    case ValueTypeString:
        return e.evaluateStringMethod(object, call, env)
    case ValueTypeDate:  // New type methods
        return e.evaluateDateMethod(object, call, env)
    default:
        return nil, fmt.Errorf("method '%s' not supported for %v", call.Method, object.Type)
    }
}
```

2. **Implement type-specific methods**:
```go
func (e *Evaluator) evaluateDateMethod(date *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
    switch call.Method {
    case "year":
        return NewNumber(float64(date.Date.Year())), nil
    case "month":
        return NewNumber(float64(date.Date.Month())), nil
    case "format":
        if len(call.Args) != 1 {
            return nil, fmt.Errorf("format method expects 1 argument")
        }
        format, err := e.EvaluateWithEnv(call.Args[0], env)
        if err != nil {
            return nil, err
        }
        if format.Type != ValueTypeString {
            return nil, fmt.Errorf("format argument must be a string")
        }
        return NewString(date.Date.Format(format.Str)), nil
    default:
        return nil, fmt.Errorf("unknown date method: %s", call.Method)
    }
}
```

### 5. Adding New Expression Types

To add support for new expression types from the parser:

1. **Add to EvaluateWithEnv**:
```go
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error) {
    if expr == nil {
        return NewNil(), nil
    }

    // Handle different expression types
    if expr.Binary != nil {
        return e.evaluateBinaryExpr(expr.Binary, env)
    }
    
    if expr.IfExpr != nil {  // New expression type
        return e.evaluateIfExpr(expr.IfExpr, env)
    }
    
    // ... other cases
}
```

2. **Implement the evaluator**:
```go
func (e *Evaluator) evaluateIfExpr(expr *parser.IfExpr, env *Environment) (*Value, error) {
    condition, err := e.EvaluateWithEnv(expr.Condition, env)
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
```

## Examples

### Basic Runtime Usage

```go
package main

import (
    "strings"
    "relay/pkg/parser"
    "relay/pkg/runtime"
)

func main() {
    // Create evaluator
    evaluator := runtime.NewEvaluator()
    
    // Parse and evaluate simple expression
    program, _ := parser.Parse("main", strings.NewReader("2 + 3"))
    result, _ := evaluator.Evaluate(program.Expressions[0])
    fmt.Println(result.String()) // "5"
    
    // Parse and evaluate function definition
    program, _ = parser.Parse("main", strings.NewReader("fn add(x, y) { x + y }"))
    evaluator.Evaluate(program.Expressions[0])
    
    // Parse and evaluate function call
    program, _ = parser.Parse("main", strings.NewReader("add(10, 20)"))
    result, _ = evaluator.Evaluate(program.Expressions[0])
    fmt.Println(result.String()) // "30"
}
```

### Custom Built-in Function

```go
func addMathBuiltins(evaluator *runtime.Evaluator) {
    // Add sqrt function
    sqrtFunc := &runtime.Value{
        Type: runtime.ValueTypeFunction,
        Function: &runtime.Function{
            Name:      "sqrt",
            IsBuiltin: true,
            Builtin: func(args []*runtime.Value) (*runtime.Value, error) {
                if len(args) != 1 {
                    return nil, fmt.Errorf("sqrt expects 1 argument, got %d", len(args))
                }
                
                if args[0].Type != runtime.ValueTypeNumber {
                    return nil, fmt.Errorf("sqrt expects a number")
                }
                
                if args[0].Number < 0 {
                    return nil, fmt.Errorf("sqrt of negative number")
                }
                
                return runtime.NewNumber(math.Sqrt(args[0].Number)), nil
            },
        },
    }
    
    evaluator.GlobalEnv().Define("sqrt", sqrtFunc)
}
```

### Custom Value Type Example

```go
// Geographic coordinate type
type Coordinate struct {
    Lat, Lng float64
}

// Add to ValueType enum
const ValueTypeCoordinate ValueType = 7

// Add to Value struct
type Value struct {
    // ... existing fields
    Coordinate *Coordinate
}

// Constructor
func NewCoordinate(lat, lng float64) *Value {
    return &Value{
        Type: ValueTypeCoordinate,
        Coordinate: &Coordinate{Lat: lat, Lng: lng},
    }
}

// Methods
func (c *Coordinate) DistanceTo(other *Coordinate) float64 {
    // Haversine formula implementation
    // ...
}
```

## Best Practices

### 1. Error Handling

- **Always validate types** before operations
- **Provide clear error messages** with context
- **Use consistent error formats** across the codebase
- **Handle edge cases** like division by zero, empty arrays, etc.

### 2. Memory Management

- **Avoid circular references** in Value structures
- **Don't mutate existing values** - create new ones instead
- **Clean up environments** when functions return
- **Use pointers efficiently** for large data structures

### 3. Type Safety

- **Check types before casting** Value fields
- **Validate argument counts** for functions
- **Provide type conversion utilities** when needed
- **Document type requirements** for new functions

### 4. Performance

- **Cache frequently used values** (like common numbers)
- **Avoid unnecessary allocations** in hot paths
- **Use efficient data structures** for large collections
- **Profile before optimizing** to identify bottlenecks

### 5. Extensibility

- **Design for extension** from the beginning
- **Use interfaces** where appropriate for pluggability
- **Document extension points** clearly
- **Maintain backward compatibility** when possible

### 6. Testing

- **Test all value types** thoroughly
- **Test error conditions** as well as success cases
- **Test scoping behavior** with complex scenarios
- **Use property-based testing** for mathematical operations

## Conclusion

The Relay runtime system provides a solid foundation for executing Relay programs with proper type safety, scoping, and error handling. Its modular design makes it easy to extend with new features while maintaining reliability and performance.

The key to successful extension is understanding the core patterns:
- Values as the universal data representation
- Environments for scoping and variable management  
- Evaluators for AST traversal and execution
- Error handling for robustness

By following these patterns and the extension guidelines in this document, you can add powerful new features to the Relay runtime while maintaining its design principles and reliability. 