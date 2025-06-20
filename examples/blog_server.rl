// Simple blog server example for HTTP JSON-RPC API
server blog_server {
    state {
        posts: [object] = [],
        next_id: number = 1
    }
    
    receive fn get_posts() -> [object] {
        state.get("posts")
    }
    
    receive fn create_post(title: string, content: string) -> object {
        set id = state.get("next_id")
        set post = {
            id: id,
            title: title,
            content: content,
            created_at: "2024-01-01T00:00:00Z"
        }
        
        set posts = state.get("posts")
        set updated_posts = posts.push(post)
        state.set("posts", updated_posts)
        state.set("next_id", id + 1)
        
        post
    }
    
    receive fn get_post(id: number) -> object {
        set posts = state.get("posts")
        set found_posts = posts.filter(fn(p) { p.get("id") == id })
        if found_posts.length() > 0 {
            found_posts.get(0)
        } else {
            nil
        }
    }
    
    receive fn update_post(id: number, title: string, content: string) -> object {
        set posts = state.get("posts")
        set updated_posts = posts.map(fn(p) {
            if p.get("id") == id {
                {
                    id: id,
                    title: title,
                    content: content,
                    created_at: p.get("created_at")
                }
            } else {
                p
            }
        })
        state.set("posts", updated_posts)
        
        get_post(id)
    }
    
    receive fn delete_post(id: number) -> bool {
        set posts = state.get("posts")
        set filtered_posts = posts.filter(fn(p) { p.get("id") != id })
        state.set("posts", filtered_posts)
        true
    }
    
    receive fn get_stats() -> object {
        {
            total_posts: state.get("posts").length(),
            next_id: state.get("next_id")
        }
    }
}

print("Blog server initialized!")
print("Server methods available:")
print("- blog_server.get_posts()")
print("- blog_server.create_post(title, content)")
print("- blog_server.get_post(id)")
print("- blog_server.update_post(id, title, content)")
print("- blog_server.delete_post(id)")
print("- blog_server.get_stats()") 