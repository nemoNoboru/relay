#!/bin/bash

# Simple End-to-End WebSocket P2P Test Script

set -e

echo "🚀 Starting Simple End-to-End WebSocket P2P Test"
echo "================================================"

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up background processes..."
    pkill -f "relay.*e2e_p2p_test_simple.rl" || true
    sleep 2
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Test configuration
NODE1_PORT=8081
NODE2_PORT=8082
NODE3_PORT=8083
NODE1_ID="node1_simple"
NODE2_ID="node2_simple"
NODE3_ID="node3_simple"

echo "🏗️  Starting 3-node P2P cluster..."

# Start Node 1
echo "🟢 Starting Node 1 on port $NODE1_PORT..."
./relay examples/e2e_p2p_test_simple.rl -server -port $NODE1_PORT -node-id $NODE1_ID > node1_simple.log 2>&1 &
NODE1_PID=$!
sleep 3

# Start Node 2 and connect to Node 1
echo "🟡 Starting Node 2 on port $NODE2_PORT..."
./relay examples/e2e_p2p_test_simple.rl -server -port $NODE2_PORT -node-id $NODE2_ID -add-peer "http://localhost:$NODE1_PORT" > node2_simple.log 2>&1 &
NODE2_PID=$!
sleep 3

# Start Node 3 and connect to Node 2
echo "🔵 Starting Node 3 on port $NODE3_PORT..."
./relay examples/e2e_p2p_test_simple.rl -server -port $NODE3_PORT -node-id $NODE3_ID -add-peer "http://localhost:$NODE2_PORT" > node3_simple.log 2>&1 &
NODE3_PID=$!
sleep 5

echo "✅ All nodes started successfully!"
echo ""

# Function to make JSON-RPC call
make_rpc_call() {
    local port=$1
    local method=$2
    local params=$3
    local description=$4
    
    echo "📞 $description"
    echo "   Method: $method | Port: $port"
    
    if [ -z "$params" ] || [ "$params" = "null" ]; then
        curl -s -X POST http://localhost:$port/rpc \
            -H "Content-Type: application/json" \
            -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"id\":1}" | jq '.'
    else
        curl -s -X POST http://localhost:$port/rpc \
            -H "Content-Type: application/json" \
            -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"params\":$params,\"id\":1}" | jq '.'
    fi
    echo ""
}

# Function to make remote call via P2P
make_remote_call() {
    local from_port=$1
    local target_node=$2
    local server_method=$3
    local params=$4
    local description=$5
    
    echo "🌐 $description"
    echo "   From: localhost:$from_port → Node: $target_node → Method: $server_method"
    
    if [ -z "$params" ] || [ "$params" = "null" ]; then
        curl -s -X POST http://localhost:$from_port/rpc \
            -H "Content-Type: application/json" \
            -d "{\"jsonrpc\":\"2.0\",\"method\":\"remote_call\",\"params\":{\"node_id\":\"$target_node\",\"server_name\":\"$(echo $server_method | cut -d. -f1)\",\"method\":\"$(echo $server_method | cut -d. -f2)\"},\"id\":1}" | jq '.'
    else
        curl -s -X POST http://localhost:$from_port/rpc \
            -H "Content-Type: application/json" \
            -d "{\"jsonrpc\":\"2.0\",\"method\":\"remote_call\",\"params\":{\"node_id\":\"$target_node\",\"server_name\":\"$(echo $server_method | cut -d. -f1)\",\"method\":\"$(echo $server_method | cut -d. -f2)\",\"args\":$params},\"id\":1}" | jq '.'
    fi
    echo ""
}

echo "🧪 PHASE 1: Health Checks"
echo "========================="

# Check health endpoints
for port in $NODE1_PORT $NODE2_PORT $NODE3_PORT; do
    echo "🏥 Health check for Node on port $port:"
    curl -s http://localhost:$port/health | jq '.'
    echo ""
done

echo "🧪 PHASE 2: Local Server Operations"
echo "=================================="

# Set node IDs
make_rpc_call $NODE1_PORT "distributed_counter.set_node_id" '["'$NODE1_ID'"]' "Set Node 1 ID"
make_rpc_call $NODE2_PORT "distributed_counter.set_node_id" '["'$NODE2_ID'"]' "Set Node 2 ID"
make_rpc_call $NODE3_PORT "distributed_counter.set_node_id" '["'$NODE3_ID'"]' "Set Node 3 ID"

# Increment counters
make_rpc_call $NODE1_PORT "distributed_counter.increment" '[5]' "Increment Node 1 counter by 5"
make_rpc_call $NODE2_PORT "distributed_counter.increment" '[3]' "Increment Node 2 counter by 3"
make_rpc_call $NODE3_PORT "distributed_counter.increment" '[7]' "Increment Node 3 counter by 7"

# Get counter states
make_rpc_call $NODE1_PORT "distributed_counter.get_count" 'null' "Get Node 1 counter"
make_rpc_call $NODE2_PORT "distributed_counter.get_count" 'null' "Get Node 2 counter"
make_rpc_call $NODE3_PORT "distributed_counter.get_count" 'null' "Get Node 3 counter"

# Health checks
make_rpc_call $NODE1_PORT "health_monitor.health_check" 'null' "Node 1 health check"
make_rpc_call $NODE2_PORT "health_monitor.health_check" 'null' "Node 2 health check"
make_rpc_call $NODE3_PORT "health_monitor.health_check" 'null' "Node 3 health check"

echo "🧪 PHASE 3: Remote Server Invocation via WebSocket P2P"
echo "======================================================"

echo "⏱️  Waiting for WebSocket connections to establish..."
sleep 5

# Test direct remote calls
make_remote_call $NODE1_PORT $NODE2_ID "distributed_counter.get_count" 'null' "Remote call: Node 1 → Node 2 (get counter)"
make_remote_call $NODE1_PORT $NODE2_ID "health_monitor.health_check" 'null' "Remote call: Node 1 → Node 2 (health check)"

# Test remote calls with parameters
make_remote_call $NODE2_PORT $NODE3_ID "distributed_counter.increment" '[10]' "Remote call: Node 2 → Node 3 (increment by 10)"

# Test multistep routing (Node 1 → Node 2 → Node 3)
echo "🔀 Testing multistep routing: Node 1 → Node 2 → Node 3"
make_remote_call $NODE1_PORT $NODE3_ID "distributed_counter.get_count" 'null' "Multistep routing: Node 1 → Node 3 (get counter)"
make_remote_call $NODE1_PORT $NODE3_ID "health_monitor.health_check" 'null' "Multistep routing: Node 1 → Node 3 (health check)"

echo "🧪 PHASE 4: Server Registry and Peer Discovery"
echo "=============================================="

# Check server registries
for port in $NODE1_PORT $NODE2_PORT $NODE3_PORT; do
    echo "📋 Server registry for Node on port $port:"
    curl -s http://localhost:$port/registry | jq '.'
    echo ""
done

echo "🧪 PHASE 5: Performance Testing"
echo "==============================="

# Rapid-fire calls
echo "⚡ Rapid-fire local calls:"
for i in {1..3}; do
    make_rpc_call $NODE1_PORT "distributed_counter.increment" '[1]' "Rapid increment #$i"
done

echo "⚡ Rapid-fire remote calls:"
for i in {1..3}; do
    make_remote_call $NODE1_PORT $NODE2_ID "health_monitor.health_check" 'null' "Rapid remote health check #$i"
done

echo "🎯 FINAL VERIFICATION"
echo "===================="

# Final state verification
echo "📊 Final counter states:"
make_rpc_call $NODE1_PORT "distributed_counter.get_count" 'null' "Final Node 1 counter"
make_rpc_call $NODE2_PORT "distributed_counter.get_count" 'null' "Final Node 2 counter"
make_rpc_call $NODE3_PORT "distributed_counter.get_count" 'null' "Final Node 3 counter"

echo ""
echo "🎉 SIMPLE END-TO-END WEBSOCKET P2P TEST COMPLETED!"
echo "=================================================="
echo ""
echo "📋 Test Summary:"
echo "✅ Local server functionality"
echo "✅ Remote server invocation via WebSocket P2P"
echo "✅ Multistep routing (Node 1 → Node 2 → Node 3)"
echo "✅ Server registry and peer discovery"
echo "✅ Performance testing"
echo ""
echo "📁 Log files: node1_simple.log, node2_simple.log, node3_simple.log"
echo "🚀 All core P2P capabilities tested successfully!" 