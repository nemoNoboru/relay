// Relay Parser - Based on concise grammar
// main -> expression+
// expression -> funcall | atom | lambda | indented_block
// funcall -> "("? name arg* ")"? | name indented_block | name arg+

class RelayParser {
  constructor(tokens) {
    this.tokens = tokens;
    this.pos = 0;
  }

  // main -> expression+
  parseMain() {
    const expressions = [];
    while (!this.isEOF()) {
      if (this.isComment() || this.isNewline()) {
        this.advance();
        continue;
      }
      expressions.push(this.parseExpression());
    }
    return { type: 'main', expressions };
  }

  // expression -> funcall | atom | lambda | indented_block
  parseExpression() {
    if (this.isIndent()) {
      return this.parseIndentedBlock();
    }
    
    if (this.isLambdaStart()) {
      return this.parseLambda();
    }
    
    if (this.isAtom()) {
      return this.parseAtom();
    }
    
    // Default to function call
    return this.parseFuncall();
  }

  // funcall -> "("? name arg* ")"? | name indented_block | name arg+
  parseFuncall() {
    const hasOpenParen = this.consume('(', false);
    const name = this.parseIdentifier();
    
    // Check for indented block form: name indented_block
    if (this.isNewlineAndIndent()) {
      const block = this.parseIndentedBlock();
      return { type: 'funcall', name, args: [block] };
    }
    
    // Parse arguments: arg*
    const args = [];
    while (this.hasMoreArgs() && !this.check(')')) {
      args.push(this.parseExpression());
    }
    
    // Optional closing paren
    if (hasOpenParen) {
      this.consume(')');
    }
    
    return { type: 'funcall', name, args };
  }

  // atom -> string | number | boolean | null | json_array | json_object | identifier
  parseAtom() {
    if (this.check('STRING')) {
      return { type: 'string', value: this.advance().value };
    }
    if (this.check('NUMBER')) {
      return { type: 'number', value: this.advance().value };
    }
    if (this.check('true') || this.check('false')) {
      return { type: 'boolean', value: this.advance().value === 'true' };
    }
    if (this.check('null')) {
      this.advance();
      return { type: 'null', value: null };
    }
    if (this.check('[')) {
      return this.parseJsonArray();
    }
    if (this.check('{')) {
      return this.parseJsonObject();
    }
    if (this.check('IDENTIFIER')) {
      return { type: 'identifier', name: this.advance().value };
    }
    
    throw new Error(`Unexpected token: ${this.currentToken()}`);
  }

  // lambda -> "{" params ":" expression "}" | params ":" expression | params ":" indented_block
  parseLambda() {
    // Brace form: {params: body}
    if (this.check('{')) {
      this.advance(); // consume '{'
      const params = this.parseParams();
      this.consume(':');
      const body = this.parseExpression();
      this.consume('}');
      return { type: 'lambda', params, body };
    }
    
    // Block form: params: body (in indented context)
    const params = this.parseParams();
    this.consume(':');
    
    let body;
    if (this.isNewlineAndIndent()) {
      body = this.parseIndentedBlock();
    } else {
      body = this.parseExpression();
    }
    
    return { type: 'lambda', params, body };
  }

  // params -> identifier ("," identifier)*
  parseParams() {
    const params = [this.parseIdentifier()];
    while (this.check(',')) {
      this.advance(); // consume ','
      params.push(this.parseIdentifier());
    }
    return params;
  }

  // indented_block -> INDENT expression+ DEDENT
  parseIndentedBlock() {
    this.consume('NEWLINE');
    this.consume('INDENT');
    
    const expressions = [];
    while (!this.check('DEDENT')) {
      if (this.isComment() || this.isNewline()) {
        this.advance();
        continue;
      }
      expressions.push(this.parseExpression());
    }
    
    this.consume('DEDENT');
    return { type: 'indented_block', expressions };
  }

  // Helper methods
  parseIdentifier() {
    if (!this.check('IDENTIFIER')) {
      throw new Error('Expected identifier');
    }
    return this.advance().value;
  }

  parseJsonArray() {
    this.consume('[');
    const elements = [];
    
    if (!this.check(']')) {
      elements.push(this.parseExpression());
      while (this.check(',')) {
        this.advance();
        elements.push(this.parseExpression());
      }
    }
    
    this.consume(']');
    return { type: 'json_array', elements };
  }

  parseJsonObject() {
    this.consume('{');
    const pairs = [];
    
    if (!this.check('}')) {
      pairs.push(this.parseJsonPair());
      while (this.check(',')) {
        this.advance();
        pairs.push(this.parseJsonPair());
      }
    }
    
    this.consume('}');
    return { type: 'json_object', pairs };
  }

  parseJsonPair() {
    const key = this.check('STRING') || this.check('IDENTIFIER') 
      ? this.advance().value 
      : (() => { throw new Error('Expected string or identifier for JSON key'); })();
    
    this.consume(':');
    const value = this.parseExpression();
    
    return { key, value };
  }

  // Utility methods
  currentToken() {
    return this.tokens[this.pos];
  }

  check(expected) {
    const token = this.currentToken();
    return token && (token.type === expected || token.value === expected);
  }

  advance() {
    return this.tokens[this.pos++];
  }

  consume(expected, required = true) {
    if (this.check(expected)) {
      this.advance();
      return true;
    }
    if (required) {
      throw new Error(`Expected ${expected}, got ${this.currentToken()?.type || 'EOF'}`);
    }
    return false;
  }

  isEOF() {
    return this.pos >= this.tokens.length;
  }

  isComment() {
    return this.check('COMMENT');
  }

  isNewline() {
    return this.check('NEWLINE');
  }

  isIndent() {
    return this.check('INDENT');
  }

  isNewlineAndIndent() {
    return this.check('NEWLINE') && 
           this.pos + 1 < this.tokens.length && 
           this.tokens[this.pos + 1].type === 'INDENT';
  }

  isLambdaStart() {
    // Check for brace form: {
    if (this.check('{')) return true;
    
    // Check for block form: identifier : (in context where lambda expected)
    if (this.check('IDENTIFIER') && 
        this.pos + 1 < this.tokens.length && 
        this.tokens[this.pos + 1].value === ':') {
      return true;
    }
    
    return false;
  }

  isAtom() {
    return this.check('STRING') || 
           this.check('NUMBER') || 
           this.check('true') || 
           this.check('false') || 
           this.check('null') || 
           this.check('[') || 
           (this.check('{') && !this.isLambdaStart()) ||
           (this.check('IDENTIFIER') && !this.isLambdaStart());
  }

  hasMoreArgs() {
    return !this.isEOF() && 
           !this.isNewline() && 
           !this.check('DEDENT') && 
           !this.check(')');
  }
}

// Usage example:
// const tokens = lexer.tokenize(sourceCode);
// const parser = new RelayParser(tokens);
// const ast = parser.parseMain();

module.exports = RelayParser; 