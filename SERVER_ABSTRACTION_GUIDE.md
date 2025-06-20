# Server Abstraction Architecture Guide

## Problem Statement

The original Relay server implementation was tightly coupled to the parser and evaluator classes, making it difficult to work on server internals independently. Key issues included:

1. **Parser Dependency**: Server functions contained `*parser.Block` AST nodes
2. **Evaluator Coupling**: Servers directly called `evaluator.CallUserFunction()`
3. **Fragile Type Assertions**: Used brittle interface{} casting patterns
4. **Mixed Concerns**: Server logic intertwined with expression evaluation

## Solution: Interface-Based Abstraction Layer

We've created a clean abstraction layer that separates server concerns from parser/evaluator implementation details.

### Key Interfaces

#### 1. ExecutionEngine
```go
type ExecutionEngine interface {
    ExecuteFunction(handler FunctionHandler, args []*Value, env *Environment) (*Value, error)
    CreateEnvironment(parent *Environment) *Environment
    EvaluateDefaultValue(expr interface{}, env *Environment) (*Value, error)
}
```
**Purpose**: Abstracts how functions are executed, allowing different execution strategies.

#### 2. FunctionHandler
```go
type FunctionHandler interface {
    GetName() string
    GetParameters() []string
    Execute(engine ExecutionEngine, args []*Value, env *Environment) (*Value, error)
}
```
**Purpose**: Represents executable functions without depending on parser AST structures.

#### 3. ServerFactory
```go
type ServerFactory interface {
    CreateServer(name string, definition ServerDefinition, env *Environment) (*Value, error)
}
```
**Purpose**: Creates servers from abstract definitions, not parser-specific types.

#### 4. ServerRegistry
```go
type ServerRegistry interface {
    RegisterServer(name string, server *Value)
    GetServer(name string) (*Value, bool)
    StopAllServers()
}
```
**Purpose**: Manages server instances independently of the evaluator.

### Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│  ServerCore (Decoupled)    │    Abstraction Interfaces      │
│  - Message handling        │    - ExecutionEngine           │
│  - State management        │    - FunctionHandler           │
│  - Concurrency            │    - ServerFactory             │
│  - Lifecycle              │    - ServerRegistry            │
├─────────────────────────────────────────────────────────────┤
│                    Adapter Layer                           │
│  - EvaluatorExecutionEngine                                │
│  - FunctionHandlerAdapter                                  │
│  - EvaluatorServerFactory                                  │
│  - EvaluatorServerRegistry                                 │
├─────────────────────────────────────────────────────────────┤
│               Current Implementation                        │
│  Evaluator  │  Parser  │  Original Server                  │
└─────────────────────────────────────────────────────────────┘
```

## Benefits

### 1. **Separation of Concerns**
- Server logic is independent of parser implementation
- Can change parser without affecting server internals
- Clear boundaries between evaluation and message handling

### 2. **Testability**
- Mock execution engines for unit testing
- Test server behavior without full parser/evaluator setup
- Isolated testing of message handling, state management, etc.

### 3. **Extensibility**
- Create custom execution engines (HTTP servers, Go actors, etc.)
- Plugin architecture for different server behaviors
- Alternative function handler implementations

### 4. **Future-Proofing**
- Ready for HTTP server implementation
- Supports planned Go Actor system
- Federation-ready architecture

## Migration Path

### Phase 1: Current (Backward Compatible)
```go
// Old way still works
evaluator := NewEvaluator()
serverExpr := &parser.ServerExpr{...}
server, err := evaluator.evaluateServerExpr(serverExpr, env)
```

### Phase 2: Using Abstractions
```go
// New abstracted way
evaluator := NewEvaluator()
engine := NewEvaluatorExecutionEngine(evaluator)
registry := NewEvaluatorServerRegistry(evaluator)
factory := NewEvaluatorServerFactory(evaluator, registry, engine)

definition := ServerDefinition{
    Name: "my_server",
    State: map[string]StateField{...},
    Receivers: map[string]FunctionHandler{...},
}

server, err := factory.CreateServer("my_server", definition, env)
```

### Phase 3: Custom Implementations
```go
// Future: Custom execution engines
httpEngine := NewHTTPExecutionEngine(port)
goActorEngine := NewGoActorExecutionEngine()

// Mix and match components
factory := NewCustomServerFactory(httpEngine, customRegistry)
```

## Working on Server Internals

With this abstraction, you can now focus on server internals without worrying about parser details:

### Example: Adding HTTP Server Support
```go
type HTTPExecutionEngine struct {
    port int
    mux  *http.ServeMux
}

func (e *HTTPExecutionEngine) ExecuteFunction(handler FunctionHandler, args []*Value, env *Environment) (*Value, error) {
    // Convert to HTTP endpoint
    endpoint := "/" + handler.GetName()
    e.mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
        // Parse HTTP request to args
        // Call handler.Execute()
        // Convert result to HTTP response
    })
    return NewString("endpoint_created"), nil
}
```

### Example: Adding Persistence
```go
type PersistentServerCore struct {
    *ServerCore
    database Database
}

func (s *PersistentServerCore) SetState(key string, value *Value) {
    s.ServerCore.SetState(key, value)
    s.database.Save(s.Name, key, value) // Persist to disk
}
```

### Example: Adding Monitoring
```go
type MonitoredExecutionEngine struct {
    engine ExecutionEngine
    metrics Metrics
}

func (e *MonitoredExecutionEngine) ExecuteFunction(handler FunctionHandler, args []*Value, env *Environment) (*Value, error) {
    start := time.Now()
    result, err := e.engine.ExecuteFunction(handler, args, env)
    e.metrics.RecordExecution(handler.GetName(), time.Since(start), err)
    return result, err
}
```

## Files Overview

- **`server_interfaces.go`**: Core abstraction interfaces
- **`server_core.go`**: Decoupled server implementation  
- **`server_adapters.go`**: Bridges to existing evaluator/parser
- **`server_usage_example.go`**: Usage examples and migration patterns

## Next Steps

1. **Gradual Migration**: Start using abstractions for new server features
2. **HTTP Implementation**: Create `HTTPExecutionEngine` for web endpoints
3. **Go Actor Integration**: Implement `GoActorExecutionEngine` 
4. **Persistence Layer**: Add persistent state management
5. **Federation Support**: Extend for service discovery and communication

This architecture provides a solid foundation for evolving Relay's server capabilities while maintaining clean separation of concerns and testability. 