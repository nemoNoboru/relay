package ast

import "time"

// Node represents any node in the AST
type Node interface {
	String() string
}

// Program represents the root of the AST
type Program struct {
	Structs   []*StructDecl
	Protocols []*ProtocolDecl
	Servers   []*ServerDecl
	Templates []*TemplateDecl
	Config    *ConfigDecl
	Auth      *AuthDecl
}

// StructDecl represents a struct declaration
type StructDecl struct {
	Name   string
	Fields []*FieldDecl
}

type FieldDecl struct {
	Name       string
	Type       Type
	Validators []Validator
}

// ProtocolDecl represents a protocol declaration
type ProtocolDecl struct {
	Name    string
	Methods []*MethodSignature
}

type MethodSignature struct {
	Name       string
	Parameters []*Parameter
	ReturnType Type
}

type Parameter struct {
	Name string
	Type Type
}

// ServerDecl represents a server declaration
type ServerDecl struct {
	Name      string
	Protocol  string // Protocol this server implements
	State     *StateDecl
	Receivers []*ReceiveDecl
}

type StateDecl struct {
	Fields []*StateField
}

type StateField struct {
	Name         string
	Type         Type
	DefaultValue Expression
}

type ReceiveDecl struct {
	Method     string
	Parameters []*Parameter
	ReturnType Type
	Body       *BlockStatement
}

// TemplateDecl represents a template declaration
type TemplateDecl struct {
	Path   string
	Method string
	Params []*Parameter
}

// ConfigDecl represents configuration
type ConfigDecl struct {
	Fields []*ConfigField
}

type ConfigField struct {
	Name  string
	Value Expression
}

// AuthDecl represents authentication configuration
type AuthDecl struct {
	Users string // "local" or "federated"
}

// Types
type Type interface {
	TypeName() string
}

type BasicType struct {
	Name string // string, number, bool, datetime
}

func (t *BasicType) TypeName() string {
	return t.Name
}

type ArrayType struct {
	ElementType Type
}

func (t *ArrayType) TypeName() string {
	return "[]" + t.ElementType.TypeName()
}

type ObjectType struct {
	KeyType   Type
	ValueType Type
}

func (t *ObjectType) TypeName() string {
	return "{" + t.KeyType.TypeName() + ": " + t.ValueType.TypeName() + "}"
}

type OptionalType struct {
	InnerType Type
}

func (t *OptionalType) TypeName() string {
	return "optional(" + t.InnerType.TypeName() + ")"
}

// Validators
type Validator interface {
	ValidatorName() string
}

type MinValidator struct {
	Value int
}

func (v *MinValidator) ValidatorName() string {
	return "min"
}

type MaxValidator struct {
	Value int
}

func (v *MaxValidator) ValidatorName() string {
	return "max"
}

type EmailValidator struct{}

func (v *EmailValidator) ValidatorName() string {
	return "email"
}

type URLValidator struct{}

func (v *URLValidator) ValidatorName() string {
	return "url"
}

type RegexValidator struct {
	Pattern string
}

func (v *RegexValidator) ValidatorName() string {
	return "regex"
}

// Statements
type Statement interface {
	Node
	statementNode()
}

type BlockStatement struct {
	Statements []Statement
}

func (s *BlockStatement) statementNode() {}
func (s *BlockStatement) String() string { return "BlockStatement" }

type ExpressionStatement struct {
	Expression Expression  
}

func (s *ExpressionStatement) statementNode() {}
func (s *ExpressionStatement) String() string { return "ExpressionStatement" }

type ReturnStatement struct {
	Value Expression
}

func (s *ReturnStatement) statementNode() {}
func (s *ReturnStatement) String() string { return "ReturnStatement" }

type LetStatement struct {
	Name  string
	Value Expression
}

func (s *LetStatement) statementNode() {}
func (s *LetStatement) String() string { return "LetStatement" }

type IfStatement struct {
	Condition Expression
	ThenBranch Statement
	ElseBranch Statement
}

func (s *IfStatement) statementNode() {}
func (s *IfStatement) String() string { return "IfStatement" }

type ForStatement struct {
	Variable   string
	Iterable   Expression
	Body       Statement
}

func (s *ForStatement) statementNode() {}
func (s *ForStatement) String() string { return "ForStatement" }

// Expressions
type Expression interface {
	Node
	expressionNode()
}

type StringLiteral struct {
	Value string
}

func (e *StringLiteral) expressionNode() {}
func (e *StringLiteral) String() string { return e.Value }

type NumberLiteral struct {
	Value float64
}

func (e *NumberLiteral) expressionNode() {}
func (e *NumberLiteral) String() string { return "" }

type BooleanLiteral struct {
	Value bool
}

func (e *BooleanLiteral) expressionNode() {}
func (e *BooleanLiteral) String() string { return "" }

type DateTimeLiteral struct {
	Value time.Time
}

func (e *DateTimeLiteral) expressionNode() {}
func (e *DateTimeLiteral) String() string { return "" }

type Identifier struct {
	Name string
}

func (e *Identifier) expressionNode() {}
func (e *Identifier) String() string { return e.Name }

type ArrayLiteral struct {
	Elements []Expression
}

func (e *ArrayLiteral) expressionNode() {}
func (e *ArrayLiteral) String() string { return "ArrayLiteral" }

type ObjectLiteral struct {
	Fields []*ObjectField
}

type ObjectField struct {
	Key   string
	Value Expression
}

func (e *ObjectLiteral) expressionNode() {}
func (e *ObjectLiteral) String() string { return "ObjectLiteral" }

type CallExpression struct {
	Function  Expression
	Arguments []Expression
}

func (e *CallExpression) expressionNode() {}
func (e *CallExpression) String() string { return "CallExpression" }

type SendExpression struct {
	Target string // Service name or protocol
	Method string
	Params Expression
}

func (e *SendExpression) expressionNode() {}
func (e *SendExpression) String() string { return "SendExpression" }

type DispatchExpression struct {
	Value Expression
	Cases []*DispatchCase
}

type DispatchCase struct {
	Pattern string
	Body    Statement
}

func (e *DispatchExpression) expressionNode() {}
func (e *DispatchExpression) String() string { return "DispatchExpression" } 