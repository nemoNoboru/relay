/**
 * Relay Language Parser - Nearley.js + Moo Implementation
 * 
 * Simple approach using Nearley's built-in Moo integration:
 * - Everything is an expression that returns a value (Lisp-like)
 * - Standard JSON literals as first-class expressions
 * - Uses @lexer directive with Moo for automatic tokenization
 */

import * as nearley from 'nearley';
import grammar from './relay-grammar.js';

// AST Node types
export interface ASTNode {
  type: string;
}

export interface ProgramNode extends ASTNode {
  type: 'program';
  expressions: ExpressionNode[];
}

export interface FuncallNode extends ASTNode {
  type: 'funcall';
  name: string;
  args: ExpressionNode[];
  children?: ExpressionNode[];  // Optional children for block syntax
}

export interface AtomNode extends ASTNode {
  type: 'atom';
  value: any;
}

export interface SequenceNode extends ASTNode {
  type: 'sequence';
  expressions: ExpressionNode[];
}

export interface JsonArrayNode extends ASTNode {
  type: 'json_array';
  elements: ExpressionNode[];
}

export interface JsonObjectNode extends ASTNode {
  type: 'json_object';
  pairs: { key: string; value: ExpressionNode }[];
}

export interface IdentifierNode extends ASTNode {
  type: 'identifier';
  name: string;
}

export type ExpressionNode = 
  | FuncallNode 
  | AtomNode 
  | SequenceNode 
  | JsonArrayNode 
  | JsonObjectNode 
  | IdentifierNode;

export class RelayParser {
  constructor() {}

  parseProgram(source: string): ProgramNode {
    // Create nearley parser with compiled grammar
    const parser = new nearley.Parser(nearley.Grammar.fromCompiled(grammar));
    
    try {
      // Feed source code directly - Nearley handles tokenization with Moo
      parser.feed(source);
      
      if (parser.results.length === 0) {
        throw new Error('No parse results');
      }
      
      if (parser.results.length > 1) {
        // Handle ambiguous parse results
        console.warn('Ambiguous parse - multiple results found, using first');
      }
      
      return parser.results[0];
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(`Parse error: ${message}`);
    }
  }
}

// Simple convenience functions
export function parse(source: string): ProgramNode {
  const parser = new RelayParser();
  return parser.parseProgram(source);
}

export interface RelayAST {
  type: 'program';
  expressions: ExpressionNode[];
}

export interface RelayStatement {
  type: string;
  [key: string]: any;
} 