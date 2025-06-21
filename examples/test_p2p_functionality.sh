#!/bin/bash

# Test script to verify P2P functionality
set -e

echo "=== P2P Functionality Test ==="
echo

# Build the relay binary if needed
if [ ! -f "../relay" ]; then
    echo "Building relay binary..."
    cd ..
    go build -o relay cmd/relay/main.go
    cd examples
fi

# Create a simple test server file
cat > test_p2p_simple.rl << 'EOF'
server test_server {
    state {
        count: number = 0,
        message: string = "Hello from P2P server"
    }
    
    receive fn hello() -> string {
        state.get("message")
    }
    
    receive fn increment() -> number {
        state.set("count", state.get("count") + 1)
        state.get("count")
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
}
EOF

echo "Created test server file: test_p2p_simple.rl"
echo

# Start first server in background
echo "Starting first P2P node (port 8080)..."
../relay test_p2p_simple.rl -server -port 8080 -node-id node1 > server1.log 2>&1 &
SERVER1_PID=$!
sleep 2

# Start second server in background
echo "Starting second P2P node (port 8081)..."
../relay test_p2p_simple.rl -server -port 8081 -node-id node2 -add-peer http://127.0.0.1:8080 > server2.log 2>&1 &
SERVER2_PID=$!
sleep 2

# Function to cleanup servers
cleanup() {
    echo
    echo "Cleaning up servers..."
    kill $SERVER1_PID 2>/dev/null || true
    kill $SERVER2_PID 2>/dev/null || true
    rm -f test_p2p_simple.rl
    rm -f server1.log server2.log
}
trap cleanup EXIT

# Test 1: Health checks
echo "Test 1: Health checks"
echo "Node 1 health:"
curl -s http://127.0.0.1:8080/health | jq . || echo "Health check failed"
echo "Node 2 health:"
curl -s http://127.0.0.1:8081/health | jq . || echo "Health check failed"
echo

# Test 2: Server info
echo "Test 2: Server info"
echo "Node 1 info:"
curl -s http://127.0.0.1:8080/info | jq . || echo "Info check failed"
echo "Node 2 info:"
curl -s http://127.0.0.1:8081/info | jq . || echo "Info check failed"
echo

# Test 3: Registry endpoints
echo "Test 3: Registry endpoints"
echo "Node 1 registry:"
curl -s http://127.0.0.1:8080/registry | jq . || echo "Registry check failed"
echo "Node 1 servers:"
curl -s http://127.0.0.1:8080/registry/servers | jq . || echo "Servers check failed"
echo "Node 1 peers:"
curl -s http://127.0.0.1:8080/registry/peers | jq . || echo "Peers check failed"
echo

# Test 4: Local server calls
echo "Test 4: Local server calls"
echo "Calling test_server.hello on node 1:"
curl -s -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"test_server.hello","id":1}' | jq . || echo "Local call failed"

echo "Calling test_server.increment on node 1:"
curl -s -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"test_server.increment","id":2}' | jq . || echo "Local call failed"

echo "Calling test_server.get_count on node 1:"
curl -s -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"test_server.get_count","id":3}' | jq . || echo "Local call failed"
echo

# Test 5: WebSocket P2P endpoint availability
echo "Test 5: WebSocket P2P endpoint availability"
echo "Testing WebSocket connection to node 1..."
timeout 5 curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" "http://127.0.0.1:8080/ws/p2p?node_id=test_client" || echo "WebSocket test completed"
echo

# Test 6: Remote server calls
echo "Test 6: Remote server calls"
echo "Attempting remote call from node 1 to node 2:"
curl -s -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "remote_call",
    "params": {
      "node_id": "node2",
      "server_name": "test_server",
      "method": "hello",
      "args": [],
      "timeout": 5.0
    },
    "id": 4
  }' | jq . || echo "Remote call failed"
echo

# Test 7: Parameter validation
echo "Test 7: Parameter validation"
echo "Testing remote_call with missing node_id:"
curl -s -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "remote_call",
    "params": {
      "server_name": "test_server",
      "method": "hello"
    },
    "id": 5
  }' | jq . || echo "Validation test failed"
echo

# Show server logs
echo "=== Server 1 Log ==="
head -20 server1.log || echo "No server1.log found"
echo
echo "=== Server 2 Log ==="
head -20 server2.log || echo "No server2.log found"
echo

echo "=== P2P Functionality Test Complete ===" 