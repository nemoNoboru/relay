package actor

import (
	"log"
	"sync"
)

// ActorMsg is the message format for all actor communication.
type ActorMsg struct {
	To   string
	From string
	Type string
	Data interface{}
}

// ReceiveFunc is the function signature for an actor's message handler.
type ReceiveFunc func(msg ActorMsg)

// Actor represents a concurrent entity in the actor model.
type Actor struct {
	Name     string
	inbox    chan ActorMsg
	router   *Router
	receive  ReceiveFunc
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// NewActor creates a new Actor.
func NewActor(name string, router *Router, receive ReceiveFunc) *Actor {
	return &Actor{
		Name:    name,
		inbox:   make(chan ActorMsg, 10), // Buffered channel
		router:  router,
		receive: receive,
	}
}

// Start begins the actor's message processing loop.
func (a *Actor) Start() {
	a.wg.Add(1)
	a.router.Register(a.Name, a)
	go func() {
		defer a.wg.Done()
		log.Printf("Actor %s started", a.Name)
		for msg := range a.inbox {
			a.receive(msg)
		}
		log.Printf("Actor %s's inbox closed", a.Name)
	}()
}

// Stop terminates the actor's message processing loop.
func (a *Actor) Stop() {
	a.stopOnce.Do(func() {
		a.router.Unregister(a.Name)
		close(a.inbox)
	})
	a.wg.Wait()
}

// Send puts a message into the actor's inbox.
func (a *Actor) Send(msg ActorMsg) {
	// Use a non-blocking send to prevent deadlocks if the inbox is full.
	select {
	case a.inbox <- msg:
	default:
		log.Printf("Actor %s: inbox is full, message dropped: %v", a.Name, msg)
	}
}

// Router is responsible for routing messages between actors.
type Router struct {
	actors map[string]*Actor
	mu     sync.RWMutex
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		actors: make(map[string]*Actor),
	}
}

// Register adds an actor to the router.
func (r *Router) Register(name string, actor *Actor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actors[name] = actor
	log.Printf("Actor %s registered with router", name)
}

// Unregister removes an actor from the router.
func (r *Router) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.actors, name)
	log.Printf("Actor %s unregistered from router", name)
}

// Send routes a message to the appropriate actor.
func (r *Router) Send(msg ActorMsg) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if actor, ok := r.actors[msg.To]; ok {
		actor.Send(msg)
	} else {
		log.Printf("Router: actor '%s' not found for message: %v", msg.To, msg)
	}
}

// StopAll stops all actors registered with the router.
func (r *Router) StopAll() {
	r.mu.Lock()
	actorsToStop := make([]*Actor, 0, len(r.actors))
	for _, actor := range r.actors {
		actorsToStop = append(actorsToStop, actor)
	}
	r.mu.Unlock()

	var wg sync.WaitGroup
	for _, actor := range actorsToStop {
		wg.Add(1)
		go func(a *Actor) {
			defer wg.Done()
			a.Stop()
		}(actor)
	}
	wg.Wait()
	log.Println("All actors stopped")
}
