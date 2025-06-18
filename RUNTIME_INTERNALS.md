# Relay Runtime Internals Documentation

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

- **`pkg/runtime/evaluator.go`** - Main evaluation coordinator
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

The expression evaluation follows a precise chain to handle operator precedence and scoping:

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

*For questions or clarifications about the runtime internals, refer to the code examples and debugging techniques outlined above.* 