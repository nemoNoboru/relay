// Node B servers
server counter_b {
    state {
        count: number = 100,
        node_name: string = "node_b"
    }
    
    receive fn increment() -> number {
        set new_count = state.get("count") + 1
        state.set("count", new_count)
        new_count
    }
    
    receive fn decrement() -> number {
        set new_count = state.get("count") - 1
        state.set("count", new_count)
        new_count
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
    
    receive fn get_info() -> string {
        "Counter B on " + state.get("node_name") + " with count: " + string(state.get("count"))
    }
}

server echo_b {
    state {
        message_count: number = 0
    }
    
    receive fn echo(msg: string) -> string {
        set count = state.get("message_count") + 1
        state.set("message_count", count)
        "Node B Echo [" + string(count) + "]: " + msg
    }
    
    receive fn reverse_echo(msg: string) -> string {
        set count = state.get("message_count") + 1
        state.set("message_count", count)
        // Simple reverse (for demo purposes)
        "Node B Reverse [" + string(count) + "]: " + msg + " (reversed)"
    }
    
    receive fn get_message_count() -> number {
        state.get("message_count")
    }
}

server data_store_b {
    state {
        data_count: number = 0
    }
    
    receive fn store(key: string, value: string) -> string {
        set count = state.get("data_count") + 1
        state.set("data_count", count)
        "Stored: " + key + " = " + value + " (item " + string(count) + ")"
    }
    
    receive fn get_count() -> number {
        state.get("data_count")
    }
    
    receive fn info() -> string {
        "Data store B with " + string(state.get("data_count")) + " items"
    }
} 