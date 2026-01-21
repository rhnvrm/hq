package parser

// ExpressionNode is the root interface for all AST nodes.
// Each node type represents a different expression construct.
type ExpressionNode interface {
	expressionNode() // marker method
}

// IdentityNode represents the identity operator (.)
type IdentityNode struct{}

func (IdentityNode) expressionNode() {}

// LiteralNode represents a literal value (number, string, bool, null)
type LiteralNode struct {
	Value any // float64, string, bool, or nil
}

func (LiteralNode) expressionNode() {}

// FieldAccessNode represents field access (.foo or .["key"])
type FieldAccessNode struct {
	Field string         // The field name
	From  ExpressionNode // nil means from current (.), otherwise chain from this
}

func (FieldAccessNode) expressionNode() {}

// IndexAccessNode represents array index access (.[n])
type IndexAccessNode struct {
	Index int            // The index (negative for from-end)
	From  ExpressionNode // nil means from current (.)
}

func (IndexAccessNode) expressionNode() {}

// SliceNode represents array/string slicing (.[start:end])
type SliceNode struct {
	Start *int           // nil means from beginning
	End   *int           // nil means to end
	From  ExpressionNode // nil means from current (.)
}

func (SliceNode) expressionNode() {}

// IteratorNode represents the iterator operator (.[] or .foo[])
type IteratorNode struct {
	From ExpressionNode // nil means from current (.)
}

func (IteratorNode) expressionNode() {}

// PipeNode represents the pipe operator (a | b)
type PipeNode struct {
	Left  ExpressionNode
	Right ExpressionNode
}

func (PipeNode) expressionNode() {}

// CommaNode represents comma operator (a, b) producing multiple outputs
type CommaNode struct {
	Expressions []ExpressionNode
}

func (CommaNode) expressionNode() {}

// BinaryOpNode represents binary operations (+, -, *, /, %, ==, !=, <, >, <=, >=, and, or)
type BinaryOpNode struct {
	Op    string
	Left  ExpressionNode
	Right ExpressionNode
}

func (BinaryOpNode) expressionNode() {}

// UnaryOpNode represents unary operations (not, -)
type UnaryOpNode struct {
	Op   string
	Expr ExpressionNode
}

func (UnaryOpNode) expressionNode() {}

// FunctionCallNode represents a function call (length, keys, select(.), etc.)
type FunctionCallNode struct {
	Name string
	Args []ExpressionNode
}

func (FunctionCallNode) expressionNode() {}

// ObjectConstructNode represents object construction ({a: .b, c: .d})
type ObjectConstructNode struct {
	Fields []ObjectFieldNode
}

func (ObjectConstructNode) expressionNode() {}

// ObjectFieldNode is a key-value pair in object construction
type ObjectFieldNode struct {
	Key       ExpressionNode // Usually LiteralNode(string) or IdentityNode for shorthand
	Value     ExpressionNode
	Shorthand bool // true for {foo} meaning {foo: .foo}
}

// ArrayConstructNode represents array construction ([.a, .b, .c])
type ArrayConstructNode struct {
	Elements ExpressionNode // Usually a comma expression or single expr
}

func (ArrayConstructNode) expressionNode() {}

// VariableNode represents a variable reference ($x)
type VariableNode struct {
	Name string
}

func (VariableNode) expressionNode() {}

// VariableBindNode represents variable binding (expr as $x | ...)
type VariableBindNode struct {
	Expr    ExpressionNode
	VarName string
	Body    ExpressionNode
}

func (VariableBindNode) expressionNode() {}

// ConditionalNode represents if-then-else
type ConditionalNode struct {
	Condition ExpressionNode
	Then      ExpressionNode
	Else      ExpressionNode // nil if no else
}

func (ConditionalNode) expressionNode() {}

// TryCatchNode represents try-catch
type TryCatchNode struct {
	Try   ExpressionNode
	Catch ExpressionNode // nil for default (empty)
}

func (TryCatchNode) expressionNode() {}

// AssignNode represents assignment (.foo = value)
type AssignNode struct {
	Path  ExpressionNode
	Op    string // "=", "|=", "+=", "-="
	Value ExpressionNode
}

func (AssignNode) expressionNode() {}

// AlternativeNode represents the alternative operator (//)
type AlternativeNode struct {
	Left  ExpressionNode
	Right ExpressionNode
}

func (AlternativeNode) expressionNode() {}

// OptionalNode represents the optional operator (expr?)
type OptionalNode struct {
	Expr ExpressionNode
}

func (OptionalNode) expressionNode() {}

// RecursiveDescentNode represents recursive descent (..)
type RecursiveDescentNode struct {
	From ExpressionNode // nil means from current
}

func (RecursiveDescentNode) expressionNode() {}

// StringInterpolationNode represents a string with embedded expressions
// e.g., "Hello, \(.name)!" has parts: ["Hello, ", expr(.name), "!"]
type StringInterpolationNode struct {
	Parts []StringPart
}

func (StringInterpolationNode) expressionNode() {}

// StringPart is either a literal string or an expression to interpolate
type StringPart struct {
	Literal string         // Non-empty if this is a literal part
	Expr    ExpressionNode // Non-nil if this is an expression part
}

// ReduceNode represents reduce expression: reduce EXPR as $VAR (INIT; UPDATE)
type ReduceNode struct {
	Expr    ExpressionNode // The iterator expression (e.g., .[])
	VarName string         // Variable name (without $)
	Init    ExpressionNode // Initial accumulator value
	Update  ExpressionNode // Update expression
}

func (ReduceNode) expressionNode() {}

// DynamicIndexNode represents dynamic index/key access .[$expr]
// The index expression is evaluated at runtime to get the key/index
type DynamicIndexNode struct {
	Index ExpressionNode // Expression that evaluates to key (string) or index (number)
	From  ExpressionNode // nil means from current (.)
}

func (DynamicIndexNode) expressionNode() {}

// DestructureBindNode represents destructuring variable binding
// e.g., .point as {x: $x, y: $y} | $x + $y
type DestructureBindNode struct {
	Expr     ExpressionNode    // The expression to destructure
	Bindings map[string]string // Maps field name to variable name (without $)
	Body     ExpressionNode    // The body to evaluate with bindings
}

func (DestructureBindNode) expressionNode() {}
