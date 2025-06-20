// Simple End-to-End WebSocket P2P Test
server distributed_counter {
    state {
        local_count: number = 0,
        node_id: string = "unknown"
    }
    
    receive fn increment(amount: number) -> object {
        set new_amount = if amount == nil { 1 } else { amount }
        state.set("local_count", state.get("local_count") + new_amount)
        
        {
            node_id: state.get("node_id"),
            local_count: state.get("local_count"),
            increment: new_amount
        }
    }
    
    receive fn get_count() -> object {
        {
            node_id: state.get("node_id"),
            local_count: state.get("local_count")
        }
    }
    
    receive fn set_node_id(id: string) -> string {
        state.set("node_id", id)
        "Node ID set to: " + id
    }
}

server health_monitor {
    state {
        check_count: number = 0
    }
    
    receive fn health_check() -> object {
        state.set("check_count", state.get("check_count") + 1)
        
        {
            status: "healthy",
            uptime: state.get("check_count"),
            timestamp: "2025-01-01T00:00:00Z"
        }
    }
}

print("Simple E2E P2P Test Servers Initialized Successfully!")
print("Available servers: distributed_counter, health_monitor") 