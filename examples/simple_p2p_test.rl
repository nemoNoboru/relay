// Simple WebSocket P2P Test
// This demonstrates basic server functionality that can be called remotely

// Simple counter server
server counter_server {
    state {
        count: number = 0
    }
    
    receive fn increment(amount: number) -> number {
        set new_amount = if amount == nil { 1 } else { amount }
        state.set("count", state.get("count") + new_amount)
        state.get("count")
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
    
    receive fn reset() -> string {
        state.set("count", 0)
        "Counter reset to 0"
    }
}

// Simple echo server
server echo_server {
    state {
        message_count: number = 0
    }
    
    receive fn echo(message: string) -> string {
        state.set("message_count", state.get("message_count") + 1)
        "Echo: " + message
    }
    
    receive fn get_stats() -> object {
        {
            message_count: state.get("message_count"),
            status: "active"
        }
    }
    
    receive fn ping() -> string {
        "pong"
    }
}

print("WebSocket P2P test servers initialized")
print("Available servers: counter_server, echo_server")
print("Try: counter_server.increment, echo_server.echo") 