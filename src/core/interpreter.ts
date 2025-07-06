// Relay Language Interpreter
// First draft - supports builtin functions and basic evaluation

import { 
  ProgramNode, 
  ExpressionNode, 
  FuncallNode, 
  AtomNode, 
  SequenceNode, 
  LambdaNode, 
  JsonArrayNode, 
  JsonObjectNode, 
  IdentifierNode 
} from './parser';

// Environment for variable and function scoping
export interface Environment {
  parent?: Environment;
  bindings: Map<string, any>;
}

// Lambda/Function representation at runtime
export interface RelayFunction {
  type: 'function';
  params: string[];
  body: ExpressionNode;
  closure: Environment;
}

// Builtin function signatures
export type EagerBuiltinFunction = (args: any[], env: Environment) => any;
export type LazyBuiltinFunction = (args: ExpressionNode[], env: Environment, evaluate: (expr: ExpressionNode, env: Environment) => any) => any;

export interface BuiltinSpec {
  fn: EagerBuiltinFunction | LazyBuiltinFunction;
  evaluateArgs: boolean; // true = eager evaluation, false = lazy evaluation
}

// Global builtins registry
export const builtins: Record<string, BuiltinSpec> = {};

// Helper function to define builtins with evaluation control
export function defineBuiltin(name: string, evaluateArgs: boolean, fn: EagerBuiltinFunction | LazyBuiltinFunction): void {
  builtins[name] = { fn, evaluateArgs };
}

// Create a new environment
export function createEnvironment(parent?: Environment): Environment {
  return {
    parent,
    bindings: new Map()
  };
}

// Look up a variable in the environment chain
export function lookupVariable(name: string, env: Environment): any {
  if (env.bindings.has(name)) {
    return env.bindings.get(name);
  }
  if (env.parent) {
    return lookupVariable(name, env.parent);
  }
  throw new Error(`Undefined variable: ${name}`);
}

// Set a variable in the current environment
export function setVariable(name: string, value: any, env: Environment): void {
  env.bindings.set(name, value);
}

// Renderable component for the show function
export interface RenderableComponent {
  type: 'component';
  name: string;
  props: Record<string, any>;
  children?: RenderableComponent[];
}

// Helper function to create a renderable component
export function createComponent(name: string, props: Record<string, any> = {}, children: RenderableComponent[] = []): RenderableComponent {
  return {
    type: 'component',
    name,
    props,
    children
  };
}

// Main interpreter class
export class RelayInterpreter {
  private globalEnv: Environment;
  private componentCollection: RenderableComponent[] = [];
  private isEvaluatingChildren: boolean = false;

  constructor() {
    this.globalEnv = createEnvironment();
    this.componentCollection = [];
    this.setupBuiltins();
  }

  // Add method to set child evaluation mode
  setChildEvaluationMode(isChild: boolean): void {
    this.isEvaluatingChildren = isChild;
  }

  // Add method to add components to the collection
  addComponent(component: RenderableComponent): void {
    // Only add to collection if not evaluating children
    if (!this.isEvaluatingChildren) {
      this.componentCollection.push(component);
    }
  }

  // Add method to set variables in the global environment
  setVariable(name: string, value: any): void {
    setVariable(name, value, this.globalEnv);
  }

  // Evaluate a complete program
  evaluate(program: ProgramNode): any {
    // Reset component collection for each evaluation
    this.componentCollection = [];
    
    let lastResult = null;
    
    for (const expression of program.expressions) {
      lastResult = this.evaluateExpression(expression, this.globalEnv);
    }
    
    // Extract event handlers from global environment
    const eventHandlers = this.globalEnv.bindings.get('_eventHandlers') || new Map();
    
    // If we have collected components, return them as a collection
    if (this.componentCollection.length > 0) {
      return {
        type: 'component_collection',
        components: this.componentCollection,
        eventHandlers: eventHandlers
      };
    }
    
    // Return result with event handlers if any exist
    if (eventHandlers.size > 0) {
      return {
        type: 'result_with_handlers',
        value: lastResult,
        eventHandlers: eventHandlers
      };
    }
    
    return lastResult;
  }

  // Evaluate a single expression (public method)
  public evaluateExpression(expr: ExpressionNode, env: Environment): any {
    switch (expr.type) {
      case 'atom':
        return this.evaluateAtom(expr as AtomNode, env);
      
      case 'identifier':
        return this.evaluateIdentifier(expr as IdentifierNode, env);
      
      case 'funcall':
        return this.evaluateFuncall(expr as FuncallNode, env);
      
      case 'sequence':
        return this.evaluateSequence(expr as SequenceNode, env);
      
      case 'lambda':
        return this.evaluateLambda(expr as LambdaNode, env);
      
      case 'json_array':
        return this.evaluateJsonArray(expr as JsonArrayNode, env);
      
      case 'json_object':
        return this.evaluateJsonObject(expr as JsonObjectNode, env);
      
      case 'comment':
        return null; // Comments evaluate to nothing
      
      default:
        throw new Error(`Unknown expression type: ${(expr as any).type}`);
    }
  }

  // Evaluate atoms (literals)
  private evaluateAtom(atom: AtomNode, env: Environment): any {
    return atom.value;
  }

  // Evaluate identifier references (variables)
  private evaluateIdentifier(identifier: IdentifierNode, env: Environment): any {
    return lookupVariable(identifier.name, env);
  }

  // Evaluate function calls
  private evaluateFuncall(funcall: FuncallNode, env: Environment): any {
    const functionName = funcall.name;
    
    // Check if it's a builtin function
    if (builtins[functionName]) {
      const builtin = builtins[functionName];
      
      if (builtin.evaluateArgs) {
        // Eager evaluation: evaluate all arguments first
        const evaluatedArgs = funcall.args.map(arg => this.evaluateExpression(arg, env));
        return (builtin.fn as EagerBuiltinFunction)(evaluatedArgs, env);
      } else {
        // Lazy evaluation: pass raw AST nodes and evaluator function
        return (builtin.fn as LazyBuiltinFunction)(funcall.args, env, (expr, env) => this.evaluateExpression(expr, env));
      }
    }
    
    // Check if it's a user-defined function or variable
    let value: any;
    try {
      value = lookupVariable(functionName, env);
    } catch (error) {
      // Variable/function not found, continue to error below
      throw new Error(`Unknown function: ${functionName}`);
    }
    
    // If it's a function, always call it (let callUserFunction handle argument validation)
    if (value && typeof value === 'object' && value.type === 'function') {
      return this.callUserFunction(value, funcall.args, env);
    }
    
    // If it's NOT a function and called with no arguments, return its value
    if (funcall.args.length === 0) {
      return value;
    }
    
    // Variable called with arguments is an error
    throw new Error(`Variable ${functionName} is not a function (called with ${funcall.args.length} arguments)`);
    
    throw new Error(`Unknown function: ${functionName}`);
  }

  // Call a user-defined function (lambda)
  private callUserFunction(func: RelayFunction, args: ExpressionNode[], env: Environment): any {
    if (args.length !== func.params.length) {
      throw new Error(`Function expects ${func.params.length} arguments, got ${args.length}`);
    }

    // Create new environment with function's closure as parent
    const callEnv = createEnvironment(func.closure);
    
    // Bind parameters to evaluated arguments
    for (let i = 0; i < func.params.length; i++) {
      const paramValue = this.evaluateExpression(args[i], env);
      setVariable(func.params[i], paramValue, callEnv);
    }
    
    // Evaluate function body
    return this.evaluateExpression(func.body, callEnv);
  }

  // Evaluate sequence blocks
  private evaluateSequence(sequence: SequenceNode, env: Environment): any {
    let lastResult = null;
    
    for (const expression of sequence.expressions) {
      lastResult = this.evaluateExpression(expression, env);
    }
    
    return lastResult;
  }

  // Evaluate lambda expressions (create functions)
  private evaluateLambda(lambda: LambdaNode, env: Environment): RelayFunction {
    return {
      type: 'function',
      params: lambda.params,
      body: lambda.body,
      closure: env // Capture current environment
    };
  }

  // Evaluate JSON arrays
  private evaluateJsonArray(array: JsonArrayNode, env: Environment): any[] {
    return array.elements.map(element => this.evaluateExpression(element, env));
  }

  // Evaluate JSON objects
  private evaluateJsonObject(object: JsonObjectNode, env: Environment): Record<string, any> {
    const result: Record<string, any> = {};
    
    for (const pair of object.pairs) {
      result[pair.key] = this.evaluateExpression(pair.value, env);
    }
    
    return result;
  }

  // Setup core builtin functions
  private setupBuiltins(): void {
    // Variable assignment: set name value (LAZY - controls first argument evaluation)
    defineBuiltin("set", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 2) {
        throw new Error("set expects exactly 2 arguments: name and value");
      }
      
      const nameArg = args[0];
      if (nameArg.type !== 'identifier') {
        throw new Error("set expects first argument to be an identifier");
      }
      
      const name = (nameArg as IdentifierNode).name;
      const value = evaluate(args[1], env);
      
      setVariable(name, value, env);
      return value;
    });

    // Conditional execution: if condition then-expr else-expr (LAZY - controls branch evaluation)
    defineBuiltin("if", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 3) {
        throw new Error("if expects exactly 3 arguments: condition, then-expr, else-expr");
      }
      
      const condition = evaluate(args[0], env);
      
      // In Relay, any non-false, non-null value is truthy
      if (condition !== false && condition !== null && condition !== 0) {
        return evaluate(args[1], env);  // evaluate then branch
      } else {
        return evaluate(args[2], env);  // evaluate else branch
      }
    });

    // Arithmetic operations (EAGER - all arguments pre-evaluated)
    defineBuiltin("+", true, (args: any[]) => {
      return args.reduce((sum, arg) => {
        if (typeof arg !== 'number') {
          throw new Error(`+ expects numbers, got ${typeof arg}`);
        }
        return sum + arg;
      }, 0);
    });

    defineBuiltin("-", true, (args: any[]) => {
      if (args.length === 0) return 0;
      if (args.length === 1) return -args[0];
      
      let result = args[0];
      for (let i = 1; i < args.length; i++) {
        if (typeof args[i] !== 'number') {
          throw new Error(`- expects numbers, got ${typeof args[i]}`);
        }
        result -= args[i];
      }
      return result;
    });

    defineBuiltin("*", true, (args: any[]) => {
      return args.reduce((product, arg) => {
        if (typeof arg !== 'number') {
          throw new Error(`* expects numbers, got ${typeof arg}`);
        }
        return product * arg;
      }, 1);
    });

    defineBuiltin("/", true, (args: any[]) => {
      if (args.length !== 2) {
        throw new Error("/ expects exactly 2 arguments");
      }
      
      const [dividend, divisor] = args;
      if (typeof dividend !== 'number' || typeof divisor !== 'number') {
        throw new Error("/ expects numbers");
      }
      
      if (divisor === 0) {
        throw new Error("Division by zero");
      }
      
      return dividend / divisor;
    });

    defineBuiltin("%", true, (args: any[]) => {
      if (args.length !== 2) {
        throw new Error("% expects exactly 2 arguments");
      }
      
      const [dividend, divisor] = args;
      if (typeof dividend !== 'number' || typeof divisor !== 'number') {
        throw new Error("% expects numbers");
      }
      
      return dividend % divisor;
    });

    // Comparison operations (EAGER)
    defineBuiltin("equal", true, (args: any[]) => {
      if (args.length !== 2) {
        throw new Error("equal expects exactly 2 arguments");
      }
      
      return args[0] === args[1];
    });

    defineBuiltin("<", true, (args: any[]) => {
      if (args.length !== 2) {
        throw new Error("< expects exactly 2 arguments");
      }
      
      return args[0] < args[1];
    });

    defineBuiltin(">", true, (args: any[]) => {
      if (args.length !== 2) {
        throw new Error("> expects exactly 2 arguments");
      }
      
      return args[0] > args[1];
    });

    // Function definition: def name params body (LAZY - controls argument evaluation)
    defineBuiltin("def", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 3) {
        throw new Error("def expects exactly 3 arguments: name, params, body");
      }
      
      // Don't evaluate the name - use it directly
      const nameArg = args[0];
      if (nameArg.type !== 'identifier') {
        throw new Error("def expects function name to be an identifier");
      }
      
      const name = (nameArg as IdentifierNode).name;
      
      // Don't evaluate params or body - store them as AST nodes
      const params = args[1];
      const body = args[2];
      
      // For now, simplified: assume params is a single identifier
      let paramNames: string[];
      if (params.type === 'identifier') {
        paramNames = [(params as IdentifierNode).name];
      } else {
        // TODO: Handle parameter lists properly
        throw new Error("def currently only supports single parameter");
      }
      
      // Create function object
      const func: RelayFunction = {
        type: 'function',
        params: paramNames,
        body: body,
        closure: env
      };
      
      setVariable(name, func, env);
      return func;
    });

    // Output function for testing (LAZY) - bind to this interpreter instance
    const interpreterInstance = this;
    defineBuiltin("show", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length === 0) {
        throw new Error("show expects at least 1 argument: component name");
      }
      
      const componentNameExpr = args[0];
      const componentName = evaluate(componentNameExpr, env);
      
      if (typeof componentName !== 'string') {
        throw new Error("show expects first argument to be a component name (string)");
      }
      
      // Extract props from remaining arguments
      const props: Record<string, any> = {};
      let children: RenderableComponent[] = [];
      
      // Handle different argument patterns
      if (args.length === 2) {
        const secondArg = args[1];
        
        // Check if it's a sequence expression (block syntax)
        if (secondArg.type === 'sequence') {
          // Handle block syntax: show grid [NEWLINE + INDENT] ... [DEDENT]
          const sequenceNode = secondArg as SequenceNode;
          children = [];
          
          // Set child evaluation mode to prevent adding children to global collection
          const wasEvaluatingChildren = interpreterInstance.isEvaluatingChildren;
          interpreterInstance.setChildEvaluationMode(true);
          
          for (const expr of sequenceNode.expressions) {
            const childResult = evaluate(expr, env);
            
            // Handle different types of results
            if (childResult && typeof childResult === 'object') {
              if (childResult.type === 'component') {
                // Single component
                children.push(childResult);
              } else if (Array.isArray(childResult)) {
                // Array of components (e.g., from 'for' function)
                for (const item of childResult) {
                  if (item && typeof item === 'object' && item.type === 'component') {
                    children.push(item);
                  }
                }
              } else if (childResult.type === 'component_collection' && Array.isArray(childResult.components)) {
                // Component collection (alternative format)
                for (const component of childResult.components) {
                  if (component && typeof component === 'object' && component.type === 'component') {
                    children.push(component);
                  }
                }
              }
            }
          }
          
          // Restore previous evaluation mode
          interpreterInstance.setChildEvaluationMode(wasEvaluatingChildren);
        } else {
          // Evaluate the second argument normally
          const secondValue = evaluate(secondArg, env);
          
          if (typeof secondValue === 'string') {
            // Simple case: show heading "Hello World"
            if (componentName === 'heading' || componentName === 'paragraph' || componentName === 'button') {
              props.text = secondValue;
            } else {
              props.content = secondValue;
            }
          } else if (typeof secondValue === 'object' && secondValue !== null) {
            // Object case: show heading {"text": "Hello", "level": 1}
            Object.assign(props, secondValue);
          }
        }
      } else if (args.length === 3) {
        // Handle 3 arguments: component name + JSON props + sequence block
        const secondArg = args[1];
        const thirdArg = args[2];
        
        // Second argument should be props (JSON object)
        const secondValue = evaluate(secondArg, env);
        if (typeof secondValue === 'object' && secondValue !== null) {
          Object.assign(props, secondValue);
        }
        
        // Third argument should be sequence block
        if (thirdArg.type === 'sequence') {
          const sequenceNode = thirdArg as SequenceNode;
          children = [];
          
          // Set child evaluation mode to prevent adding children to global collection
          const wasEvaluatingChildren = interpreterInstance.isEvaluatingChildren;
          interpreterInstance.setChildEvaluationMode(true);
          
          for (const expr of sequenceNode.expressions) {
            const childResult = evaluate(expr, env);
            
            // Handle different types of results
            if (childResult && typeof childResult === 'object') {
              if (childResult.type === 'component') {
                // Single component
                children.push(childResult);
              } else if (Array.isArray(childResult)) {
                // Array of components (e.g., from 'for' function)
                for (const item of childResult) {
                  if (item && typeof item === 'object' && item.type === 'component') {
                    children.push(item);
                  }
                }
              } else if (childResult.type === 'component_collection' && Array.isArray(childResult.components)) {
                // Component collection (alternative format)
                for (const component of childResult.components) {
                  if (component && typeof component === 'object' && component.type === 'component') {
                    children.push(component);
                  }
                }
              }
            }
          }
          
          // Restore previous evaluation mode
          interpreterInstance.setChildEvaluationMode(wasEvaluatingChildren);
        }
      } else if (args.length > 3) {
        // Multiple arguments - evaluate them and join as text
        const evaluatedArgs = args.slice(1).map(arg => evaluate(arg, env));
        props.text = evaluatedArgs.join(' ');
      }
      
      const component = createComponent(componentName, props, children);
      
      // Add to component collection
      interpreterInstance.addComponent(component);
      
      // Also log for debugging
      console.log(`[SHOW] ${componentName}:`, { props, children: children.length });
      
      return component;
    });

    // Define JavaScript builtin functions at runtime (LAZY - controls argument evaluation)
    defineBuiltin("def-js", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 3) {
        throw new Error("def-js expects exactly 3 arguments: name, evaluateArgs, jsCode");
      }
      
      // Get the function name
      const nameArg = args[0];
      if (nameArg.type !== 'identifier') {
        throw new Error("def-js expects first argument to be an identifier");
      }
      const name = (nameArg as IdentifierNode).name;
      
      // Get the evaluation strategy (eager/lazy)
      const evaluateArgsValue = evaluate(args[1], env);
      if (typeof evaluateArgsValue !== 'boolean') {
        throw new Error("def-js expects second argument to be a boolean (true for eager, false for lazy)");
      }
      
      // Get the JavaScript code
      const jsCode = evaluate(args[2], env);
      if (typeof jsCode !== 'string') {
        throw new Error("def-js expects third argument to be a string containing JavaScript code");
      }
      
      try {
        // Create the JavaScript function
        let jsFunction: EagerBuiltinFunction | LazyBuiltinFunction;
        
        if (evaluateArgsValue) {
          // Eager function: (args, env) => { ... }
          jsFunction = new Function('args', 'env', `
            ${jsCode}
          `) as EagerBuiltinFunction;
        } else {
          // Lazy function: (args, env, evaluate) => { ... }
          jsFunction = new Function('args', 'env', 'evaluate', `
            ${jsCode}
          `) as LazyBuiltinFunction;
        }
        
        // Register the new builtin
        defineBuiltin(name, evaluateArgsValue, jsFunction);
        
        return `Defined builtin function: ${name}`;
        
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        throw new Error(`Failed to create JavaScript function: ${errorMessage}`);
      }
    });

    // Get function for object property access (EAGER)
    defineBuiltin("get", true, (args: any[]) => {
      if (args.length !== 2) {
        throw new Error("get expects exactly 2 arguments: object and key");
      }
      
      const obj = args[0];
      const key = args[1];
      
      if (typeof obj !== 'object' || obj === null) {
        throw new Error("get expects first argument to be an object, got: " + typeof obj);
      }
      
      if (typeof key !== 'string' && typeof key !== 'number') {
        throw new Error("get expects second argument to be a string key or number index, got: " + typeof key);
      }
      
      return obj[key];
    });

    // Concat function for string concatenation (EAGER)
    defineBuiltin("concat", true, (args: any[]) => {
      if (args.length === 0) {
        return "";
      }
      
      // Convert all arguments to strings and concatenate
      return args.map(arg => {
        if (arg === null || arg === undefined) {
          return "";
        }
        return String(arg);
      }).join("");
    });

    // For function for dynamic component generation (LAZY)
    defineBuiltin("for", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 2) {
        throw new Error("for expects exactly 2 arguments: list and component-generator function");
      }
      
      // Evaluate the list argument
      const list = evaluate(args[0], env);
      
      // Evaluate the function argument
      const func = evaluate(args[1], env);
      
      // Check if first argument is an array
      if (!Array.isArray(list)) {
        throw new Error("for expects first argument to be a list/array, got: " + typeof list);
      }
      
      // Check if second argument is a function
      if (!func || typeof func !== 'object' || func.type !== 'function') {
        throw new Error("for expects second argument to be a function, got: " + typeof func);
      }
      
      // Generate components for each item
      const components: RenderableComponent[] = [];
      
      for (let i = 0; i < list.length; i++) {
        const item = list[i];
        
        // Create a new environment for function execution
        const funcEnv: Environment = {
          parent: func.closure || env,
          bindings: new Map()
        };
        
        // Set the parameter values in the function environment
        if (func.params && func.params.length > 0) {
          funcEnv.bindings.set(func.params[0], item);
          
          // If there's a second parameter, pass the index
          if (func.params.length > 1) {
            funcEnv.bindings.set(func.params[1], i);
          }
        }
        
        try {
          // Evaluate the function body to get the component
          const component = evaluate(func.body, funcEnv);
          
          // If the result is a component, add it to our collection
          if (component && typeof component === 'object' && component.type === 'component') {
            components.push(component);
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : String(error);
          throw new Error(`Error generating component for item ${i}: ${errorMessage}`);
        }
      }
      
      // Return the components as an array so they can be used as children
      return components;
    });

    // When function for event handling (LAZY)
    defineBuiltin("when", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length < 2) {
        throw new Error("when expects at least 2 arguments: event-name and handler-body");
      }
      
      // Get the event name - can be identifier or string literal
      const eventNameArg = args[0];
      let eventName: string;
      
      if (eventNameArg.type === 'identifier') {
        eventName = (eventNameArg as IdentifierNode).name;
      } else if (eventNameArg.type === 'atom' && typeof (eventNameArg as AtomNode).value === 'string') {
        eventName = (eventNameArg as AtomNode).value as string;
      } else {
        throw new Error("when expects first argument to be an event name (identifier or string)");
      }
      
      // Get the data parameter name (optional)
      let dataParamName = 'data';
      if (args.length > 2) {
        const dataArg = args[1];
        if (dataArg.type === 'identifier') {
          dataParamName = (dataArg as IdentifierNode).name;
        }
      }
      
      // Get the handler body (last argument)
      const handlerBody = args[args.length - 1];
      
      // Create event handler function
      const eventHandler = {
        type: 'event_handler',
        eventName,
        dataParamName,
        handler: handlerBody,
        env
      };
      
      // Store the event handler in the environment
      // This allows the renderer to access it later
      if (!env.bindings.has('_eventHandlers')) {
        env.bindings.set('_eventHandlers', new Map());
      }
      const eventHandlers = env.bindings.get('_eventHandlers');
      eventHandlers.set(eventName, eventHandler);
      
      console.log(`[EVENT] Registered handler for: ${eventName}`);
      return eventHandler;
    });

    // Form function for creating forms (LAZY)
    defineBuiltin("form", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length < 1) {
        throw new Error("form expects at least 1 argument: form-name");
      }
      
      // Get the form name - can be identifier or string literal
      const formNameArg = args[0];
      let formName: string;
      
      if (formNameArg.type === 'identifier') {
        formName = (formNameArg as IdentifierNode).name;
      } else if (formNameArg.type === 'atom' && typeof (formNameArg as AtomNode).value === 'string') {
        formName = (formNameArg as AtomNode).value as string;
      } else {
        throw new Error("form expects first argument to be a form name (identifier or string)");
      }
      
      // Create form component with the name
      const component = createComponent('form', { name: formName, onSubmit: formName });
      
      // Add to component collection
      interpreterInstance.addComponent(component);
      
      console.log(`[FORM] Created form: ${formName}`);
      return component;
    });

    // State function for state management (LAZY)
    defineBuiltin("state", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 2) {
        throw new Error("state expects exactly 2 arguments: variable-name and initial-value");
      }
      
      // Get the variable name
      const varNameArg = args[0];
      if (varNameArg.type !== 'identifier') {
        throw new Error("state expects first argument to be a variable name identifier");
      }
      const varName = (varNameArg as IdentifierNode).name;
      
      // Check if variable already exists, if so return its current value
      if (env.bindings.has(varName)) {
        const currentValue = env.bindings.get(varName);
        console.log(`[STATE] Variable ${varName} already exists, current value: `, currentValue);
        return currentValue;
      }
      
      // Evaluate the initial value only if variable doesn't exist
      const initialValue = evaluate(args[1], env);
      
      // Set the variable in the environment
      env.bindings.set(varName, initialValue);
      
      console.log(`[STATE] Initialized state variable: ${varName} = `, initialValue);
      return initialValue;
    });

    // Set function for updating state (LAZY)
    defineBuiltin("set", false, (args: ExpressionNode[], env: Environment, evaluate) => {
      if (args.length !== 2) {
        throw new Error("set expects exactly 2 arguments: variable-name and new-value");
      }
      
      // Get the variable name
      const varNameArg = args[0];
      if (varNameArg.type !== 'identifier') {
        throw new Error("set expects first argument to be a variable name identifier");
      }
      const varName = (varNameArg as IdentifierNode).name;
      
      // Evaluate the new value
      const newValue = evaluate(args[1], env);
      
      // Update the variable in the environment
      env.bindings.set(varName, newValue);
      
      console.log(`[STATE] Updated state variable: ${varName} = `, newValue);
      return newValue;
    });

    // List function for creating lists (EAGER)
    defineBuiltin("list", true, (args: any[]) => {
      // If only one argument and it's an array, return it directly
      if (args.length === 1 && Array.isArray(args[0])) {
        return args[0];
      }
      // Otherwise return all arguments as an array
      return args;
    });

    // Add function for adding items to lists (EAGER)
    defineBuiltin("add", true, (args: any[]) => {
      if (args.length < 2) {
        throw new Error("add expects at least 2 arguments: list and item(s)");
      }
      
      const list = args[0];
      const itemsToAdd = args.slice(1);
      
      if (!Array.isArray(list)) {
        throw new Error("add expects first argument to be a list/array");
      }
      
      // Return a new array with the items added
      return [...list, ...itemsToAdd];
    });
  }
}

// Convenience function to run Relay code
export function runRelay(program: ProgramNode): any {
  const interpreter = new RelayInterpreter();
  return interpreter.evaluate(program);
} 