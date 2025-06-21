package runtime

import "fmt"

// defineBuiltins adds built-in functions to the global environment
func (e *Evaluator) defineBuiltins() {
	e.definePrintFunction()
	e.defineMessageFunction()
	e.defineLenFunction()
	e.defineStringFunction()
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

// defineLenFunction adds the len function
func (e *Evaluator) defineLenFunction() {
	lenFunc := &Value{
		Type: ValueTypeFunction,
		Function: &Function{
			Name:      "len",
			IsBuiltin: true,
			Builtin: func(args []*Value) (*Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("len() expected 1 argument, got %d", len(args))
				}
				switch args[0].Type {
				case ValueTypeString:
					return NewNumber(float64(len(args[0].Str))), nil
				case ValueTypeArray:
					return NewNumber(float64(len(args[0].Array))), nil
				case ValueTypeObject:
					return NewNumber(float64(len(args[0].Object))), nil
				default:
					return nil, fmt.Errorf("len() not supported for type %s", args[0].Type)
				}
			},
		},
	}
	e.globalEnv.Define("len", lenFunc)
}

// SetGlobal defines a new variable in the evaluator's global environment.
// This is used to inject values or functions from outside the runtime,
// for example, an actor-aware 'send' function.
func (e *Evaluator) SetGlobal(name string, value *Value) {
	e.globalEnv.Define(name, value)
}

// defineStringFunction adds the string function
func (e *Evaluator) defineStringFunction() {
	stringFunc := &Value{
		Type: ValueTypeFunction,
		Function: &Function{
			Name:      "string",
			IsBuiltin: true,
			Builtin: func(args []*Value) (*Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("string() expected 1 argument, got %d", len(args))
				}
				// If it's already a string, return it to avoid adding extra quotes.
				if args[0].Type == ValueTypeString {
					return args[0], nil
				}
				return NewString(args[0].String()), nil
			},
		},
	}
	e.globalEnv.Define("string", stringFunc)
}
