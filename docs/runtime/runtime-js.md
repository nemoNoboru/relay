# JavaScript Runtime

the javascript runtime makes relay work in browsers. you write relay, it creates web pages. simple as that.

## how it works

relay says "show a button" → runtime creates `<button>` in the DOM
relay says "fetch some data" → runtime makes HTTP request
user clicks button → runtime tells relay what happened

the user writes this:
```relay
state counter 0

show card "Counter App"
  show text "Count: " (get counter)
  show button "Increment"
    when click
      set counter (add counter 1)
```

the runtime turns it into a working web page with a card, some text, and a clickable button. no build process, no webpack, no configuration files.

## what the runtime handles

**UI components** - turns relay's `show button` into actual DOM elements with proper styling (using tailwind)

**events** - when user clicks, types, submits forms, the runtime captures these and tells relay

**networking** - handles `fetch` requests, websockets, all the async web API stuff

**state management** - when relay updates state, runtime updates the page automatically

**local data** - localStorage, IndexedDB, anything the browser can persist

## example: a simple chat app

```relay
state messages []
state new-message ""

show card "Chat"
  for message in messages
    show text (get message content)
  
  show input new-message "Type a message..."
  show button "Send"
    when click
      fetch "/api/messages"
        method POST
        body { content: new-message }
      
      add messages { content: new-message }
      set new-message ""
```

the runtime:
- renders a card with message list and input
- handles typing in the input field  
- makes POST request when send is clicked
- updates the message list automatically
- clears the input field

all the user had to do was describe what they wanted. the runtime figured out how to make it work.

## networking made simple

when relay code does:
```relay
fetch "https://api.example.com/users"
  set users (get response data)
```

the runtime:
1. makes the HTTP request using the browser's fetch API
2. waits for response
3. extracts the data
4. tells relay to update the `users` state
5. automatically re-renders any UI that uses `users`

error handling, CORS, timeouts - the runtime deals with all of that. if something goes wrong, it just shows a reasonable error message.

## fallbacks

not every browser supports everything. the runtime handles graceful degradation:

```relay
show date-picker
  fallback show input "Enter date (YYYY-MM-DD)"
```

modern browsers get a nice date picker widget. older browsers get a text input with helpful placeholder text. either way, the app works.

## why this approach works

**developers focus on the app, not the platform** - you describe what you want, runtime handles how to build it

**instant feedback** - change relay code, see results immediately in the browser

**web standards** - uses normal HTML, CSS, DOM events under the hood. inspect element works like you'd expect

**progressive enhancement** - works on old browsers, better on new ones

**no build step** - relay code runs directly, no compilation or bundling required

## what gets generated

the runtime creates clean, semantic HTML:
- relay cards become `<div class="card">`
- relay buttons become `<button>` with event listeners
- relay forms become actual `<form>` elements
- everything styled with tailwind classes

accessibility features work automatically. screen readers understand the generated HTML. keyboard navigation just works.

## implementation notes

the JS runtime is basically:
1. **relay parser** - converts relay syntax to function calls
2. **DOM renderer** - creates/updates HTML elements  
3. **event system** - captures user interactions
4. **state manager** - keeps UI in sync with data
5. **network layer** - handles fetch requests and websockets

but all of this is invisible to the developer. they just write relay and get a working web app.

the goal is to make building web apps feel like writing instructions to a helpful assistant, not fighting with frameworks and build tools.

## what's next

this runtime should be so good that people say "why would I use anything else for simple web apps?"

once that works, the patterns can be applied to other runtimes - CLI, mobile, server. but browser first, because that's where most people expect to see their apps running. 