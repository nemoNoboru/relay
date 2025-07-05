import { 
  tokenize, 
  parse, 
  RelayLexer, 
  RelayParser, 
  FuncallNode, 
  AtomNode, 
  SequenceNode, 
  LambdaNode, 
  JsonArrayNode, 
  JsonObjectNode,
  IdentifierNode 
} from '../parser';

describe('RelayLexer', () => {
  test('tokenizes simple identifiers', () => {
    const lexer = new RelayLexer('set working true');
    const tokens = lexer.tokenize();
    
    expect(tokens).toHaveLength(4); // set, working, true, EOF
    expect(tokens[0].type).toBe('IDENTIFIER');
    expect(tokens[0].value).toBe('set');
    expect(tokens[1].type).toBe('IDENTIFIER'); 
    expect(tokens[1].value).toBe('working');
    expect(tokens[2].type).toBe('BOOLEAN');
    expect(tokens[2].value).toBe(true);
    expect(tokens[3].type).toBe('EOF');
  });

  test('tokenizes strings', () => {
    const lexer = new RelayLexer('show "hello world"');
    const tokens = lexer.tokenize();
    
    expect(tokens[1]).toEqual({ type: 'STRING', value: 'hello world', line: 1, column: 1 });
  });

  test('tokenizes numbers', () => {
    const lexer = new RelayLexer('set count 42');
    const tokens = lexer.tokenize();
    
    expect(tokens[2]).toEqual({ type: 'NUMBER', value: 42, line: 1, column: 1 });
  });

  test('tokenizes float numbers', () => {
    const lexer = new RelayLexer('set price 19.99');
    const tokens = lexer.tokenize();
    
    expect(tokens[2]).toEqual({ type: 'NUMBER', value: 19.99, line: 1, column: 1 });
  });

  test('tokenizes operators', () => {
    const lexer = new RelayLexer('+ - * / %');
    const tokens = lexer.tokenize();
    
    expect(tokens[0]).toEqual({ type: 'OPERATOR', value: '+', line: 1, column: 1 });
    expect(tokens[1]).toEqual({ type: 'OPERATOR', value: '-', line: 1, column: 1 });
    expect(tokens[2]).toEqual({ type: 'OPERATOR', value: '*', line: 1, column: 1 });
    expect(tokens[3]).toEqual({ type: 'OPERATOR', value: '/', line: 1, column: 1 });
    expect(tokens[4]).toEqual({ type: 'OPERATOR', value: '%', line: 1, column: 1 });
  });

  test('tokenizes JSON brackets', () => {
    const lexer = new RelayLexer('[1, 2, 3]');
    const tokens = lexer.tokenize();
    
    expect(tokens[0]).toEqual({ type: 'LBRACKET', value: '[', line: 1, column: 1 });
    expect(tokens[4]).toEqual({ type: 'RBRACKET', value: ']', line: 1, column: 1 });
  });

  test('handles indentation', () => {
    const code = `if user
    show card`;
    const lexer = new RelayLexer(code);
    const tokens = lexer.tokenize();
    
    const tokenTypes = tokens.map(t => t.type);
    expect(tokenTypes).toContain('INDENT');
    expect(tokenTypes).toContain('DEDENT');
  });

  test('handles comments', () => {
    const lexer = new RelayLexer('# This is a comment\nset x 1');
    const tokens = lexer.tokenize();
    
    expect(tokens[0]).toEqual({ type: 'COMMENT', value: '# This is a comment', line: 1, column: 1 });
  });
});

describe('RelayParser', () => {
  test('parses simple function call', () => {
    const ast = parse('set working true');
    
    expect(ast.type).toBe('program');
    expect(ast.expressions).toHaveLength(1);
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('set');
    expect(expr.args).toHaveLength(2);
    
    const arg1 = expr.args[0] as IdentifierNode;
    expect(arg1.type).toBe('identifier');
    expect(arg1.name).toBe('working');
    
    const arg2 = expr.args[1] as AtomNode;
    expect(arg2.type).toBe('atom');
    expect(arg2.value).toBe(true);
  });

  test('parses function with sequence', () => {
    const code = `if user
    show card user
    show heading "welcome!"`;
    
    const ast = parse(code);
    const expr = ast.expressions[0] as FuncallNode;
    
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('if');
    expect(expr.args).toHaveLength(2);
    
    const userArg = expr.args[0] as IdentifierNode;
    expect(userArg.type).toBe('identifier');
    expect(userArg.name).toBe('user');
    
    const sequenceArg = expr.args[1] as SequenceNode;
    expect(sequenceArg.type).toBe('sequence');
    expect(sequenceArg.expressions).toHaveLength(2);
    
    const firstCall = sequenceArg.expressions[0] as FuncallNode;
    expect(firstCall.type).toBe('funcall');
    expect(firstCall.name).toBe('show');
    expect(firstCall.args).toHaveLength(2);
  });

  test('parses lambda with braces', () => {
    const ast = parse('map users {user: get user name}');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('map');
    expect(expr.args).toHaveLength(2);
    
    const lambdaArg = expr.args[1] as LambdaNode;
    expect(lambdaArg.type).toBe('lambda');
    expect(lambdaArg.params).toEqual(['user']);
    
    const body = lambdaArg.body as FuncallNode;
    expect(body.type).toBe('funcall');
    expect(body.name).toBe('get');
  });

  test('parses function definition', () => {
    const code = `def is_even? n
    equal (% n 2) 0`;
    
    const ast = parse(code);
    const expr = ast.expressions[0] as FuncallNode;
    
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('def');
    expect(expr.args).toHaveLength(3);
    
    const nameArg = expr.args[0] as IdentifierNode;
    expect(nameArg.name).toBe('is_even?');
    
    const paramArg = expr.args[1] as IdentifierNode;
    expect(paramArg.name).toBe('n');
    
    const bodyArg = expr.args[2] as SequenceNode;
    expect(bodyArg.type).toBe('sequence');
    expect(bodyArg.expressions).toHaveLength(1);
  });

  test('parses JSON array', () => {
    const ast = parse('set numbers [1, 2, 3]');
    
    const expr = ast.expressions[0] as FuncallNode;
    const arrayArg = expr.args[1] as JsonArrayNode;
    
    expect(arrayArg.type).toBe('json_array');
    expect(arrayArg.elements).toHaveLength(3);
    
    const firstElement = arrayArg.elements[0] as AtomNode;
    expect(firstElement.value).toBe(1);
  });

  test('parses JSON object', () => {
    const ast = parse('set user {"name": "john", "age": 30}');
    
    const expr = ast.expressions[0] as FuncallNode;
    const objectArg = expr.args[1] as JsonObjectNode;
    
    expect(objectArg.type).toBe('json_object');
    expect(objectArg.pairs).toHaveLength(2);
    
    expect(objectArg.pairs[0].key).toBe('name');
    const nameValue = objectArg.pairs[0].value as AtomNode;
    expect(nameValue.value).toBe('john');
    
    expect(objectArg.pairs[1].key).toBe('age');
    const ageValue = objectArg.pairs[1].value as AtomNode;
    expect(ageValue.value).toBe(30);
  });

  test('parses parenthesized function call', () => {
    const ast = parse('if user ((show card user) (show heading "welcome!"))');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('if');
    expect(expr.args).toHaveLength(3);
    
    const firstCall = expr.args[1] as FuncallNode;
    expect(firstCall.name).toBe('show');
    
    const secondCall = expr.args[2] as FuncallNode;
    expect(secondCall.name).toBe('show');
  });

  test('parses operators as atoms', () => {
    const ast = parse('+ 1 2');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('+');
    expect(expr.args).toHaveLength(2);
    
    const arg1 = expr.args[0] as AtomNode;
    expect(arg1.value).toBe(1);
    
    const arg2 = expr.args[1] as AtomNode;
    expect(arg2.value).toBe(2);
  });

  test('parses nested function calls', () => {
    const ast = parse('equal (% n 2) 0');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('equal');
    expect(expr.args).toHaveLength(2);
    
    const modCall = expr.args[0] as FuncallNode;
    expect(modCall.name).toBe('%');
    expect(modCall.args).toHaveLength(2);
    
    const nArg = modCall.args[0] as IdentifierNode;
    expect(nArg.name).toBe('n');
    
    const twoArg = modCall.args[1] as AtomNode;
    expect(twoArg.value).toBe(2);
  });

  test('parses multiple expressions', () => {
    const code = `set x 1
set y 2
+ x y`;
    
    const ast = parse(code);
    expect(ast.expressions).toHaveLength(3);
    
    const firstExpr = ast.expressions[0] as FuncallNode;
    expect(firstExpr.name).toBe('set');
    
    const secondExpr = ast.expressions[1] as FuncallNode;
    expect(secondExpr.name).toBe('set');
    
    const thirdExpr = ast.expressions[2] as FuncallNode;
    expect(thirdExpr.name).toBe('+');
  });

  test('handles comments in expressions', () => {
    const code = `# This is a comment
set x 1 # Another comment`;
    
    const ast = parse(code);
    // Comments are skipped at the top level
    expect(ast.expressions).toHaveLength(1);
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('set');
  });

  test('parses complex nested example', () => {
    const code = `def factorial n
    if (equal n 0)
        1
    else
        * n (factorial (- n 1))`;
    
    const ast = parse(code);
    const expr = ast.expressions[0] as FuncallNode;
    
    expect(expr.name).toBe('def');
    expect(expr.args).toHaveLength(3);
    
    const bodyArg = expr.args[2] as SequenceNode;
    expect(bodyArg.type).toBe('sequence');
    
    const ifExpr = bodyArg.expressions[0] as FuncallNode;
    expect(ifExpr.name).toBe('if');
    expect(ifExpr.args).toHaveLength(3);
  });

  test('handles empty input', () => {
    const ast = parse('');
    expect(ast.type).toBe('program');
    expect(ast.expressions).toHaveLength(0);
  });

  test('handles whitespace-only input', () => {
    const ast = parse('   \n  \t  \n  ');
    expect(ast.type).toBe('program');
    expect(ast.expressions).toHaveLength(0);
  });
});

describe('Error handling', () => {
  test('throws error for unterminated string', () => {
    expect(() => {
      const lexer = new RelayLexer('set message "hello world');
      lexer.tokenize();
    }).toThrow('Unterminated string');
  });

  test('throws error for unexpected character', () => {
    expect(() => {
      const lexer = new RelayLexer('set x @');
      lexer.tokenize();
    }).toThrow('Unexpected character: @');
  });

  test('throws error for indentation mismatch', () => {
    expect(() => {
      const code = `if user
    show card
  invalid_indent`;
      const lexer = new RelayLexer(code);
      lexer.tokenize();
    }).toThrow('Indentation mismatch');
  });

  test('throws error for missing closing brace', () => {
    expect(() => {
      parse('map users {user: get user name');
    }).toThrow('Expected RBRACE');
  });

  test('throws error for missing closing parenthesis', () => {
    expect(() => {
      parse('if user (show card user');
    }).toThrow('Expected RPAREN');
  });
});

describe('Edge cases', () => {
  test('handles identifiers with question marks', () => {
    const ast = parse('is_even? 42');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('is_even?');
    expect(expr.args).toHaveLength(1);
  });

  test('handles identifiers with hyphens', () => {
    const ast = parse('some-function arg');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('some-function');
  });

  test('handles negative numbers', () => {
    const ast = parse('set x -42');
    
    const expr = ast.expressions[0] as FuncallNode;
    const numArg = expr.args[1] as AtomNode;
    expect(numArg.value).toBe(-42);
  });

  test('handles escaped strings', () => {
    const ast = parse('set message "hello\\nworld"');
    
    const expr = ast.expressions[0] as FuncallNode;
    const strArg = expr.args[1] as AtomNode;
    expect(strArg.value).toBe('hello\nworld');
  });

  test('handles empty JSON structures', () => {
    const ast = parse('set empty_array []');
    
    const expr = ast.expressions[0] as FuncallNode;
    const arrayArg = expr.args[1] as JsonArrayNode;
    expect(arrayArg.elements).toHaveLength(0);
  });
});

describe('Integration tests', () => {
  test('parses complete Relay program', () => {
    const code = `
# Example Relay program
def greet name
    concat "Hello, " name "!"

set users ["Alice", "Bob", "Charlie"]
set greetings (map users greet)

show "Greetings:"
for greetings
    greeting: show greeting
`;
    
    const ast = parse(code);
    expect(ast.type).toBe('program');
    expect(ast.expressions.length).toBeGreaterThan(0);
    
    // Check that we have function definitions and calls
    const funcalls = ast.expressions.filter(expr => expr.type === 'funcall') as FuncallNode[];
    const functionNames = funcalls.map(call => call.name);
    
    expect(functionNames).toContain('def');
    expect(functionNames).toContain('set');
    expect(functionNames).toContain('show');
    expect(functionNames).toContain('for');
  });
}); 