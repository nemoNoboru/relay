package parser

import (
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestStructDefinition(t *testing.T) {
	src := `struct User {
		username: string,
		email: string,
		created_at: datetime
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestProtocolDefinition(t *testing.T) {
	src := `protocol UserService {
		get_user(id: string) -> User
		create_user(user: User) -> User
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestServerDefinition(t *testing.T) {
	src := `server user_service {
		state {
			users: [User] = [],
			count: number = 0
		}
		
		receive fn get_user(id: string) -> User {
			state.get("users").find(fn (u) { u.get("id") == id })
		}
		
		receive fn create_user(user: User) -> User {
			state.set("users", state.get("users").add(user))
			user
		}
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestOptionalTypes(t *testing.T) {
	src := `struct User {
		name: string,
		bio: optional(string)
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestArrayTypes(t *testing.T) {
	src := `struct Post {
		tags: [string],
		comments: [Comment]
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestStructConstructorFunctionCall(t *testing.T) {
	src := `
	server blog_service {
		state {
			posts: [Post] = [],
			next_id: number = 1
		}
		
		receive fn create_post(title: string, content: string) -> Post {
			set post = Post{
				id: state.get(:next_id),
				title: title,
				content: content
			}
			post
		}
	}
	
	`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}
