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

// Main interpreter class
export class RelayInterpreter {
  private globalEnv: Environment;

  constructor() {
    this.globalEnv = createEnvironment();
    this.setupBuiltins();
  }

  // Evaluate a complete program
  evaluate(program: ProgramNode): any {
    let lastResult = null;
    
    for (const expression of program.expressions) {
      lastResult = this.evaluateExpression(expression, this.globalEnv);
    }
    
    return lastResult;
  }

  // Evaluate a single expression
  evaluateExpression(expr: ExpressionNode, env: Environment): any {
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

    // Output function for testing (EAGER)
    defineBuiltin("show", true, (args: any[]) => {
      const output = args.map(arg => {
        if (typeof arg === 'string') return arg;
        if (typeof arg === 'object') return JSON.stringify(arg);
        return String(arg);
      }).join(' ');
      
      console.log(output);
      return output;
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
  }
}

// Convenience function to run Relay code
export function runRelay(program: ProgramNode): any {
  const interpreter = new RelayInterpreter();
  return interpreter.evaluate(program);
} 