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
			update state.users = state.get("users").add(user)
			user
		}
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestFunctionDefinition(t *testing.T) {
	src := `fn calculate_total(items: [object]) -> number {
		set total = 0
		for item in items {
			update total = total + item.get("price")
		}
		total
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestLambdaExpression(t *testing.T) {
	src := `set doubled = numbers.map(fn (x) { x * 2 })`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestFieldAccess(t *testing.T) {
	src := `set name = user.get("name")`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestMethodChaining(t *testing.T) {
	src := `set result = users.filter(fn (u) { u.get("active") })
		.map(fn (u) { u.get("name") })
		.sort_by(fn (u) { u })`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestSymbolLiterals(t *testing.T) {
	src := `dispatch action {
		:create: fn (data) { "created" },
		:update: fn (data) { "updated" },
		:delete: fn (data) { "deleted" }
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestUpdateStatement(t *testing.T) {
	src := `update state.count = state.get("count") + 1`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestBlockExpression(t *testing.T) {
	src := `set result = {
		set x = 10
		set y = 20
		x + y
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestIfExpression(t *testing.T) {
	src := `set message = if count > 0 { "items found" } else { "no items" }`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestSendExpression(t *testing.T) {
	src := `set posts = send "blog_service" get_posts {}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestTemplateDeclaration(t *testing.T) {
	src := `template "index.html" from get_posts()`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestTemplateWithParams(t *testing.T) {
	src := `template "post.html" from get_post(id: string)`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestConfigDefinition(t *testing.T) {
	src := `config {
		app_name: "myblog",
		port: 8080,
		federation: {
			auto_discover: true,
			cache_duration: "1h"
		}
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestControlFlow(t *testing.T) {
	src := `for user in users {
		set processed = user.get("name")
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestErrorHandling(t *testing.T) {
	src := `try {
		set result = risky_operation()
	} catch error {
		set result = "failed"
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

func TestFunctionTypes(t *testing.T) {
	src := `fn apply_operation(op: fn(number) -> number, value: number) -> number {
		op(value)
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestNullCoalesceOperator(t *testing.T) {
	src := `set value = optional_value ?? "default"`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

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
	
	server blog_service {
		state {
			posts: [Post] = [],
			next_id: number = 1
		}
		
		receive fn get_posts() -> [Post] {
			state.get("posts")
		}
		
		receive fn create_post(title: string, content: string) -> Post {
			set post = Post{
				id: state.get("next_id").toString(),
				title: title,
				content: content
			}
			
			update state.posts = state.get("posts").add(post)
			update state.next_id = state.get("next_id") + 1
			
			post
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
