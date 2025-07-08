import { 
  parse, 
  RelayParser, 
  FuncallNode, 
  AtomNode, 
  SequenceNode, 
  JsonArrayNode, 
  JsonObjectNode,
  IdentifierNode 
} from '../parser';

describe('RelayParser - Simple Vision', () => {
  test('parses basic show call', () => {
    const ast = parse('show "hello"');
    
    expect(ast.type).toBe('program');
    expect(ast.expressions).toHaveLength(1);
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('show');
    expect(expr.args).toHaveLength(1);
    
    const arg = expr.args[0] as AtomNode;
    expect(arg.type).toBe('atom');
    expect(arg.value).toBe('hello');
  });

  test('parses show with simple JSON', () => {
    const ast = parse('show "container" {"class": "bold"}');
    
    expect(ast.type).toBe('program');
    expect(ast.expressions).toHaveLength(1);
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('show');
    expect(expr.args).toHaveLength(2);
    
    const jsonArg = expr.args[1] as JsonObjectNode;
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(1);
    expect(jsonArg.pairs[0].key).toBe('class');
    expect((jsonArg.pairs[0].value as AtomNode).value).toBe('bold');
  });

  test('parses math operations', () => {
    const ast = parse('+ 1 2');
    
    expect(ast.expressions).toHaveLength(1);
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('+');
    expect(expr.args).toHaveLength(2);
    
    expect((expr.args[0] as AtomNode).value).toBe(1);
    expect((expr.args[1] as AtomNode).value).toBe(2);
  });

  test('parses state function', () => {
    const ast = parse('state count 0');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('state');
    expect(expr.args).toHaveLength(2);
    
    expect((expr.args[0] as IdentifierNode).name).toBe('count');
    expect((expr.args[1] as AtomNode).value).toBe(0);
  });

  test('parses nested function call in parentheses', () => {
    const ast = parse('set total (+ count 1)');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.type).toBe('funcall');
    expect(expr.name).toBe('set');
    expect(expr.args).toHaveLength(2);
    
    const nestedCall = expr.args[1] as FuncallNode;
    expect(nestedCall.type).toBe('funcall');
    expect(nestedCall.name).toBe('+');
  });

  test('parses JSON arrays', () => {
    const ast = parse('tags ["tech", "web"]');
    
    const expr = ast.expressions[0] as FuncallNode;
    expect(expr.name).toBe('tags');
    
    const arrayArg = expr.args[0] as JsonArrayNode;
    expect(arrayArg.type).toBe('json_array');
    expect(arrayArg.elements).toHaveLength(2);
    expect((arrayArg.elements[0] as AtomNode).value).toBe('tech');
    expect((arrayArg.elements[1] as AtomNode).value).toBe('web');
  });

  test('parses multiple statements', () => {
    const ast = parse(`
state count 0
show "hello"
set count 1
    `);
    
    expect(ast.expressions).toHaveLength(3);
    
    const firstExpr = ast.expressions[0] as FuncallNode;
    expect(firstExpr.name).toBe('state');
    
    const secondExpr = ast.expressions[1] as FuncallNode;
    expect(secondExpr.name).toBe('show');
    
    const thirdExpr = ast.expressions[2] as FuncallNode;
    expect(thirdExpr.name).toBe('set');
  });
});

describe('Simple error handling', () => {
  test('throws error for invalid syntax', () => {
    expect(() => {
      parse('show "hello');
    }).toThrow();
  });
}); 