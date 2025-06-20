// WebSocket P2P and Remote Server Invocation Demo
// This example demonstrates real-time peer-to-peer communication between Relay nodes

// Define a distributed counter server
counter_server = server {
    state = {count: 0, node_peers: []}
    
    // Increment counter locally
    increment = method(amount) {
        state.count = state.count + amount
        return state.count
    }
    
    // Get current count
    get_count = method() {
        return state.count
    }
    
    // Sync counter with peer nodes
    sync_with_peers = method(peer_nodes) {
        state.node_peers = peer_nodes
        total_count = state.count
        
        // This would call remote servers in a real implementation
        // For now, just return local count
        return {
            local_count: state.count,
            total_count: total_count,
            peers: state.node_peers
        }
    }
    
    // Reset counter
    reset = method() {
        state.count = 0
        return "Counter reset"
    }
}

// Define a discovery service for peer management
discovery_server = server {
    state = {known_peers: [], last_discovery: 0}
    
    // Register a new peer
    register_peer = method(node_id, address) {
        peer = {node_id: node_id, address: address, registered_at: time()}
        state.known_peers = append(state.known_peers, peer)
        return "Peer registered: " + node_id
    }
    
    // Get list of known peers
    get_peers = method() {
        return state.known_peers
    }
    
    // Announce this node to peers
    announce = method(my_node_id) {
        return {
            message: "Node announcement",
            node_id: my_node_id,
            timestamp: time(),
            peer_count: len(state.known_peers)
        }
    }
    
    // Health check for peers
    health_check = method() {
        return {
            status: "healthy",
            peer_count: len(state.known_peers),
            last_discovery: state.last_discovery
        }
    }
}

// Define a message relay server for P2P communication
relay_server = server {
    state = {message_history: [], connected_peers: []}
    
    // Send message to specific peer
    send_message = method(target_node, message_type, payload) {
        message = {
            from: "local",
            to: target_node,
            type: message_type,
            payload: payload,
            timestamp: time(),
            id: generate_id()
        }
        
        state.message_history = append(state.message_history, message)
        
        return {
            status: "message_queued",
            message_id: message.id,
            target: target_node
        }
    }
    
    // Broadcast message to all peers
    broadcast = method(message_type, payload) {
        message = {
            from: "local",
            type: message_type,
            payload: payload,
            timestamp: time(),
            id: generate_id()
        }
        
        state.message_history = append(state.message_history, message)
        
        return {
            status: "broadcast_queued",
            message_id: message.id,
            peer_count: len(state.connected_peers)
        }
    }
    
    // Get message history
    get_history = method(limit) {
        if limit == nil {
            limit = 10
        }
        
        history_len = len(state.message_history)
        start_idx = max(0, history_len - limit)
        
        return slice(state.message_history, start_idx, history_len)
    }
    
    // Handle incoming message
    handle_message = method(message) {
        state.message_history = append(state.message_history, message)
        
        return {
            status: "message_received",
            from: message.from,
            type: message.type
        }
    }
}

// Define a distributed task server
task_server = server {
    state = {tasks: [], completed_tasks: [], node_load: 0}
    
    // Add a new task
    add_task = method(task_name, task_data, priority) {
        if priority == nil {
            priority = 1
        }
        
        task = {
            id: generate_id(),
            name: task_name,
            data: task_data,
            priority: priority,
            status: "pending",
            created_at: time(),
            assigned_node: nil
        }
        
        state.tasks = append(state.tasks, task)
        return task
    }
    
    // Get pending tasks
    get_pending_tasks = method() {
        pending = []
        for task in state.tasks {
            if task.status == "pending" {
                pending = append(pending, task)
            }
        }
        return pending
    }
    
    // Assign task to node
    assign_task = method(task_id, node_id) {
        for i, task in state.tasks {
            if task.id == task_id {
                task.status = "assigned"
                task.assigned_node = node_id
                task.assigned_at = time()
                state.tasks[i] = task
                state.node_load = state.node_load + 1
                return task
            }
        }
        return nil
    }
    
    // Complete a task
    complete_task = method(task_id, result) {
        for i, task in state.tasks {
            if task.id == task_id {
                task.status = "completed"
                task.result = result
                task.completed_at = time()
                state.tasks[i] = task
                state.completed_tasks = append(state.completed_tasks, task)
                state.node_load = max(0, state.node_load - 1)
                return task
            }
        }
        return nil
    }
    
    // Get node statistics
    get_stats = method() {
        pending_count = 0
        assigned_count = 0
        completed_count = len(state.completed_tasks)
        
        for task in state.tasks {
            if task.status == "pending" {
                pending_count = pending_count + 1
            } else if task.status == "assigned" {
                assigned_count = assigned_count + 1
            }
        }
        
        return {
            pending: pending_count,
            assigned: assigned_count,
            completed: completed_count,
            node_load: state.node_load,
            total_tasks: len(state.tasks)
        }
    }
}

// Utility functions
generate_id = func() {
    return "id_" + string(time()) + "_" + string(random(1000, 9999))
}

time = func() {
    // This would return current timestamp in a real implementation
    return 1234567890
}

random = func(min, max) {
    // This would return a random number in a real implementation
    return min + ((max - min) / 2)
}

max = func(a, b) {
    if a > b {
        return a
    }
    return b
}

min = func(a, b) {
    if a < b {
        return a
    }
    return b
}

slice = func(array, start, end) {
    // This would slice the array in a real implementation
    return array
}

append = func(array, item) {
    // This would append to array in a real implementation
    return array
}

len = func(array) {
    // This would return array length in a real implementation
    return 0
}

// Initialize servers and show startup message
startup_message = {
    message: "WebSocket P2P Demo initialized",
    servers: ["counter_server", "discovery_server", "relay_server", "task_server"],
    features: [
        "Real-time peer-to-peer communication",
        "Distributed counter synchronization", 
        "Message relay and broadcasting",
        "Distributed task management",
        "Peer discovery and health monitoring"
    ],
    usage: {
        local_calls: "Use JSON-RPC: counter_server.increment, discovery_server.get_peers, etc.",
        remote_calls: "Use remote_call method with node_id, server_name, method, and args",
        websocket: "Connect peers via WebSocket at /ws/p2p endpoint"
    }
}

startup_message 