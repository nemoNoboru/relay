# P2P to Hub-and-Spoke Federation Migration Plan

## Overview

This document outlines the migration strategy from the existing P2P mesh architecture to the new hub-and-spoke federation model. The migration preserves existing functionality while transforming the architecture to support real-world NAT traversal and scalable federation.

## Migration Phases

### Phase 1: Core Architecture Refactoring
*Goal: Transform P2P components to federation-aware components*

#### Task 1.1: Refactor WebSocketP2P to FederationRouter
**Implementation Strategy:**
- Rename `WebSocketP2P` class to `FederationRouter`
- Add `nodeType` field to distinguish Main vs Home Relays
- Split connection management logic based on node type
- Preserve existing message handling infrastructure
- Maintain backward compatibility during transition

**Acceptance Criteria:**
- [ ] `FederationRouter` class replaces `WebSocketP2P` with same interface
- [ ] Node type detection works: `NodeTypeMain` and `NodeTypeHome` constants
- [ ] Existing WebSocket endpoints continue to work
- [ ] Message routing logic preserved for current P2P functionality
- [ ] All existing tests pass with new class name
- [ ] Connection management split into Main/Home specific methods

**Test Cases:**
```bash
# Test 1: Basic FederationRouter startup
./relay -server -node-type main
# Expected: FederationRouter initializes as Main Relay

# Test 2: Node type detection
./relay blog.rl -node-type home -connect ws://localhost:8080
# Expected: FederationRouter initializes as Home Relay

# Test 3: Backward compatibility
# Run existing P2P test with new router
./examples/test_websocket_p2p.sh
# Expected: All tests pass with FederationRouter
```

#### Task 1.2: Update Message Protocol for Federation
**Implementation Strategy:**
- Rename `P2PMessage` to `FederationMessage`
- Add `SourceNode` and `DestinationNode` with `NodeAddress` structure
- Replace simple node IDs with compound addressing (type + ID + main relay)
- Add federation-specific fields: `ServerName`, `Priority`, `Timeout`
- Update message serialization/deserialization

**Acceptance Criteria:**
- [ ] `FederationMessage` struct includes all required federation fields
- [ ] `NodeAddress` supports compound addressing: `{type, id, main_relay}`
- [ ] Message serialization preserves backward compatibility
- [ ] All message types work with new structure: `registration`, `heartbeat`, `request`, `response`
- [ ] JSON marshaling/unmarshaling works correctly
- [ ] Message validation rejects malformed federation messages

**Test Cases:**
```go
// Test 1: Message structure validation
msg := &FederationMessage{
    Type: "request",
    SourceNode: NodeAddress{Type: "home", ID: "alice.home", MainRelay: "relay.com"},
    DestinationNode: NodeAddress{Type: "home", ID: "bob.home", MainRelay: "relay.com"},
    ServerName: "blog_service",
    Method: "get_posts",
    TTL: 10,
}
// Expected: Message validates and serializes correctly

// Test 2: Backward compatibility
oldStyleMessage := map[string]interface{}{
    "type": "ping",
    "from": "node1", 
    "to": "node2",
}
// Expected: Message adapter converts to new format

// Test 3: Compound addressing
addr := NodeAddress{
    Type: "home",
    ID: "alice.home",
    MainRelay: "relay.example.com",
}
// Expected: Address parsing and validation works
```

#### Task 1.3: Split Connection Management by Node Type
**Implementation Strategy:**
- Create `HomeRelayConnectionManager` for outbound-only connections
- Create `MainRelayConnectionManager` for incoming + peer connections
- Move connection health monitoring to type-specific managers
- Implement automatic reconnection for Home Relays
- Add peer connection management for Main Relays

**Acceptance Criteria:**
- [ ] `HomeRelayConnectionManager` handles single outbound Main Relay connection
- [ ] `MainRelayConnectionManager` handles multiple Home + peer connections
- [ ] Connection health monitoring specific to connection type
- [ ] Automatic reconnection works for Home Relays (exponential backoff)
- [ ] Peer connection establishment works for Main Relays
- [ ] Connection cleanup properly handles both connection types

**Test Cases:**
```bash
# Test 1: Home Relay connection management
./relay blog.rl -node-type home -connect ws://relay.com:8080
# Expected: Single outbound connection established and maintained

# Test 2: Main Relay connection management
./relay -server -node-type main -peers ws://peer.relay.com:8080
# Expected: Accepts incoming Home connections + establishes peer connections

# Test 3: Connection recovery
# Start Home Relay, stop Main Relay, restart Main Relay
./relay blog.rl -node-type home -connect ws://localhost:8080 &
./relay -server -node-type main & 
kill %2  # Kill main relay
./relay -server -node-type main &  # Restart
# Expected: Home Relay automatically reconnects
```

### Phase 2: Federation-Specific Logic
*Goal: Implement hub-and-spoke routing and service discovery*

#### Task 2.1: Implement Node Type Auto-Detection
**Implementation Strategy:**
- Detect node type from CLI flags: `-server` implies main, `-connect` implies home
- Add explicit `-node-type` flag to override detection
- Generate node IDs automatically if not specified
- Validate node type consistency with other flags
- Update help text and error messages

**Acceptance Criteria:**
- [ ] `-server` without `-connect` auto-detects as Main Relay
- [ ] `-connect` without `-server` auto-detects as Home Relay
- [ ] `-node-type main` or `-node-type home` overrides detection
- [ ] Error if incompatible flags used together
- [ ] Auto-generated node IDs are unique and valid
- [ ] Node type logged clearly at startup

**Test Cases:**
```bash
# Test 1: Main Relay auto-detection
./relay -server
# Expected: "Starting as Main Relay with auto-generated node ID"

# Test 2: Home Relay auto-detection  
./relay blog.rl -connect ws://relay.com:8080
# Expected: "Starting as Home Relay with auto-generated node ID"

# Test 3: Explicit node type
./relay -server -node-type home
# Expected: Error - incompatible flags

# Test 4: Custom node ID
./relay -server -node-id custom.relay.com
# Expected: Uses custom node ID instead of auto-generated
```

#### Task 2.2: Implement Federation Service Registry
**Implementation Strategy:**
- Replace `ExposableServerRegistry` with `FederationServiceRegistry`
- Add separate storage for local vs federated services
- Implement service propagation between Main Relays
- Add TTL-based expiration for federated services
- Create service lookup with local-first preference

**Acceptance Criteria:**
- [ ] `FederationServiceRegistry` separates local and federated services
- [ ] Service registration from Home Relays updates local registry
- [ ] Service propagation between Main Relays works automatically
- [ ] TTL expiration removes stale federated services (10 minute default)
- [ ] Service lookup checks local first, then federated
- [ ] Registry endpoints show both local and federated services

**Test Cases:**
```bash
# Setup: 2 Main Relays + 2 Home Relays
./relay -server -node-type main -node-id relay-a.com -port 8080 &
./relay -server -node-type main -node-id relay-b.com -port 8081 \
  -peers ws://relay-a.com:8080 &

./relay blog.rl -node-type home -node-id alice.home \
  -connect ws://relay-a.com:8080 &
./relay shop.rl -node-type home -node-id bob.home \
  -connect ws://relay-b.com:8081 &

# Test 1: Local service registration
curl http://relay-a.com:8080/registry
# Expected: Shows alice.home services as local

# Test 2: Federated service discovery
curl http://relay-a.com:8080/federation/registry  
# Expected: Shows both alice.home (local) and bob.home (federated) services

# Test 3: Service propagation
curl http://relay-b.com:8081/federation/registry
# Expected: Shows both bob.home (local) and alice.home (federated) services
```

#### Task 2.3: Implement Hub-and-Spoke Routing Logic
**Implementation Strategy:**
- Create role-specific routing in `FederationRouter`
- Home Relays: all remote calls go to Main Relay
- Main Relays: route to local Home or peer Main based on destination
- Implement service location lookup before routing
- Add routing path tracking for debugging
- Remove complex P2P flooding algorithms

**Acceptance Criteria:**
- [ ] Home Relay routing sends all remote calls to connected Main Relay
- [ ] Main Relay routing distinguishes between local Home and peer Main destinations
- [ ] Service location lookup determines correct routing target
- [ ] Routing path tracked in messages for debugging
- [ ] Maximum 3 hops for any message (Home → Main → Main → Home)
- [ ] Routing errors properly reported back to caller

**Test Cases:**
```bash
# Test 1: Home to Home (same Main)
# alice.home calls bob.home (both on relay-a.com)
curl -X POST http://relay-a.com:8080/rpc -d '{
  "jsonrpc": "2.0",
  "method": "remote_call", 
  "params": {
    "node_id": "bob.home@relay-a.com",
    "server_name": "shop_service",
    "method": "get_products"
  },
  "id": 1
}'
# Expected: 2-hop routing (alice → relay-a → bob)

# Test 2: Home to Home (different Main)
# alice.home calls charlie.home (on relay-b.com)
curl -X POST http://relay-a.com:8080/rpc -d '{
  "jsonrpc": "2.0",
  "method": "remote_call",
  "params": {
    "node_id": "charlie.home@relay-b.com", 
    "server_name": "news_service",
    "method": "get_headlines"
  },
  "id": 2
}'
# Expected: 3-hop routing (alice → relay-a → relay-b → charlie)

# Test 3: Service not found
curl -X POST http://relay-a.com:8080/rpc -d '{
  "jsonrpc": "2.0",
  "method": "remote_call",
  "params": {
    "node_id": "nonexistent.home",
    "server_name": "fake_service", 
    "method": "test"
  },
  "id": 3
}'
# Expected: Service not found error
```

### Phase 3: Advanced Federation Features
*Goal: Add production-ready features and optimizations*

#### Task 3.1: Implement Main Relay Peer Discovery and Health
**Implementation Strategy:**
- Add peer Main Relay discovery through DNS or configuration
- Implement health checking between Main Relays
- Add automatic peer connection recovery
- Create peer status monitoring and reporting
- Implement peer failover for service routing

**Acceptance Criteria:**
- [ ] Main Relays can discover peers through configuration or DNS
- [ ] Health checking between Main Relays works (30 second intervals)
- [ ] Automatic peer reconnection with exponential backoff
- [ ] Peer status visible in federation health endpoint
- [ ] Service routing fails over to healthy peers only
- [ ] Peer connection metrics tracked and reported

**Test Cases:**
```bash
# Test 1: Peer discovery and connection
./relay -server -node-type main -node-id relay-a.com \
  -peers ws://relay-b.com:8081,ws://relay-c.com:8082
# Expected: Attempts connections to both peers, logs success/failure

# Test 2: Peer health monitoring
curl http://relay-a.com:8080/federation/peers
# Expected: Shows health status of all configured peers

# Test 3: Peer failover
# Stop relay-b, verify relay-a routes through relay-c
# Expected: Automatic failover to healthy peer for cross-federation calls
```

#### Task 3.2: Add Load Balancing and Service Selection
**Implementation Strategy:**
- Implement service instance tracking (multiple nodes can provide same service)
- Add load balancing algorithms: round-robin, least-connections, random
- Create service health scoring based on response times
- Implement service preference (local > federated)
- Add service call metrics tracking

**Acceptance Criteria:**
- [ ] Multiple instances of same service tracked separately
- [ ] Round-robin load balancing distributes calls evenly
- [ ] Health-based selection excludes unhealthy instances
- [ ] Local services preferred over federated when available
- [ ] Service call metrics collected (latency, success rate)
- [ ] Load balancing configuration per service

**Test Cases:**
```bash
# Setup: Multiple instances of same service
./relay blog1.rl -node-type home -node-id alice1.home -connect ws://relay-a.com:8080 &
./relay blog2.rl -node-type home -node-id alice2.home -connect ws://relay-a.com:8080 &

# Test 1: Load balancing across instances
for i in {1..10}; do
  curl -X POST http://relay-a.com:8080/rpc -d '{
    "jsonrpc": "2.0",
    "method": "remote_call",
    "params": {
      "service_name": "blog_service",
      "method": "get_posts"
    },
    "id": '$i'
  }'
done
# Expected: Calls distributed between alice1.home and alice2.home

# Test 2: Health-based selection
# Stop alice1.home
curl -X POST http://relay-a.com:8080/rpc -d '{
  "jsonrpc": "2.0", 
  "method": "remote_call",
  "params": {
    "service_name": "blog_service",
    "method": "get_posts" 
  },
  "id": 1
}'
# Expected: Calls only go to healthy alice2.home
```

#### Task 3.3: Enhanced Error Handling and Monitoring
**Implementation Strategy:**
- Add comprehensive error codes for federation failures
- Implement request tracing across federation hops
- Create federation health dashboards and metrics
- Add timeout handling with configurable timeouts
- Implement circuit breaker patterns for failed services

**Acceptance Criteria:**
- [ ] Federation-specific error codes for common failure modes
- [ ] Request tracing shows full path through federation
- [ ] Health metrics available via `/federation/health` endpoint
- [ ] Configurable timeouts per service call (default 60s for federation)
- [ ] Circuit breaker temporarily blocks calls to failed services
- [ ] Detailed logging for troubleshooting federation issues

**Test Cases:**
```bash
# Test 1: Request tracing
curl -X POST http://relay-a.com:8080/rpc -d '{
  "jsonrpc": "2.0",
  "method": "remote_call",
  "params": {
    "node_id": "bob.home@relay-b.com",
    "server_name": "shop_service",
    "method": "get_products",
    "trace": true
  },
  "id": 1
}'
# Expected: Response includes routing path and timing information

# Test 2: Circuit breaker
# Configure shop_service to fail, make 10 calls, verify circuit opens
# Expected: After threshold failures, circuit breaker blocks subsequent calls

# Test 3: Federation health dashboard
curl http://relay-a.com:8080/federation/health
# Expected: 
# {
#   "status": "healthy",
#   "connected_homes": 5,
#   "connected_peers": 2, 
#   "total_services": 12,
#   "failed_services": 1,
#   "average_latency": "45ms"
# }
```

### Phase 4: Production Readiness and Testing
*Goal: Comprehensive testing and production deployment preparation*

#### Task 4.1: Comprehensive Integration Testing
**Implementation Strategy:**
- Create automated test suite for full federation scenarios
- Test failure recovery: node restarts, network partitions, timeouts
- Performance testing with realistic loads
- Memory and connection leak testing
- Security testing for federation endpoints

**Acceptance Criteria:**
- [ ] Automated test covers 3-Main + 6-Home Relay federation
- [ ] Failure recovery tests pass: node restart, network partition, timeout
- [ ] Performance test: 1000 concurrent cross-federation calls complete
- [ ] Memory usage stable during 24-hour test run
- [ ] Security scan passes for all federation endpoints
- [ ] Test suite runs in CI/CD pipeline

**Test Cases:**
```bash
# Test 1: Full federation test
./test_full_federation.sh
# Expected: Creates 3 Main + 6 Home federation, runs all scenarios

# Test 2: Chaos testing
./test_federation_chaos.sh --duration 1h --failure-rate 10%
# Expected: Random failures injected, system recovers automatically

# Test 3: Load testing
./test_federation_load.sh --calls 1000 --concurrency 50
# Expected: All calls complete within 60 seconds, no errors
```

#### Task 4.2: Documentation and Examples
**Implementation Strategy:**
- Create comprehensive federation setup guide
- Write troubleshooting guide for common issues
- Create example applications showcasing federation
- Document API endpoints and CLI flags
- Create architecture diagrams and sequence diagrams

**Acceptance Criteria:**
- [ ] Federation quickstart guide enables setup in under 15 minutes
- [ ] Troubleshooting guide covers 90% of common issues
- [ ] Example applications demonstrate real-world federation use cases
- [ ] API documentation complete with examples and error codes
- [ ] Architecture diagrams accurately reflect implementation
- [ ] Documentation tested by following guides exactly

**Test Cases:**
```bash
# Test 1: Quickstart guide validation
# Follow federation-quickstart.md step by step
# Expected: Working federation in under 15 minutes

# Test 2: Example applications
./relay examples/federated-blog.rl -node-type home -connect ws://relay.com:8080
./relay examples/federated-shop.rl -node-type home -connect ws://relay.com:8080
# Call blog from shop across federation
# Expected: Cross-federation calls work as documented

# Test 3: Troubleshooting guide accuracy
# Introduce common problems, follow troubleshooting steps
# Expected: Documented solutions resolve 90% of issues
```

#### Task 4.3: Production Deployment and Monitoring
**Implementation Strategy:**
- Create Docker images for Main and Home Relays
- Add Kubernetes deployment configurations
- Implement Prometheus metrics integration
- Create Grafana dashboards for federation monitoring
- Add health check endpoints for load balancers

**Acceptance Criteria:**
- [ ] Docker images build and run correctly for both relay types
- [ ] Kubernetes deployments scale Main Relays horizontally
- [ ] Prometheus metrics exported for all key federation metrics
- [ ] Grafana dashboards provide real-time federation visibility
- [ ] Health check endpoints work with standard load balancers
- [ ] Production deployment guide covers cloud deployment

**Test Cases:**
```bash
# Test 1: Docker deployment
docker run -p 8080:8080 relay-main:latest -server -node-type main
docker run relay-home:latest blog.rl -node-type home -connect ws://relay-main:8080
# Expected: Federation works in containerized environment

# Test 2: Kubernetes deployment
kubectl apply -f k8s/relay-main-deployment.yaml
kubectl apply -f k8s/relay-home-deployment.yaml
# Expected: Pods start and federation establishes automatically

# Test 3: Monitoring integration
curl http://relay-main:8080/metrics
# Expected: Prometheus metrics in correct format
```

## Migration Timeline

### Week 1-2: Phase 1 (Core Refactoring)
- Refactor WebSocketP2P to FederationRouter
- Update message protocol
- Split connection management

### Week 3-4: Phase 2 (Federation Logic)
- Implement node type detection
- Create federation service registry
- Build hub-and-spoke routing

### Week 5-6: Phase 3 (Advanced Features)
- Add peer discovery and health
- Implement load balancing
- Enhance error handling

### Week 7-8: Phase 4 (Production Ready)
- Comprehensive testing
- Documentation
- Production deployment

## Risk Mitigation

### Technical Risks
- **Breaking Changes**: Maintain backward compatibility interfaces during transition
- **Performance Regression**: Benchmark each phase against baseline
- **Data Migration**: Provide migration tools for existing deployments

### Deployment Risks
- **Service Interruption**: Blue-green deployment strategy for Main Relays
- **Configuration Errors**: Comprehensive validation and error messages
- **Rollback Plan**: Ability to revert to P2P mode if needed

## Success Criteria

The migration is successful when:

1. **Functional Compatibility**: All existing P2P functionality works with federation
2. **Performance**: Federation performance meets or exceeds P2P performance
3. **Scalability**: Support for 1000+ Home Relays per Main Relay
4. **Reliability**: 99.9% uptime for federation services
5. **Usability**: Setup time reduced from P2P complexity to simple federation
6. **Documentation**: Complete guides enable external adoption

This migration plan provides a systematic approach to transforming the P2P architecture into a production-ready federation system while minimizing risk and ensuring comprehensive testing at each phase.