package types

// Context holds evaluation state during expression evaluation.
type Context struct {
	// MatchingNodes are the current nodes being processed.
	// Most operations work on all matching nodes and produce new matching nodes.
	MatchingNodes []*CandidateNode

	// Variables holds bound variables ($x, $y, etc.).
	Variables map[string]any

	// ReadOnlyVariables are variables that cannot be reassigned.
	ReadOnlyVariables map[string]any
}

// NewContext creates a new evaluation context from input data.
func NewContext(input any) *Context {
	return &Context{
		MatchingNodes:     []*CandidateNode{NewCandidateNode(input)},
		Variables:         make(map[string]any),
		ReadOnlyVariables: make(map[string]any),
	}
}

// Clone creates a copy of the context with new MatchingNodes slice.
// Variables are shared (intentionally - for lexical scoping).
func (c *Context) Clone() *Context {
	nodes := make([]*CandidateNode, len(c.MatchingNodes))
	copy(nodes, c.MatchingNodes)
	return &Context{
		MatchingNodes:     nodes,
		Variables:         c.Variables,
		ReadOnlyVariables: c.ReadOnlyVariables,
	}
}

// SetMatchingNodes sets the matching nodes for subsequent operations.
func (c *Context) SetMatchingNodes(nodes []*CandidateNode) {
	c.MatchingNodes = nodes
}

// SingleNode returns true if there's exactly one matching node.
func (c *Context) SingleNode() bool {
	return len(c.MatchingNodes) == 1
}

// GetVariable returns a variable value, checking ReadOnlyVariables first.
func (c *Context) GetVariable(name string) (any, bool) {
	if v, ok := c.ReadOnlyVariables[name]; ok {
		return v, true
	}
	v, ok := c.Variables[name]
	return v, ok
}
