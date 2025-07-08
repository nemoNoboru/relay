// Generated automatically by nearley, version 2.20.1
// http://github.com/Hardmath123/nearley
import * as moo from 'moo';

const grammar = (function () {
function id(x) { return x[0]; }

/**
 * Relay Language - Complete Implementation
 * "Basically a lisp with python indents and json"
 */

// Custom indentation-aware lexer for Python-style blocks
class IndentationLexer {
  constructor() {
    this.indentStack = [0]; // Stack of indentation levels
    this.tokens = [];
    this.position = 0;
    
    // Base Moo lexer for individual tokens
    this.baseLexer = moo.compile({
      // Whitespace (not newlines) - handled specially for indentation
      WS: /[ \t]+/,
      
      // Comments - everything from # to end of line  
      comment: {
        match: /#[^\r\n]*/,
        value: (s) => s
      },
      
      // Newlines with line break tracking
      newline: {
        match: /\r?\n/,
        lineBreaks: true,
        value: () => '\n'
      },
  
  // Numbers - integers and floats, including negative
  number: {
    match: /-?(?:0|[1-9]\d*)(?:\.\d+)?/,
    value: (s) => s.includes('.') ? parseFloat(s) : parseInt(s, 10)
  },
  
  // Strings - double quoted with escape sequences
  string: {
    match: /"(?:\\["\\ntr]|[^\r\n"\\])*"/,
    value: (s) => {
      // Remove quotes and process escape sequences
      const content = s.slice(1, -1);
      return content.replace(/\\(.)/g, (match, char) => {
        switch (char) {
          case 'n': return '\n';
          case 't': return '\t';
          case 'r': return '\r';
          case '\\': return '\\';
          case '"': return '"';
          default: return char;
        }
      });
    }
  },
  
  // Keywords and identifiers
  identifier: {
    match: /[a-zA-Z_][a-zA-Z0-9_\-?]*/,
    type: moo.keywords({
      'true': 'true',
      'false': 'false', 
      'null': 'null'
    })
  },
  
  // Operators
  operator: /[+\-*\/%<>=!]/,
  
      // Single character tokens
      lparen: '(',
      rparen: ')',
      lbracket: '[',
      rbracket: ']',
      lbrace: '{',
      rbrace: '}',
      colon: ':',
      comma: ',',
    });
  }
  
  reset(text) {
    this.text = text;
    this.tokens = this.tokenize(text);
    this.position = 0;
    this.indentStack = [0]; // Reset indent stack
    return this;
  }
  
  next() {
    if (this.position >= this.tokens.length) {
      return undefined;
    }
    return this.tokens[this.position++];
  }
  
  save() {
    return this.position;
  }
  
  formatError(token) {
    return `Unexpected ${token.type} token: "${token.value}".`;
  }
  
  has(type) {
    return true; // Accept all token types
  }
  
  tokenize(text) {
    const lines = text.split('\n');
    const tokens = [];
    this.indentStack = [0]; // Reset for fresh tokenization
    
    for (let lineNum = 0; lineNum < lines.length; lineNum++) {
      const line = lines[lineNum];
      
      // Skip empty lines  
      if (!line.trim()) {
        if (line.includes('\n') || lineNum < lines.length - 1) {
          tokens.push({ type: 'newline', value: '\n', line: lineNum + 1, col: 1 });
        }
        continue;
      }
      
      // Calculate indentation level (spaces only)
      const indentMatch = line.match(/^[ ]*/);
      const indentLevel = indentMatch ? indentMatch[0].length : 0;
      const content = line.trim();
      
      // Only handle indentation for lines that start with identifiers (Relay block syntax)
      // Skip indentation handling for JSON content (lines starting with quotes, braces, etc.)
      const isRelayBlockLine = /^[a-zA-Z_][a-zA-Z0-9_\-?]*/.test(content);
      
      if (isRelayBlockLine) {
        // Handle indentation changes for Relay block syntax
        const currentIndent = this.indentStack[this.indentStack.length - 1];
        
        if (indentLevel > currentIndent) {
          // Increase indentation - add INDENT token
          this.indentStack.push(indentLevel);
          tokens.push({ type: 'INDENT', value: '', line: lineNum + 1, col: 1 });
        } else if (indentLevel < currentIndent) {
          // Decrease indentation - add DEDENT token(s)
          while (this.indentStack.length > 1 && this.indentStack[this.indentStack.length - 1] > indentLevel) {
            this.indentStack.pop();
            tokens.push({ type: 'DEDENT', value: '', line: lineNum + 1, col: 1 });
          }
        }
      }
      
      // Tokenize the line content
      if (content) {
        this.baseLexer.reset(content);
        let token;
        while ((token = this.baseLexer.next())) {
          if (token.type !== 'WS') { // Skip whitespace within lines
            tokens.push({
              ...token,
              line: lineNum + 1,
              col: token.col
            });
          }
        }
        
        // Add newline token
        tokens.push({ type: 'newline', value: '\n', line: lineNum + 1, col: content.length + 1 });
      }
    }
    
    // Add final DEDENT tokens to close all indentation levels
    while (this.indentStack.length > 1) {
      this.indentStack.pop();
      tokens.push({ type: 'DEDENT', value: '', line: lines.length, col: 1 });
    }
    
    return tokens;
  }
}

// Create lexer instance
const lexer = new IndentationLexer();

// Everything is an expression that returns a value
function expr(type, data) {
  return { type, ...data };
}
return {
    Lexer: lexer,
    ParserRules: [
    {"name": "main", "symbols": ["_", "program", "_"], "postprocess": d => d[1]},
    {"name": "program$ebnf$1", "symbols": []},
    {"name": "program$ebnf$1", "symbols": ["program$ebnf$1", "statement"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "program", "symbols": ["program$ebnf$1"], "postprocess": d => expr('program', { expressions: d[0].filter(e => e) })},
    {"name": "statement", "symbols": ["block_statement"], "postprocess": id},
    {"name": "statement", "symbols": ["simple_statement"], "postprocess": id},
    {"name": "statement", "symbols": [(lexer.has("comment") ? {type: "comment"} : comment)], "postprocess": () => null},
    {"name": "statement", "symbols": [(lexer.has("newline") ? {type: "newline"} : newline)], "postprocess": () => null},
    {"name": "block_statement$ebnf$1", "symbols": ["arg_list"], "postprocess": id},
    {"name": "block_statement$ebnf$1", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "block_statement", "symbols": ["name", "block_statement$ebnf$1", "_", (lexer.has("newline") ? {type: "newline"} : newline), (lexer.has("INDENT") ? {type: "INDENT"} : INDENT), "block_body", (lexer.has("DEDENT") ? {type: "DEDENT"} : DEDENT)], "postprocess": d => expr('funcall', { name: d[0], args: d[1] || [], children: d[5] })},
    {"name": "simple_statement", "symbols": ["name", "arg_list", "_", (lexer.has("newline") ? {type: "newline"} : newline)], "postprocess": d => expr('funcall', { name: d[0], args: d[1] })},
    {"name": "simple_statement", "symbols": ["name", "_", (lexer.has("newline") ? {type: "newline"} : newline)], "postprocess": d => expr('identifier', { name: d[0] })},
    {"name": "block_body$ebnf$1", "symbols": ["statement"]},
    {"name": "block_body$ebnf$1", "symbols": ["block_body$ebnf$1", "statement"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "block_body", "symbols": ["block_body$ebnf$1"], "postprocess": d => d[0].filter(s => s)},
    {"name": "_$ebnf$1", "symbols": []},
    {"name": "_$ebnf$1$subexpression$1", "symbols": [(lexer.has("newline") ? {type: "newline"} : newline)]},
    {"name": "_$ebnf$1$subexpression$1", "symbols": [(lexer.has("comment") ? {type: "comment"} : comment)]},
    {"name": "_$ebnf$1", "symbols": ["_$ebnf$1", "_$ebnf$1$subexpression$1"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "_", "symbols": ["_$ebnf$1"], "postprocess": () => null},
    {"name": "arg_list$ebnf$1", "symbols": ["arg"]},
    {"name": "arg_list$ebnf$1", "symbols": ["arg_list$ebnf$1", "arg"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "arg_list", "symbols": ["arg_list$ebnf$1"], "postprocess": d => d[0]},
    {"name": "nested_funcall", "symbols": ["name", "arg_list"], "postprocess": d => expr('funcall', { name: d[0], args: d[1] })},
    {"name": "nested_funcall", "symbols": ["name"], "postprocess": d => expr('funcall', { name: d[0], args: [] })},
    {"name": "name", "symbols": [(lexer.has("identifier") ? {type: "identifier"} : identifier)], "postprocess": d => d[0].value},
    {"name": "name", "symbols": [(lexer.has("operator") ? {type: "operator"} : operator)], "postprocess": d => d[0].value},
    {"name": "arg", "symbols": ["atom"], "postprocess": id},
    {"name": "arg", "symbols": ["parenthesized"], "postprocess": id},
    {"name": "parenthesized", "symbols": [(lexer.has("lparen") ? {type: "lparen"} : lparen), "_", "expression", "_", (lexer.has("rparen") ? {type: "rparen"} : rparen)], "postprocess": d => d[2]},
    {"name": "expression", "symbols": ["nested_funcall"], "postprocess": id},
    {"name": "expression", "symbols": ["parenthesized"], "postprocess": id},
    {"name": "expression", "symbols": ["atom"], "postprocess": id},
    {"name": "atom", "symbols": [(lexer.has("string") ? {type: "string"} : string)], "postprocess": d => expr('atom', { value: d[0].value })},
    {"name": "atom", "symbols": [(lexer.has("number") ? {type: "number"} : number)], "postprocess": d => expr('atom', { value: d[0].value })},
    {"name": "atom", "symbols": [{"literal":"true"}], "postprocess": d => expr('atom', { value: true })},
    {"name": "atom", "symbols": [{"literal":"false"}], "postprocess": d => expr('atom', { value: false })},
    {"name": "atom", "symbols": [{"literal":"null"}], "postprocess": d => expr('atom', { value: null })},
    {"name": "atom", "symbols": [(lexer.has("identifier") ? {type: "identifier"} : identifier)], "postprocess": d => expr('identifier', { name: d[0].value })},
    {"name": "atom", "symbols": [(lexer.has("operator") ? {type: "operator"} : operator)], "postprocess": d => expr('identifier', { name: d[0].value })},
    {"name": "atom", "symbols": ["json_array"], "postprocess": id},
    {"name": "atom", "symbols": ["json_object"], "postprocess": id},
    {"name": "json_object", "symbols": [(lexer.has("lbrace") ? {type: "lbrace"} : lbrace), "_", "json_pairs", "_", (lexer.has("rbrace") ? {type: "rbrace"} : rbrace)], "postprocess": d => ({ type: 'json_object', pairs: d[2] })},
    {"name": "json_object", "symbols": [(lexer.has("lbrace") ? {type: "lbrace"} : lbrace), "_", (lexer.has("rbrace") ? {type: "rbrace"} : rbrace)], "postprocess": () => ({ type: 'json_object', pairs: [] })},
    {"name": "json_pairs$ebnf$1", "symbols": []},
    {"name": "json_pairs$ebnf$1$subexpression$1", "symbols": ["json_separator", "json_pair"]},
    {"name": "json_pairs$ebnf$1", "symbols": ["json_pairs$ebnf$1", "json_pairs$ebnf$1$subexpression$1"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "json_pairs", "symbols": ["json_pair", "json_pairs$ebnf$1"], "postprocess": d => [d[0], ...d[1].map(a => a[1])]},
    {"name": "json_separator", "symbols": ["_", (lexer.has("comma") ? {type: "comma"} : comma), "_"]},
    {"name": "json_separator", "symbols": ["_", (lexer.has("newline") ? {type: "newline"} : newline), "_"]},
    {"name": "json_pair", "symbols": [(lexer.has("string") ? {type: "string"} : string), "_", (lexer.has("colon") ? {type: "colon"} : colon), "_", "expression"], "postprocess": d => ({ key: d[0].value, value: d[4] })},
    {"name": "json_array", "symbols": [(lexer.has("lbracket") ? {type: "lbracket"} : lbracket), "_", "json_list", "_", (lexer.has("rbracket") ? {type: "rbracket"} : rbracket)], "postprocess": d => ({ type: 'json_array', elements: d[2] })},
    {"name": "json_array", "symbols": [(lexer.has("lbracket") ? {type: "lbracket"} : lbracket), "_", (lexer.has("rbracket") ? {type: "rbracket"} : rbracket)], "postprocess": () => ({ type: 'json_array', elements: [] })},
    {"name": "json_list$ebnf$1", "symbols": []},
    {"name": "json_list$ebnf$1$subexpression$1", "symbols": ["json_separator", "expression"]},
    {"name": "json_list$ebnf$1", "symbols": ["json_list$ebnf$1", "json_list$ebnf$1$subexpression$1"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "json_list", "symbols": ["expression", "json_list$ebnf$1"], "postprocess": d => [d[0], ...d[1].map(a => a[1])]}
]
  , ParserStart: "main"
};
})();

export default grammar;
