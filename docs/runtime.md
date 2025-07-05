# Relay runtime abstraction

relay should run anywhere. same code, different places. want your app in a browser? done. need it as a command line tool? done. mobile app? server? done.

the idea is simple: you write relay once, it runs everywhere.

## why runtime abstraction matters

imagine you build a local library catalog in relay. starts as a web app for browsing books. later you want:
- a CLI tool for librarians to add books quickly
- a mobile app for patrons
- a server component for other libraries to connect to

with runtime abstraction, you write the core logic once. each runtime handles the platform-specific stuff.

```relay
# this code works everywhere
state books []

show card "Add Book"
  show input title
  show input author
  show button "Add"
    when click
      add books { title: title, author: author }

for book in books
  show card
    show text (get book title)
    show text (get book author)
```

**browser runtime:** renders as HTML with forms and cards
**CLI runtime:** shows menu options and text prompts  
**mobile runtime:** native iOS/Android UI components
**server runtime:** REST API endpoints

same relay code. totally different user experiences.

## how it works under the hood

relay is fundamentally functional. everything is a function call. the indented syntax is just easier to read than lisp parentheses.

when you write:
```relay
show button "Click me"
  when click
    set counter (add counter 1)
```

the parser converts this to function calls and sends them to the runtime. the runtime decides how to actually show the button based on its platform.

but you don't need to think about any of this. you just write relay and it works.

## what each runtime provides

runtimes handle the platform-specific parts:

- **UI components** - buttons, forms, cards that make sense for the platform
- **IO** - reading files, databases, storage
- **networking** - HTTP, websockets, platform APIs
- **events** - clicks, keypresses, touch, whatever the platform supports

the beauty is that relay doesn't care. it says "show a button" and the runtime figures out how.

## extending runtimes with def-js

relay includes `def-js` for defining custom JavaScript builtin functions at runtime. this lets you extend any runtime with custom functionality.

```relay
# define eager function (all args evaluated first)
def-js max true "return Math.max(...args);"
max 1 5 3  # returns 5

# define lazy function (control evaluation yourself)
def-js unless false "
  const condition = evaluate(args[0], env);
  if (!condition) {
    return evaluate(args[1], env);
  }
  return null;
"

unless (> x 20) (show "x is not big enough")
```

**eager functions** get pre-evaluated arguments, perfect for:
- math operations: `abs`, `sqrt`, `round`
- string manipulation: `concat`, `length`, `trim`
- data processing: `sort`, `filter`, `map`

**lazy functions** control argument evaluation, perfect for:
- control flow: `unless`, `while`, `repeat`
- error handling: `try`, `safe`
- short-circuiting: `and`, `or`

this means communities can extend relay without waiting for core changes. need database functions? write them. want special UI components? add them. the runtime becomes exactly what you need.

## graceful degradation

not every platform supports everything. that's fine.

```relay
show date-picker
  fallback show input "Enter date (YYYY-MM-DD)"

show notification "Hello!"
  fallback show text "Hello!"
```

if a runtime can't do date pickers, it falls back to text input. if it can't do notifications, it shows regular text. your app still works.

## why this serves the vision

relay's vision is simple web development for everyone. runtime abstraction supports this because:

**you learn once, deploy anywhere** - write relay, run it on whatever platform makes sense

**no vendor lock-in** - start on one runtime, move to another if you need to

**communities choose their infrastructure** - browser for public sites, CLI for quick tools, servers for performance

**your code stays simple** - platform complexity is hidden from you

## implementation priorities

start with one really good runtime (probably javascript/browser) that just works perfectly. add others when needed.

the goal isn't to build every possible runtime. it's to prove the concept and let communities build what they need.

## what's next

the first priority is making relay so good in the browser that people want to use it everywhere else. nail the developer experience first. runtime abstraction is the technical foundation that makes this possible, but user experience is what makes it worth doing. 