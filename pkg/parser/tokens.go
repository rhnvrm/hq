// Package parser provides the expression parser for hq.
// It uses Participle for lexing and builds an AST for evaluation.
package parser

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer for hq expressions
var hqLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Whitespace (skip)
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},

	// Keywords (must come before Ident)
	{Name: "Keyword", Pattern: `\b(if|then|elif|else|end|as|and|or|not|true|false|null|try|catch|reduce|foreach|def|empty)\b`},

	// Operators (multi-char first)
	{Name: "Operator", Pattern: `==|!=|<=|>=|\|=|\+=|-=|\*=|//=|//|\.\.|<|>|\||\+|-|\*|/|%|=`},

	// Punctuation
	{Name: "Punct", Pattern: `[.,;:?\[\]{}()]`},

	// Numbers (including negative and float)
	{Name: "Number", Pattern: `-?[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?`},

	// String (double-quoted with escapes)
	{Name: "String", Pattern: `"([^"\\]|\\.)*"`},

	// Variable ($name)
	{Name: "Variable", Pattern: `\$[a-zA-Z_][a-zA-Z0-9_]*`},

	// Identifier (field names, function names)
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
})
