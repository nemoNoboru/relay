package parser

import (
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestCompleteProgram(t *testing.T) {
	src := `struct Post {
		id: string,
		title: string,
		content: string
	}
	
	protocol BlogService {
		get_posts() -> [Post]
		create_post(title: string, content: string) -> Post
	}
	
	server blog_services {
		state {
			posts: [Post] = [],
			next_id: number = 1
		}
		
		receive fn get_posts() -> [Post] {
			state.get("posts")
		}
		
		receive fn create_post(title: string, content: string) -> Post {
			set post = Post{
				id: state.get(:next_id),
				title: title,
				content: content
			}
			
			state.set("posts", state.get("posts").add(post))
			state.set("next_id", state.get("next_id") + 1)
			
			return post
		}
	}
	
	template "index.html" from get_posts()
	
	config {
		app_name: "simple_blog",
		port: 8080
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}
