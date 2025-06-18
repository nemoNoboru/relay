# Relay Runtime Quick Reference

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