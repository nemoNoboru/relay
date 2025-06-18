package runtime

import "fmt"

// defineBuiltins adds built-in functions to the global environment
func (e *Evaluator) defineBuiltins() {
	e.definePrintFunction()
	e.defineMessageFunction()
}

// definePrintFunction adds the print function
func (e *Evaluator) definePrintFunction() {
	printFunc := &Value{
		Type: ValueTypeFunction,
		Function: &Function{
			Name:      "print",
			IsBuiltin: true,
			Builtin: func(args []*Value) (*Value, error) {
				for i, arg := range args {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(arg.String())
				}
				fmt.Println()
				return NewNil(), nil
			},
		},
	}
	e.globalEnv.Define("print", printFunc)
}

// defineMessageFunction adds the message function for server communication
func (e *Evaluator) defineMessageFunction() {
	messageFunc := &Value{
		Type: ValueTypeFunction,
		Function: &Function{
			Name:      "message",
			IsBuiltin: true,
			Builtin: func(args []*Value) (*Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("message function expects at least 2 arguments: server_name, method_name, [args...]")
				}

				// First argument should be server name
				if args[0].Type != ValueTypeString {
					return nil, fmt.Errorf("first argument to message must be server name (string)")
				}
				serverName := args[0].Str

				// Second argument should be method name
				if args[1].Type != ValueTypeString {
					return nil, fmt.Errorf("second argument to message must be method name (string)")
				}
				methodName := args[1].Str

				// Remaining arguments are parameters
				methodArgs := args[2:]

				// Find the server
				server, exists := e.GetServer(serverName)
				if !exists {
					return nil, fmt.Errorf("server '%s' not found", serverName)
				}

				// Send message to server (synchronous call)
				result, err := server.Server.SendMessage(methodName, methodArgs, true)
				if err != nil {
					return nil, err
				}

				return result, nil
			},
		},
	}
	e.globalEnv.Define("message", messageFunc)
}
