// Package types provides the core data structures for hq expression evaluation.
package types

// CandidateNode wraps a value with metadata for evaluation.
// Based on yq's CandidateNode pattern, but simplified for HUML.
type CandidateNode struct {
	// Value is the actual Go value (from huml.Unmarshal).
	// Can be: map[string]any, []any, string, float64, bool, nil
	Value any

	// Path is the path from root to this value, for debugging/error messages.
	// Elements are either string (field name) or int (array index).
	Path []any

	// Document is the source document index (0 for single document).
	// Used for multi-document operations.
	Document int
}

// NewCandidateNode creates a new CandidateNode wrapping the given value.
func NewCandidateNode(value any) *CandidateNode {
	return &CandidateNode{
		Value:    value,
		Path:     nil,
		Document: 0,
	}
}

// WithPath returns a new CandidateNode with the path appended.
func (n *CandidateNode) WithPath(elem any) *CandidateNode {
	newPath := make([]any, len(n.Path)+1)
	copy(newPath, n.Path)
	newPath[len(n.Path)] = elem
	return &CandidateNode{
		Value:    n.Value,
		Path:     newPath,
		Document: n.Document,
	}
}
