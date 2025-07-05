# Relay Grammar Implementation Guide (Expression-Based)

This document provides practical guidance for implementing the Relay language parser, where **everything is an expression that returns a value**.

## Overview

The Relay language is fundamentally **expression-based**:
- **Everything is an expression** - no statements, only expressions that return values
- **Function calls are expressions** - they return values that can be used in other expressions
- **Blocks are expressions** - they return the value of the last expression
- **Composable** - expressions can be nested and combined naturally
- **Consistent** - same evaluation model applies everywhere

## Key Philosophy: Everything Returns a Value

Unlike languages with statements and expressions, Relay has **only expressions**:

```relay
# These are all expressions that return values:
set working true                    # returns true
if working "active" "inactive"      # returns "active" or "inactive"
def greet name (concat "Hello, " name)  # returns the function
map users {user: get user name}     # returns list of names

# Even blocks are expressions:
if working
    set status "active"
    show notification status
    status                          # block returns "active"
```

## Grammar Structure

The grammar is built around expressions:

```ebnf
program = expression_list ;
expression = function_call_expression | lambda_expression | infix_expression | primary_expression ;
function_call_expression = identifier argument_list ;
expression_block = indent expression_list dedent ;
```

## Implementation Considerations

### 1. Expression Evaluation Model

Every expression must return a value:

```javascript
function evaluateExpression(expr) {
  switch (expr.type) {
    case 'function_call_expression':
      return evaluateFunctionCall(expr);
    case 'lambda_expression':
      return createLambda(expr);
    case 'literal':
      return expr.value;
    case 'identifier':
      return getVariable(expr.name);
    case 'expression_block':
      return evaluateExpressionBlock(expr);
    // ... etc
  }
}

function evaluateExpressionBlock(block) {
  let result = null;
  for (const expr of block.expressions) {
    result = evaluateExpression(expr);
  }
  return result; // Return value of last expression
}
```

### 2. Function Call Expressions

Function calls are expressions that return values:

```javascript
function evaluateFunctionCall(expr) {
  const func = getFunction(expr.name);
  const args = expr.args.map(arg => evaluateExpression(arg));
  return func(...args); // Returns whatever the function returns
}

// Core functions that return values
const coreFunctions = {
  set: (name, value) => {
    setVariable(name, value);
    return value; // 'set' returns the value that was set
  },
  
  if: (condition, thenExpr, elseExpr) => {
    return condition ? thenExpr : elseExpr;
  },
  
  def: (name, ...params) => {
    const func = createFunction(params);
    setVariable(name, func);
    return func; // 'def' returns the function
  },
  
  map: (collection, lambda) => {
    return collection.map(item => lambda(item));
  },
  
  get: (obj, key) => {
    return obj[key];
  }
};
```

### 3. Lambda Expressions

Lambda expressions create function values:

```javascript
function createLambda(lambdaExpr) {
  return (...args) => {
    // Bind parameters to arguments
    const bindings = {};
    lambdaExpr.params.forEach((param, i) => {
      bindings[param] = args[i];
    });
    
    // Evaluate body in new scope
    pushScope(bindings);
    const result = evaluateExpression(lambdaExpr.body);
    popScope();
    
    return result;
  };
}
```

### 4. Expression Blocks

Blocks are expressions that return the last expression's value:

```javascript
function parseExpressionBlock() {
  consumeToken(); // indent
  
  const expressions = [];
  while (currentToken.type !== 'DEDENT') {
    expressions.push(parseExpression());
  }
  
  consumeToken(); // dedent
  
  return {
    type: 'expression_block',
    expressions
  };
}
```

## PEG.js Grammar (Expression-Based)

```pegjs
start = program

program = expressions:expression_list { 
  return { type: 'program', expressions }; 
}

expression_list = expressions:(expression)* { 
  return expressions.filter(e => e != null); 
}

expression = 
  function_call_expression /
  lambda_expression /
  infix_expression /
  primary_expression /
  comment

function_call_expression = name:identifier _ args:argument_list {
  return { type: 'function_call_expression', name, args };
}

argument_list = 
  indented_arguments /
  inline_arguments

indented_arguments = 
  NEWLINE INDENT args:argument_expression_list DEDENT { 
    return { type: 'indented_args', args }; 
  }

argument_expression_list = 
  expressions:(argument_expression)+ { 
    return expressions; 
  }

argument_expression = 
  lambda_expression_block /
  expression

lambda_expression_block = 
  params:parameter_list _ ":" _ body:lambda_body {
    return { type: 'lambda_expression', params, body };
  }

lambda_body = 
  expression /
  (NEWLINE expression_block)

expression_block = 
  INDENT expressions:expression_list DEDENT {
    return { type: 'expression_block', expressions };
  }

primary_expression = 
  parenthesized_expression /
  literal /
  identifier /
  json_literal

parenthesized_expression = 
  "(" _ expressions:expression_list _ ")" {
    return { type: 'parenthesized_expression', expressions };
  }

// ... rest of grammar
```

## AST Structure

The AST represents everything as expressions:

```typescript
interface RelayProgram {
  type: 'program';
  expressions: Expression[];
}

interface Expression {
  type: 'function_call_expression' | 'lambda_expression' | 'expression_block' | 'literal' | 'identifier' | 'infix_expression' | 'parenthesized_expression';
}

interface FunctionCallExpression extends Expression {
  type: 'function_call_expression';
  name: string;
  args: Expression[];
}

interface LambdaExpression extends Expression {
  type: 'lambda_expression';
  params: string[];
  body: Expression;
}

interface ExpressionBlock extends Expression {
  type: 'expression_block';
  expressions: Expression[];
}

interface LiteralExpression extends Expression {
  type: 'literal';
  value: any;
}
```

## Testing Examples

```javascript
const testCases = [
  // Simple expression
  {
    name: 'simple_expression',
    input: 'set working true',
    expected: {
      type: 'function_call_expression',
      name: 'set',
      args: ['working', true]
    },
    evaluatesTo: true
  },
  
  // Nested expressions
  {
    name: 'nested_expressions',
    input: 'set name (get user name)',
    expected: {
      type: 'function_call_expression',
      name: 'set',
      args: [
        'name',
        {
          type: 'function_call_expression',
          name: 'get',
          args: ['user', 'name']
        }
      ]
    },
    evaluatesTo: 'user\'s name value'
  },
  
  // Block expression
  {
    name: 'block_expression',
    input: `if working
    set status "active"
    show notification status
    status`,
    expected: {
      type: 'function_call_expression',
      name: 'if',
      args: [
        'working',
        {
          type: 'expression_block',
          expressions: [
            { type: 'function_call_expression', name: 'set', args: ['status', 'active'] },
            { type: 'function_call_expression', name: 'show', args: ['notification', 'status'] },
            { type: 'identifier', name: 'status' }
          ]
        }
      ]
    },
    evaluatesTo: 'active'
  },
  
  // Lambda expression
  {
    name: 'lambda_expression',
    input: `map users
    user: get user name`,
    expected: {
      type: 'function_call_expression',
      name: 'map',
      args: [
        'users',
        {
          type: 'lambda_expression',
          params: ['user'],
          body: { type: 'function_call_expression', name: 'get', args: ['user', 'name'] }
        }
      ]
    },
    evaluatesTo: ['array', 'of', 'names']
  }
];
```

## Expression Evaluation Rules

### 1. Function Call Expressions
- Evaluate all arguments first
- Call the function with evaluated arguments
- Return the function's return value

### 2. Block Expressions
- Evaluate each expression in sequence
- Return the value of the last expression
- Each expression can use values from previous expressions

### 3. Lambda Expressions
- Create a closure capturing current scope
- Return a function that can be called later
- When called, evaluate body in new scope with parameters bound

### 4. Literal Expressions
- Return the literal value directly

### 5. Identifier Expressions
- Look up the variable in current scope
- Return the variable's value

## Core Runtime Functions

All core functions must return values:

```javascript
const coreRuntime = {
  // Variable operations
  set: (name, value) => {
    setVariable(name, value);
    return value; // Returns the value that was set
  },
  
  get: (obj, key) => {
    return obj[key]; // Returns the property value
  },
  
  // Control flow
  if: (condition, thenExpr, elseExpr) => {
    return condition ? thenExpr : (elseExpr || null);
  },
  
  // Function definition
  def: (name, params, body) => {
    const func = createFunction(params, body);
    setVariable(name, func);
    return func; // Returns the function
  },
  
  // Collection operations
  map: (collection, lambda) => {
    return collection.map(item => lambda(item));
  },
  
  filter: (collection, lambda) => {
    return collection.filter(item => lambda(item));
  },
  
  // UI operations (return elements or handles)
  show: (component, props) => {
    const element = createUIElement(component, props);
    renderElement(element);
    return element; // Returns the element
  },
  
  // Utility functions
  concat: (...args) => {
    return args.join('');
  },
  
  equal: (a, b) => {
    return a === b;
  },
  
  add: (a, b) => {
    return a + b;
  }
};
```

## Error Handling

Expressions can fail, so handle errors gracefully:

```javascript
function safeEvaluateExpression(expr) {
  try {
    return evaluateExpression(expr);
  } catch (error) {
    throw new Error(`Expression evaluation failed: ${error.message}
      Expression: ${JSON.stringify(expr, null, 2)}`);
  }
}
```

## Key Benefits

1. **Consistent model** - everything follows the same evaluation rules
2. **Composable** - expressions can be nested arbitrarily
3. **Functional** - natural support for functional programming patterns
4. **Predictable** - every expression returns a value
5. **Testable** - easy to test individual expressions

## Implementation Priority

1. **Build expression evaluator** (core feature)
2. **Implement expression blocks** with proper return values
3. **Add function call expressions** with argument evaluation
4. **Handle lambda expressions** as first-class values
5. **Test expression composition** thoroughly

This expression-based approach makes Relay both powerful and consistent - everything is composable because everything returns a value! 