// Simple test server for HTTP API demonstration
server test_server {
    state {
        counter: number = 0,
        message: string = "Hello from Relay!"
    }
    
    receive fn hello() -> string {
        "Hello, World!"
    }
    
    receive fn get_counter() -> number {
        state.get("counter")
    }
    
    receive fn increment() -> number {
        set current = state.get("counter")
        set new_value = current + 1
        state.set("counter", new_value)
        new_value
    }
    
    receive fn echo(msg: string) -> string {
        msg
    }
    
    receive fn add(a: number, b: number) -> number {
        a + b
    }
}

print("Test server loaded! Available methods:")
print("- test_server.hello()")
print("- test_server.get_counter()")
print("- test_server.increment()") 
print("- test_server.echo(message)")
print("- test_server.add(a, b)") 