# Relay Server Goroutine Architecture & Communication

## Overview

Relay implements a pure **Actor Model** where each server runs in its own dedicated goroutine. This document explains where these goroutines are created, how they communicate with the main HTTP server, and the complete message-passing architecture.

## Server Goroutine Creation

### 1. Server Definition and Instantiation

Server goroutines are created during the server definition evaluation process:

**Location**: `pkg/runtime/servers.go` - `evaluateServerExpr()`

```go
func (e *Evaluator) evaluateServerExpr(expr *parser.ServerExpr, env *Environment) (*Value, error) {
    // 1. Initialize server state and receivers
    state := make(map[string]*Value)
    receivers := make(map[string]*Function)
    
    // 2. Process server body (state fields, receive functions)
    // ... state and receiver initialization ...
    
    // 3. Create server instance
    serverValue := NewServer(expr.Name, state, receivers, env)
    
    // 4. Store server in registry
    e.servers[expr.Name] = serverValue
    
    // 5. Start the server goroutine
    serverValue.Server.Start(e)  // <-- GOROUTINE CREATED HERE
    
    return serverValue, nil
}
```

### 2. Goroutine Start Function

**Location**: `pkg/runtime/value.go` - `Server.Start()`

```go
// Start starts the server goroutine
func (s *Server) Start(evaluator interface{}) {
    if s.Running {
        return
    }
    
    s.Running = true
    go s.runServerLoop(evaluator)  // <-- NEW GOROUTINE STARTS HERE
}
```

### 3. Server Message Loop (Goroutine Entry Point)

**Location**: `pkg/runtime/value.go` - `runServerLoop()`

```go
// runServerLoop runs the main server message handling loop
func (s *Server) runServerLoop(evaluator interface{}) {
    for s.Running {
        select {
        case message, ok := <-s.MessageChan:  // <-- BLOCKING WAIT FOR MESSAGES
            if !ok {
                return // Channel closed
            }
            s.handleMessage(message, evaluator)  // <-- SEQUENTIAL MESSAGE PROCESSING
        }
    }
}
```

## Server Actor Architecture

### Core Components

Each Relay server actor consists of:

1. **Dedicated Goroutine**: Runs `runServerLoop()`
2. **Message Channel**: Buffered channel (`chan *Message`) with capacity 100
3. **Private State**: `map[string]*Value` accessible only within the goroutine
4. **Receiver Functions**: Methods that can be called via messages
5. **Sequential Processing**: Messages processed one at a time (no race conditions)

### Message Structure

```go
type Message struct {
    Method string        // Method name to call
    Args   []*Value     // Arguments for the method
    Reply  chan *Value  // Channel for synchronous reply (optional)
}
```

## Communication Patterns

### 1. HTTP Server → Relay Server Communication

The HTTP server acts as a **Federation Proxy Actor** that routes JSON-RPC requests to Relay servers.

**Flow**:
1. HTTP request arrives at `/rpc` endpoint.
2. The `handleJSONRPC` method parses the request into a `JSONRPCRequest` struct.
3. It calls `processRPCCall` with the full request object.
4. `processRPCCall` inspects the method name.
   - If the method is `remote_call`, it's handled by the P2P system.
   - Otherwise, it parses the `server_name.method_name` from the `method` field.
5. It looks up the server in the registry.
6. It converts the JSON params to Relay values.
7. It sends a `Message` to the server's `MessageChan` and waits for a reply.
8. The result is converted back to JSON and sent in the HTTP response.

**Location**: `pkg/runtime/http_server.go` - `processRPCCall()`

```go
// processRPCCall processes a JSON-RPC method call
func (h *HTTPServer) processRPCCall(request JSONRPCRequest) (interface{}, error) {
	// Check if this is a remote server call
	if request.Method == "remote_call" {
		return h.processRemoteCall(request)
	}

	// Local server call
	serverName, methodName, err := h.parseMethod(request.Method)
	if err != nil {
		return nil, &JSONRPCError{Code: -32600, Message: "Invalid method format"}
	}

	server, exists := h.evaluator.GetServer(serverName)
	if !exists {
		return nil, &JSONRPCError{Code: -32601, Message: "Server not found"}
	}

	// Convert parameters to Relay values
	args, err := h.convertParams(request.Params)
	if err != nil {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params"}
	}

	// Send message to the server
	result, err := server.Server.SendMessage(methodName, args, true)
	if err != nil {
		return nil, &JSONRPCError{Code: -32000, Message: "Server error"}
	}

	// Convert result back to JSON
	return h.convertValueToJSON(result), nil
}
```

### 2. Relay Code → Relay Server Communication

Relay code uses the built-in `message()` function to communicate with servers.

**Location**: `pkg/runtime/builtins.go` - `defineMessageFunction()`

```go
messageFunc := &Value{
    Type: ValueTypeFunction,
    Function: &Function{
        Name:      "message",
        IsBuiltin: true,
        Builtin: func(args []*Value) (*Value, error) {
            serverName := args[0].Str
            methodName := args[1].Str
            methodArgs := args[2:]
            
            // Find server in registry
            server, exists := e.GetServer(serverName)
            if !exists {
                return nil, fmt.Errorf("server '%s' not found", serverName)
            }
            
            // Send message synchronously
            result, err := server.Server.SendMessage(methodName, methodArgs, true)
            return result, err
        },
    },
}
```

### 3. Server → Server Communication

Servers can communicate with each other through the `message()` function:

```relay
server producer {
    receive fn produce_item() -> string {
        set item = "item_" + generate_id()
        
        // Send to consumer server
        set result = message("consumer", "consume", item)
        
        item
    }
}
```

## Message Processing Flow

### 1. Message Sending (`SendMessage`)

**Location**: `pkg/runtime/value.go` - `SendMessage()`

```go
func (s *Server) SendMessage(method string, args []*Value, waitForReply bool) (*Value, error) {
    // 1. Create reply channel (if synchronous)
    var replyChan chan *Value
    if waitForReply {
        replyChan = make(chan *Value, 1)
    }
    
    // 2. Create message
    message := &Message{
        Method: method,
        Args:   args,
        Reply:  replyChan,
    }
    
    // 3. Send to server's message channel (non-blocking with timeout)
    select {
    case s.MessageChan <- message:
        if waitForReply {
            // 4. Wait for reply (blocking with timeout)
            select {
            case reply := <-replyChan:
                return reply, nil
            case <-time.After(5 * time.Second):
                return nil, fmt.Errorf("timeout waiting for reply")
            }
        }
        return NewNil(), nil
    case <-time.After(1 * time.Second):
        return nil, fmt.Errorf("failed to send message: channel full")
    }
}
```

### 2. Message Handling (`handleMessage`)

**Location**: `pkg/runtime/value.go` - `handleMessage()`

```go
func (s *Server) handleMessage(message *Message, evaluator interface{}) {
    // 1. Find receiver function
    receiver, exists := s.Receivers[message.Method]
    if !exists {
        if message.Reply != nil {
            message.Reply <- NewNil() // Send error response
        }
        return
    }
    
    // 2. Create server environment with state access
    serverEnv := NewEnvironment(s.Environment)
    
    // 3. Add 'state' object to environment (actor-safe, no mutex needed)
    stateValue := NewServerStateActorSafe(&s.State)
    serverEnv.Define("state", stateValue)
    
    // 4. Execute receiver function
    result, err := evaluator.CallUserFunction(receiver, message.Args, serverEnv)
    
    // 5. Send reply (if requested)
    if message.Reply != nil {
        if err != nil {
            message.Reply <- NewNil()
        } else {
            message.Reply <- result
        }
    }
}
```

## Thread Safety & Concurrency

### 1. Actor Isolation

- **Each server has its own goroutine**: No shared state between servers
- **Sequential message processing**: Messages processed one at a time per server
- **No mutexes needed**: Actor model eliminates race conditions within servers
- **Atomic state updates**: State changes within a message handler are atomic

### 2. Thread-Safe Communication

- **Channel-based messaging**: Go channels provide thread-safe communication
- **Buffered channels**: Prevent blocking on message sends (capacity: 100)
- **Timeout mechanisms**: Prevent deadlocks and hanging operations
- **Graceful shutdown**: Channels closed properly when servers stop

### 3. State Management

**Location**: `pkg/runtime/server_state_methods.go`

```go
// State access in actor model (no mutex needed)
func (h *ServerStateMethodHandler) getMethod(target *Value, args []*Value) (*Value, error) {
    key := args[0].Str
    
    // No mutex needed - server processes messages sequentially
    if value, exists := (*target.ServerState.State)[key]; exists {
        return value, nil
    }
    return NewNil(), nil
}
```

## Lifecycle Management

### Server Startup

1. **Definition**: Server structure parsed from Relay code
2. **Creation**: `NewServer()` creates server instance with channel
3. **Registration**: Server stored in evaluator's registry
4. **Goroutine Start**: `server.Start()` launches dedicated goroutine
5. **Message Loop**: Goroutine enters blocking message processing loop

### Server Shutdown

**Location**: `pkg/runtime/value.go` - `Stop()`

```go
func (s *Server) Stop() {
    if !s.Running {
        return
    }
    
    s.Running = false
    close(s.MessageChan)  // Closes channel, causing message loop to exit
}
```

**Graceful shutdown in main**:
```go
func main() {
    evaluator := runtime.NewEvaluator()
    defer evaluator.StopAllServers()  // Ensures all servers stop
    
    // ... rest of application ...
}
```

## Performance Characteristics

### 1. Concurrency Model

- **Multiple servers run concurrently**: Each in its own goroutine
- **Sequential processing per server**: No internal race conditions
- **Non-blocking sends**: Channel sends have timeouts to prevent blocking
- **Efficient scheduling**: Go runtime handles goroutine scheduling

### 2. Memory Model

- **Isolated state**: Each server has private state map
- **Shared read-only data**: Receiver functions and environment are read-only
- **Message copying**: Arguments copied when sent between servers
- **Garbage collection**: Unused servers and messages are garbage collected

### 3. Scalability

- **Horizontal scaling**: Easy to add more servers
- **Load distribution**: Different servers can handle different concerns
- **Federation ready**: Architecture supports distributed deployment
- **Resource efficient**: Goroutines are lightweight (2KB stack)

## HTTP Server Integration

The HTTP server acts as a **Federation Proxy Actor** that:

1. **Receives JSON-RPC requests** from external clients
2. **Routes requests** to appropriate Relay servers
3. **Handles type conversion** between JSON and Relay values
4. **Manages timeouts** and error responses
5. **Provides standard endpoints** (health, info, RPC)

**Key insight**: The HTTP server doesn't run in a separate goroutine for each request. Instead, it uses Go's standard HTTP server which handles requests in separate goroutines, but the HTTP server actor logic runs synchronously within each request handler.

## Summary

Relay's goroutine architecture implements a pure Actor Model where:

- **Server goroutines are created** during server definition evaluation in `evaluateServerExpr()`
- **Each server runs independently** in its own goroutine with a message processing loop
- **Communication is channel-based** using buffered channels for message passing
- **The HTTP server acts as a proxy** routing external requests to internal server actors
- **Thread safety is achieved** through actor isolation rather than locks
- **Scalability is built-in** through the concurrent, isolated server design

This architecture provides the foundation for Relay's federation capabilities, allowing servers to be transparently distributed across multiple processes or machines while maintaining the same programming model. 