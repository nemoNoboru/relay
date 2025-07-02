# Networking in relay

networking in relay is simple. you say what you want, it happens.

## basic HTTP requests

```relay
# get some data
fetch "https://api.example.com/users"
  set users (get response data)

# post some data  
fetch "https://api.example.com/users"
  method POST
  body { name: "Alice", email: "alice@example.com" }
  set new-user (get response data)
  show notification "User created!"

# handle errors
fetch "https://api.example.com/login"
  method POST
  body { username: username, password: password }
  set auth-token (get response token)
  navigate "/dashboard"

catch error
  show notification "Login failed: " (get error message)
```

that's it. no promises, no callbacks, no error handling boilerplate. just describe what you want to happen.

## real-time with websockets

```relay
# connect to a chat server
connect websocket "wss://chat.example.com"
  when message
    add messages (get data content)

# send messages
when send-button-clicked
  send websocket { type: "message", content: message-text }
  set message-text ""
```

websockets are just as simple. connect, listen for messages, send messages. the runtime handles connection management, reconnection, all the messy details.

## what the runtime does for you

**handles async automatically** - you write synchronous-looking code, runtime manages all the promises and callbacks

**deals with errors gracefully** - network timeouts, CORS issues, server errors all get handled reasonably

**manages connections** - websockets reconnect automatically, requests timeout appropriately

**formats data correctly** - JSON parsing, proper headers, content types all handled

**provides helpful defaults** - reasonable timeouts, retry logic, error messages

## authentication made simple

```relay
# set auth token once
set auth-token "your-jwt-token"

# all requests automatically include it
fetch "https://api.example.com/protected-data"
  set secure-data (get response data)
```

set your auth token once, every request includes it automatically. no need to remember to add Authorization headers everywhere.

## fallbacks for offline

```relay
fetch "https://api.example.com/data"
  set data (get response data)

catch network-error
  set data (get local-storage "cached-data")
  show notification "Using cached data (offline)"
```

when network fails, gracefully fall back to cached data. users get a working app even when connectivity is spotty.

## real example: a simple blog

```relay
state posts []
state new-post ""

# load posts when page loads
when page-load
  fetch "https://api.myblog.com/posts"
    set posts (get response data)

# show the posts
for post in posts
  show card
    show text (get post title)
    show text (get post content)

# add new posts
show input new-post "Write a post..."
show button "Publish"
  when click
    fetch "https://api.myblog.com/posts"
      method POST
      body { content: new-post }
    
    add posts { content: new-post }
    set new-post ""
    show notification "Posted!"
```

this creates a working blog interface. loads posts, displays them, lets you add new ones. all the HTTP handling is automatic.

## relay federation (future)

```relay
# connect to friend's relay instance
connect relay "relay://alice.local:8080"
  sync book-reviews  # share book reviews with alice

# when alice adds a review, you see it
when external-review-added review
  add book-reviews review
  show notification "Alice reviewed: " (get review book-title)
```

eventually relay instances will be able to connect to each other and share data. same simple syntax, but now your apps can be part of a decentralized network.

## why this works

**you focus on what, not how** - describe what data you need, not how to fetch it

**errors are handled sensibly** - network problems don't crash your app

**works offline** - graceful degradation when network is unavailable  

**no build tools** - no webpack configs for handling different environments

**scales naturally** - same syntax works for simple API calls and complex real-time apps

the goal is making networked apps feel as simple as local ones. the complexity of HTTP, websockets, error handling, etc. is hidden behind simple relay syntax.

write relay code that describes your app's behavior. the runtime makes it work on the real, messy internet. 