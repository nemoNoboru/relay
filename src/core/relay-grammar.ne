@{%
/**
 * Relay Language - Complete Implementation
 * "Basically a lisp with python indents and json"
 */

const moo = require('moo');

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
%}

# Use the Moo lexer
@lexer lexer

# Relay = Lisp + Python indents + JSON

main -> _ program _ {% d => d[1] %}

program -> statement:* {% d => expr('program', { expressions: d[0].filter(e => e) }) %}

# Statements are processed in order of precedence - first match wins
statement ->
    block_statement {% id %}
  | simple_statement {% id %}
  | %comment {% () => null %}
  | %newline {% () => null %}

# Block statement: function call followed by indented children
block_statement -> 
    name arg_list:? _ %newline %INDENT block_body %DEDENT 
    {% d => expr('funcall', { name: d[0], args: d[1] || [], children: d[5] }) %}

# Simple statement: function call or standalone identifier (no indentation)  
simple_statement ->
    name arg_list _ %newline
    {% d => expr('funcall', { name: d[0], args: d[1] }) %}
  | name _ %newline  
    {% d => expr('identifier', { name: d[0] }) %}

# Block body contains child statements
block_body -> statement:+ {% d => d[0].filter(s => s) %}

# Optional whitespace (including newlines for JSON)
_ -> (%newline | %comment):* {% () => null %}

# Argument list - one or more arguments
arg_list -> arg:+ {% d => d[0] %}

# For nested calls (inside parentheses)
nested_funcall -> name arg_list {% d => expr('funcall', { name: d[0], args: d[1] }) %}
             | name {% d => expr('funcall', { name: d[0], args: [] }) %}

name -> %identifier {% d => d[0].value %} | %operator {% d => d[0].value %}

arg -> 
    atom {% id %}
  | parenthesized {% id %}

parenthesized -> %lparen _ expression _ %rparen {% d => d[2] %}

expression ->
    nested_funcall {% id %}
  | parenthesized {% id %}
  | atom {% id %}

# Atoms (JSON + identifiers + operators)
atom ->
    %string {% d => expr('atom', { value: d[0].value }) %}
  | %number {% d => expr('atom', { value: d[0].value }) %}
  | "true" {% d => expr('atom', { value: true }) %}
  | "false" {% d => expr('atom', { value: false }) %}
  | "null" {% d => expr('atom', { value: null }) %}
  | %identifier {% d => expr('identifier', { name: d[0].value }) %}
  | %operator {% d => expr('identifier', { name: d[0].value }) %}
  | json_array {% id %}
  | json_object {% id %}

# JSON Objects
json_object -> %lbrace _ json_pairs _ %rbrace {% d => ({ type: 'json_object', pairs: d[2] }) %}
           | %lbrace _ %rbrace {% () => ({ type: 'json_object', pairs: [] }) %}

json_pairs -> json_pair (json_separator json_pair):* {% d => [d[0], ...d[1].map(a => a[1])] %}

json_separator -> _ %comma _ | _ %newline _

json_pair -> %string _ %colon _ expression {% d => ({ key: d[0].value, value: d[4] }) %}

# JSON Arrays
json_array -> %lbracket _ json_list _ %rbracket {% d => ({ type: 'json_array', elements: d[2] }) %}
           | %lbracket _ %rbracket {% () => ({ type: 'json_array', elements: [] }) %}

json_list -> expression (json_separator expression):* {% d => [d[0], ...d[1].map(a => a[1])] %} 