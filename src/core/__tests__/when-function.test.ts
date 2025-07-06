import { parse } from '../parser'
import { RelayInterpreter } from '../interpreter'

describe('When Function and Event Handling', () => {
  let interpreter: RelayInterpreter

  beforeEach(() => {
    interpreter = new RelayInterpreter()
  })

  test('should register when event handler', () => {
    const code = `
    when "button-click" data
      show "paragraph" "Button clicked!"
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    // Should not throw an error
    expect(result).toBeDefined()
  })

  test('should handle state management', () => {
    const code = `
    state books (list [
      {"title": "Book 1", "author": "Author 1"}
    ])
    
    books
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    expect(result).toEqual([
      {"title": "Book 1", "author": "Author 1"}
    ])
  })

  test('should update state with set', () => {
    const code = `
    state counter 0
    set counter 5
    counter
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    expect(result).toBe(5)
  })

  test('should create lists with list function', () => {
    const code = `
    list ["item1", "item2", "item3"]
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    expect(result).toEqual(["item1", "item2", "item3"])
  })

  test('should add items to lists', () => {
    const code = `
    state items (list ["a", "b"])
    set items (add items "c" "d")
    items
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    expect(result).toEqual(["a", "b", "c", "d"])
  })

  test('should handle form creation', () => {
    const code = `
    form "contact-form"
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    // Should create a form component as part of component collection
    expect(result).toBeDefined()
    expect(result.type).toBe('component_collection')
    expect(result.components).toHaveLength(1)
    expect(result.components[0].type).toBe('component')
    expect(result.components[0].name).toBe('form')
  })

  test('should handle complex state operations', () => {
    const code = `
    state books (list [])
    set books (add books {"title": "Book 1", "author": "Author 1"})
    set books (add books {"title": "Book 2", "author": "Author 2"})
    get (get books 0) "title"
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    expect(result).toBe("Book 1")
  })

  test('should handle event handlers with data parameters', () => {
    const code = `
    when "form-submit" formData
      show "paragraph" "Form submitted"
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    // Should not throw an error and return result with handlers
    expect(result).toBeDefined()
    expect(result.type).toBe('result_with_handlers')
    expect(result.eventHandlers).toBeDefined()
    expect(result.eventHandlers.has('form-submit')).toBe(true)
    
    const handler = result.eventHandlers.get('form-submit')
    expect(handler.eventName).toBe('form-submit')
    expect(handler.dataParamName).toBe('formData')
  })

  test('should handle multiple event handlers', () => {
    const code = `
    when "click-1" data
      show "paragraph" "First click"
    
    when "click-2" data
      show "paragraph" "Second click"
    `
    
    const ast = parse(code)
    const result = interpreter.evaluate(ast)
    
    // Should not throw an error
    expect(result).toBeDefined()
  })
}) 