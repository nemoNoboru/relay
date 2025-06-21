#!/bin/bash

# End-to-End Multi-Node Test Script
# Tests the unified architecture with real multi-node communication

set -e  # Exit on any error

echo "=== Relay Multi-Node End-to-End Test ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
NODE_A_PORT=8083
NODE_B_PORT=8084
NODE_A_ID="node_a_$(date +%s)"
NODE_B_ID="node_b_$(date +%s)"

# Build Relay
echo -e "${BLUE}Building Relay...${NC}"
cd .. && go build -o relay cmd/relay/main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to build Relay${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Relay built successfully${NC}"
echo

# Start Node A
echo -e "${BLUE}Starting Node A on port $NODE_A_PORT...${NC}"
./relay examples/e2e_multi_node_test.rl -server -port $NODE_A_PORT -node-id $NODE_A_ID &
NODE_A_PID=$!

# Start Node B  
echo -e "${BLUE}Starting Node B on port $NODE_B_PORT...${NC}"
./relay examples/e2e_multi_node_test_b.rl -server -port $NODE_B_PORT -node-id $NODE_B_ID &
NODE_B_PID=$!

# Wait for servers to start
echo -e "${BLUE}Waiting for nodes to start...${NC}"
sleep 5

echo -e "${GREEN}Node A PID: $NODE_A_PID${NC}"
echo -e "${GREEN}Node B PID: $NODE_B_PID${NC}"
echo

# Function to make JSON-RPC call
make_rpc_call() {
    local port=$1
    local method=$2
    local params=$3
    local id=$4
    
    if [ -z "$params" ] || [ "$params" = "null" ]; then
        params="null"
    fi
    
    curl -s -X POST "http://localhost:$port/rpc" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"params\":$params,\"id\":$id}"
}

# Function to check if response contains expected result
check_result() {
    local response=$1
    local expected=$2
    local test_name=$3
    
    if echo "$response" | grep -q "\"result\":$expected"; then
        echo -e "${GREEN}✅ $test_name - PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name - FAILED${NC}"
        echo -e "${YELLOW}Expected: $expected${NC}"
        echo -e "${YELLOW}Response: $response${NC}"
        return 1
    fi
}

# Function to check if response contains error
check_error() {
    local response=$1
    local test_name=$2
    
    if echo "$response" | grep -q "\"error\""; then
        echo -e "${GREEN}✅ $test_name - PASSED (expected error)${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name - FAILED (expected error but got success)${NC}"
        echo -e "${YELLOW}Response: $response${NC}"
        return 1
    fi
}

# Test counter
TESTS_PASSED=0
TESTS_TOTAL=0

run_test() {
    local test_name=$1
    local test_function=$2
    
    echo -e "${BLUE}🔍 Test: $test_name${NC}"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    if $test_function; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
    echo
}

# === BASIC CONNECTIVITY TESTS ===

test_node_a_health() {
    local response=$(curl -s "http://localhost:$NODE_A_PORT/health")
    if echo "$response" | grep -q "healthy"; then
        echo -e "${GREEN}✅ Node A health check - PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ Node A health check - FAILED${NC}"
        return 1
    fi
}

test_node_b_health() {
    local response=$(curl -s "http://localhost:$NODE_B_PORT/health")
    if echo "$response" | grep -q "healthy"; then
        echo -e "${GREEN}✅ Node B health check - PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ Node B health check - FAILED${NC}"
        return 1
    fi
}

test_node_a_counter_increment() {
    local response=$(make_rpc_call $NODE_A_PORT "counter_a.increment" "null" 1)
    check_result "$response" "1" "Node A counter increment"
}

test_node_a_echo() {
    local response=$(make_rpc_call $NODE_A_PORT "echo_a.echo" "[\"Hello from test\"]" 3)
    if echo "$response" | grep -q "Node A Echo.*Hello from test"; then
        echo -e "${GREEN}✅ Node A echo - PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ Node A echo - FAILED${NC}"
        return 1
    fi
}

test_node_b_counter_increment() {
    local response=$(make_rpc_call $NODE_B_PORT "counter_b.increment" "null" 5)
    check_result "$response" "101" "Node B counter increment"
}

test_nonexistent_server() {
    local response=$(make_rpc_call $NODE_A_PORT "nonexistent_server.test" "null" 20)
    check_error "$response" "Call to non-existent server"
}

# === RUN ALL TESTS ===

echo -e "${YELLOW}=== BASIC CONNECTIVITY TESTS ===${NC}"
run_test "Node A Health Check" test_node_a_health
run_test "Node B Health Check" test_node_b_health

echo -e "${YELLOW}=== LOCAL SERVER TESTS ===${NC}"
run_test "Node A Counter Increment" test_node_a_counter_increment
run_test "Node A Echo" test_node_a_echo
run_test "Node B Counter Increment" test_node_b_counter_increment

echo -e "${YELLOW}=== ERROR HANDLING TESTS ===${NC}"
run_test "Non-existent Server" test_nonexistent_server

# === CLEANUP ===

echo -e "${BLUE}🧹 Cleaning up...${NC}"
kill $NODE_A_PID $NODE_B_PID 2>/dev/null || true
wait $NODE_A_PID $NODE_B_PID 2>/dev/null || true
echo -e "${GREEN}✅ Servers stopped${NC}"

# === RESULTS ===

echo
echo -e "${YELLOW}=== TEST RESULTS ===${NC}"
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${BLUE}Tests Total:  $TESTS_TOTAL${NC}"

if [ $TESTS_PASSED -eq $TESTS_TOTAL ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
    exit 0
else
    TESTS_FAILED=$((TESTS_TOTAL - TESTS_PASSED))
    echo -e "${RED}❌ $TESTS_FAILED tests failed${NC}"
    exit 1
fi 