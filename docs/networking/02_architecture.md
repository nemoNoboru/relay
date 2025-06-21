# Unified Concurrency and Networking Architecture

## Overview

The Relay runtime and HTTP server use a **unified message routing architecture** built on a pure **Actor Model**. Each server runs in its own dedicated goroutine, and all communication—whether from local Relay code, HTTP JSON-RPC, or WebSocket P2P—is handled by a central `MessageRouter` actor.

This design provides complete transport transparency, eliminates the need for complex mutexes, and creates a robust, scalable foundation for concurrent applications.

## Key Benefits

### 🚀 **Transport Transparency**
- Relay servers don't care whether they're called via HTTP JSON-RPC or WebSocket P2P
- Same server code works for local calls, HTTP requests, and P2P messages
- No transport-specific logic in server implementations

### 🎯 **Actor-Based Architecture**
- Eliminates RW mutexes by using dedicated actors for shared state
- MessageRouter actor handles all server routing and registry management
- Thread-safe by design using Go channels and goroutines

### 🔧 **Simplified Integration**
- Single unified API for all server communication
- Consistent error handling across all transport mechanisms
- Easy to add new transport protocols (gRPC, TCP, etc.)

## Architecture Components

### 1. MessageRouter (Central Actor)

The `MessageRouter` is the heart of the system - a dedicated actor that handles all server communication.

```go
type MessageRouter struct {
    // Core routing
    servers    map[string]*Value // Local servers
    serversMux sync.RWMutex      // Temporary until fully actor-based
    
    // Message channels
    routeChannel    chan *RouteRequest
    registerChannel chan *RegisterRequest
    running         bool
    
    // P2P integration
    p2pNodes map[string]*P2PNode
    nodesMux sync.RWMutex
    
    // Response tracking
    pendingResponses map[string]chan *RouteResponse
    responsesMux     sync.RWMutex
}
```

**Key Features:**
- **Thread-safe server registry** - No locks needed for server operations
- **Unified routing** - Handles local and remote server calls identically
- **Response correlation** - Tracks async responses with correlation IDs
- **P2P node management** - Manages peer connections and health

### 2. TransportAdapter (Protocol Abstraction)

The `TransportAdapter` converts different transport protocols to unified `RouteRequest` messages.

```go
type TransportAdapter struct {
    router   *MessageRouter
    upgrader websocket.Upgrader
    nodeID   string
}
```

**Supported Transports:**
- **HTTP JSON-RPC 2.0** - Standard web API calls
- **WebSocket P2P** - Real-time peer-to-peer communication
- **Future**: gRPC, TCP, UDP, etc.

### 3. UnifiedHTTPServer (Integration Layer)

The `UnifiedHTTPServer` ties everything together. It initializes the `MessageRouter` and `TransportAdapter` and exposes the necessary HTTP and WebSocket endpoints.

```go
type UnifiedHTTPServer struct {
    config           *HTTPServerConfig
    evaluator        *Evaluator
    messageRouter    *MessageRouter
    transportAdapter *TransportAdapter
    server           *http.Server
    running          bool
    nodeID           string
}
```

## Server Goroutine Creation and Lifecycle

The actor model is brought to life when a Relay server is defined and instantiated.

### 1. Server Definition Evaluation

**Location**: `pkg/runtime/servers.go` - `evaluateServerExpr()`

When the evaluator encounters a `server` block, it:
1.  Initializes the server's private state and `receive` functions.
2.  Creates a `Server` instance containing this information.
3.  Registers the new server with the runtime.
4.  Calls `server.Start()` to launch the actor.

### 2. Goroutine Startup

**Location**: `pkg/runtime/value.go` - `Server.Start()`

The `Start` method launches the server's main loop in a new goroutine.

```go
func (s *Server) Start(evaluator interface{}) {
    go s.runServerLoop(evaluator)
}
```

### 3. The Actor's Heartbeat: The Message Loop

**Location**: `pkg/runtime/value.go` - `runServerLoop()`

This function is the entry point for the server's dedicated goroutine. It runs an infinite loop, waiting for messages on its private `MessageChan`.

```go
func (s *Server) runServerLoop(evaluator interface{}) {
    for message := range s.MessageChan {
        s.handleMessage(message, evaluator) // Process one message at a time
    }
}
```
This sequential, single-threaded processing of messages is the key to eliminating race conditions for server state.

## Message Flow

All communication is normalized into a `RouteRequest` and processed by the `MessageRouter`.

### HTTP JSON-RPC Request Flow

```
1. HTTP Request → TransportAdapter.HandleHTTPJSONRPC()
2. Parse JSON-RPC → Convert to RouteRequest
3. MessageRouter.RouteMessage() → Route to local/remote server
4. Server's MessageChan → Server processes message → Returns result
5. Result passed back through channels → JSON-RPC response
```

### WebSocket P2P Request Flow

```
1. WebSocket Message → TransportAdapter.HandleWebSocket()
2. Parse WebSocket message → Convert to RouteRequest  
3. MessageRouter.RouteMessage() → Route to local/remote server
4. Server's MessageChan → Server processes message → Returns result
5. Result passed back through channels → WebSocket response
```

### Local Server Call Flow

```
1. Relay code: message("server_name", "method_name", ...args)
2. Builtin 'message' function → Creates RouteRequest
3. MessageRouter.RouteMessage() → Route to local server
4. Server's MessageChan → Server processes message → Returns result
5. Result returned to caller
```

## API Reference

### MessageRouter API

```go
// Create and start router
router := NewMessageRouter()
router.Start()

// Register servers
router.RegisterServer("my_server", serverValue)

// Call any server (local or remote)
result, err := router.CallServer("node_id", "server_name", "method", args, timeout)

// Add P2P peers
router.AddP2PNode("peer_node_id", "peer_address")
```

### UnifiedHTTPServer API

```go
// Create server
config := &HTTPServerConfig{
    Host: "0.0.0.0",
    Port: 8080,
    EnableCORS: true,
    NodeID: "my_node",
}
server := NewUnifiedHTTPServer(evaluator, config)

// Start server
server.Start()

// Add peers
server.AddPeer("peer_node", "http://peer:8080")

// Call servers
result, err := server.CallServer("", "local_server", "method", args, 30*time.Second)
```

## HTTP Endpoints

### Core Endpoints

- `POST /rpc` - JSON-RPC 2.0 endpoint for server calls
- `GET /health` - Health check with server count
- `GET /info` - Server information and capabilities
- `GET /registry` - Server registry with local servers
- `GET /registry/servers` - List of available servers
- `GET /registry/peers` - List of peer nodes

### WebSocket Endpoints

- `ws://host:port/ws/p2p?node_id=peer_id` - P2P WebSocket connection

## Configuration

```go
type HTTPServerConfig struct {
    Host              string
    Port              int
    EnableCORS        bool
    ReadTimeout       time.Duration
    WriteTimeout      time.Duration
    Headers           map[string]string
    NodeID            string
    EnableRegistry    bool
    DiscoveryInterval time.Duration
}
```

## Usage Examples

### Basic Server Setup

```relay
server counter_server {
    state {
        count: number = 0
    }
    
    receive fn increment() -> number {
        set new_count = state.get("count") + 1
        state.set("count", new_count)
        new_count
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
}
```

### Start HTTP Server

```bash
# Start with unified architecture
./relay my_servers.rl -server -port 8080

# With P2P enabled
./relay my_servers.rl -server -port 8080 -add-peer http://peer:8081
```

### HTTP API Calls

```bash
# Call server method via HTTP JSON-RPC
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"counter_server.increment","id":1}'

# Get server registry
curl http://localhost:8080/registry

# Health check
curl http://localhost:8080/health
```

### Remote Server Calls

```bash
# Call server on remote node
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "remote_call",
    "params": {
      "node_id": "remote_node_123",
      "server_name": "counter_server", 
      "method": "increment",
      "args": [],
      "timeout": 30
    },
    "id": 1
  }'
```

## Implementation Details

### Thread Safety

- **MessageRouter**: Uses actor model with channels - no mutexes needed for core operations
- **Server Registry**: Managed by MessageRouter actor - thread-safe by design
- **Response Tracking**: Correlation IDs with timeout handling
- **P2P Connections**: Each connection handled by separate goroutine

### Error Handling

- **JSON-RPC Errors**: Standard error codes (-32700 to -32603)
- **Transport Errors**: Proper HTTP status codes and WebSocket error frames
- **Timeout Handling**: Configurable timeouts with proper cleanup
- **Connection Errors**: Automatic retry and reconnection logic

### Performance Characteristics

- **Concurrent Requests**: Unlimited concurrent HTTP requests
- **Server Calls**: Sequential processing per server (actor model)
- **Memory Usage**: Bounded channels prevent memory leaks
- **Latency**: Minimal overhead from unified routing

## Testing

### Run Unified Architecture Tests

```bash
cd examples
./test_unified_architecture.sh
```

### Test Coverage

- ✅ HTTP JSON-RPC calls
- ✅ WebSocket P2P connections  
- ✅ Server registry endpoints
- ✅ Error handling
- ✅ Health monitoring
- ✅ Remote server calls
- ✅ Transport transparency

## Migration Guide

### From Old HTTP Server

```go
// Old way
httpServer := runtime.NewHTTPServer(evaluator, config)

// New way (unified architecture)
httpServer := runtime.NewUnifiedHTTPServer(evaluator, config)
```

### API Compatibility

The new unified server maintains full API compatibility with the old server:
- Same HTTP endpoints
- Same JSON-RPC protocol
- Same configuration options
- Same CLI flags

## Future Enhancements

### Planned Features

1. **Full Actor-Based Registry** - Remove remaining mutexes
2. **gRPC Transport** - Add gRPC support via TransportAdapter
3. **TCP/UDP Transports** - Direct binary protocol support
4. **Load Balancing** - Distribute calls across multiple server instances
5. **Service Mesh** - Integration with Istio/Linkerd
6. **Metrics & Monitoring** - Prometheus/Grafana integration

### Extension Points

- **Custom Transports**: Implement new protocols via TransportAdapter
- **Custom Routing**: Extend MessageRouter for advanced routing logic
- **Middleware**: Add request/response middleware to TransportAdapter
- **Authentication**: Add auth middleware to UnifiedHTTPServer

## Conclusion

The unified architecture provides a clean, scalable foundation for Relay's HTTP server infrastructure. By making transport mechanisms transparent to Relay servers and using the actor model throughout, we've created a system that's both simple to use and powerful enough for complex distributed applications.

The architecture supports:
- **Seamless scaling** from single-node to multi-node deployments
- **Transport flexibility** with easy protocol additions
- **Developer simplicity** with consistent APIs across all transport mechanisms
- **Production reliability** with proper error handling and monitoring 