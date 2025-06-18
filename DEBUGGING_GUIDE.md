# Relay Runtime Debugging Guide

This guide provides practical debugging techniques specifically for the Relay runtime, with real examples and step-by-step troubleshooting procedures.

## Quick Debugging Checklist

When you encounter a runtime error, check these common issues first:

- [ ] **Environment Passing**: Are you using `EvaluateWithEnv(expr, env)` instead of `Evaluate(expr)`?
- [ ] **Function Arguments**: Are function arguments evaluated with the correct environment?
- [ ] **Parameter Binding**: Are function parameters properly bound in the function environment?
- [ ] **Parser Coverage**: Does the runtime handle all expression types from the parser?
- [ ] **Method Calls**: Are method calls using the correct environment for argument evaluation?

## Step-by-Step Debugging Process

### 1. Reproduce the Issue

#### Create a Minimal Test Case
```bash
# Create the simplest possible failing case
echo 'fn test(a) { helper(a) }' > debug.rl
echo 'fn helper(x) { x * 2 }' >> debug.rl
echo 'test(21)' >> debug.rl

# Test it
relay debug.rl
```

#### Use the REPL for Interactive Testing
```bash
# Load and test interactively
relay debug.rl -repl

# In REPL, test components individually:
# > helper(21)
# > test(21)
```

### 2. Add Environment Tracing

#### Add Debug Helpers (temporarily)
```go
// Add to pkg/runtime/functions.go or wherever needed
func debugEnvironment(env *Environment, context string) {
    fmt.Printf("DEBUG [%s]: Environment depth: %d\n", context, environmentDepth(env))
    fmt.Printf("DEBUG [%s]: Variables: %v\n", context, getEnvironmentKeys(env))
}

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

#### Trace Function Calls
```go
// In evaluateFuncCall or evaluateLiteralFuncCall
func (e *Evaluator) evaluateFuncCall(expr *parser.FuncCallExpr, env *Environment) (*Value, error) {
    debugEnvironment(env, fmt.Sprintf("FuncCall[%s]", expr.Name))
    
    // Look up the function
    funcValue, exists := env.Get(expr.Name)
    // ... rest of function
}
```

#### Trace Parameter Binding
```go
// In CallUserFunction
func (e *Evaluator) CallUserFunction(fn *Function, args []*Value, parentEnv *Environment) (*Value, error) {
    debugEnvironment(parentEnv, "CallUserFunction:before")
    
    // Create function environment...
    
    // Bind parameters
    for i, param := range fn.Parameters {
        fmt.Printf("DEBUG: Binding param '%s' = %v\n", param, args[i])
        funcEnv.Define(param, args[i])
    }
    
    debugEnvironment(funcEnv, "CallUserFunction:after_binding")
    
    // Execute function body...
}
```

### 3. Trace Expression Evaluation

#### Add Evaluation Tracing
```go
// In EvaluateWithEnv
func (e *Evaluator) EvaluateWithEnv(expr *parser.Expression, env *Environment) (*Value, error) {
    fmt.Printf("DEBUG: EvaluateWithEnv - Expression type: %T\n", getExpressionType(expr))
    debugEnvironment(env, "EvaluateWithEnv")
    
    // ... normal evaluation
}

func getExpressionType(expr *parser.Expression) string {
    if expr.Binary != nil { return "Binary" }
    if expr.SetExpr != nil { return "SetExpr" }
    if expr.FunctionExpr != nil { return "FunctionExpr" }
    // ... etc
    return "Unknown"
}
```

#### Trace Identifier Lookups
```go
// In evaluateBaseExpr
if expr.Identifier != nil {
    fmt.Printf("DEBUG: Looking up identifier '%s'\n", *expr.Identifier)
    debugEnvironment(env, fmt.Sprintf("Identifier[%s]", *expr.Identifier))
    
    value, exists := env.Get(*expr.Identifier)
    if !exists {
        fmt.Printf("DEBUG: Identifier '%s' NOT FOUND!\n", *expr.Identifier)
        return nil, fmt.Errorf("undefined variable: %s", *expr.Identifier)
    }
    return value, nil
}
```

### 4. Common Error Patterns and Solutions

#### Pattern 1: "undefined variable" in function calls
```bash
Error: undefined variable: myParam
```

**Diagnosis**: Function parameter not found during evaluation

**Debug Steps**:
1. Check parameter binding in `CallUserFunction`
2. Check environment passing in argument evaluation
3. Look for `Evaluate()` calls that should be `EvaluateWithEnv()`

**Common Fix**:
```go
// WRONG
value, err := e.Evaluate(arg)

// RIGHT  
value, err := e.EvaluateWithEnv(arg, env)
```

#### Pattern 2: "undefined function" for function parameters
```bash
Error: undefined function: f
```

**Diagnosis**: Function parameter treated as function name lookup instead of parameter

**Debug Steps**:
1. Check if function is being looked up in wrong environment
2. Verify function parameter is bound as a value, not a function definition
3. Check if function call is going through correct evaluation path

#### Pattern 3: Variables work in simple access but not in expressions
```bash
# This works:
fn test(a) { a }

# This fails:
fn test(a) { helper(a) }
```

**Diagnosis**: Environment not passed to sub-expressions

**Debug Steps**:
1. Trace where `a` is evaluated in `helper(a)`
2. Check environment in function call argument evaluation
3. Look for missing environment passing

### 5. Testing Specific Components

#### Test Environment Chaining
```go
func TestEnvironmentChaining() {
    // Create environment chain
    global := NewEnvironment(nil)
    global.Define("global_var", NewString("global"))
    
    local := NewEnvironment(global)
    local.Define("local_var", NewString("local"))
    
    // Test lookups
    val, exists := local.Get("local_var")  // Should find in local
    val, exists := local.Get("global_var") // Should find in parent
    val, exists := local.Get("missing")    // Should not find
}
```

#### Test Function Parameter Binding
```go
func TestParameterBinding() {
    code := `
        fn test(param1, param2) {
            param1 + param2
        }
        test(10, 20)
    `
    result := evalCode(t, code)
    assert.Equal(t, 30.0, result.Number)
}
```

#### Test Nested Function Calls
```go
func TestNestedFunctionCalls() {
    code := `
        fn inner(x) { x * 2 }
        fn outer(y) { inner(y) + 1 }
        outer(5)
    `
    result := evalCode(t, code)
    assert.Equal(t, 11.0, result.Number) // (5 * 2) + 1
}
```

## Advanced Debugging Techniques

### 1. AST Inspection

Use the REPL's AST mode to see how code is parsed:

```bash
relay -repl
> :astmode
> fn test(a) { helper(a) }
```

This shows the exact AST structure and helps identify parser vs runtime issues.

### 2. Runtime State Inspection

Add temporary inspection code:

```go
// Inspect function object
fmt.Printf("DEBUG: Function info - Name: %s, Params: %v, IsBuiltin: %v\n", 
    funcValue.Function.Name, 
    funcValue.Function.Parameters, 
    funcValue.Function.IsBuiltin)

// Inspect environment state
current := env
level := 0
for current != nil {
    fmt.Printf("DEBUG: Env Level %d: %d variables\n", level, len(current.variables))
    for key, value := range current.variables {
        fmt.Printf("  %s: %s\n", key, value.String())
    }
    current = current.parent
    level++
}
```

### 3. Differential Debugging

Compare working vs non-working cases:

```bash
# Working case
echo 'fn test(a) { a }' > working.rl
echo 'test(42)' >> working.rl

# Broken case  
echo 'fn test(a) { helper(a) }' > broken.rl
echo 'fn helper(x) { x }' >> broken.rl
echo 'test(42)' >> broken.rl

# Compare execution with debug output
relay working.rl  # Should work
relay broken.rl   # Shows the difference
```

### 4. Binary Search Debugging

For complex expressions, simplify step by step:

```bash
# Original failing case
fn complex(a, b, c) { 
    helper1(a) + helper2(b, c) 
}

# Test 1: Simplify to single call
fn complex(a, b, c) { 
    helper1(a) 
}

# Test 2: Test with constants
fn complex(a, b, c) { 
    helper1(42) 
}

# Test 3: Test helper directly
helper1(42)
```

## Debugging Workflow Example

Let's walk through debugging a real issue:

### Problem
```bash
$ relay test.rl
Error: undefined variable: x
```

### Step 1: Isolate
```bash
# Create minimal failing case
echo 'fn double(x) { x * 2 }' > debug.rl
echo 'fn apply(f, val) { f(val) }' >> debug.rl  
echo 'apply(double, 5)' >> debug.rl

relay debug.rl
# Error: undefined variable: val
```

### Step 2: Add Debug Output
```go
// In evaluateLiteralFuncCall (or relevant function)
func (e *Evaluator) evaluateLiteralFuncCall(funcCall *parser.FuncCall, env *Environment) (*Value, error) {
    fmt.Printf("DEBUG: FuncCall %s with env: %v\n", funcCall.Name, getEnvironmentKeys(env))
    
    for i, arg := range funcCall.Args {
        fmt.Printf("DEBUG: Evaluating arg %d\n", i)
        value, err := e.EvaluateWithEnv(arg, env) // Check this line!
        if err != nil {
            fmt.Printf("DEBUG: Arg %d failed: %v\n", i, err)
            return nil, err
        }
    }
}
```

### Step 3: Run with Debug
```bash
relay debug.rl
```

Output:
```
DEBUG: FuncCall double with env: [L0:f L0:val L1:double L1:apply]
DEBUG: Evaluating arg 0
DEBUG: FuncCall apply with env: [L0:apply L0:double]  # Wrong env!
DEBUG: Arg 0 failed: undefined variable: val
```

### Step 4: Identify Root Cause
The environment changes from `[L0:f L0:val ...]` to `[L0:apply L0:double]` - the function parameters disappear!

### Step 5: Find the Bug
The issue is in argument evaluation - it's using the global environment instead of the function environment.

### Step 6: Fix
```go
// Change from:
value, err := e.Evaluate(arg) // Uses global env

// To:
value, err := e.EvaluateWithEnv(arg, env) // Uses function env
```

### Step 7: Test Fix
```bash
relay debug.rl
# Result: 10  ✅ Fixed!
```

### Step 8: Clean Up
Remove debug output and run full test suite to ensure no regressions.

## Prevention Strategies

### 1. Code Review Checklist
- [ ] All `Evaluate()` calls should be `EvaluateWithEnv()` with proper environment
- [ ] Function argument evaluation uses the calling environment
- [ ] New expression types have corresponding runtime handlers
- [ ] Environment creation properly sets parent chains

### 2. Testing Guidelines
- Always test function parameters as arguments to other functions
- Test nested function calls and closures
- Test variable scoping edge cases
- Create negative test cases for error conditions

### 3. Documentation
- Document environment flow for new features
- Update this guide when adding new expression types
- Include debugging examples in code comments

---

Remember: The most common runtime bugs in Relay are environment scoping issues. When in doubt, trace the environment chain and verify that variables are accessible where they're being used. 