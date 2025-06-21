package runtime

// ServerInitData contains the data needed to initialize a new RelayServerActor
// when the `server` keyword is encountered by the evaluator.
type ServerInitData struct {
	Name     string
	State    map[string]*Value
	Receives map[string]*Function
}
