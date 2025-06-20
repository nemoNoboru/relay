# Peer-to-Peer Server Registry

## Overview

The Relay HTTP server now includes an **Exposable Server Registry** that enables peer-to-peer functionality. This allows Relay nodes to discover and communicate with servers running on other nodes, providing the foundation for building distributed Relay applications.

## Key Features

### 1. Server Discovery
- **Local Server Registration**: Automatically registers local servers with metadata
- **Peer Discovery**: Discover servers running on remote Relay nodes
- **Health Monitoring**: Track the health status of peer nodes
- **Automatic Updates**: Periodic discovery and health checks

### 2. HTTP Endpoints

The registry exposes several HTTP endpoints for peer-to-peer communication:

#### GET /registry
Returns complete registry information including local servers and peer nodes.

**Response:**
```json
{
  "node_id": "a1b2c3d4e5f6g7h8",
  "node_address": "localhost:8080",
  "servers": {
    "test_counter": {
      "name": "test_counter",
      "node_id": "a1b2c3d4e5f6g7h8",
      "node_address": "localhost:8080",
      "methods": ["increment", "get_count", "get_info", "reset"],
      "last_seen": "2024-01-01T12:00:00Z",
      "is_local": true
    }
  },
  "peers": {
    "b2c3d4e5f6g7h8i9": {
      "node_id": "b2c3d4e5f6g7h8i9",
      "address": "192.168.1.100:8080",
      "last_seen": "2024-01-01T12:00:00Z",
      "is_healthy": true,
      "server_count": 3
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### GET /registry/servers
Returns just the server list with metadata.

#### GET /registry/peers
Returns information about peer nodes.

#### POST /registry/peers/add
Add a new peer node.

**Request:**
```json
{
  "node_id": "peer_node_123",
  "address": "192.168.1.100:8080"
}
```

#### DELETE /registry/peers/remove
Remove a peer node.

**Request:**
```json
{
  "node_id": "peer_node_123"
}
```

#### GET /registry/health
Health check endpoint that returns node information.

**Response:**
```json
{
  "status": "healthy",
  "node_id": "a1b2c3d4e5f6g7h8",
  "node_address": "localhost:8080",
  "server_count": 2,
  "peer_count": 1,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Configuration

### HTTP Server Configuration

The `HTTPServerConfig` now includes P2P-related options:

```go
config := &runtime.HTTPServerConfig{
    Host:              "0.0.0.0",
    Port:              8080,
    EnableCORS:        true,
    ReadTimeout:       15 * time.Second,
    WriteTimeout:      15 * time.Second,
    Headers:           make(map[string]string),
    
    // P2P Configuration
    NodeID:            "custom_node_id",      // Auto-generated if empty
    EnableRegistry:    true,                  // Enable P2P registry
    DiscoveryInterval: 30 * time.Second,      // Peer discovery interval
}
```

### CLI Flags

New command-line flags support P2P functionality:

```bash
# Start server with custom node ID
./relay -server -node-id "my_custom_node"

# Add a peer node on startup
./relay -server -add-peer "http://192.168.1.100:8080"

# Disable registry functionality
./relay -server -disable-registry

# Start with Relay file and P2P enabled
./relay blog.rl -server -node-id "blog_node"
```

## Usage Examples

### 1. Basic Server Discovery

**Start Node 1:**
```bash
./relay examples/test_p2p_registry.rl -server -port 8080
```

**Start Node 2:**
```bash
./relay examples/test_p2p_registry.rl -server -port 8081
```

**Add Node 1 as peer to Node 2:**
```bash
curl -X POST http://localhost:8081/registry/peers/add \
  -H "Content-Type: application/json" \
  -d '{"node_id": "node1", "address": "localhost:8080"}'
```

**Discover servers on Node 1 from Node 2:**
```bash
curl http://localhost:8081/registry
```

### 2. Server Information Query

**Get all servers:**
```bash
curl http://localhost:8080/registry/servers
```

**Get peer information:**
```bash
curl http://localhost:8080/registry/peers
```

**Health check:**
```bash
curl http://localhost:8080/registry/health
```

### 3. JSON-RPC with Server Discovery

Once peers are discovered, you can call methods on remote servers:

```bash
# Call local server
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_counter.increment",
    "id": 1
  }'

# In the future, this will work for remote servers too:
# "method": "remote_node.test_counter.increment"
```

## Architecture

### Core Components

1. **ExposableServerRegistry**: Main registry class that wraps the standard ServerRegistry
2. **ServerInfo**: Metadata about servers including methods, node location, and health
3. **PeerNode**: Information about peer Relay nodes
4. **HTTP Endpoints**: RESTful API for registry operations

### Data Flow

```
Local Server Creation
        ↓
ExposableServerRegistry.RegisterServer()
        ↓
Update Local Server Metadata
        ↓
Expose via HTTP Endpoints
        ↓
Peer Discovery Process
        ↓
Remote Registry Queries
        ↓
Update Peer Information
```

### Thread Safety

- **Mutex Protection**: All registry operations are protected by read/write mutexes
- **Concurrent Access**: Multiple peers can query the registry simultaneously
- **Atomic Updates**: Server registration and peer updates are atomic operations

## API Reference

### ExposableServerRegistry Methods

```go
// Core ServerRegistry interface
RegisterServer(name string, server *Value)
GetServer(name string) (*Value, bool)
StopAllServers()

// P2P-specific methods
AddPeer(nodeID, address string)
RemovePeer(nodeID string)
GetPeers() map[string]*PeerNode
GetAllServers() map[string]*ServerInfo
AddPeerFromURL(peerURL string) error
DiscoverPeers() error
StartPeriodicDiscovery(interval time.Duration)

// Information methods
GetRegistryInfo() *ServerRegistryResponse
GetNodeInfo() map[string]interface{}
IsLocalServer(serverName string) bool
GetServerInfo(serverName string) (*ServerInfo, bool)
```

### HTTPServer P2P Methods

```go
GetExposableRegistry() *ExposableServerRegistry
GetNodeID() string
AddPeer(nodeID, address string)
RemovePeer(nodeID string)
```

## Future Enhancements

### 1. Remote Server Invocation
Currently, the registry can discover remote servers but cannot invoke methods on them. Future versions will support:

```relay
// Call method on remote server
set result = message("remote_node.server_name", "method_name", args...)
```

### 2. Automatic Peer Discovery
- **Broadcast Discovery**: Use UDP broadcast for local network discovery
- **DNS-SD**: Service discovery via DNS
- **Gossip Protocol**: Peer-to-peer gossip for distributed discovery

### 3. Load Balancing
- **Round-Robin**: Distribute calls across multiple instances of the same server
- **Health-Based**: Route to healthy nodes only
- **Geographic**: Route based on node location

### 4. Security
- **Authentication**: Secure peer-to-peer communication
- **Authorization**: Control which peers can access which servers
- **Encryption**: TLS for all peer communication

## Testing

### Test File: examples/test_p2p_registry.rl

The test file creates two servers:
- `test_counter`: A simple counter with increment/get operations
- `discovery_test`: Tracks peer discovery events

**Run the test:**
```bash
# Terminal 1
./relay examples/test_p2p_registry.rl -server -port 8080

# Terminal 2  
./relay examples/test_p2p_registry.rl -server -port 8081

# Terminal 3 - Add peer and test
curl -X POST http://localhost:8081/registry/peers/add \
  -H "Content-Type: application/json" \
  -d '{"node_id": "node1", "address": "localhost:8080"}'

curl http://localhost:8081/registry
```

## Troubleshooting

### Common Issues

1. **Peer Connection Failed**
   - Check network connectivity between nodes
   - Verify port accessibility
   - Check firewall settings

2. **Registry Not Available**
   - Ensure `EnableRegistry` is true in config
   - Check that registry endpoints are accessible
   - Verify HTTP server is running

3. **Server Not Discovered**
   - Check that server is properly registered
   - Verify peer discovery is running
   - Check network connectivity to peer nodes

### Debug Commands

```bash
# Check registry health
curl http://localhost:8080/registry/health

# View all servers
curl http://localhost:8080/registry/servers | jq

# View peer status
curl http://localhost:8080/registry/peers | jq

# Test server method
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "test_counter.get_info", "id": 1}' | jq
```

## Summary

The Exposable Server Registry provides the foundation for peer-to-peer Relay applications by:

1. **Automatic Server Discovery**: Servers are automatically registered and discoverable
2. **HTTP API**: RESTful endpoints for registry operations
3. **Health Monitoring**: Track peer node health and availability
4. **Extensible Architecture**: Foundation for future distributed features

This enables building distributed Relay applications where servers can be spread across multiple nodes while maintaining the same programming model and actor-based communication patterns. 