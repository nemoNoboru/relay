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

By following these patterns and the extension guidelines in this document, you can add powerful new features to the Relay runtime while maintaining its design principles and reliability. # Relay Runtime Internals Documentation

This document provides a comprehensive guide to understanding the Relay language runtime, expression evaluation, environment scoping, and debugging techniques. It's designed for developers who need to maintain, extend, or debug the Relay interpreter.

## Table of Contents

1. [Runtime Architecture Overview](#runtime-architecture-overview)
2. [Environment and Scoping System](#environment-and-scoping-system)
3. [Expression Evaluation Chain](#expression-evaluation-chain)
4. [Function Call Mechanism](#function-call-mechanism)
5. [Parser Integration](#parser-integration)
6. [Debugging Techniques](#debugging-techniques)
7. [Common Bug Patterns](#common-bug-patterns)
8. [Code Organization](#code-organization)
9. [Testing Strategy](#testing-strategy)

---

## Runtime Architecture Overview

### Core Components

The Relay runtime consists of several interconnected components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Parser      │───▶│    Evaluator    │───▶│   Environment   │
│  (AST Creation) │    │ (Code Execution)│    │ (Variable Store)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │     Values      │
                    │ (Runtime Data)  │
                    └─────────────────┘
```

### Key Files and Their Responsibilities

- **`pkg/runtime/evaluator.go`** - Main evaluation coordinator, responsible for initialization and high-level state.
- **`pkg/runtime/core.go`** - The unified expression evaluation system. Contains the core logic for evaluating all language constructs.
- **`pkg/runtime/expressions.go`** - Binary/unary expression evaluation
- **`pkg/runtime/functions.go`** - Function call and definition handling
- **`pkg/runtime/literals.go`** - Literal value evaluation
- **`pkg/runtime/value.go`** - Value types and environment management
- **`pkg/runtime/methods.go`** - Built-in method implementations

---

## Environment and Scoping System

### Environment Structure

The environment is a hierarchical scoping system that manages variable bindings:

```go
type Environment struct {
    variables map[string]*Value  // Local variable bindings
    parent    *Environment       // Parent scope (for lexical scoping)
}
```

### Scoping Rules

#### 1. Global Environment
- Created by `NewEvaluator()`
- Contains built-in functions (`print`, `message`, etc.)
- Root of the environment chain

#### 2. Function Environment  
- Created for each function call
- Contains parameter bindings
- Parent points to closure environment or global environment

#### 3. Environment Chain Resolution

```go
func (e *Environment) Get(name string) (*Value, bool) {
    // 1. Check local scope
    value, exists := e.variables[name]
    if !exists && e.parent != nil {
        // 2. Walk up the parent chain
        return e.parent.Get(name)
    }
    return value, exists
}
```

### Critical Environment Patterns

#### ✅ **Correct**: Passing Environment Through Evaluation
```go
// Always pass the current environment to child evaluations
func (e *Evaluator) someEvaluation(expr *parser.SomeExpr, env *Environment) (*Value, error) {
    childResult, err := e.EvaluateWithEnv(expr.Child, env) // ✅ Correct
    return childResult, err
}
```

#### ❌ **WRONG**: Using Global Environment
```go
// This breaks scoping - child can't see local variables!
func (e *Evaluator) someEvaluation(expr *parser.SomeExpr, env *Environment) (*Value, error) {
    childResult, err := e.Evaluate(expr.Child) // ❌ Wrong - uses global env
    return childResult, err
}
```

---

## Expression Evaluation Chain

### Evaluation Flow

```
Expression → Binary/Set/Function/Return
    ↓
UnaryExpr → Primary → Base
    ↓
Literal/Identifier/FuncCall/etc.
```

The core of this logic resides in `pkg/runtime/core.go`.

### Core Evaluation Methods

#### Main Entry Point
The `Evaluator` exposes two main entry points, which then delegate to the core evaluation engine.

```go
func (e *Evaluator) Evaluate(expr *parser.Expression) (*Value, error)
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error)
```

#### Expression Types (in `core.go`)
The `core.go` file contains the specific functions for evaluating each type of expression:

```go
// The unified entry point in core.go
func (e *Evaluator) EvaluateExpression(expr *parser.Expression, env *Environment) (*Value, error)

// Binary expressions (arithmetic, comparisons, logical)
func (e *Evaluator) evaluateBinary(expr *parser.BinaryExpr, env *Environment) (*Value, error)

// Variable assignments
func (e *Evaluator) evaluateSet(expr *parser.SetExpr, env *Environment) (*Value, error)

// Function definitions
func (e *Evaluator) evaluateFunction(expr *parser.FunctionExpr, env *Environment) (*Value, error)

// Return statements
func (e *Evaluator) evaluateReturn(expr *parser.ReturnExpr, env *Environment) (*Value, error)
```

### Binary Operations

The runtime supports comprehensive binary operations with type checking:

```
Expression
    ↓
EvaluateWithEnv(expr, env)
    ↓
Binary Expression
    ↓
Unary Expression  
    ↓
Primary Expression
    ↓
Base Expression
    ↓
Specific Evaluators (FuncCall, Literal, etc.)
```

### Detailed Flow Diagram

```go
// 1. Entry Point
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error)

// 2. Route to Expression Type
if expr.Binary != nil {
    return e.evaluateBinaryExpr(expr.Binary, env) // Handles operators like +, -, &&, ||
}

// 3. Binary Expression Processing
func (e *Evaluator) evaluateBinaryExpr(expr *parser.BinaryExpr, env *Environment) (*Value, error) {
    left, err := e.evaluateUnaryExpr(expr.Left, env)  // Evaluate left operand
    // Process right operands with operators...
}

// 4. Unary Expression Processing  
func (e *Evaluator) evaluateUnaryExpr(expr *parser.UnaryExpr, env *Environment) (*Value, error) {
    primary, err := e.evaluatePrimaryExpr(expr.Primary, env) // Evaluate base
    // Apply unary operators like !, -
}

// 5. Primary Expression Processing
func (e *Evaluator) evaluatePrimaryExpr(expr *parser.PrimaryExpr, env *Environment) (*Value, error) {
    base, err := e.evaluateBaseExpr(expr.Base, env) // Get the core value
    // Apply method calls like .get(), .set()
}

// 6. Base Expression Routing
func (e *Evaluator) evaluateBaseExpr(expr *parser.BaseExpr, env *Environment) (*Value, error) {
    if expr.FuncCall != nil {
        return e.evaluateFuncCall(expr.FuncCall, env) // Function calls
    }
    if expr.Identifier != nil {
        return env.Get(*expr.Identifier) // Variable lookup
    }
    if expr.Literal != nil {
        return e.evaluateLiteral(expr.Literal, env) // Literal values
    }
    // ... other expression types
}
```

### Environment Passing Rules

**CRITICAL**: Every evaluation function MUST pass the environment to child evaluations:

```go
// ✅ CORRECT Pattern
func (e *Evaluator) someEvaluation(expr *SomeExpr, env *Environment) (*Value, error) {
    for _, child := range expr.Children {
        result, err := e.EvaluateWithEnv(child, env) // Pass env through!
        if err != nil {
            return nil, err
        }
    }
}

// ❌ BROKEN Pattern  
func (e *Evaluator) someEvaluation(expr *SomeExpr, env *Environment) (*Value, error) {
    for _, child := range expr.Children {
        result, err := e.Evaluate(child) // Missing env - breaks scoping!
        if err != nil {
            return nil, err
        }
    }
}
```

---

## Function Call Mechanism

### Two Function Call Paths

The parser creates two different types of function calls that follow different evaluation paths:

#### 1. **BaseExpr.FuncCall** (Most common)
```go
// Path: BaseExpr → evaluateFuncCall  
func (e *Evaluator) evaluateFuncCall(expr *parser.FuncCallExpr, env *Environment) (*Value, error)
```

#### 2. **Literal.FuncCall** (Less common, but crucial)
```go
// Path: Literal → evaluateLiteralFuncCall
func (e *Evaluator) evaluateLiteralFuncCall(funcCall *parser.FuncCall, env *Environment) (*Value, error)
```

### Function Call Steps

#### Step 1: Function Lookup
```go
// Look up the function in the current environment
funcValue, exists := env.Get(expr.Name)
if !exists {
    return nil, fmt.Errorf("undefined function: %s", expr.Name)
}
```

#### Step 2: Argument Evaluation  
```go
// ⚠️ CRITICAL: Must use EvaluateWithEnv, not Evaluate!
args := make([]*Value, 0, len(expr.Args))
for _, arg := range expr.Args {
    value, err := e.EvaluateWithEnv(arg, env) // ✅ Correct - preserves scoping
    if err != nil {
        return nil, err  
    }
    args = append(args, value)
}
```

#### Step 3: Function Execution
```go
// Create new environment for function execution
var funcEnv *Environment
if fn.ClosureEnv != nil {
    funcEnv = NewEnvironment(fn.ClosureEnv) // For closures - capture environment
} else {
    funcEnv = NewEnvironment(parentEnv)     // For regular functions
}

// Bind parameters
for i, param := range fn.Parameters {
    funcEnv.Define(param, args[i])
}

// Execute function body with the function environment
result, err := e.evaluateBlock(fn.Body, funcEnv)
```

### The Bug We Fixed

**Problem**: In `evaluateLiteralFuncCall`, the argument evaluation was:
```go
value, err := e.Evaluate(arg) // ❌ WRONG - uses global environment
```

**Solution**: Changed to:
```go  
value, err := e.EvaluateWithEnv(arg, env) // ✅ CORRECT - uses function environment
```

This allowed function parameters to be used as arguments to other function calls.

---

## Parser Integration

### AST Structure

The parser creates an Abstract Syntax Tree (AST) that the runtime evaluates:

```go
type Expression struct {
    StructExpr   *StructExpr   // struct definitions
    ServerExpr   *ServerExpr   // server blocks  
    FunctionExpr *FunctionExpr // function definitions
    SetExpr      *SetExpr      // variable assignments
    ReturnExpr   *ReturnExpr   // return statements
    Binary       *BinaryExpr   // operators and expressions
}

type BinaryExpr struct {
    Left  *UnaryExpr  // Left operand
    Right []*BinaryOp // Chain of operations
}

type BaseExpr struct {
    FuncCall          *FuncCallExpr      // Function calls: foo(a, b)
    StructConstructor *StructConstructor // Struct creation: User{name: "John"}
    ObjectLit         *ObjectLit         // Object literals: {key: value}
    Literal           *Literal           // Values: 42, "string", [1,2,3]
    Identifier        *string            // Variables: myVar
}
```

### Critical Parser Mappings

1. **Function Calls**: `foo(arg)` → `BaseExpr.FuncCall`
2. **Variables**: `myVar` → `BaseExpr.Identifier`  
3. **Literals**: `42`, `"hello"` → `BaseExpr.Literal`
4. **Arrays**: `[1, 2, 3]` → `Literal.Array`
5. **Function calls in literals**: Some function calls → `Literal.FuncCall`

---

## Debugging Techniques

### 1. Environment Debugging

Add temporary debug output to track environment issues:

```go
// In any evaluation function
func (e *Evaluator) someEvaluation(expr *SomeExpr, env *Environment) (*Value, error) {
    fmt.Printf("DEBUG: Function called with env depth: %d\n", environmentDepth(env))
    fmt.Printf("DEBUG: Environment variables: %v\n", getEnvironmentKeys(env))
    
    // Your evaluation logic...
}

// Helper functions for debugging
func environmentDepth(env *Environment) int {
    depth := 0
    current := env
    for current != nil {
        depth++
        current = current.parent
    }
    return depth
}

func getEnvironmentKeys(env *Environment) []string {
    var keys []string
    current := env
    level := 0
    for current != nil {
        for key := range current.variables {
            keys = append(keys, fmt.Sprintf("L%d:%s", level, key))
        }
        current = current.parent
        level++
    }
    return keys
}
```

### 2. Expression Evaluation Tracing

```go
// Add to EvaluateWithEnv for full tracing
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error) {
    fmt.Printf("DEBUG: Evaluating expression type: %T\n", expr)
    fmt.Printf("DEBUG: Environment depth: %d\n", environmentDepth(env))
    
    // Normal evaluation...
}
```

### 3. Function Call Debugging

```go
// Add to function call evaluators
func (e *Evaluator) evaluateFuncCall(expr *parser.FuncCallExpr, env *Environment) (*Value, error) {
    fmt.Printf("DEBUG: Calling function '%s' with %d args\n", expr.Name, len(expr.Args))
    fmt.Printf("DEBUG: Function environment: %v\n", getEnvironmentKeys(env))
    
    // Normal evaluation...
}
```

### 4. Using the File Execution Feature for Debugging

Create test files to isolate problems:

```bash
# Create a minimal test case
echo 'fn test(a) { id(a) }' > debug_issue.rl
echo 'fn id(x) { x }' >> debug_issue.rl  
echo 'test(42)' >> debug_issue.rl

# Run with debugging
relay debug_issue.rl

# Or load into REPL for interactive debugging
relay debug_issue.rl -repl
```

---

## Common Bug Patterns

### 1. Environment Scope Bugs

#### Symptom
```
Error: undefined variable: myVar
```

#### Common Causes
- Using `e.Evaluate(expr)` instead of `e.EvaluateWithEnv(expr, env)`
- Creating new environment without proper parent chaining
- Missing environment passing in recursive calls

#### Debugging Steps
1. Add environment tracing to see where variables disappear
2. Check that every evaluation call passes the environment
3. Verify environment parent chain is correct

### 2. Function Parameter Bugs

#### Symptom  
```
Error: undefined function: f
Error: undefined variable: param
```

#### Common Causes
- Function arguments evaluated with wrong environment
- Parameter binding happening in wrong environment
- Closure environment not properly captured

#### Debugging Steps
1. Trace parameter binding in `CallUserFunction`
2. Check argument evaluation in function call handlers
3. Verify environment creation for function execution

### 3. Parser/Runtime Mismatches

#### Symptom
```
Error: unsupported expression type
Error: nil pointer dereference in evaluation
```

#### Common Causes
- Parser creates AST nodes that runtime doesn't handle
- Missing cases in `EvaluateWithEnv` switch statement
- Type mismatches between parser and runtime

#### Debugging Steps
1. Add logging to see what expression types are being created
2. Check that all parser expression types have runtime handlers
3. Verify AST structure matches runtime expectations

### 4. Method Call Issues

#### Symptom
```
Error: method 'methodName' not supported for valueType
```

#### Common Causes
- Method calls on wrong value types
- Missing method implementations
- Incorrect argument passing to methods

### 5. Closure and Scoping Issues

#### Symptom
```
Error: captured variable not found
Error: closure environment incorrect
```

#### Common Causes
- `ClosureEnv` not set correctly during function creation
- Wrong environment used as closure parent
- Named functions incorrectly capturing environment

---

## Code Organization

### File Responsibilities

#### `evaluator.go` - **Coordination**
- Main entry points: `Evaluate()`, `EvaluateWithEnv()`
- Expression type routing
- Block evaluation
- Return statement handling

#### `expressions.go` - **Expression Processing**  
- Binary expression evaluation (`+`, `-`, `&&`, `||`, etc.)
- Unary expression evaluation (`!`, `-`)
- Primary expression coordination
- Base expression routing
- Object literal creation

#### `functions.go` - **Function System**
- Function call evaluation (`evaluateFuncCall`)
- User-defined function execution (`CallUserFunction`)
- Function definition handling (`evaluateFunctionExpr`)
- Parameter binding and environment creation

#### `literals.go` - **Value Processing**
- Literal value evaluation (numbers, strings, booleans)
- Array literal processing
- **Function calls from literals** (`evaluateLiteralFuncCall`)

#### `value.go` - **Data and Environment**
- Value type definitions
- Environment structure and methods
- Built-in value constructors
- Environment chain management

#### `methods.go` - **Built-in Methods**
- Array methods (`.get()`, `.set()`, `.length()`, etc.)
- Object methods
- String methods
- Type-specific method dispatch

### Adding New Features

#### 1. New Expression Types
1. **Add to parser**: Define AST structure in `parser.go`
2. **Add to runtime**: Add case in `EvaluateWithEnv()` in `evaluator.go`
3. **Implement evaluator**: Create evaluation function in appropriate file
4. **Write tests**: Add comprehensive test cases

#### 2. New Value Types
1. **Define type**: Add to `ValueType` enum in `value.go`
2. **Add constructors**: Create `New*()` functions
3. **Update methods**: Add `.String()`, `.IsTruthy()`, `.IsEqual()` implementations
4. **Handle in expressions**: Update relevant evaluators

#### 3. New Built-in Functions
1. **Add to builtins**: Modify `defineBuiltins()` in evaluator
2. **Implement function**: Create function implementation
3. **Handle arguments**: Validate argument types and count
4. **Write tests**: Test all edge cases

---

## Testing Strategy

### Test Categories

#### 1. **Unit Tests** - Individual Components
```go
func TestEnvironmentChaining(t *testing.T) {
    parent := NewEnvironment(nil)
    parent.Define("global", NewString("global_value"))
    
    child := NewEnvironment(parent)
    child.Define("local", NewString("local_value"))
    
    // Test local access
    value, exists := child.Get("local")
    assert.True(t, exists)
    assert.Equal(t, "local_value", value.Str)
    
    // Test parent chain access
    value, exists = child.Get("global") 
    assert.True(t, exists)
    assert.Equal(t, "global_value", value.Str)
}
```

#### 2. **Integration Tests** - Full Expression Evaluation
```go
func TestFunctionParameterScoping(t *testing.T) {
    code := `
        fn outer(a) {
            fn inner(b) { a + b }
            inner(10)
        }
        outer(5)
    `
    result := evalCode(t, code)
    assert.Equal(t, ValueTypeNumber, result.Type)
    assert.Equal(t, 15.0, result.Number)
}
```

#### 3. **File Execution Tests** - Real-world Usage
```go
func TestFileExecution(t *testing.T) {
    // Create test file
    content := `fn double(x) { x * 2 }\ndouble(21)`
    err := ioutil.WriteFile("test.rl", []byte(content), 0644)
    require.NoError(t, err)
    defer os.Remove("test.rl")
    
    // Test execution
    cmd := exec.Command("go", "run", "cmd/relay/main.go", "test.rl")
    output, err := cmd.Output()
    require.NoError(t, err)
    assert.Contains(t, string(output), "Result: 42")
}
```

---

## Conclusion

Understanding the Relay runtime requires grasping three core concepts:

1. **Environment Chain**: How variables are stored and resolved through scoped environments
2. **Expression Evaluation Flow**: The precise order of evaluation from expressions down to values
3. **Function Call Mechanism**: How functions are called and how their environments are managed

The most critical rule for maintaining the runtime is:

> **Always pass the current environment to child evaluations using `EvaluateWithEnv(expr, env)`**

When this rule is broken, you get scoping bugs where variables "disappear" or functions can't find their parameters.

The debugging techniques and patterns in this document will help you quickly identify and fix similar issues in the future. The file execution feature (`relay file.rl`) makes it easy to create minimal test cases for debugging complex problems.

---

*For questions or clarifications about the runtime internals, refer to the code examples and debugging techniques outlined above.* # Relay Runtime Quick Reference

A quick reference guide for debugging and maintaining the Relay runtime.

## 🚨 Most Common Bugs

### 1. Environment Scoping Issue
**Error**: `undefined variable: param`

**Cause**: Using `e.Evaluate(expr)` instead of `e.EvaluateWithEnv(expr, env)`

**Fix**: Always pass environment to child evaluations:
```go
// ❌ WRONG
value, err := e.Evaluate(arg)

// ✅ CORRECT  
value, err := e.EvaluateWithEnv(arg, env)
```

### 2. Function Parameter Not Found
**Error**: `undefined function: f` (where f is a parameter)

**Cause**: Function call going through wrong evaluation path

**Debug**: Check if function call uses correct environment for argument evaluation

### 3. Variables Work Alone But Not in Expressions
**Symptom**: `fn test(a) { a }` works, `fn test(a) { helper(a) }` fails

**Cause**: Sub-expressions not receiving proper environment

**Fix**: Trace environment passing through the evaluation chain

## 🔧 Quick Debugging Steps

1. **Isolate**: Create minimal failing case in `.rl` file
2. **Test**: Use `relay debug.rl` or `relay debug.rl -repl`
3. **Trace**: Add environment debugging to relevant functions
4. **Fix**: Ensure all child evaluations receive proper environment
5. **Verify**: Run full test suite to check for regressions

## 📝 Environment Debugging Template

```go
// Add temporarily to any evaluation function
func debugEnvironment(env *Environment, context string) {
    fmt.Printf("DEBUG [%s]: Env depth: %d\n", context, environmentDepth(env))
    current := env
    level := 0
    for current != nil {
        fmt.Printf("  L%d: %d vars\n", level, len(current.variables))
        for key := range current.variables {
            fmt.Printf("    %s\n", key)
        }
        current = current.parent
        level++
    }
}
```

## 🏗️ Architecture at a Glance

```
Parser → AST → Evaluator → Expression Chain → Environment Lookup
                     ↓
              Value Creation
```

### Key Evaluation Chain
1. `EvaluateWithEnv()` - Entry point
2. `evaluateBinaryExpr()` - Operators  
3. `evaluateUnaryExpr()` - Unary operators
4. `evaluatePrimaryExpr()` - Method calls
5. `evaluateBaseExpr()` - Core routing
6. Specific handlers (functions, literals, etc.)

## 🎯 Critical Rules

1. **Always use `EvaluateWithEnv(expr, env)` for child evaluations**
2. **Function arguments must be evaluated in calling environment**
3. **New environments must have proper parent chain**
4. **Closure environments must be captured during function creation**

## 📂 File Guide

- `evaluator.go` - Main coordination
- `expressions.go` - Expression processing
- `functions.go` - Function calls and definitions
- `literals.go` - Literal values and some function calls
- `value.go` - Data types and environment
- `methods.go` - Built-in methods

## 🧪 Testing Patterns

### Environment Chain Test
```go
parent := NewEnvironment(nil)
parent.Define("global", NewString("value"))
child := NewEnvironment(parent)
child.Define("local", NewString("value"))
// Test both local and parent access
```

### Function Parameter Test
```go
code := `fn test(param) { helper(param) }; fn helper(x) { x }; test(42)`
result := evalCode(t, code)
assert.Equal(t, 42.0, result.Number)
```

### Closure Test
```go
code := `fn outer(a) { fn inner() { a }; inner() }; outer(42)`
result := evalCode(t, code)
assert.Equal(t, 42.0, result.Number)
```

## 🆘 Emergency Debugging

If everything is broken:

1. **Check recent changes to environment passing**
2. **Look for new `Evaluate()` calls that should be `EvaluateWithEnv()`**
3. **Verify function call argument evaluation**
4. **Test basic variable access first, then build up complexity**
5. **Use `relay file.rl -repl` to test interactively**

## 🎨 Common Fixes

### Fix Function Call Arguments
```go
// In evaluateLiteralFuncCall or evaluateFuncCall
for _, arg := range expr.Args {
    value, err := e.EvaluateWithEnv(arg, env) // Not e.Evaluate(arg)!
    // ...
}
```

### Fix Environment Creation
```go
// Create function environment with proper parent
var funcEnv *Environment
if fn.ClosureEnv != nil {
    funcEnv = NewEnvironment(fn.ClosureEnv) // Use closure environment
} else {
    funcEnv = NewEnvironment(callingEnv)    // Use calling environment
}
```

### Fix Parameter Binding
```go
// In CallUserFunction
for i, param := range fn.Parameters {
    funcEnv.Define(param, args[i]) // Bind each parameter
}
```

---

**Remember**: Most runtime bugs are environment scoping issues. When in doubt, trace the environment chain! 