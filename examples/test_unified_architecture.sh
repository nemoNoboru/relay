#!/bin/bash

# Test script for unified HTTP server architecture
# This tests the new MessageRouter and TransportAdapter system

echo "=== Relay Unified Architecture Test ==="
echo

# Build the relay binary
echo "Building Relay..."
cd .. && go build -o relay cmd/relay/main.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build Relay"
    exit 1
fi
echo "✅ Relay built successfully"
echo

# Start the server in background
echo "Starting Relay HTTP server with unified architecture..."
./relay examples/test_unified_architecture.rl -server -port 8082 &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Server started with PID: $SERVER_PID"
echo

# Test 1: Health check
echo "🔍 Test 1: Health check"
HEALTH_RESPONSE=$(curl -s http://localhost:8082/health)
echo "Response: $HEALTH_RESPONSE"
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
fi
echo

# Test 2: Server info
echo "🔍 Test 2: Server info"
INFO_RESPONSE=$(curl -s http://localhost:8082/info)
echo "Response: $INFO_RESPONSE"
if echo "$INFO_RESPONSE" | grep -q "unified_message_router"; then
    echo "✅ Server info shows unified architecture"
else
    echo "❌ Server info doesn't show unified architecture"
fi
echo

# Test 3: Server registry
echo "🔍 Test 3: Server registry"
REGISTRY_RESPONSE=$(curl -s http://localhost:8082/registry)
echo "Response: $REGISTRY_RESPONSE"
if echo "$REGISTRY_RESPONSE" | grep -q "test_server"; then
    echo "✅ Server registry shows test_server"
else
    echo "❌ Server registry doesn't show test_server"
fi
echo

# Test 4: JSON-RPC call to test_server.hello
echo "🔍 Test 4: JSON-RPC call to test_server.hello"
HELLO_RESPONSE=$(curl -s -X POST http://localhost:8082/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"test_server.hello","id":1}')
echo "Response: $HELLO_RESPONSE"
if echo "$HELLO_RESPONSE" | grep -q "Hello from unified architecture"; then
    echo "✅ test_server.hello call successful"
else
    echo "❌ test_server.hello call failed"
fi
echo

# Test 5: JSON-RPC call to test_server.increment
echo "🔍 Test 5: JSON-RPC call to test_server.increment"
INCREMENT_RESPONSE=$(curl -s -X POST http://localhost:8082/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"test_server.increment","id":2}')
echo "Response: $INCREMENT_RESPONSE"
if echo "$INCREMENT_RESPONSE" | grep -q '"result":1'; then
    echo "✅ test_server.increment call successful"
else
    echo "❌ test_server.increment call failed"
fi
echo

# Test 6: JSON-RPC call to echo_server.ping
echo "🔍 Test 6: JSON-RPC call to echo_server.ping"
PING_RESPONSE=$(curl -s -X POST http://localhost:8082/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"echo_server.ping","id":3}')
echo "Response: $PING_RESPONSE"
if echo "$PING_RESPONSE" | grep -q "pong"; then
    echo "✅ echo_server.ping call successful"
else
    echo "❌ echo_server.ping call failed"
fi
echo

# Test 7: WebSocket endpoint availability
echo "🔍 Test 7: WebSocket endpoint availability"
# Use curl to test WebSocket upgrade (will fail but show endpoint exists)
WS_RESPONSE=$(curl -s -I -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost:8082/ws/p2p 2>&1)
if echo "$WS_RESPONSE" | grep -q "400\|426"; then
    echo "✅ WebSocket endpoint is available (expected 400/426 for curl)"
else
    echo "❌ WebSocket endpoint not available"
fi
echo

# Test 8: Error handling for non-existent server
echo "🔍 Test 8: Error handling for non-existent server"
ERROR_RESPONSE=$(curl -s -X POST http://localhost:8082/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"nonexistent_server.test","id":4}')
echo "Response: $ERROR_RESPONSE"
if echo "$ERROR_RESPONSE" | grep -q "error"; then
    echo "✅ Proper error handling for non-existent server"
else
    echo "❌ Error handling failed"
fi
echo

# Cleanup
echo "🧹 Cleaning up..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null
echo "✅ Server stopped"
echo

echo "=== Unified Architecture Test Complete ===" 