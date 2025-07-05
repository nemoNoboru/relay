# Relay Language Implementation Guide - Final

This guide provides practical implementation advice for the Relay language based on our refined grammar understanding.

## Core Principles

1. **Everything is an expression** that returns a value
2. **Indented blocks are syntactic sugar** for sequence expressions
3. **The runtime is uniform** - it only evaluates expressions
4. **Parentheses are optional** and used only for disambiguation

## Grammar Summary

```ebnf
program = expression*
expression = funcall | atom | lambda | sequence | comment
funcall = identifier argument_list
sequence = indent expression+ dedent
```

## AST Structure

```typescript
interface Expression {
  type: 'funcall' | 'atom' | 'lambda' | 'sequence' | 'json_array' | 'json_object' | 'comment';
}

interface FuncallExpression extends Expression {
  type: 'funcall';
  name: string;
  args: Expression[];
}

interface SequenceExpression extends Expression {
  type: 'sequence';
  expressions: Expression[];
}

interface LambdaExpression extends Expression {
  type: 'lambda';
  params: string[];
  body: Expression;
}

interface AtomExpression extends Expression {
  type: 'atom';
  value: any;
}
```

## Parser Implementation

```javascript
class RelayParser {
  // program = expression*
  parseProgram() {
    const expressions = [];
    while (!this.isEOF()) {
      if (this.isComment() || this.isNewline()) {
        this.advance();
        continue;
      }
      expressions.push(this.parseExpression());
    }
    return { type: 'program', expressions };
  }

  // expression = funcall | atom | lambda | sequence | comment
  parseExpression() {
    if (this.isIndent()) {
      return this.parseSequence();
    }
    if (this.isLambdaStart()) {
      return this.parseLambda();
    }
    if (this.isAtom()) {
      return this.parseAtom();
    }
    return this.parseFuncall();
  }

  // funcall = identifier argument_list
  parseFuncall() {
    const name = this.parseIdentifier();
    const args = this.parseArgumentList();
    return { type: 'funcall', name, args };
  }

  // argument_list = inline_args | sequence_arg | parenthesized_args
  parseArgumentList() {
    // Check for sequence argument (indented)
    if (this.isNewlineAndIndent()) {
      return [this.parseSequence()];
    }
    
    // Check for parenthesized arguments
    if (this.check('(')) {
      this.advance(); // consume '('
      const args = [];
      while (!this.check(')')) {
        args.push(this.parseExpression());
      }
      this.consume(')');
      return args;
    }
    
    // Parse inline arguments
    const args = [];
    while (this.hasMoreArgs()) {
      args.push(this.parseExpression());
    }
    return args;
  }

  // sequence = indent expression+ dedent
  parseSequence() {
    this.consume('NEWLINE');
    this.consume('INDENT');
    
    const expressions = [];
    while (!this.check('DEDENT')) {
      if (this.isComment() || this.isNewline()) {
        this.advance();
        continue;
      }
      expressions.push(this.parseExpression());
    }
    
    this.consume('DEDENT');
    return { type: 'sequence', expressions };
  }

  // lambda = brace_lambda | block_lambda
  parseLambda() {
    if (this.check('{')) {
      return this.parseBraceLambda();
    }
    return this.parseBlockLambda();
  }

  parseBraceLambda() {
    this.consume('{');
    const params = this.parseParameters();
    this.consume(':');
    const body = this.parseExpression();
    this.consume('}');
    return { type: 'lambda', params, body };
  }

  parseBlockLambda() {
    const params = this.parseParameters();
    this.consume(':');
    
    let body;
    if (this.isNewlineAndIndent()) {
      body = this.parseSequence();
    } else {
      body = this.parseExpression();
    }
    
    return { type: 'lambda', params, body };
  }

  // atom = string | number | boolean | null | json_literal | identifier
  parseAtom() {
    if (this.check('STRING')) {
      return { type: 'atom', value: this.advance().value };
    }
    if (this.check('NUMBER')) {
      return { type: 'atom', value: this.advance().value };
    }
    if (this.check('true') || this.check('false')) {
      return { type: 'atom', value: this.advance().value === 'true' };
    }
    if (this.check('null')) {
      this.advance();
      return { type: 'atom', value: null };
    }
    if (this.check('[')) {
      return this.parseJsonArray();
    }
    if (this.check('{') && !this.isLambdaStart()) {
      return this.parseJsonObject();
    }
    if (this.check('IDENTIFIER')) {
      return { type: 'atom', value: this.advance().value };
    }
    
    throw new Error(`Unexpected token: ${this.currentToken()}`);
  }
}
```

## Runtime Implementation

```javascript
class RelayRuntime {
  constructor() {
    this.scopes = [{}];
    this.builtins = this.createBuiltins();
  }

  // Main evaluation function
  eval(expr) {
    switch (expr.type) {
      case 'funcall':
        return this.evalFuncall(expr);
      case 'sequence':
        return this.evalSequence(expr);
      case 'lambda':
        return this.evalLambda(expr);
      case 'atom':
        return this.evalAtom(expr);
      case 'json_array':
        return this.evalJsonArray(expr);
      case 'json_object':
        return this.evalJsonObject(expr);
      case 'comment':
        return null; // Comments are ignored
      default:
        throw new Error(`Unknown expression type: ${expr.type}`);
    }
  }

  evalFuncall(expr) {
    // Look up function
    const func = this.builtins[expr.name] || this.getVariable(expr.name);
    if (!func) {
      throw new Error(`Unknown function: ${expr.name}`);
    }
    
    // Evaluate arguments
    const args = expr.args.map(arg => this.eval(arg));
    
    // Call function
    return func(args);
  }

  evalSequence(expr) {
    let result = null;
    for (const subExpr of expr.expressions) {
      result = this.eval(subExpr);
    }
    return result; // Return value of last expression
  }

  evalLambda(expr) {
    // Capture current scope
    const capturedScope = { ...this.currentScope() };
    
    return (args) => {
      this.pushScope(capturedScope);
      
      // Bind parameters
      expr.params.forEach((param, i) => {
        this.setVariable(param, args[i]);
      });
      
      // Evaluate body
      const result = this.eval(expr.body);
      
      this.popScope();
      return result;
    };
  }

  evalAtom(expr) {
    if (typeof expr.value === 'string' && this.hasVariable(expr.value)) {
      return this.getVariable(expr.value);
    }
    return expr.value;
  }

  evalJsonArray(expr) {
    return expr.elements.map(elem => this.eval(elem));
  }

  evalJsonObject(expr) {
    const obj = {};
    for (const pair of expr.pairs) {
      obj[pair.key] = this.eval(pair.value);
    }
    return obj;
  }

  // Built-in functions
  createBuiltins() {
    return {
      // Variable assignment
      set: (args) => {
        const [name, value] = args;
        this.setVariable(name, value);
        return value;
      },

      // Conditional
      if: (args) => {
        const [condition, positive, negative] = args;
        if (condition) {
          return positive;
        }
        return negative || null;
      },

      // Function definition
      def: (args) => {
        const name = args[0];
        const params = args.slice(1, -1);
        const body = args[args.length - 1];

        const func = (callArgs) => {
          this.pushScope();
          params.forEach((param, i) => {
            this.setVariable(param, callArgs[i]);
          });
          const result = this.eval(body);
          this.popScope();
          return result;
        };

        this.setVariable(name, func);
        return func;
      },

      // Collection operations
      map: (args) => {
        const [collection, lambda] = args;
        return collection.map(item => lambda([item]));
      },

      filter: (args) => {
        const [collection, lambda] = args;
        return collection.filter(item => lambda([item]));
      },

      // Utility functions
      get: (args) => {
        const [obj, key] = args;
        return obj[key];
      },

      equal: (args) => {
        const [a, b] = args;
        return a === b;
      },

      concat: (args) => {
        return args.join('');
      },

      // Math operations
      '+': (args) => args.reduce((a, b) => a + b),
      '-': (args) => args.reduce((a, b) => a - b),
      '*': (args) => args.reduce((a, b) => a * b),
      '/': (args) => args.reduce((a, b) => a / b),
      '%': (args) => args[0] % args[1],

      // UI functions (example)
      show: (args) => {
        console.log('SHOW:', ...args);
        return args[args.length - 1]; // Return last argument
      }
    };
  }

  // Scope management
  pushScope(initialScope = {}) {
    this.scopes.push({ ...this.currentScope(), ...initialScope });
  }

  popScope() {
    if (this.scopes.length > 1) {
      this.scopes.pop();
    }
  }

  currentScope() {
    return this.scopes[this.scopes.length - 1];
  }

  setVariable(name, value) {
    this.currentScope()[name] = value;
  }

  getVariable(name) {
    for (let i = this.scopes.length - 1; i >= 0; i--) {
      if (name in this.scopes[i]) {
        return this.scopes[i][name];
      }
    }
    throw new Error(`Undefined variable: ${name}`);
  }

  hasVariable(name) {
    for (let i = this.scopes.length - 1; i >= 0; i--) {
      if (name in this.scopes[i]) {
        return true;
      }
    }
    return false;
  }
}
```

## Usage Example

```javascript
// Example Relay code:
const relayCode = `
def is_even? n
    set result (% n 2)
    equal result 0

set numbers [1, 2, 3, 4, 5]
set evens (filter numbers is_even?)
show evens
`;

// Parse and execute
const tokens = lexer.tokenize(relayCode);
const parser = new RelayParser(tokens);
const ast = parser.parseProgram();

const runtime = new RelayRuntime();
for (const expr of ast.expressions) {
  const result = runtime.eval(expr);
  console.log('Result:', result);
}
```

## Key Insights

1. **Sequences are just expressions** - they evaluate all sub-expressions and return the last value
2. **Functions receive evaluated arguments** - no special handling needed
3. **Everything composes naturally** - because everything returns a value
4. **The runtime is simple** - just a switch statement for expression types
5. **Indentation is purely syntactic** - the runtime only sees sequence expressions

This design makes Relay both **powerful and simple** - the grammar is minimal but the expressiveness comes from composition of expressions. 