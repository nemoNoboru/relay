package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// evaluateServerExpr handles server definitions and starts the server
func (e *Evaluator) evaluateServerExpr(expr *parser.ServerExpr, env *Environment) (*Value, error) {
	// Initialize server state and receivers
	state := make(map[string]*Value)
	receivers := make(map[string]*Function)

	// Process server body elements
	for _, element := range expr.Body.Elements {
		if element.State != nil {
			// Process state fields
			for _, field := range element.State.Fields {
				var defaultValue *Value
				if field.DefaultValue != nil {
					val, err := e.evaluateLiteral(field.DefaultValue, env)
					if err != nil {
						return nil, fmt.Errorf("error evaluating default value for state field %s: %v", field.Name, err)
					}
					defaultValue = val
				} else {
					defaultValue = NewNil()
				}
				state[field.Name] = defaultValue
			}
		}

		if element.Receive != nil {
			// Process receive functions
			paramNames := make([]string, len(element.Receive.Parameters))
			for i, param := range element.Receive.Parameters {
				paramNames[i] = param.Name
			}

			receiver := &Function{
				Name:       element.Receive.Name,
				Parameters: paramNames,
				Body:       element.Receive.Body,
				IsBuiltin:  false,
				ClosureEnv: env, // Capture environment for receive functions
			}

			receivers[element.Receive.Name] = receiver
		}
	}

	// Create server instance
	serverValue := NewServer(expr.Name, state, receivers, env)

	// Store server in registry
	e.servers[expr.Name] = serverValue

	// Start the server
	serverValue.Server.Start(e)

	return serverValue, nil
}

// GetServer retrieves a server by name
func (e *Evaluator) GetServer(name string) (*Value, bool) {
	server, exists := e.servers[name]
	return server, exists
}

// StopAllServers stops all running servers
func (e *Evaluator) StopAllServers() {
	for _, server := range e.servers {
		if server.Server.Running {
			server.Server.Stop()
		}
	}
}
