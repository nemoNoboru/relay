server test_server {
    state {
        count: number = 0,
        message: string = "Hello from unified architecture!"
    }
    
    receive fn hello() -> string {
        state.get("message")
    }
    
    receive fn increment() -> number {
        set new_count = state.get("count") + 1
        state.set("count", new_count)
        new_count
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
}

server echo_server {
    receive fn echo(msg: string) -> string {
        msg
    }
    
    receive fn ping() -> string {
        "pong"
    }
} 