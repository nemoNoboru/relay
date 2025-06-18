# Relay Runtime Package

The runtime package provides the core evaluation engine for the Relay programming language. It implements a unified, well-documented architecture that handles all language constructs including functions, structs, servers, and expressions.

## Architecture Overview

The runtime follows a hierarchical evaluation pattern:
```
Expression → Binary → Unary → Primary → Base → Literals/Identifiers
```

### Core Files

- **`core.go`** - Unified expression evaluation engine
- **`evaluator.go`** - Main evaluator type and initialization  
- **`value.go`** - Value system and environment management
- **`operations.go`** - Arithmetic and logical operations
- **`methods.go`** - Object and array method dispatch
- **`functions.go`** - Function calling and built-ins
- **`structs.go`** - Struct definitions and instantiation
- **`servers.go`** - Server infrastructure and state management

## Key Features

### ✅ Unified Expression Evaluation
All expression types are handled through a single, consistent evaluation path in `core.go`:
- Binary operations (arithmetic, logical, comparisons)
- Unary operations (negation, logical NOT)
- Function calls and method chaining
- Field access and array indexing
- Variable assignment and scoping

### ✅ First-Class Functions with Closures
Functions are true closures that capture their lexical environment:
```relay
fn make_counter(start) {
  set count = start
  fn() { 
    set count = count + 1
    count 
  }
}

set counter = make_counter(10)
counter()  // Returns 11
counter()  // Returns 12 - state preserved!
```

### ✅ Immutable-by-Default Semantics
All operations return new values rather than modifying existing ones:
```relay
set user = {name: "John", age: 30}
set updated = user.set("age", 31)  // Returns new object
// user is unchanged, updated has new age
```

### ✅ Struct System
Complete struct definition and instantiation:
```relay
struct User {
  name: string,
  email: string,
  age: number
}

set user = User{name: "John", email: "john@test.com", age: 30}
set name = user.get("name")  // Field access
```

### ✅ Server State Management
Stateful servers with automatic concurrency control:
```relay
server counter {
  state {
    count: number = 0
  }
  
  receive fn increment() -> number {
    state.set("count", state.get("count") + 1)
    state.get("count")
  }
}
```

## Value System

The runtime uses a tagged union approach for values:

```go
type Value struct {
    Type        ValueType         // Determines which field is valid
    Number      float64           // For numbers
    Str         string            // For strings
    Bool        bool              // For booleans
    Array       []*Value          // For arrays
    Object      map[string]*Value // For objects
    Function    *Function         // For functions (includes closures)
    Struct      *Struct           // For struct instances
    Server      *Server           // For server instances
    ServerState *ServerState      // For mutable server state
}
```

## Environment System

Lexical scoping with parent chain support:
- Variables are looked up in current scope first, then parent scopes
- Function closures capture their defining environment
- Each function call creates a new environment with parameter bindings

## Error Handling

The runtime uses Go's error system with special handling for:
- **Return values**: Wrapped in `ReturnValue` type for early returns
- **Type errors**: Clear messages about type mismatches
- **Undefined variables/functions**: Specific error messages
- **Server errors**: Concurrent error handling for message passing

## Testing

The runtime has comprehensive test coverage with **739 tests passing**:
- Expression evaluation tests
- Function and closure tests  
- Struct definition and usage tests
- Server state management tests
- Array and object method tests
- Error handling tests

## Performance Considerations

- **Unified evaluation**: Single code path reduces method call overhead
- **Immutable values**: Prevents defensive copying in most cases
- **Closure optimization**: Environments are shared where possible
- **Server concurrency**: Proper mutex usage for state access

## Usage Example

```go
// Create evaluator
evaluator := runtime.NewEvaluator()

// Parse and evaluate expression
result, err := evaluator.Evaluate(expr)
if err != nil {
    log.Fatal(err)
}

// Use result
fmt.Println(result.String())
```

## Future Improvements

The unified architecture provides a solid foundation for:
- Additional expression types
- Performance optimizations
- Better error messages
- Debugging support
- Code generation backends

This runtime successfully balances simplicity with power, providing a clean foundation for the Relay language while maintaining excellent test coverage and performance. 