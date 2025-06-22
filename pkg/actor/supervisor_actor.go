package actor

import (
	"log"
	"relay/pkg/runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SupervisorActor is responsible for managing the lifecycle of other actors.
type SupervisorActor struct {
	*Actor
	children map[string]*Actor
	mu       sync.RWMutex
	router   *Router
}

// NewSupervisorActor creates a new SupervisorActor.
func NewSupervisorActor(name string, router *Router) *SupervisorActor {
	s := &SupervisorActor{
		children: make(map[string]*Actor),
		router:   router,
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
	for _, a := range s.children {
		wg.Add(1)
		go func(act *Actor) {
			defer wg.Done()
			act.Stop()
		}(a)
	}
	wg.Wait()
	log.Printf("Supervisor %s and all its actors stopped", s.Name)
}

// Receive handles messages for the SupervisorActor.
func (s *SupervisorActor) Receive(msg ActorMsg) {
	switch {
	case strings.HasPrefix(msg.Type, "create_child:"):
		childTypeName := strings.TrimPrefix(msg.Type, "create_child: ")

		// Generate a unique name for the child actor
		baseName := s.Name + "-" + childTypeName
		childName := baseName + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)

		log.Printf("ChildType: %s", childTypeName)

		var newChild *Actor
		if childTypeName == "RelayServerActor" {
			log.Printf("Creating RelayServerActor")
			// The data can be a gateway name (string) or ServerInitData.
			var gatewayName string
			var initData *runtime.ServerInitData

			if data, ok := msg.Data.(runtime.ServerInitData); ok {
				initData = &data
				childName = data.Name // Use the name from the server definition
			} else if name, ok := msg.Data.(string); ok {
				gatewayName = name
			}

			newChildActor := NewRelayServerActor(childName, gatewayName, s.Name, s.router, initData)
			newChildActor.Start()
			newChild = newChildActor.Actor
		} else {
			log.Printf("Supervisor %s cannot create child of unknown type: %s", s.Name, msg.Type)
			return
		}

		s.mu.Lock()
		s.children[childName] = newChild
		s.mu.Unlock()
		log.Printf("Supervisor %s created child %s", s.Name, childName)

		if newChild != nil && msg.ReplyChan != nil {
			reply := ActorMsg{
				To:   msg.From,
				From: s.Name,
				Type: "child_created",
				Data: newChild.Name,
			}
			msg.ReplyChan <- reply
		}

	case msg.Type == "stop_child":
		childName, ok := msg.Data.(string)
		if !ok {
			log.Printf("Supervisor %s: invalid data for 'stop_child', expected string", s.Name)
			return
		}
		s.mu.Lock()
		child, exists := s.children[childName]
		if exists {
			child.Stop()
			delete(s.children, childName)
			log.Printf("Supervisor %s stopped child %s", s.Name, childName)
		} else {
			log.Printf("Supervisor %s: child %s not found to stop", s.Name, childName)
		}
		s.mu.Unlock()

	case msg.Type == "stop":
		s.Stop()
	default:
		log.Printf("Supervisor %s received unhandled message type: %s", s.Name, msg.Type)
	}
}
