import { parse } from '../parser';
import { RelayInterpreter } from '../interpreter';

describe('Block Syntax', () => {
  test('handles simple block syntax', () => {
    const code = `
show "row"
  show "card" {"title": "Card 1", "content": "Left"}
  show "card" {"title": "Card 2", "content": "Right"}
`;
    
    const ast = parse(code);
    const interpreter = new RelayInterpreter();
    const result = interpreter.evaluate(ast);
    
    expect(result).toHaveProperty('type', 'component_collection');
    expect(result.components).toHaveLength(1);
    
    const rowComponent = result.components[0];
    expect(rowComponent.name).toBe('row');
    expect(rowComponent.children).toHaveLength(2);
    
    const firstCard = rowComponent.children[0];
    expect(firstCard.name).toBe('card');
    expect(firstCard.props.title).toBe('Card 1');
    expect(firstCard.props.content).toBe('Left');
    
    const secondCard = rowComponent.children[1];
    expect(secondCard.name).toBe('card');
    expect(secondCard.props.title).toBe('Card 2');
    expect(secondCard.props.content).toBe('Right');
  });

  test('handles nested block syntax', () => {
    const code = `
show "container"
  show "row"
    show "card" {"title": "Card 1", "content": "Left"}
    show "card" {"title": "Card 2", "content": "Right"}
`;
    
    const ast = parse(code);
    const interpreter = new RelayInterpreter();
    const result = interpreter.evaluate(ast);
    
    expect(result).toHaveProperty('type', 'component_collection');
    expect(result.components).toHaveLength(1);
    
    const containerComponent = result.components[0];
    expect(containerComponent.name).toBe('container');
    expect(containerComponent.children).toHaveLength(1);
    
    const rowComponent = containerComponent.children[0];
    expect(rowComponent.name).toBe('row');
    expect(rowComponent.children).toHaveLength(2);
  });

  test('handles column layout', () => {
    const code = `
show "col"
  show "heading" "Column Test"
  show "paragraph" "Paragraph in column"
  show "button" "Button in column"
`;
    
    const ast = parse(code);
    const interpreter = new RelayInterpreter();
    const result = interpreter.evaluate(ast);
    
    expect(result).toHaveProperty('type', 'component_collection');
    expect(result.components).toHaveLength(1);
    
    const colComponent = result.components[0];
    expect(colComponent.name).toBe('col');
    expect(colComponent.children).toHaveLength(3);
    
    expect(colComponent.children[0].name).toBe('heading');
    expect(colComponent.children[1].name).toBe('paragraph');
    expect(colComponent.children[2].name).toBe('button');
  });

  test('handles grid layout', () => {
    const code = `
show "grid"
  show "card" {"title": "Grid 1", "content": "Item 1"}
  show "card" {"title": "Grid 2", "content": "Item 2"}
  show "card" {"title": "Grid 3", "content": "Item 3"}
`;
    
    const ast = parse(code);
    const interpreter = new RelayInterpreter();
    const result = interpreter.evaluate(ast);
    
    expect(result).toHaveProperty('type', 'component_collection');
    expect(result.components).toHaveLength(1);
    
    const gridComponent = result.components[0];
    expect(gridComponent.name).toBe('grid');
    expect(gridComponent.children).toHaveLength(3);
    
    gridComponent.children.forEach((child, index) => {
      expect(child.name).toBe('card');
      expect(child.props.title).toBe(`Grid ${index + 1}`);
      expect(child.props.content).toBe(`Item ${index + 1}`);
    });
  });

  test('handles mixed inline and block syntax', () => {
    const code = `
show "heading" "Mixed Syntax Test"
show "row"
  show "card" {"title": "Card 1", "content": "Block child"}
  show "card" {"title": "Card 2", "content": "Another block child"}
show "paragraph" "This is inline"
`;
    
    const ast = parse(code);
    const interpreter = new RelayInterpreter();
    const result = interpreter.evaluate(ast);
    
    expect(result).toHaveProperty('type', 'component_collection');
    expect(result.components).toHaveLength(3);
    
    expect(result.components[0].name).toBe('heading');
    expect(result.components[1].name).toBe('row');
    expect(result.components[1].children).toHaveLength(2);
    expect(result.components[2].name).toBe('paragraph');
  });
}); 