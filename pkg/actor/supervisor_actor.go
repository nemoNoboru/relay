package actor

import (
	"log"
	"strconv"
	"sync"
	"time"
)

// SupervisorActor manages the lifecycle of all other actors, including RelayServerActors.
type SupervisorActor struct {
	Actor  *Actor
	actors map[string]*Actor
	mu     sync.RWMutex
}

// NewSupervisorActor creates a new SupervisorActor.
func NewSupervisorActor(name string, router *Router) *SupervisorActor {
	s := &SupervisorActor{
		actors: make(map[string]*Actor),
	}
	s.Actor = NewActor(name, router, s.Receive)
	return s
}

// Start begins the supervisor's message processing loop.
func (s *SupervisorActor) Start() {
	s.Actor.Start()
}

// Stop terminates the supervisor and all actors it manages.
func (s *SupervisorActor) Stop() {
	s.Actor.Stop() // This will wait until the actor's loop is done.

	s.mu.Lock()
	defer s.mu.Unlock()

	var wg sync.WaitGroup
	for _, a := range s.actors {
		wg.Add(1)
		go func(act *Actor) {
			defer wg.Done()
			act.Stop()
		}(a)
	}
	wg.Wait()
	log.Printf("Supervisor %s and all its actors stopped", s.Actor.Name)
}

// Receive handles incoming messages for the supervisor.
func (s *SupervisorActor) Receive(msg ActorMsg) {
	// For now, we'll just log the received message.
	// Later, this will involve creating/routing to other actors.
	log.Printf("Supervisor %s received message: %v", s.Actor.Name, msg)

	switch msg.Type {
	case "create_child":
		childTypeName, ok := msg.Data.(string)
		if !ok {
			log.Printf("Supervisor %s: invalid data for 'create_child', expected string", s.Actor.Name)
			return
		}

		// Generate a unique name for the child
		childName := s.Actor.Name + "-" + childTypeName + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)

		var newChildActor *RelayServerActor
		switch childTypeName {
		case "RelayServerActor":
			newChildActor = NewRelayServerActor(childName, s.Actor.router)
		default:
			log.Printf("Supervisor %s: unknown child type '%s'", s.Actor.Name, childTypeName)
			return
		}

		newChildActor.Start()
		s.mu.Lock()
		s.actors[childName] = newChildActor.actor
		s.mu.Unlock()
		log.Printf("Supervisor %s created child %s", s.Actor.Name, childName)

		// Reply to the requester with the name of the created child if a reply
		// channel was provided.
		if msg.ReplyChan != nil {
			reply := ActorMsg{
				To:   msg.From,
				From: s.Actor.Name,
				Type: "child_created",
				Data: childName,
			}
			select {
			case msg.ReplyChan <- reply:
			case <-time.After(2 * time.Second):
				log.Printf("Supervisor %s: timeout sending reply for 'create_child' to %s", s.Actor.Name, msg.From)
			}
		}

	case "stop_child":
		childName, ok := msg.Data.(string)
		if !ok {
			log.Printf("Supervisor %s: invalid data for stop_child, expected string", s.Actor.Name)
			return
		}

		s.mu.Lock()
		childActor, exists := s.actors[childName]
		if exists {
			delete(s.actors, childName)
		}
		s.mu.Unlock()

		if exists {
			childActor.Stop()
			log.Printf("Supervisor %s stopped child %s", s.Actor.Name, childName)
		} else {
			log.Printf("Supervisor %s: child %s not found", s.Actor.Name, childName)
		}

	case "stop":
		s.Stop()
	default:
		s.mu.RLock()
		targetActor, ok := s.actors[msg.To]
		s.mu.RUnlock()

		if ok {
			targetActor.Send(msg)
		} else {
			// In the future, we might create the actor here if it doesn't exist.
			log.Printf("Supervisor %s: actor %s not found", s.Actor.Name, msg.To)
		}
	}
}
