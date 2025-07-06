import { parse } from '../parser';

describe('Multiline JSON Support', () => {
  test('parses single line JSON object', () => {
    const program = parse('show "card" {"title": "Test", "content": "Single line"}');
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    expect(funcall.type).toBe('funcall');
    expect(funcall.name).toBe('show');
    
    const jsonArg = funcall.args[1];
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(2);
    expect(jsonArg.pairs[0].key).toBe('title');
    expect(jsonArg.pairs[1].key).toBe('content');
  });

  test('parses multiline JSON object', () => {
    const program = parse(`show "card" {
  "title": "Test",
  "content": "Multiline"
}`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    expect(funcall.type).toBe('funcall');
    expect(funcall.name).toBe('show');
    
    const jsonArg = funcall.args[1];
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(2);
    expect(jsonArg.pairs[0].key).toBe('title');
    expect(jsonArg.pairs[1].key).toBe('content');
  });

  test('parses multiline JSON object without commas', () => {
    const program = parse(`show "card" {
  "title": "Test"
  "content": "No commas"
}`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    const jsonArg = funcall.args[1];
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(2);
    expect(jsonArg.pairs[0].key).toBe('title');
    expect(jsonArg.pairs[1].key).toBe('content');
  });

  test('parses multiline JSON array', () => {
    const program = parse(`set "colors" [
  "red",
  "green",
  "blue"
]`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    expect(funcall.type).toBe('funcall');
    expect(funcall.name).toBe('set');
    
    const arrayArg = funcall.args[1];
    expect(arrayArg.type).toBe('json_array');
    expect(arrayArg.elements).toHaveLength(3);
    expect(arrayArg.elements[0].value).toBe('red');
    expect(arrayArg.elements[1].value).toBe('green');
    expect(arrayArg.elements[2].value).toBe('blue');
  });

  test('parses multiline JSON array without commas', () => {
    const program = parse(`set "items" [
  "first"
  "second"
  "third"
]`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    const arrayArg = funcall.args[1];
    expect(arrayArg.type).toBe('json_array');
    expect(arrayArg.elements).toHaveLength(3);
    expect(arrayArg.elements[0].value).toBe('first');
    expect(arrayArg.elements[1].value).toBe('second');
    expect(arrayArg.elements[2].value).toBe('third');
  });

  test('parses nested multiline JSON', () => {
    const program = parse(`set "user" {
  "name": "John",
  "address": {
    "street": "123 Main St",
    "city": "Anytown"
  },
  "hobbies": [
    "reading",
    "coding"
  ]
}`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    const jsonArg = funcall.args[1];
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(3);
    
    // Check nested object
    const addressPair = jsonArg.pairs[1];
    expect(addressPair.key).toBe('address');
    expect(addressPair.value.type).toBe('json_object');
    expect(addressPair.value.pairs).toHaveLength(2);
    
    // Check nested array
    const hobbiesPair = jsonArg.pairs[2];
    expect(hobbiesPair.key).toBe('hobbies');
    expect(hobbiesPair.value.type).toBe('json_array');
    expect(hobbiesPair.value.elements).toHaveLength(2);
  });

  test('parses multiline JSON with comments', () => {
    const program = parse(`show "card" {
  # Title property
  "title": "Test",
  # Content property
  "content": "With comments"
}`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    const jsonArg = funcall.args[1];
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(2);
    expect(jsonArg.pairs[0].key).toBe('title');
    expect(jsonArg.pairs[1].key).toBe('content');
  });

  test('parses mixed single line and multiline JSON', () => {
    const program = parse(`show "card" {"title": "Single", "props": {
  "nested": "multiline",
  "value": 42
}}`);
    expect(program.expressions).toHaveLength(1);
    
    const funcall = program.expressions[0] as any;
    const jsonArg = funcall.args[1];
    expect(jsonArg.type).toBe('json_object');
    expect(jsonArg.pairs).toHaveLength(2);
    
    // Check nested multiline object
    const propsPair = jsonArg.pairs[1];
    expect(propsPair.key).toBe('props');
    expect(propsPair.value.type).toBe('json_object');
    expect(propsPair.value.pairs).toHaveLength(2);
  });
}); 