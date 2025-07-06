/**
 * Relay Language Parser - TypeScript Implementation
 * 
 * Implements the final grammar:
 * - Everything is an expression that returns a value
 * - Indented blocks are syntactic sugar for sequence expressions  
 * - Three forms of function calls: inline, sequence, parenthesized
 * - Lambda expressions in brace and block forms
 * - JSON literals as first-class expressions
 */

// Token types
export type TokenType = 
  | 'IDENTIFIER' | 'STRING' | 'NUMBER' | 'BOOLEAN' | 'NULL'
  | 'LPAREN' | 'RPAREN' | 'LBRACKET' | 'RBRACKET' | 'LBRACE' | 'RBRACE'
  | 'COLON' | 'COMMA' | 'NEWLINE' | 'INDENT' | 'DEDENT' | 'COMMENT' | 'EOF'
  | 'OPERATOR'; // For operators like +, -, *, %, etc.

export interface Token {
  type: TokenType;
  value: any;
  line: number;
  column: number;
}

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
}

export interface AtomNode extends ASTNode {
  type: 'atom';
  value: any;
}

export interface SequenceNode extends ASTNode {
  type: 'sequence';
  expressions: ExpressionNode[];
}

export interface LambdaNode extends ASTNode {
  type: 'lambda';
  params: string[];
  body: ExpressionNode;
}

export interface JsonArrayNode extends ASTNode {
  type: 'json_array';
  elements: ExpressionNode[];
}

export interface JsonObjectNode extends ASTNode {
  type: 'json_object';
  pairs: { key: string; value: ExpressionNode }[];
}

export interface CommentNode extends ASTNode {
  type: 'comment';
  value: string;
}

export interface IdentifierNode extends ASTNode {
  type: 'identifier';
  name: string;
}

export type ExpressionNode = 
  | FuncallNode 
  | AtomNode 
  | SequenceNode 
  | LambdaNode 
  | JsonArrayNode 
  | JsonObjectNode 
  | CommentNode
  | IdentifierNode;

export class RelayLexer {
  private source: string;
  private pos: number = 0;
  private line: number = 1;
  private column: number = 1;
  private indentStack: number[] = [0];
  private tokens: Token[] = [];

  constructor(source: string) {
    this.source = source;
  }

  tokenize(): Token[] {
    while (this.pos < this.source.length) {
      this.skipWhitespace();
      
      if (this.pos >= this.source.length) break;
      
      const char = this.source[this.pos];
      
      // Handle newlines and indentation
      if (char === '\n' || char === '\r') {
        this.handleNewline();
        continue;
      }
      
      // Handle comments
      if (char === '#') {
        this.handleComment();
        continue;
      }
      
      // Handle strings
      if (char === '"') {
        this.handleString();
        continue;
      }
      
      // Handle numbers
      if (this.isDigit(char) || (char === '-' && this.isDigit(this.peek()))) {
        this.handleNumber();
        continue;
      }
      
      // Handle identifiers and keywords
      if (this.isIdentifierStart(char)) {
        this.handleIdentifier();
        continue;
      }
      
      // Handle operators
      if (this.isOperator(char)) {
        this.handleOperator();
        continue;
      }
      
      // Handle single-character tokens
      switch (char) {
        case '(':
          this.addToken('LPAREN', '(');
          break;
        case ')':
          this.addToken('RPAREN', ')');
          break;
        case '[':
          this.addToken('LBRACKET', '[');
          break;
        case ']':
          this.addToken('RBRACKET', ']');
          break;
        case '{':
          this.addToken('LBRACE', '{');
          break;
        case '}':
          this.addToken('RBRACE', '}');
          break;
        case ':':
          this.addToken('COLON', ':');
          break;
        case ',':
          this.addToken('COMMA', ',');
          break;
        default:
          throw new Error(`Unexpected character: ${char} at line ${this.line}, column ${this.column}`);
      }
      
      this.advance();
    }
    
    // Handle final dedents
    while (this.indentStack.length > 1) {
      this.indentStack.pop();
      this.addToken('DEDENT', '');
    }
    
    this.addToken('EOF', '');
    return this.tokens;
  }

  private handleNewline(): void {
    const start = this.pos;
    
    // Skip newline characters
    while (this.pos < this.source.length && (this.source[this.pos] === '\n' || this.source[this.pos] === '\r')) {
      if (this.source[this.pos] === '\n') {
        this.line++;
        this.column = 1;
      }
      this.pos++;
    }
    
    // Don't emit newline tokens for empty lines or lines with only comments
    if (this.pos < this.source.length && this.source[this.pos] === '#') {
      return;
    }
    
    this.addToken('NEWLINE', this.source.slice(start, this.pos));
    
    // Handle indentation after newline
    this.handleIndentation();
  }

  private handleIndentation(): void {
    const start = this.pos;
    let indent = 0;
    
    // Count spaces and tabs
    while (this.pos < this.source.length && (this.source[this.pos] === ' ' || this.source[this.pos] === '\t')) {
      if (this.source[this.pos] === '\t') {
        indent += 4; // Tab = 4 spaces
      } else {
        indent += 1;
      }
      this.pos++;
    }
    
    // Skip empty lines
    if (this.pos < this.source.length && (this.source[this.pos] === '\n' || this.source[this.pos] === '\r' || this.source[this.pos] === '#')) {
      return;
    }
    
    const currentIndent = this.indentStack[this.indentStack.length - 1];
    
    if (indent > currentIndent) {
      // Increase indentation
      this.indentStack.push(indent);
      this.addToken('INDENT', '');
    } else if (indent < currentIndent) {
      // Decrease indentation
      while (this.indentStack.length > 1 && this.indentStack[this.indentStack.length - 1] > indent) {
        this.indentStack.pop();
        this.addToken('DEDENT', '');
      }
      
      if (this.indentStack[this.indentStack.length - 1] !== indent) {
        throw new Error(`Indentation mismatch at line ${this.line}`);
      }
    }
    
    this.column += (this.pos - start);
  }

  private handleComment(): void {
    const start = this.pos;
    
    // Skip until end of line
    while (this.pos < this.source.length && this.source[this.pos] !== '\n' && this.source[this.pos] !== '\r') {
      this.pos++;
    }
    
    this.addToken('COMMENT', this.source.slice(start, this.pos));
  }

  private handleString(): void {
    const start = this.pos;
    this.advance(); // Skip opening quote
    
    let value = '';
    
    while (this.pos < this.source.length && this.source[this.pos] !== '"') {
      if (this.source[this.pos] === '\\' && this.pos + 1 < this.source.length) {
        // Handle escape sequences
        this.advance();
        const escaped = this.source[this.pos];
        switch (escaped) {
          case 'n': value += '\n'; break;
          case 't': value += '\t'; break;
          case 'r': value += '\r'; break;
          case '\\': value += '\\'; break;
          case '"': value += '"'; break;
          default: value += escaped; break;
        }
      } else {
        value += this.source[this.pos];
      }
      this.advance();
    }
    
    if (this.pos >= this.source.length) {
      throw new Error(`Unterminated string at line ${this.line}`);
    }
    
    this.advance(); // Skip closing quote
    this.addToken('STRING', value);
  }

  private handleNumber(): void {
    const start = this.pos;
    
    // Handle negative sign
    if (this.source[this.pos] === '-') {
      this.advance();
    }
    
    // Handle integer part
    while (this.pos < this.source.length && this.isDigit(this.source[this.pos])) {
      this.advance();
    }
    
    // Handle decimal part
    if (this.pos < this.source.length && this.source[this.pos] === '.') {
      this.advance();
      while (this.pos < this.source.length && this.isDigit(this.source[this.pos])) {
        this.advance();
      }
    }
    
    const text = this.source.slice(start, this.pos);
    const value = text.includes('.') ? parseFloat(text) : parseInt(text, 10);
    
    this.addToken('NUMBER', value);
  }

  private handleIdentifier(): void {
    const start = this.pos;
    const startColumn = this.column;
    
    // First character (letter or underscore)
    this.advance();
    
    // Rest of identifier
    while (this.pos < this.source.length && this.isIdentifierPart(this.source[this.pos])) {
      this.advance();
    }
    
    const text = this.source.slice(start, this.pos);
    
    // Check for keywords
    switch (text) {
      case 'true':
        this.addToken('BOOLEAN', true, startColumn);
        break;
      case 'false':
        this.addToken('BOOLEAN', false, startColumn);
        break;
      case 'null':
        this.addToken('NULL', null, startColumn);
        break;
      default:
        this.addToken('IDENTIFIER', text, startColumn);
        break;
    }
  }

  private handleOperator(): void {
    const char = this.source[this.pos];
    this.addToken('OPERATOR', char);
    this.advance();
  }

  private skipWhitespace(): void {
    while (this.pos < this.source.length && this.isWhitespace(this.source[this.pos])) {
      this.advance();
    }
  }

  private isWhitespace(char: string): boolean {
    return char === ' ' || char === '\t';
  }

  private isDigit(char: string): boolean {
    return char >= '0' && char <= '9';
  }

  private isIdentifierStart(char: string): boolean {
    return (char >= 'a' && char <= 'z') || 
           (char >= 'A' && char <= 'Z') || 
           char === '_';
  }

  private isIdentifierPart(char: string): boolean {
    return this.isIdentifierStart(char) || 
           this.isDigit(char) || 
           char === '-' || 
           char === '?';
  }

  private isOperator(char: string): boolean {
    return '+-*/%<>=!'.includes(char);
  }

  private peek(): string {
    return this.pos + 1 < this.source.length ? this.source[this.pos + 1] : '\0';
  }

  private advance(): void {
    if (this.pos < this.source.length) {
      this.pos++;
      this.column++;
    }
  }

  private addToken(type: TokenType, value: any, startPos?: number): void {
    this.tokens.push({
      type,
      value,
      line: this.line,
      column: startPos || this.column
    });
  }
}

export class RelayParser {
  private tokens: Token[];
  private pos: number = 0;

  constructor(tokens: Token[]) {
    this.tokens = tokens;
  }

  // program = expression*
  parseProgram(): ProgramNode {
    const expressions: ExpressionNode[] = [];
    
    while (!this.isEOF()) {
      // Skip newlines, comments, and empty indentation at top level
      if (this.check('NEWLINE') || this.check('COMMENT') || this.check('INDENT') || this.check('DEDENT')) {
        this.advance();
        continue;
      }
      
      expressions.push(this.parseExpression());
    }
    
    return {
      type: 'program',
      expressions
    };
  }

  // expression = funcall | atom | lambda | sequence | comment
  parseExpression(): ExpressionNode {
    // Handle comments
    if (this.check('COMMENT')) {
      const comment = this.advance();
      return {
        type: 'comment',
        value: comment.value
      };
    }
    
    // Handle sequences (indented blocks)
    if (this.check('INDENT')) {
      return this.parseSequence();
    }
    
    // Handle lambda expressions
    if (this.isLambdaStart()) {
      return this.parseLambda();
    }
    
    // Handle JSON arrays
    if (this.check('LBRACKET')) {
      return this.parseJsonArray();
    }
    
    // Handle JSON objects (that are not lambdas)
    if (this.check('LBRACE') && !this.isLambdaStart()) {
      return this.parseJsonObject();
    }
    
    // Handle parenthesized expressions
    if (this.check('LPAREN')) {
      this.advance(); // consume '('
      const expr = this.parseExpression();
      this.consume('RPAREN');
      return expr;
    }
    
    // Handle atoms that aren't function calls
    if (this.isAtomOnly()) {
      return this.parseAtom();
    }
    
    // Check for infix operator pattern: identifier operator ...
    if (this.isInfixExpression()) {
      return this.parseInfixExpression();
    }
    
    // Default to function call for identifiers and operators
    // In Relay, all identifiers are function calls (no bare identifier references)
    return this.parseFuncall();
  }

  // funcall = identifier argument_list
  parseFuncall(): FuncallNode {
    const name = this.parseIdentifier();
    const args = this.parseArgumentList();
    
    return {
      type: 'funcall',
      name,
      args
    };
  }

  // argument_list = mixed inline and parenthesized args | sequence_arg | parenthesized_args
  parseArgumentList(): ExpressionNode[] {
    // Check for sequence argument ONLY (newline + indent)
    if (this.check('NEWLINE') && this.peekNext() === 'INDENT') {
      this.advance(); // consume NEWLINE
      return [this.parseSequence()];
    }
    
    // Parse mixed inline and parenthesized arguments
    const args: ExpressionNode[] = [];
    
    // Handle case where arguments are wrapped in parentheses: func (arg1 arg2 arg3)
    if (this.check('LPAREN') && this.isFullyParenthesizedArgs()) {
      this.advance(); // consume '('
      
      while (!this.check('RPAREN') && !this.isEOF()) {
        // Skip newlines inside parentheses
        if (this.check('NEWLINE')) {
          this.advance();
          continue;
        }
        
        // Parse arguments inline - same logic as mixed inline arguments
        if (this.check('LPAREN')) {
          // Parenthesized expression - parse as function call
          this.advance(); // consume '('
          const expr = this.parseExpression();
          this.consume('RPAREN');
          args.push(expr);
        } else if (this.check('LBRACKET')) {
          // JSON array
          args.push(this.parseJsonArray());
        } else if (this.check('LBRACE') && !this.isLambdaStart()) {
          // JSON object
          args.push(this.parseJsonObject());
        } else if (this.isLambdaStart()) {
          // Lambda expression
          args.push(this.parseLambda());
        } else if (this.check('IDENTIFIER')) {
          // Check if this identifier has arguments following it (function call vs identifier ref)
          const startPos = this.pos;
          this.advance(); // tentatively consume identifier
          
          if (this.hasMoreInlineArgs() && !this.check('RPAREN')) {
            // This is a function call - reset and parse as funcall
            this.pos = startPos;
            args.push(this.parseFuncall());
          } else {
            // This is an identifier reference - reset and parse as identifier ref
            this.pos = startPos;
            args.push(this.parseIdentifierRef());
          }
        } else if (this.check('OPERATOR')) {
          // Operator references (like + used as a value)
          args.push(this.parseIdentifierRef());
        } else {
          // Everything else in inline context is an atom
          args.push(this.parseAtom());
        }
      }
      
      this.consume('RPAREN');
      return args;
    }
    
    // Parse mixed inline arguments (including parenthesized expressions)
    while (this.hasMoreInlineArgs()) {
      if (this.check('LPAREN')) {
        // Parenthesized expression - parse as function call
        this.advance(); // consume '('
        const expr = this.parseExpression();
        this.consume('RPAREN');
        args.push(expr);
      } else if (this.check('LBRACKET')) {
        // JSON array
        args.push(this.parseJsonArray());
      } else if (this.check('LBRACE') && !this.isLambdaStart()) {
        // JSON object
        args.push(this.parseJsonObject());
      } else if (this.isLambdaStart()) {
        // Lambda expression
        args.push(this.parseLambda());
      } else if (this.isIdentifierRef()) {
        // Identifier references (parameters, variables)
        args.push(this.parseIdentifierRef());
      } else {
        // Everything else in inline context is an atom
        args.push(this.parseAtom());
      }
    }
    
    // CRITICAL FIX: Check for indented block AFTER inline arguments
    // This handles cases like: def factorial n [NEWLINE + INDENT]
    if (this.check('NEWLINE') && this.peekNext() === 'INDENT') {
      this.advance(); // consume NEWLINE
      args.push(this.parseSequence());
    }
    
    return args;
  }

  // sequence = indent expression+ dedent
  parseSequence(): SequenceNode {
    this.consume('INDENT');
    
    const expressions: ExpressionNode[] = [];
    while (!this.check('DEDENT') && !this.isEOF()) {
      // Skip newlines and comments inside sequences
      if (this.check('NEWLINE') || this.check('COMMENT')) {
        this.advance();
        continue;
      }
      
      expressions.push(this.parseExpression());
    }
    
    this.consume('DEDENT');
    
    return {
      type: 'sequence',
      expressions
    };
  }

  // lambda = brace_lambda | block_lambda
  parseLambda(): LambdaNode {
    // Brace form: {params: body}
    if (this.check('LBRACE')) {
      this.advance(); // consume '{'
      const params = this.parseParameters();
      this.consume('COLON');
      const body = this.parseExpression();
      this.consume('RBRACE');
      
      return {
        type: 'lambda',
        params,
        body
      };
    }
    
    // Block form: params: body
    const params = this.parseParameters();
    this.consume('COLON');
    
    let body: ExpressionNode;
    if (this.check('NEWLINE') && this.peekNext() === 'INDENT') {
      this.advance(); // consume NEWLINE
      body = this.parseSequence();
    } else {
      body = this.parseExpression();
    }
    
    return {
      type: 'lambda',
      params,
      body
    };
  }

  // parameters = identifier ("," identifier)*
  parseParameters(): string[] {
    const params = [this.parseIdentifier()];
    
    while (this.check('COMMA')) {
      this.advance(); // consume ','
      params.push(this.parseIdentifier());
    }
    
    return params;
  }

  // atom = string | number | boolean | null (literals only)
  parseAtom(): AtomNode {
    // Skip comments
    if (this.check('COMMENT')) {
      this.advance();
      return this.parseAtom(); // Try again after skipping comment
    }
    
    if (this.check('STRING')) {
      const token = this.advance();
      return {
        type: 'atom',
        value: token.value
      };
    }
    
    if (this.check('NUMBER')) {
      const token = this.advance();
      return {
        type: 'atom',
        value: token.value
      };
    }
    
    if (this.check('BOOLEAN')) {
      const token = this.advance();
      return {
        type: 'atom',
        value: token.value
      };
    }
    
    if (this.check('NULL')) {
      this.advance();
      return {
        type: 'atom',
        value: null
      };
    }
    
    throw new Error(`Unexpected token: ${this.currentToken()?.type} at line ${this.currentToken()?.line}`);
  }

  // Parse identifier reference (not a literal)
  parseIdentifierRef(): IdentifierNode {
    if (!this.check('IDENTIFIER') && !this.check('OPERATOR')) {
      throw new Error(`Expected identifier at line ${this.currentToken()?.line}`);
    }
    const token = this.advance();
    return {
      type: 'identifier',
      name: token.value
    };
  }

  // json_array = "[" (expression ("," expression)*)? "]"
  parseJsonArray(): JsonArrayNode {
    this.consume('LBRACKET');
    
    const elements: ExpressionNode[] = [];
    
    // Skip newlines and indentation after opening bracket
    this.skipJsonWhitespace();
    
    if (!this.check('RBRACKET')) {
      elements.push(this.parseExpression());
      
      // Handle both comma-separated and newline-separated elements
      while (this.check('COMMA') || this.check('NEWLINE') || this.check('INDENT')) {
        if (this.check('COMMA')) {
          this.advance(); // consume ','
          this.skipJsonWhitespace(); // Skip whitespace after comma
        } else {
          this.skipJsonWhitespace(); // Skip newlines and indentation
        }
        
        // Check if we've reached the end after skipping whitespace
        if (this.check('RBRACKET')) {
          break;
        }
        
        elements.push(this.parseExpression());
      }
    }
    
    // Skip any trailing whitespace before closing bracket
    this.skipJsonWhitespace();
    this.consume('RBRACKET');
    
    return {
      type: 'json_array',
      elements
    };
  }

  // json_object = "{" (json_pair ("," json_pair)*)? "}"
  parseJsonObject(): JsonObjectNode {
    this.consume('LBRACE');
    
    const pairs: { key: string; value: ExpressionNode }[] = [];
    
    // Skip newlines and indentation after opening brace
    this.skipJsonWhitespace();
    
    if (!this.check('RBRACE')) {
      pairs.push(this.parseJsonPair());
      
      // Handle both comma-separated and newline-separated pairs
      while (this.check('COMMA') || this.check('NEWLINE') || this.check('INDENT')) {
        if (this.check('COMMA')) {
          this.advance(); // consume ','
          this.skipJsonWhitespace(); // Skip whitespace after comma
        } else {
          this.skipJsonWhitespace(); // Skip newlines and indentation
        }
        
        // Check if we've reached the end after skipping whitespace
        if (this.check('RBRACE')) {
          break;
        }
        
        pairs.push(this.parseJsonPair());
      }
    }
    
    // Skip any trailing whitespace before closing brace
    this.skipJsonWhitespace();
    this.consume('RBRACE');
    
    return {
      type: 'json_object',
      pairs
    };
  }

  // Helper method to skip whitespace tokens inside JSON objects
  private skipJsonWhitespace(): void {
    while (this.check('NEWLINE') || this.check('INDENT') || this.check('DEDENT') || this.check('COMMENT')) {
      this.advance();
    }
  }

  // json_pair = (string | identifier) ":" expression
  parseJsonPair(): { key: string; value: ExpressionNode } {
    let key: string;
    
    if (this.check('STRING')) {
      key = this.advance().value;
    } else if (this.check('IDENTIFIER')) {
      key = this.advance().value;
    } else {
      throw new Error(`Expected string or identifier for JSON key at line ${this.currentToken()?.line}`);
    }
    
    this.consume('COLON');
    const value = this.parseExpression();
    
    return { key, value };
  }

  // Helper methods
  parseIdentifier(): string {
    if (!this.check('IDENTIFIER') && !this.check('OPERATOR')) {
      throw new Error(`Expected identifier at line ${this.currentToken()?.line}`);
    }
    return this.advance().value;
  }

  currentToken(): Token | undefined {
    return this.tokens[this.pos];
  }

  check(expectedType: TokenType): boolean {
    const token = this.currentToken();
    return token ? token.type === expectedType : false;
  }

  advance(): Token {
    return this.tokens[this.pos++];
  }

  consume(expectedType: TokenType): Token {
    if (this.check(expectedType)) {
      return this.advance();
    }
    throw new Error(`Expected ${expectedType}, got ${this.currentToken()?.type || 'EOF'} at line ${this.currentToken()?.line || 'EOF'}`);
  }

  isEOF(): boolean {
    return this.pos >= this.tokens.length || this.check('EOF');
  }

  peekNext(): TokenType | null {
    return this.pos + 1 < this.tokens.length ? this.tokens[this.pos + 1].type : null;
  }

  isLambdaStart(): boolean {
    // Check for brace form: {param: body}
    if (this.check('LBRACE')) {
      // Look ahead to see if this is a lambda or JSON object
      // Lambda: {param: body} - starts with identifier
      // JSON: {"key": value} - starts with string or identifier followed by colon
      const nextPos = this.pos + 1;
      if (nextPos < this.tokens.length) {
        const nextToken = this.tokens[nextPos];
        // If it's an identifier followed by colon, it's a lambda
        if (nextToken.type === 'IDENTIFIER' && 
            nextPos + 1 < this.tokens.length && 
            this.tokens[nextPos + 1].type === 'COLON') {
          return true;
        }
        // If it's a string followed by colon, it's JSON
        if (nextToken.type === 'STRING' && 
            nextPos + 1 < this.tokens.length && 
            this.tokens[nextPos + 1].type === 'COLON') {
          return false;
        }
        // If it's just an identifier (no colon), it's JSON
        if (nextToken.type === 'IDENTIFIER') {
          return false;
        }
      }
      return false;
    }
    
    // Check for block form: identifier : (but not JSON object)
    if (this.check('IDENTIFIER') && this.peekNext() === 'COLON') {
      return true;
    }
    
    return false;
  }

  isAtom(): boolean {
    return this.check('STRING') || 
           this.check('NUMBER') || 
           this.check('BOOLEAN') || 
           this.check('NULL') || 
           this.check('LBRACKET') || 
           (this.check('LBRACE') && !this.isLambdaStart());
  }

  isAtomOnly(): boolean {
    return this.check('STRING') || 
           this.check('NUMBER') || 
           this.check('BOOLEAN') || 
           this.check('NULL');
  }

  isIdentifierRef(): boolean {
    return (this.check('IDENTIFIER') || this.check('OPERATOR')) && !this.isLambdaStart();
  }

  hasMoreInlineArgs(): boolean {
    // Skip comments when checking for more args
    if (this.check('COMMENT')) {
      return false;
    }
    
    return !this.isEOF() && 
           !this.check('NEWLINE') && 
           !this.check('DEDENT') && 
           !this.check('RPAREN') && 
           !this.check('RBRACKET') && 
           !this.check('RBRACE') && 
           !this.check('COMMA') && 
           !this.check('COLON') &&
           !this.check('EOF') &&
           this.currentToken() != null;
  }

  hasArgumentsAfter(): boolean {
    // Look ahead to see if there are more tokens after the current one
    const nextPos = this.pos + 1;
    if (nextPos >= this.tokens.length) return false;
    
    const nextToken = this.tokens[nextPos];
    return nextToken && 
           nextToken.type !== 'NEWLINE' && 
           nextToken.type !== 'DEDENT' && 
           nextToken.type !== 'RPAREN' && 
           nextToken.type !== 'RBRACKET' && 
           nextToken.type !== 'RBRACE' && 
           nextToken.type !== 'COMMA' && 
           nextToken.type !== 'COLON' &&
           nextToken.type !== 'EOF';
  }

  isInfixExpression(): boolean {
    // Check for pattern: identifier operator ...
    if (this.check('IDENTIFIER') && 
        this.pos + 1 < this.tokens.length && 
        this.tokens[this.pos + 1].type === 'OPERATOR') {
      return true;
    }
    return false;
  }

  parseInfixExpression(): FuncallNode {
    // Parse infix expression: identifier operator ... 
    // Convert to prefix: operator identifier ...
    
    const leftOperand = this.parseIdentifierRef(); // consume identifier
    const operator = this.advance(); // consume operator
    
    // Parse the rest as arguments
    const args: ExpressionNode[] = [leftOperand];
    
    // Parse remaining arguments
    while (this.hasMoreInlineArgs()) {
      if (this.check('LPAREN')) {
        // Parenthesized expression
        this.advance(); // consume '('
        const expr = this.parseExpression();
        this.consume('RPAREN');
        args.push(expr);
      } else if (this.check('LBRACKET')) {
        // JSON array
        args.push(this.parseJsonArray());
      } else if (this.check('LBRACE') && !this.isLambdaStart()) {
        // JSON object
        args.push(this.parseJsonObject());
      } else if (this.isLambdaStart()) {
        // Lambda expression
        args.push(this.parseLambda());
      } else if (this.check('IDENTIFIER') && this.hasMoreInlineArgs()) {
        // Check if this is another infix expression or just an atom
        const nextPos = this.pos + 1;
        if (nextPos < this.tokens.length && this.tokens[nextPos].type === 'OPERATOR') {
          // This is another infix expression - parse it recursively
          args.push(this.parseInfixExpression());
        } else {
          // Just an identifier reference
          args.push(this.parseIdentifierRef());
        }
      } else if (this.isIdentifierRef()) {
        // Identifier reference
        args.push(this.parseIdentifierRef());
      } else {
        // Atom
        args.push(this.parseAtom());
      }
    }
    
    return {
      type: 'funcall',
      name: operator.value,
      args
    };
  }

  isFullyParenthesizedArgs(): boolean {
    // Check if this is a case like: func (arg1 arg2 arg3)
    // vs: func (expr) arg2 arg3
    // We look ahead to see if there are tokens after the closing paren
    
    if (!this.check('LPAREN')) return false;
    
    let depth = 0;
    let pos = this.pos;
    
    // Find the matching closing paren
    while (pos < this.tokens.length) {
      const token = this.tokens[pos];
      if (token.type === 'LPAREN') depth++;
      else if (token.type === 'RPAREN') depth--;
      
      if (depth === 0) {
        // Found matching paren, check what's after
        const nextPos = pos + 1;
        if (nextPos >= this.tokens.length) return true;
        
        const nextToken = this.tokens[nextPos];
        // If there are more non-terminating tokens, it's mixed args
        return nextToken.type === 'NEWLINE' || 
               nextToken.type === 'DEDENT' || 
               nextToken.type === 'EOF' ||
               nextToken.type === 'RPAREN' ||
               nextToken.type === 'RBRACKET' ||
               nextToken.type === 'RBRACE';
      }
      pos++;
    }
    
    return false;
  }
}

// Usage functions
export function tokenize(source: string): Token[] {
  const lexer = new RelayLexer(source);
  return lexer.tokenize();
}

export function parse(source: string): ProgramNode {
  const tokens = tokenize(source);
  const parser = new RelayParser(tokens);
  return parser.parseProgram();
}

// Legacy interface compatibility
export interface RelayAST {
  type: 'program';
  expressions: ExpressionNode[];
}

export interface RelayStatement {
  type: string;
  [key: string]: any;
} 