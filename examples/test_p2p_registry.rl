// Test file for peer-to-peer server registry functionality

server test_counter {
    state {
        count: number = 0,
        name: string = "P2P Test Counter"
    }
    
    receive fn increment() -> number {
        state.set("count", state.get("count") + 1)
        state.get("count")
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
    
    receive fn get_info() -> object {
        {
            name: state.get("name"),
            count: state.get("count"),
            message: "This server can be discovered by peers!"
        }
    }
    
    receive fn reset() -> number {
        state.set("count", 0)
        0
    }
}

server discovery_test {
    state {
        peers_discovered: number = 0,
        last_discovery: string = "none"
    }
    
    receive fn peer_discovered(peer_info: string) -> string {
        state.set("peers_discovered", state.get("peers_discovered") + 1)
        state.set("last_discovery", peer_info)
        "Peer registered: " + peer_info
    }
    
    receive fn get_discovery_stats() -> object {
        {
            peers_discovered: state.get("peers_discovered"),
            last_discovery: state.get("last_discovery")
        }
    }
}

print("P2P test servers created successfully!")
print("test_counter server: increment, get_count, get_info, reset")
print("discovery_test server: peer_discovered, get_discovery_stats")
print("")
print("Start the server with: ./relay test_p2p_registry.rl -server")
print("View registry at: http://localhost:8080/registry")
print("Test JSON-RPC at: http://localhost:8080/rpc") 