import { RelayInterpreter } from '../interpreter';
import { parse } from '../parser';

describe('Built-in Functions', () => {
  let interpreter: RelayInterpreter;

  beforeEach(() => {
    interpreter = new RelayInterpreter();
  });

  describe('get function', () => {
    test('gets object property', () => {
      const program = parse('get {"name": "Alice", "age": 25} "name"');
      const result = interpreter.evaluate(program);
      expect(result).toBe('Alice');
    });

    test('gets numeric property', () => {
      const program = parse('get {"name": "Alice", "age": 25} "age"');
      const result = interpreter.evaluate(program);
      expect(result).toBe(25);
    });

    test('throws error for non-object', () => {
      expect(() => {
        const program = parse('get "not an object" "key"');
        interpreter.evaluate(program);
      }).toThrow('get expects first argument to be an object');
    });

    test('accepts numeric keys for array access', () => {
      const program = parse('get ["Alice", "Bob"] 0');
      const result = interpreter.evaluate(program);
      expect(result).toBe('Alice');
    });
  });

  describe('for function', () => {
    test('creates components from string list', () => {
      const program = parse(`set names ["Alice", "Bob"]
def name_card name (show "card" {"title": name})
for names name_card`);
      const result = interpreter.evaluate(program);
      
      expect(result.type).toBe('component_collection');
      expect(result.components).toHaveLength(2);
      expect(result.components[0]).toEqual({
        type: 'component',
        name: 'card',
        props: { title: 'Alice' },
        children: []
      });
      expect(result.components[1]).toEqual({
        type: 'component',
        name: 'card',
        props: { title: 'Bob' },
        children: []
      });
    });

    test('creates components from object list using get', () => {
      const program = parse(`set users [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}]
def user_card user (show "card" {"title": (get user "name"), "content": (get user "age")})
for users user_card`);
      const result = interpreter.evaluate(program);
      
      expect(result.type).toBe('component_collection');
      expect(result.components).toHaveLength(2);
      expect(result.components[0]).toEqual({
        type: 'component',
        name: 'card',
        props: { title: 'Alice', content: 25 },
        children: []
      });
      expect(result.components[1]).toEqual({
        type: 'component',
        name: 'card',
        props: { title: 'Bob', content: 30 },
        children: []
      });
    });

    test('handles single parameter functions', () => {
      const program = parse(`set items ["A", "B"]
def simple_card item (show "card" {"title": item})
for items simple_card`);
      const result = interpreter.evaluate(program);
      
      expect(result.type).toBe('component_collection');
      expect(result.components).toHaveLength(2);
      expect(result.components[0]).toEqual({
        type: 'component',
        name: 'card',
        props: { title: 'A' },
        children: []
      });
      expect(result.components[1]).toEqual({
        type: 'component',
        name: 'card',
        props: { title: 'B' },
        children: []
      });
    });

    test('throws error for non-array', () => {
      expect(() => {
        const program = parse(`def card_maker item (show "card" {"title": item})
for "not an array" card_maker`);
        interpreter.evaluate(program);
      }).toThrow('for expects first argument to be a list/array');
    });

    test('throws error for non-function', () => {
      expect(() => {
        const program = parse('for ["a", "b"] "not a function"');
        interpreter.evaluate(program);
      }).toThrow('for expects second argument to be a function');
    });
  });

  describe('integration test', () => {
    test('for and get work together in show block syntax', () => {
      const program = parse(`set products [{"name": "Laptop", "price": 999}, {"name": "Phone", "price": 599}]
def product_card product (show "card" {
  "title": (get product "name"),
  "content": (concat "Price: $" (get product "price"))
})
show "grid"
  for products product_card`);
      
      const result = interpreter.evaluate(program);
      
      // The result should be a component collection
      expect(result).toEqual({
        type: 'component_collection',
        components: expect.arrayContaining([
          expect.objectContaining({
            type: 'component',
            name: 'grid',
            children: expect.arrayContaining([
              expect.objectContaining({
                type: 'component',
                name: 'card',
                props: { title: 'Laptop', content: 'Price: $999' }
              }),
              expect.objectContaining({
                type: 'component',
                name: 'card',
                props: { title: 'Phone', content: 'Price: $599' }
              })
            ])
          })
        ]),
        eventHandlers: expect.any(Map)
      });
    });
  });
}); 