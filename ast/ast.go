package ast

// Node is the base interface for all AST nodes.
type Node interface {
	nodeType() string
}

// Statement nodes
type (
	// Program is the root node of every AST.
	Program struct {
		Statements []Node
	}

	// VariableDeclaration represents: store x = expr
	VariableDeclaration struct {
		Name  string
		Value Node
		Line  int
	}

	// AssignmentExpression represents: x = expr or x[i] = expr
	AssignmentExpression struct {
		Target Node // Identifier or IndexExpression
		Value  Node
		Line   int
	}

	// FunctionDeclaration represents: fn name(params) { body }
	FunctionDeclaration struct {
		Name   string
		Params []string
		Body   []Node
		Line   int
	}

	// ReturnStatement represents: return expr
	ReturnStatement struct {
		Value Node
		Line  int
	}

	// IfStatement represents: when condition { body } otherwise { body }
	IfStatement struct {
		Condition Node
		Body      []Node
		ElseBody  []Node
		Line      int
	}

	// RepeatStatement represents all loop forms:
	// repeat N { body }
	// repeat item in collection { body }
	// repeat condition { body }
	RepeatStatement struct {
		// For numeric repeat: Count is set, Iterator/Collection/Condition are nil
		Count Node

		// For for-each: Iterator and Collection are set
		Iterator   string
		Collection Node

		// For conditional: Condition is set
		Condition Node

		Body []Node
		Line int
	}

	// SayStatement represents: say expr (without parentheses)
	SayStatement struct {
		Value Node
		Line  int
	}

	// UseStatement represents: use module
	UseStatement struct {
		Module string
		Line   int
	}

	// BreakStatement represents: break
	BreakStatement struct {
		Line int
	}

	// ContinueStatement represents: continue
	ContinueStatement struct {
		Line int
	}
)

// Expression nodes
type (
	// Identifier represents a variable or function name.
	Identifier struct {
		Name string
		Line int
	}

	// NumberLiteral represents an integer or float number.
	NumberLiteral struct {
		Value float64
		IsInt bool
		Line  int
	}

	// StringLiteral represents a string value.
	StringLiteral struct {
		Value string
		Line  int
	}

	// BooleanLiteral represents true or false.
	BooleanLiteral struct {
		Value bool
		Line  int
	}

	// NullLiteral represents null.
	NullLiteral struct {
		Line int
	}

	// ArrayLiteral represents: [elem1, elem2, ...]
	ArrayLiteral struct {
		Elements []Node
		Line     int
	}

	// ObjectLiteral represents: { "key": value, ... }
	ObjectLiteral struct {
		Keys   []Node
		Values []Node
		Line   int
	}

	// BinaryExpression represents: left op right
	BinaryExpression struct {
		Left     Node
		Operator string
		Right    Node
		Line     int
	}

	// UnaryExpression represents: op expr (e.g., -x, !x)
	UnaryExpression struct {
		Operator string
		Operand  Node
		Line     int
	}

	// CallExpression represents: callee(args)
	CallExpression struct {
		Callee    Node
		Arguments []Node
		Line      int
	}

	// IndexExpression represents: object[index] or object.field
	IndexExpression struct {
		Object Node
		Index  Node
		Line   int
	}

	// MemberExpression represents: object.field (dot access)
	MemberExpression struct {
		Object   Node
		Property string
		Line     int
	}
)

// nodeType implementations
func (p *Program) nodeType() string             { return "Program" }
func (v *VariableDeclaration) nodeType() string  { return "VariableDeclaration" }
func (a *AssignmentExpression) nodeType() string { return "AssignmentExpression" }
func (f *FunctionDeclaration) nodeType() string  { return "FunctionDeclaration" }
func (r *ReturnStatement) nodeType() string      { return "ReturnStatement" }
func (i *IfStatement) nodeType() string          { return "IfStatement" }
func (r *RepeatStatement) nodeType() string      { return "RepeatStatement" }
func (s *SayStatement) nodeType() string         { return "SayStatement" }
func (u *UseStatement) nodeType() string         { return "UseStatement" }
func (b *BreakStatement) nodeType() string       { return "BreakStatement" }
func (c *ContinueStatement) nodeType() string    { return "ContinueStatement" }
func (i *Identifier) nodeType() string           { return "Identifier" }
func (n *NumberLiteral) nodeType() string        { return "NumberLiteral" }
func (s *StringLiteral) nodeType() string        { return "StringLiteral" }
func (b *BooleanLiteral) nodeType() string       { return "BooleanLiteral" }
func (n *NullLiteral) nodeType() string          { return "NullLiteral" }
func (a *ArrayLiteral) nodeType() string         { return "ArrayLiteral" }
func (o *ObjectLiteral) nodeType() string        { return "ObjectLiteral" }
func (b *BinaryExpression) nodeType() string     { return "BinaryExpression" }
func (u *UnaryExpression) nodeType() string      { return "UnaryExpression" }
func (c *CallExpression) nodeType() string       { return "CallExpression" }
func (i *IndexExpression) nodeType() string      { return "IndexExpression" }
func (m *MemberExpression) nodeType() string     { return "MemberExpression" }
