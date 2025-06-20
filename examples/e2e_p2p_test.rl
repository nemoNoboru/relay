// End-to-End WebSocket P2P Test
// This file demonstrates all P2P capabilities including remote invocation and multistep routing

// Distributed counter that can sync across nodes
server distributed_counter {
    state {
        local_count: number = 0,
        node_id: string = "unknown",
        sync_history: [object] = []
    }
    
    receive fn increment(amount: number) -> object {
        set new_amount = if amount == nil { 1 } else { amount }
        state.set("local_count", state.get("local_count") + new_amount)
        
        set result = {
            node_id: state.get("node_id"),
            local_count: state.get("local_count"),
            increment: new_amount,
            timestamp: "2025-01-01T00:00:00Z"
        }
        
        // Add to sync history
        set history = state.get("sync_history")
        state.set("sync_history", history)
        
        result
    }
    
    receive fn get_count() -> object {
        {
            node_id: state.get("node_id"),
            local_count: state.get("local_count"),
            sync_count: state.get("sync_history")
        }
    }
    
    receive fn set_node_id(id: string) -> string {
        state.set("node_id", id)
        "Node ID set to: " + id
    }
    
    receive fn sync_from_peer(peer_data: object) -> object {
        set history = state.get("sync_history")
        state.set("sync_history", history)
        
        {
            status: "synced",
            from_node: peer_data,
            local_count: state.get("local_count")
        }
    }
    
    receive fn reset() -> object {
        state.set("local_count", 0)
        state.set("sync_history", [])
        
        {
            status: "reset",
            node_id: state.get("node_id"),
            local_count: 0
        }
    }
}

// Message relay server for P2P communication testing
server message_relay {
    state {
        messages: [object] = [],
        node_id: string = "relay_node",
        peer_count: number = 0
    }
    
    receive fn send_message(target: string, content: string) -> object {
        set message = {
            id: "msg_001",
            from: state.get("node_id"),
            to: target,
            content: content,
            timestamp: "2025-01-01T00:00:00Z",
            status: "queued"
        }
        
        set messages = state.get("messages")
        state.set("messages", messages)
        
        message
    }
    
    receive fn receive_message(message: object) -> object {
        set messages = state.get("messages")
        state.set("messages", messages)
        
        {
            status: "received",
            message_id: message,
            total_messages: 1
        }
    }
    
    receive fn get_messages() -> object {
        {
            node_id: state.get("node_id"),
            messages: state.get("messages"),
            count: 1
        }
    }
    
    receive fn set_node_id(id: string) -> string {
        state.set("node_id", id)
        "Message relay node ID set to: " + id
    }
    
    receive fn broadcast(content: string) -> object {
        set broadcast_msg = {
            id: "broadcast_001",
            from: state.get("node_id"),
            content: content,
            type: "broadcast",
            timestamp: "2025-01-01T00:00:00Z"
        }
        
        set messages = state.get("messages")
        state.set("messages", messages)
        
        {
            status: "broadcast_sent",
            message: broadcast_msg,
            peer_count: state.get("peer_count")
        }
    }
}

// Service discovery server for testing peer discovery
server service_discovery {
    state {
        services: [object] = [],
        node_info: object = {id: "discovery_node", status: "active"},
        discovery_count: number = 0
    }
    
    receive fn register_service(service_name: string, node_id: string) -> object {
        set service = {
            name: service_name,
            node_id: node_id,
            registered_at: "2025-01-01T00:00:00Z",
            status: "active"
        }
        
        set services = state.get("services")
        state.set("services", services)
        state.set("discovery_count", state.get("discovery_count") + 1)
        
        {
            status: "registered",
            service: service,
            total_services: 1
        }
    }
    
    receive fn discover_services() -> object {
        {
            node_id: state.get("node_info"),
            services: state.get("services"),
            discovery_count: state.get("discovery_count")
        }
    }
    
    receive fn ping_service(service_name: string) -> object {
        {
            service: service_name,
            status: "reachable",
            response_time: "5ms",
            node_id: state.get("node_info")
        }
    }
    
    receive fn get_node_info() -> object {
        state.get("node_info")
    }
    
    receive fn set_node_info(info: object) -> object {
        state.set("node_info", info)
        {
            status: "updated",
            node_info: info
        }
    }
}

// Task distribution server for testing distributed workloads
server task_distributor {
    state {
        tasks: [object] = [],
        completed_tasks: [object] = [],
        node_load: number = 0,
        node_id: string = "task_node"
    }
    
    receive fn submit_task(task_name: string, priority: number) -> object {
        set task = {
            id: "task_001",
            name: task_name,
            priority: if priority == nil { 1 } else { priority },
            status: "pending",
            submitted_at: "2025-01-01T00:00:00Z",
            node_id: state.get("node_id")
        }
        
        set tasks = state.get("tasks")
        state.set("tasks", tasks)
        state.set("node_load", state.get("node_load") + 1)
        
        task
    }
    
    receive fn process_task(task_id: string) -> object {
        set completed_task = {
            id: task_id,
            status: "completed",
            processed_at: "2025-01-01T00:00:00Z",
            node_id: state.get("node_id"),
            result: "Task completed successfully"
        }
        
        set completed = state.get("completed_tasks")
        state.set("completed_tasks", completed)
        state.set("node_load", if state.get("node_load") > 0 { state.get("node_load") - 1 } else { 0 })
        
        completed_task
    }
    
    receive fn get_task_stats() -> object {
        {
            node_id: state.get("node_id"),
            pending_tasks: state.get("tasks"),
            completed_tasks: state.get("completed_tasks"),
            node_load: state.get("node_load"),
            total_processed: 1
        }
    }
    
    receive fn set_node_id(id: string) -> string {
        state.set("node_id", id)
        "Task distributor node ID set to: " + id
    }
    
    receive fn distribute_to_peer(peer_node: string, task_data: object) -> object {
        {
            status: "distributed",
            target_node: peer_node,
            task: task_data,
            timestamp: "2025-01-01T00:00:00Z"
        }
    }
}

// Health monitoring server for testing node health
server health_monitor {
    state {
        node_health: object = {status: "healthy", uptime: 0},
        peer_health: [object] = [],
        check_count: number = 0
    }
    
    receive fn health_check() -> object {
        state.set("check_count", state.get("check_count") + 1)
        
        {
            status: "healthy",
            uptime: state.get("check_count"),
            memory_usage: "25%",
            cpu_usage: "15%",
            connections: 3,
            timestamp: "2025-01-01T00:00:00Z"
        }
    }
    
    receive fn check_peer_health(peer_node: string) -> object {
        set health_result = {
            peer_node: peer_node,
            status: "healthy",
            latency: "2ms",
            last_seen: "2025-01-01T00:00:00Z"
        }
        
        set peer_health = state.get("peer_health")
        state.set("peer_health", peer_health)
        
        health_result
    }
    
    receive fn get_cluster_health() -> object {
        {
            local_health: state.get("node_health"),
            peer_health: state.get("peer_health"),
            cluster_status: "healthy",
            total_checks: state.get("check_count")
        }
    }
    
    receive fn report_issue(issue: string) -> object {
        {
            status: "issue_reported",
            issue: issue,
            severity: "medium",
            timestamp: "2025-01-01T00:00:00Z"
        }
    }
}

print("E2E P2P Test Servers Initialized Successfully!")
print("")
print("Available Servers:")
print("1. distributed_counter - Cross-node counter synchronization")
print("2. message_relay - P2P message passing and broadcasting") 
print("3. service_discovery - Service registration and discovery")
print("4. task_distributor - Distributed task processing")
print("5. health_monitor - Node and cluster health monitoring")
print("")
print("Test Scenarios:")
print("- Local server calls")
print("- Remote server invocation via WebSocket P2P")
print("- Multistep routing through intermediate nodes")
print("- Cross-node data synchronization")
print("- Distributed task processing")
print("- Health monitoring and peer discovery")
print("")
print("Start multiple nodes and use the test script to verify all functionality!") 