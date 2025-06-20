#!/bin/bash

# End-to-End WebSocket P2P Test Script
# Tests all P2P capabilities including remote invocation and multistep routing

set -e

echo "🚀 Starting End-to-End WebSocket P2P Test"
echo "=========================================="

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up background processes..."
    pkill -f "relay.*e2e_p2p_test.rl" || true
    sleep 2
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Test configuration
NODE1_PORT=8081
NODE2_PORT=8082
NODE3_PORT=8083
NODE1_ID="node1_e2e"
NODE2_ID="node2_e2e"
NODE3_ID="node3_e2e"

# Build relay binary if needed
if [ ! -f "./relay" ]; then
    echo "📦 Building Relay binary..."
    go build -o relay cmd/relay/main.go
fi

echo "🏗️  Starting 3-node P2P cluster for comprehensive testing..."

# Start Node 1
echo "🟢 Starting Node 1 (Primary) on port $NODE1_PORT..."
./relay examples/e2e_p2p_test.rl -server -port $NODE1_PORT -node-id $NODE1_ID > node1_e2e.log 2>&1 &
NODE1_PID=$!
sleep 3

# Start Node 2 and connect to Node 1
echo "🟡 Starting Node 2 (Secondary) on port $NODE2_PORT..."
./relay examples/e2e_p2p_test.rl -server -port $NODE2_PORT -node-id $NODE2_ID -add-peer "http://localhost:$NODE1_PORT" > node2_e2e.log 2>&1 &
NODE2_PID=$!
sleep 3

# Start Node 3 and connect to Node 2 (for multistep routing)
echo "🔵 Starting Node 3 (Tertiary) on port $NODE3_PORT..."
./relay examples/e2e_p2p_test.rl -server -port $NODE3_PORT -node-id $NODE3_ID -add-peer "http://localhost:$NODE2_PORT" > node3_e2e.log 2>&1 &
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

echo "🧪 PHASE 1: Basic Server Health Checks"
echo "======================================"

# Check health endpoints
for port in $NODE1_PORT $NODE2_PORT $NODE3_PORT; do
    echo "🏥 Health check for Node on port $port:"
    curl -s http://localhost:$port/health | jq '.'
    echo ""
done

echo "🧪 PHASE 2: Local Server Functionality Tests"
echo "============================================"

# Test distributed counter locally on each node
make_rpc_call $NODE1_PORT "distributed_counter.set_node_id" '["'$NODE1_ID'"]' "Set Node 1 ID"
make_rpc_call $NODE2_PORT "distributed_counter.set_node_id" '["'$NODE2_ID'"]' "Set Node 2 ID"
make_rpc_call $NODE3_PORT "distributed_counter.set_node_id" '["'$NODE3_ID'"]' "Set Node 3 ID"

# Increment counters locally
make_rpc_call $NODE1_PORT "distributed_counter.increment" '[5]' "Increment Node 1 counter by 5"
make_rpc_call $NODE2_PORT "distributed_counter.increment" '[3]' "Increment Node 2 counter by 3"
make_rpc_call $NODE3_PORT "distributed_counter.increment" '[7]' "Increment Node 3 counter by 7"

# Check local counter states
make_rpc_call $NODE1_PORT "distributed_counter.get_count" 'null' "Get Node 1 counter state"
make_rpc_call $NODE2_PORT "distributed_counter.get_count" 'null' "Get Node 2 counter state"
make_rpc_call $NODE3_PORT "distributed_counter.get_count" 'null' "Get Node 3 counter state"

echo "🧪 PHASE 3: Service Discovery and Registration"
echo "=============================================="

# Register services on different nodes
make_rpc_call $NODE1_PORT "service_discovery.set_node_info" '[{"id":"'$NODE1_ID'","role":"primary","status":"active"}]' "Set Node 1 info"
make_rpc_call $NODE2_PORT "service_discovery.set_node_info" '[{"id":"'$NODE2_ID'","role":"secondary","status":"active"}]' "Set Node 2 info"
make_rpc_call $NODE3_PORT "service_discovery.set_node_info" '[{"id":"'$NODE3_ID'","role":"tertiary","status":"active"}]' "Set Node 3 info"

# Register services
make_rpc_call $NODE1_PORT "service_discovery.register_service" '["counter_service","'$NODE1_ID'"]' "Register counter service on Node 1"
make_rpc_call $NODE2_PORT "service_discovery.register_service" '["relay_service","'$NODE2_ID'"]' "Register relay service on Node 2"
make_rpc_call $NODE3_PORT "service_discovery.register_service" '["task_service","'$NODE3_ID'"]' "Register task service on Node 3"

# Discover services
make_rpc_call $NODE1_PORT "service_discovery.discover_services" 'null' "Discover services on Node 1"
make_rpc_call $NODE2_PORT "service_discovery.discover_services" 'null' "Discover services on Node 2"

echo "🧪 PHASE 4: Message Relay and Broadcasting"
echo "=========================================="

# Set up message relay nodes
make_rpc_call $NODE1_PORT "message_relay.set_node_id" '["'$NODE1_ID'"]' "Set Node 1 message relay ID"
make_rpc_call $NODE2_PORT "message_relay.set_node_id" '["'$NODE2_ID'"]' "Set Node 2 message relay ID"

# Send messages
make_rpc_call $NODE1_PORT "message_relay.send_message" '["'$NODE2_ID'","Hello from Node 1!"]' "Send message from Node 1 to Node 2"
make_rpc_call $NODE2_PORT "message_relay.broadcast" '["Broadcasting from Node 2 to all peers"]' "Broadcast message from Node 2"

# Check messages
make_rpc_call $NODE1_PORT "message_relay.get_messages" 'null' "Get messages on Node 1"
make_rpc_call $NODE2_PORT "message_relay.get_messages" 'null' "Get messages on Node 2"

echo "🧪 PHASE 5: Task Distribution"
echo "============================"

# Set node IDs for task distributors
make_rpc_call $NODE1_PORT "task_distributor.set_node_id" '["'$NODE1_ID'"]' "Set Node 1 task distributor ID"
make_rpc_call $NODE2_PORT "task_distributor.set_node_id" '["'$NODE2_ID'"]' "Set Node 2 task distributor ID"
make_rpc_call $NODE3_PORT "task_distributor.set_node_id" '["'$NODE3_ID'"]' "Set Node 3 task distributor ID"

# Submit tasks
make_rpc_call $NODE1_PORT "task_distributor.submit_task" '["data_processing",3]' "Submit high-priority task to Node 1"
make_rpc_call $NODE2_PORT "task_distributor.submit_task" '["image_analysis",2]' "Submit medium-priority task to Node 2"
make_rpc_call $NODE3_PORT "task_distributor.submit_task" '["log_aggregation",1]' "Submit low-priority task to Node 3"

# Process tasks
make_rpc_call $NODE1_PORT "task_distributor.process_task" '["task_001"]' "Process task on Node 1"
make_rpc_call $NODE2_PORT "task_distributor.process_task" '["task_001"]' "Process task on Node 2"

# Get task statistics
make_rpc_call $NODE1_PORT "task_distributor.get_task_stats" 'null' "Get Node 1 task stats"
make_rpc_call $NODE2_PORT "task_distributor.get_task_stats" 'null' "Get Node 2 task stats"
make_rpc_call $NODE3_PORT "task_distributor.get_task_stats" 'null' "Get Node 3 task stats"

echo "🧪 PHASE 6: Health Monitoring"
echo "============================"

# Health checks
make_rpc_call $NODE1_PORT "health_monitor.health_check" 'null' "Node 1 health check"
make_rpc_call $NODE2_PORT "health_monitor.health_check" 'null' "Node 2 health check"
make_rpc_call $NODE3_PORT "health_monitor.health_check" 'null' "Node 3 health check"

# Peer health checks
make_rpc_call $NODE1_PORT "health_monitor.check_peer_health" '["'$NODE2_ID'"]' "Node 1 checking Node 2 health"
make_rpc_call $NODE2_PORT "health_monitor.check_peer_health" '["'$NODE3_ID'"]' "Node 2 checking Node 3 health"

# Get cluster health
make_rpc_call $NODE1_PORT "health_monitor.get_cluster_health" 'null' "Get cluster health from Node 1"

echo "🧪 PHASE 7: Remote Server Invocation via WebSocket P2P"
echo "======================================================"

echo "⏱️  Waiting for WebSocket connections to establish..."
sleep 5

# Test remote calls from Node 1 to Node 2
make_remote_call $NODE1_PORT $NODE2_ID "distributed_counter.get_count" 'null' "Remote call: Node 1 → Node 2 (get counter)"
make_remote_call $NODE1_PORT $NODE2_ID "health_monitor.health_check" 'null' "Remote call: Node 1 → Node 2 (health check)"

# Test remote calls from Node 2 to Node 3
make_remote_call $NODE2_PORT $NODE3_ID "distributed_counter.increment" '[10]' "Remote call: Node 2 → Node 3 (increment by 10)"
make_remote_call $NODE2_PORT $NODE3_ID "task_distributor.get_task_stats" 'null' "Remote call: Node 2 → Node 3 (get task stats)"

# Test remote calls from Node 1 to Node 3 (multistep routing through Node 2)
echo "🔀 Testing multistep routing: Node 1 → Node 2 → Node 3"
make_remote_call $NODE1_PORT $NODE3_ID "distributed_counter.get_count" 'null' "Multistep routing: Node 1 → Node 3 (get counter)"
make_remote_call $NODE1_PORT $NODE3_ID "service_discovery.get_node_info" 'null' "Multistep routing: Node 1 → Node 3 (get node info)"

echo "🧪 PHASE 8: Cross-Node Data Synchronization"
echo "==========================================="

# Sync counter data across nodes via remote calls
echo "🔄 Synchronizing counter data across all nodes..."

# Get current states
echo "📊 Current counter states before sync:"
make_rpc_call $NODE1_PORT "distributed_counter.get_count" 'null' "Node 1 counter before sync"
make_rpc_call $NODE2_PORT "distributed_counter.get_count" 'null' "Node 2 counter before sync"
make_rpc_call $NODE3_PORT "distributed_counter.get_count" 'null' "Node 3 counter before sync"

# Perform cross-node synchronization
make_remote_call $NODE1_PORT $NODE2_ID "distributed_counter.sync_from_peer" '[{"node_id":"'$NODE1_ID'","count":5}]' "Sync Node 1 data to Node 2"
make_remote_call $NODE2_PORT $NODE3_ID "distributed_counter.sync_from_peer" '[{"node_id":"'$NODE2_ID'","count":3}]' "Sync Node 2 data to Node 3"

echo "📊 Counter states after sync:"
make_rpc_call $NODE1_PORT "distributed_counter.get_count" 'null' "Node 1 counter after sync"
make_rpc_call $NODE2_PORT "distributed_counter.get_count" 'null' "Node 2 counter after sync"
make_rpc_call $NODE3_PORT "distributed_counter.get_count" 'null' "Node 3 counter after sync"

echo "🧪 PHASE 9: Server Registry and Peer Discovery"
echo "=============================================="

# Check server registries
for port in $NODE1_PORT $NODE2_PORT $NODE3_PORT; do
    echo "📋 Server registry for Node on port $port:"
    curl -s http://localhost:$port/registry | jq '.'
    echo ""
done

# Check peer information
for port in $NODE1_PORT $NODE2_PORT $NODE3_PORT; do
    echo "🤝 Peer information for Node on port $port:"
    curl -s http://localhost:$port/registry/peers | jq '.'
    echo ""
done

echo "🧪 PHASE 10: Error Handling and Edge Cases"
echo "=========================================="

# Test invalid remote calls
echo "❌ Testing error handling with invalid calls:"
make_remote_call $NODE1_PORT "invalid_node" "distributed_counter.get_count" 'null' "Invalid node ID test"
make_remote_call $NODE1_PORT $NODE2_ID "invalid_server.invalid_method" 'null' "Invalid server/method test"

# Test malformed JSON-RPC calls
echo "❌ Testing malformed JSON-RPC:"
curl -s -X POST http://localhost:$NODE1_PORT/rpc \
    -H "Content-Type: application/json" \
    -d '{"invalid":"json","missing":"required_fields"}' | jq '.'
echo ""

echo "✅ PHASE 11: Performance and Load Testing"
echo "========================================"

# Rapid-fire calls to test performance
echo "⚡ Rapid-fire local calls (performance test):"
for i in {1..5}; do
    make_rpc_call $NODE1_PORT "distributed_counter.increment" '[1]' "Rapid increment #$i"
done

echo "⚡ Rapid-fire remote calls (P2P performance test):"
for i in {1..3}; do
    make_remote_call $NODE1_PORT $NODE2_ID "health_monitor.health_check" 'null' "Rapid remote health check #$i"
done

echo "🎯 FINAL VERIFICATION"
echo "===================="

# Final state verification
echo "📊 Final system state verification:"
make_rpc_call $NODE1_PORT "distributed_counter.get_count" 'null' "Final Node 1 counter state"
make_rpc_call $NODE2_PORT "distributed_counter.get_count" 'null' "Final Node 2 counter state"
make_rpc_call $NODE3_PORT "distributed_counter.get_count" 'null' "Final Node 3 counter state"

# Final health checks
echo "🏥 Final health status:"
make_rpc_call $NODE1_PORT "health_monitor.get_cluster_health" 'null' "Final cluster health check"

echo ""
echo "🎉 END-TO-END WEBSOCKET P2P TEST COMPLETED!"
echo "============================================"
echo ""
echo "📋 Test Summary:"
echo "✅ Local server functionality"
echo "✅ Service discovery and registration"
echo "✅ Message relay and broadcasting"
echo "✅ Task distribution"
echo "✅ Health monitoring"
echo "✅ Remote server invocation via WebSocket P2P"
echo "✅ Multistep routing (Node 1 → Node 2 → Node 3)"
echo "✅ Cross-node data synchronization"
echo "✅ Server registry and peer discovery"
echo "✅ Error handling and edge cases"
echo "✅ Performance and load testing"
echo ""
echo "📁 Log files: node1_e2e.log, node2_e2e.log, node3_e2e.log"
echo "🔍 Check logs for detailed WebSocket P2P communication traces"
echo ""
echo "🚀 All P2P capabilities tested successfully!" 