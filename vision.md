# The Relay Vision

## Simple Web Development for Everyone

### The Problem

The web has been captured. What started as a decentralized network for sharing human knowledge has become a handful of walled gardens controlled by mega-corporations. Today's internet is really just 5 websites owned by 5 companies, all hosted on 3 cloud providers.

Building dynamic websites has become the exclusive domain of well-funded corporations and technical elites. The gap between "I can write HTML" and "I can build a real web application" has become a chasm filled with frameworks, build tools, deployment pipelines, and database setup nightmares that require teams of specialists and enterprise budgets.

Meanwhile, we've lost the magic of the early web - where anyone could create an HTML file, add some PHP, upload it to a server, and have a working dynamic website. The barrier to entry has grown so high that millions of potential creators, communities, and independent voices have been pushed to the sidelines, forced to build their digital presence on platforms they don't control.

### The Relay Solution

Relay is more than a development platform - it's a tool for digital independence. We're building technology that puts the power of web creation back into the hands of individuals, communities, and small organizations.

Relay brings back the simplicity of early web development while embracing modern capabilities. It's designed to break the corporate stranglehold on web development by making it accessible to everyone, not just venture-funded startups and Fortune 500 companies.

**Write in Markdown + Relay**

```relay
# My Book Club
relay {
state books (list [
    {"title": "The Great Gatsby", "author": "F. Scott Fitzgerald", "rating": 4},
    {"title": "To Kill a Mockingbird", "author": "Harper Lee", "rating": 5}
])

form add-book
	show card
		show text-input title "Write your book title"
		show text-input author "The author"
		show number-input rating "and your rating"

for book in books
    show book-card
        set title (get book title)
        set author (get book author)
        set rating (get book rating)

}
---
And on your server...

relay {

when add-book data
	set books
		add books 
			get data title
			get data author
			get data rating

}

you can also write and extend relay using javascript

"""
This is the documentation for filter_by_ratings function
param list of objects that have a rating
param number of minimum rating to be filtered by.

return list of objects with rating

---
# docstrings after the delimiter are valid relay code to be used as example and tests.

set correct-list
	list [{rating:4},{rating:3.5}]

set books 
	list [{rating: 2}, {rating: 3.5}, {rating: 1}, {rating:4}]

test 
	equals (filter_by_ratings books) correct-list 
"""
def-js filter_by_ratings 
	"
	(list, minimum_rating) => {
		return list.filter(l => l.rating > minimum_rating)
	}
	"

and use it,

filter_by_ratings books 3.0 

```

**See Results Instantly** As you type, watch your website come to life in the live preview. No build steps, no compilation errors, no "it works on my machine" problems.

**Share Components Naturally**

```relay
# Load someone else's creation
load https://gist.github.com/bookclub-components/reading-list.rl

# Use it immediately
show reading-list books
```

**Deploy with One Click** When you're ready, click publish and your site goes live on the Relay federated cloud - no servers to configure, no databases to set up, no deployment pipelines to debug.

### Core Principles

**English-Like Syntax** Relay reads like instructions to a human, not commands to a machine. If you can write a to-do list, you can learn Relay.

**Everything Just Works** Database included. Styling handled by Tailwind. Server logic runs automatically. Forms process correctly. You focus on your idea, not the infrastructure.

**Your Web in Your Computer** you host your apps and sites directly on your computer and people across the globe can securely access your website

**Community Owned** The Relay cloud is open source and federated. Communities can run their own instances, connect to others, or stay completely independent. No single company controls the infrastructure. Your digital presence belongs to you, not to shareholders.

**Escape hatch ready** Start with simple Relay syntax. Add JavaScript when you need more power. 

**Quality Built-In** Tests live alongside your code in documentation strings. Automatic mocking makes testing trivial. Share quality components, not broken examples.

### The Complete Workflow

1. **Write** - Create `.relay` files that mix markdown and logic
2. **Preview** - See your site update live as you type
3. **Test** - Built-in testing with automatic mocks
4. **Share** - Publish components as GitHub gists
5. **Deploy** - One-click publishing to Relay cloud
6. **Scale** - Federated hosting grows with your needs

### Who This Is For

**Local Libraries** who want to catalog books without paying subscription fees to corporate platforms

**Neighborhood Groups** organizing events without surrendering data to social media giants

**Independent Artists** selling work without platform fees eating their profits

**Small Nonprofits** building tools without enterprise software budgets

**Student Organizations** creating websites without IT department approval

**Hobby Communities** sharing knowledge without algorithmic interference

**Local Businesses** connecting with customers without advertising monopolies

**Activists and Organizers** communicating without corporate censorship

### Technical Foundation

- **Desktop app** - Easy to install and get started with visually
- **JavaScript Under the Hood** - Familiar runtime, full ecosystem access
- **Server-First Rendering** - Fast, SEO-friendly by default
- **Key-Value JSON Database** - Simple to understand, easy to work with
- **Tailwind Styling** - Professional appearance without CSS expertise
- **GitHub Gist Integration** - Instant component ecosystem
- **Federated Hosting** - Community-owned, not corporate-controlled

### Why This Matters

The centralization of the web isn't just a technical problem - it's a democratic crisis. When only corporations can afford to build digital tools, only corporate voices get amplified. When all hosting flows through three cloud providers, we have three points of failure for human communication.

We're watching the digital commons get enclosed, piece by piece. Platform fees extract value from creators. Algorithmic feeds control what ideas spread. Terms of service changes can destroy communities overnight. Surveillance capitalism monetizes human connection.

But technology isn't destiny. The same infrastructure that enables corporate dominance can enable digital independence - if we build the right tools.

### The Future We're Building

Imagine a web where:

- Anyone can build the digital tools their community needs
- Small organizations compete with corporations on equal technological footing
- Communities own their digital infrastructure instead of renting it from Big Tech
- Innovation happens in bedrooms and coffee shops, not just venture-funded offices
- Digital creators keep the value they create instead of feeding platform shareholders
- The internet serves human flourishing, not just corporate profit
- Diversity of voices leads to diversity of solutions

This isn't nostalgia for the past - it's a vision for a better digital future. One where technology serves democracy instead of undermining it.

### A Call to Action

The web belongs to all of us, but we have to fight to keep it that way. Every community that builds with Relay instead of surrendering to platform capitalism is a victory for digital independence.

Every component shared as a gist instead of locked behind corporate APIs is a contribution to the digital commons.

Every federated Relay instance is a node in a web that serves people, not profit.

The future of the internet isn't predetermined. We can still choose decentralization over consolidation, community ownership over corporate control, human agency over algorithmic manipulation.

But we need tools that make that choice practical, not just idealistic.

Relay is one of those tools.

---

_The web was built by people, for people. Let's take it back._