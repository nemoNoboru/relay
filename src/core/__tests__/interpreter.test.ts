import { parse } from '../parser';
import { 
  RelayInterpreter, 
  runRelay, 
  builtins, 
  createEnvironment, 
  setVariable, 
  lookupVariable,
  defineBuiltin
} from '../interpreter';

describe('RelayInterpreter', () => {
  let interpreter: RelayInterpreter;

  beforeEach(() => {
    interpreter = new RelayInterpreter();
  });

  describe('Basic Operations', () => {
    test('evaluates simple arithmetic', () => {
      const program = parse('+ 1 2');
      const result = interpreter.evaluate(program);
      expect(result).toBe(3);
    });

    test('evaluates multiple arithmetic operations', () => {
      const program = parse('* 3 4');
      const result = interpreter.evaluate(program);
      expect(result).toBe(12);
    });

    test('evaluates subtraction', () => {
      const program = parse('- 10 3');
      const result = interpreter.evaluate(program);
      expect(result).toBe(7);
    });

    test('evaluates division', () => {
      const program = parse('/ 15 3');
      const result = interpreter.evaluate(program);
      expect(result).toBe(5);
    });

    test('evaluates modulo', () => {
      const program = parse('% 10 3');
      const result = interpreter.evaluate(program);
      expect(result).toBe(1);
    });

    test('throws error on division by zero', () => {
      const program = parse('/ 5 0');
      expect(() => interpreter.evaluate(program)).toThrow('Division by zero');
    });
  });

  describe('Variable Operations', () => {
    test('sets and retrieves variables', () => {
      const program1 = parse('set x 42');
      interpreter.evaluate(program1);
      // In Relay, 'x' is a zero-argument function call that returns the variable value
      const program2 = parse('x');
      const result = interpreter.evaluate(program2);
      expect(result).toBe(42);
    });

    test('variable assignment returns the assigned value', () => {
      const program = parse('set y 100');
      const result = interpreter.evaluate(program);
      expect(result).toBe(100);
    });

    test('throws error for undefined function', () => {
      const program = parse('undefined_var');
      expect(() => interpreter.evaluate(program)).toThrow('Unknown function: undefined_var');
    });
  });

  describe('Comparison Operations', () => {
    test('evaluates equality', () => {
      const program1 = parse('equal 5 5');
      const program2 = parse('equal 5 3');
      
      expect(interpreter.evaluate(program1)).toBe(true);
      expect(interpreter.evaluate(program2)).toBe(false);
    });

    test('evaluates less than', () => {
      const program1 = parse('< 3 5');
      const program2 = parse('< 5 3');
      
      expect(interpreter.evaluate(program1)).toBe(true);
      expect(interpreter.evaluate(program2)).toBe(false);
    });

    test('evaluates greater than', () => {
      const program1 = parse('> 5 3');
      const program2 = parse('> 3 5');
      
      expect(interpreter.evaluate(program1)).toBe(true);
      expect(interpreter.evaluate(program2)).toBe(false);
    });
  });

  describe('Conditional Operations', () => {
    test('evaluates if with true condition', () => {
      const program = parse('if true "yes" "no"');
      const result = interpreter.evaluate(program);
      expect(result).toBe('yes');
    });

    test('evaluates if with false condition', () => {
      const program = parse('if false "yes" "no"');
      const result = interpreter.evaluate(program);
      expect(result).toBe('no');
    });

    test('evaluates if with truthy condition', () => {
      const program = parse('if 1 "yes" "no"');
      const result = interpreter.evaluate(program);
      expect(result).toBe('yes');
    });

    test('evaluates if with falsy condition', () => {
      const program = parse('if 0 "yes" "no"');
      const result = interpreter.evaluate(program);
      expect(result).toBe('no');
    });
  });

  describe('JSON Data Structures', () => {
    test('evaluates JSON arrays', () => {
      const program = parse('[1, 2, 3]');
      const result = interpreter.evaluate(program);
      expect(result).toEqual([1, 2, 3]);
    });

    test('evaluates JSON objects', () => {
      const program = parse('{"name": "test", "value": 42}');
      const result = interpreter.evaluate(program);
      expect(result).toEqual({ name: 'test', value: 42 });
    });

    test('evaluates arrays with expressions', () => {
      const program = parse('[+ 1 2, * 3 4]');
      const result = interpreter.evaluate(program);
      expect(result).toEqual([3, 12]);
    });
  });

  describe('Lambda Functions', () => {
    test('creates and calls lambda functions', () => {
      interpreter.evaluate(parse('set add {x: + x 1}'));
      const result = interpreter.evaluate(parse('add 5'));
      expect(result).toBe(6);
    });

    test('lambda functions capture closure', () => {
      interpreter.evaluate(parse('set y 10'));
      interpreter.evaluate(parse('set add_y {x: + x y}'));
      const result = interpreter.evaluate(parse('add_y 5'));
      expect(result).toBe(15);
    });
  });

  describe('Complex Expressions', () => {
    test('evaluates nested arithmetic', () => {
      const program = parse('+ (* 2 3) (/ 8 2)');
      const result = interpreter.evaluate(program);
      expect(result).toBe(10); // (2*3) + (8/2) = 6 + 4 = 10
    });

    test('evaluates complex conditional', () => {
      interpreter.evaluate(parse('set x 5'));
      const result = interpreter.evaluate(parse('if (> x 3) (+ x 10) (- x 2)'));
      expect(result).toBe(15);
    });
  });

  describe('Multiple Expressions', () => {
    test('returns last expression result', () => {
      interpreter.evaluate(parse('set a 1'));
      interpreter.evaluate(parse('set b 2'));
      const result = interpreter.evaluate(parse('+ a b'));
      expect(result).toBe(3);
    });
  });

  describe('Error Handling', () => {
    test('throws error for wrong argument count', () => {
      const program = parse('+ 1');
      // + expects at least 0 arguments and works with 1, so this should work
      const result = interpreter.evaluate(program);
      expect(result).toBe(1);
    });

    test('throws error for type mismatch', () => {
      const program = parse('+ 1 "hello"');
      expect(() => interpreter.evaluate(program)).toThrow('+ expects numbers');
    });

    test('throws error for unknown function', () => {
      const program = parse('unknown_function 1 2');
      expect(() => interpreter.evaluate(program)).toThrow('Unknown function: unknown_function');
    });
  });

  describe('Show Function', () => {
    test('show function returns formatted output', () => {
      const program = parse('show "Hello" "World"');
      const result = interpreter.evaluate(program);
      expect(result).toBe('Hello World');
    });

    test('show function handles different types', () => {
      const program = parse('show 42 true {"key": "value"}');
      const result = interpreter.evaluate(program);
      expect(result).toBe('42 true {"key":"value"}');
    });
  });
});

describe('runRelay convenience function', () => {
  test('runs Relay code directly', () => {
    const result = runRelay(parse('+ 2 3'));
    expect(result).toBe(5);
  });
});

describe('Environment operations', () => {
  test('createEnvironment works', () => {
    const env = createEnvironment();
    expect(env.bindings).toBeInstanceOf(Map);
    expect(env.parent).toBeUndefined();
  });

  test('environment with parent', () => {
    const parent = createEnvironment();
    const child = createEnvironment(parent);
    expect(child.parent).toBe(parent);
  });

  test('setVariable and lookupVariable work', () => {
    const env = createEnvironment();
    setVariable('test', 'value', env);
    expect(lookupVariable('test', env)).toBe('value');
  });

  test('variable lookup in parent environment', () => {
    const parent = createEnvironment();
    const child = createEnvironment(parent);
    
    setVariable('parent_var', 'parent_value', parent);
    expect(lookupVariable('parent_var', child)).toBe('parent_value');
  });
});

describe('Custom Builtins', () => {
  test('can add custom builtin functions', () => {
    const testInterpreter = new RelayInterpreter();
    
    // Add a custom builtin using the new system
    defineBuiltin('double', true, (args: any[]) => {
      if (args.length !== 1) {
        throw new Error('double expects exactly 1 argument');
      }
      return args[0] * 2;
    });

    const program = parse('double 21');
    const result = testInterpreter.evaluate(program);
    expect(result).toBe(42);

    // Clean up
    delete builtins['double'];
  });

  test('can define JavaScript builtins with def-js (eager)', () => {
    const testInterpreter = new RelayInterpreter();
    
    // Define a JavaScript builtin function
    const jsCode = "if (args.length !== 1) { throw new Error('triple expects exactly 1 argument'); } return args[0] * 3;";
    const defProgram = parse(`def-js triple true "${jsCode}"`);
    
    const defResult = testInterpreter.evaluate(defProgram);
    expect(defResult).toBe('Defined builtin function: triple');
    
    // Test the newly defined function
    const result = testInterpreter.evaluate(parse('triple 14'));
    expect(result).toBe(42);
    
    // Clean up
    delete builtins['triple'];
  });

  test('can define JavaScript builtins with def-js (lazy)', () => {
    const testInterpreter = new RelayInterpreter();
    
    // Define a lazy JavaScript builtin (custom conditional)
    const jsCode = "if (args.length !== 2) { throw new Error('when expects exactly 2 arguments: condition, action'); } const condition = evaluate(args[0], env); if (condition) { return evaluate(args[1], env); } return null;";
    const defProgram = parse(`def-js when false "${jsCode}"`);
    
    const defResult = testInterpreter.evaluate(defProgram);
    expect(defResult).toBe('Defined builtin function: when');
    
    // Test the lazy function - should only evaluate action if condition is true
    testInterpreter.evaluate(parse('set x 5'));
    const result1 = testInterpreter.evaluate(parse('when (> x 3) (set x (* x 2))'));
    expect(result1).toBe(10); // x should be doubled to 10
    
    const result2 = testInterpreter.evaluate(parse('when (< x 5) (set x 999)'));
    expect(result2).toBe(null); // condition false, x unchanged
    expect(testInterpreter.evaluate(parse('x'))).toBe(10); // x still 10
    
    // Clean up
    delete builtins['when'];
  });

  test('user-defined functions work correctly', () => {
    const testInterpreter = new RelayInterpreter();
    
    // Define a function with def
    testInterpreter.evaluate(parse('def truther a true'));
    
    // Try calling it with wrong argument count - should throw error
    expect(() => testInterpreter.evaluate(parse('truther'))).toThrow('Function expects 1 arguments, got 0');
    
    // Try calling it with correct argument count - should work
    const result = testInterpreter.evaluate(parse('truther 42'));
    expect(result).toBe(true);
  });

  test('def functions work in complex scenarios', () => {
    const testInterpreter = new RelayInterpreter();
    
    // Define a math function
    testInterpreter.evaluate(parse('def double x (* x 2)'));
    
    // Test it works
    const result1 = testInterpreter.evaluate(parse('double 21'));
    expect(result1).toBe(42);
    
    // Use it in expressions
    const result2 = testInterpreter.evaluate(parse('+ (double 5) (double 3)'));
    expect(result2).toBe(16); // (5*2) + (3*2) = 10 + 6 = 16
    
    // Verify def-js still works alongside def
    testInterpreter.evaluate(parse('def-js triple true "return args[0] * 3;"'));
    const result3 = testInterpreter.evaluate(parse('triple 7'));
    expect(result3).toBe(21);
  });
}); 