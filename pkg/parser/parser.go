package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

// Parser is the hq expression parser.
type Parser struct {
	lexer *lexer.StatefulDefinition
}

// New creates a new Parser.
func New() *Parser {
	return &Parser{
		lexer: hqLexer,
	}
}

// Parse parses an hq expression string into an AST.
func (p *Parser) Parse(expr string) (ExpressionNode, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Tokenize the expression
	lex, err := p.lexer.LexString("", expr)
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	// Collect tokens (filtering whitespace)
	var tokens []lexer.Token
	for {
		tok, err := lex.Next()
		if err != nil {
			return nil, fmt.Errorf("tokenize error: %w", err)
		}
		if tok.EOF() {
			break
		}
		// Skip whitespace tokens
		if p.lexer.Symbols()["Whitespace"] == tok.Type {
			continue
		}
		tokens = append(tokens, tok)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty expression after tokenization")
	}

	// Parse tokens into AST
	return p.parseExpression(tokens, 0)
}

// parseExpression is the main parsing entry point.
// It handles pipe operator (lowest precedence) and dispatches to sub-parsers.
func (p *Parser) parseExpression(tokens []lexer.Token, minPrec int) (ExpressionNode, error) {
	// Parse the left-hand side
	left, rest, err := p.parsePrimary(tokens)
	if err != nil {
		return nil, err
	}

	// Handle binary operators with precedence climbing
	for len(rest) > 0 {
		tok := rest[0]
		tokVal := tok.Value

		// Check if it's an operator and get its precedence
		prec, rightAssoc := p.getOperatorPrecedence(tokVal)
		if prec < minPrec {
			break
		}

		// Consume operator
		rest = rest[1:]

		// Adjust precedence for right associativity
		nextMinPrec := prec
		if !rightAssoc {
			nextMinPrec = prec + 1
		}

		// Parse right-hand side
		var right ExpressionNode
		right, rest, err = p.parseExpressionTokens(rest, nextMinPrec)
		if err != nil {
			return nil, err
		}

		// Build node based on operator
		left = p.buildBinaryNode(tokVal, left, right)
	}

	if len(rest) > 0 {
		// Check if remaining tokens are valid continuations
		// For now, return error for unexpected tokens
		// (We'll handle this better as we add more operators)
	}

	return left, nil
}

// parseExpressionTokens parses tokens and returns remaining tokens.
func (p *Parser) parseExpressionTokens(tokens []lexer.Token, minPrec int) (ExpressionNode, []lexer.Token, error) {
	left, rest, err := p.parsePrimary(tokens)
	if err != nil {
		return nil, nil, err
	}

	for len(rest) > 0 {
		tok := rest[0]
		tokVal := tok.Value

		prec, rightAssoc := p.getOperatorPrecedence(tokVal)
		if prec < minPrec {
			break
		}

		rest = rest[1:]

		nextMinPrec := prec
		if !rightAssoc {
			nextMinPrec = prec + 1
		}

		var right ExpressionNode
		right, rest, err = p.parseExpressionTokens(rest, nextMinPrec)
		if err != nil {
			return nil, nil, err
		}

		left = p.buildBinaryNode(tokVal, left, right)
	}

	return left, rest, nil
}

// parsePrimary parses a primary expression (identity, literal, field access, etc.)
func (p *Parser) parsePrimary(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("unexpected end of expression")
	}

	tok := tokens[0]

	switch {
	// Identity or field access starting with .
	case tok.Value == ".":
		return p.parseDotExpression(tokens)

	// Parenthesized expression
	case tok.Value == "(":
		return p.parseParenExpr(tokens)

	// Array construction
	case tok.Value == "[":
		return p.parseArrayConstruct(tokens)

	// Object construction
	case tok.Value == "{":
		return p.parseObjectConstruct(tokens)

	// Number literal
	case p.isTokenType(tok, "Number"):
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid number: %s", tok.Value)
		}
		return &LiteralNode{Value: val}, tokens[1:], nil

	// String literal
	case p.isTokenType(tok, "String"):
		// Remove quotes and unescape
		s := tok.Value[1 : len(tok.Value)-1]
		s = unescapeString(s)
		return &LiteralNode{Value: s}, tokens[1:], nil

	// Boolean/null keywords
	case tok.Value == "true":
		return &LiteralNode{Value: true}, tokens[1:], nil
	case tok.Value == "false":
		return &LiteralNode{Value: false}, tokens[1:], nil
	case tok.Value == "null":
		return &LiteralNode{Value: nil}, tokens[1:], nil

	// Variable
	case p.isTokenType(tok, "Variable"):
		return &VariableNode{Name: tok.Value[1:]}, tokens[1:], nil

	// Function call or keyword
	case p.isTokenType(tok, "Ident") || p.isTokenType(tok, "Keyword"):
		return p.parseFunctionOrKeyword(tokens)

	default:
		return nil, nil, fmt.Errorf("unexpected token: %s", tok.Value)
	}
}

// parseDotExpression handles expressions starting with .
func (p *Parser) parseDotExpression(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Consume the .
	rest := tokens[1:]
	var node ExpressionNode = &IdentityNode{}

	// Check what follows the .
	for len(rest) > 0 {
		tok := rest[0]

		switch {
		// Field access: .foo
		case p.isTokenType(tok, "Ident"):
			node = &FieldAccessNode{Field: tok.Value, From: node}
			rest = rest[1:]

		// Bracket access: .[...] or .["key"]
		case tok.Value == "[":
			var err error
			node, rest, err = p.parseBracketAccess(node, rest)
			if err != nil {
				return nil, nil, err
			}

		default:
			// End of dot expression chain
			return node, rest, nil
		}
	}

	return node, rest, nil
}

// parseBracketAccess handles .[n], .["key"], .[start:end], .[]
func (p *Parser) parseBracketAccess(from ExpressionNode, tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Consume [
	rest := tokens[1:]

	if len(rest) == 0 {
		return nil, nil, fmt.Errorf("unexpected end of expression after [")
	}

	// Check for empty brackets (iterator)
	if rest[0].Value == "]" {
		return &IteratorNode{From: from}, rest[1:], nil
	}

	// Check for string key
	if p.isTokenType(rest[0], "String") {
		key := rest[0].Value[1 : len(rest[0].Value)-1]
		key = unescapeString(key)
		rest = rest[1:]
		if len(rest) == 0 || rest[0].Value != "]" {
			return nil, nil, fmt.Errorf("expected ] after bracket key")
		}
		return &FieldAccessNode{Field: key, From: from}, rest[1:], nil
	}

	// Helper to parse a possibly-negative number
	parseNumber := func(tokens []lexer.Token) (*int, []lexer.Token, bool) {
		if len(tokens) == 0 {
			return nil, tokens, false
		}

		negative := false
		rest := tokens

		// Check for minus sign
		if rest[0].Value == "-" {
			negative = true
			rest = rest[1:]
			if len(rest) == 0 {
				return nil, tokens, false
			}
		}

		// Check for number
		if !p.isTokenType(rest[0], "Number") {
			if negative {
				return nil, tokens, false // Had minus but no number
			}
			return nil, rest, false
		}

		idx, err := strconv.Atoi(rest[0].Value)
		if err != nil {
			f, _ := strconv.ParseFloat(rest[0].Value, 64)
			idx = int(f)
		}
		if negative {
			idx = -idx
		}
		return &idx, rest[1:], true
	}

	// Check for number index or slice
	var startIdx *int
	var hasStart bool

	startIdx, rest, hasStart = parseNumber(rest)

	// Check for : (slice)
	if len(rest) > 0 && rest[0].Value == ":" {
		rest = rest[1:]

		var endIdx *int
		endIdx, rest, _ = parseNumber(rest)

		if len(rest) == 0 || rest[0].Value != "]" {
			return nil, nil, fmt.Errorf("expected ] after slice")
		}
		return &SliceNode{Start: startIdx, End: endIdx, From: from}, rest[1:], nil
	}

	// Regular index access
	if hasStart {
		if len(rest) == 0 || rest[0].Value != "]" {
			return nil, nil, fmt.Errorf("expected ] after index")
		}
		return &IndexAccessNode{Index: *startIdx, From: from}, rest[1:], nil
	}

	return nil, nil, fmt.Errorf("unexpected token in bracket expression: %s", rest[0].Value)
}

// parseParenExpr handles parenthesized expressions
func (p *Parser) parseParenExpr(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Consume (
	rest := tokens[1:]

	// Find matching )
	depth := 1
	end := 0
	for i, tok := range rest {
		if tok.Value == "(" {
			depth++
		} else if tok.Value == ")" {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if depth != 0 {
		return nil, nil, fmt.Errorf("unmatched parenthesis")
	}

	// Parse inner expression
	inner := rest[:end]
	node, _, err := p.parseExpressionTokens(inner, 0)
	if err != nil {
		return nil, nil, err
	}

	return node, rest[end+1:], nil
}

// parseArrayConstruct handles array construction [...]
func (p *Parser) parseArrayConstruct(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Consume [
	rest := tokens[1:]

	// Check for empty array
	if len(rest) > 0 && rest[0].Value == "]" {
		return &ArrayConstructNode{Elements: nil}, rest[1:], nil
	}

	// Find matching ]
	depth := 1
	end := 0
	for i, tok := range rest {
		if tok.Value == "[" || tok.Value == "{" || tok.Value == "(" {
			depth++
		} else if tok.Value == "]" || tok.Value == "}" || tok.Value == ")" {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if depth != 0 {
		return nil, nil, fmt.Errorf("unmatched bracket")
	}

	// Parse inner expression
	inner := rest[:end]
	elements, _, err := p.parseExpressionTokens(inner, 0)
	if err != nil {
		return nil, nil, err
	}

	return &ArrayConstructNode{Elements: elements}, rest[end+1:], nil
}

// parseObjectConstruct handles object construction {...}
func (p *Parser) parseObjectConstruct(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Consume {
	rest := tokens[1:]

	// Check for empty object
	if len(rest) > 0 && rest[0].Value == "}" {
		return &ObjectConstructNode{Fields: nil}, rest[1:], nil
	}

	// Find matching }
	depth := 1
	end := 0
	for i, tok := range rest {
		if tok.Value == "[" || tok.Value == "{" || tok.Value == "(" {
			depth++
		} else if tok.Value == "]" || tok.Value == "}" || tok.Value == ")" {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if depth != 0 {
		return nil, nil, fmt.Errorf("unmatched brace")
	}

	// For now, parse as simple key: value pairs
	// TODO: Full implementation with shorthand, computed keys, etc.
	inner := rest[:end]
	fields, err := p.parseObjectFields(inner)
	if err != nil {
		return nil, nil, err
	}

	return &ObjectConstructNode{Fields: fields}, rest[end+1:], nil
}

// parseObjectFields parses key: value pairs in object construction
func (p *Parser) parseObjectFields(tokens []lexer.Token) ([]ObjectFieldNode, error) {
	var fields []ObjectFieldNode

	for len(tokens) > 0 {
		// Parse key
		var key ExpressionNode
		var shorthand bool

		tok := tokens[0]
		if p.isTokenType(tok, "Ident") {
			// Could be shorthand {foo} or key {foo: ...}
			if len(tokens) > 1 && tokens[1].Value == ":" {
				// Full form: foo: expr
				key = &LiteralNode{Value: tok.Value}
				tokens = tokens[2:] // skip ident and :
			} else if len(tokens) == 1 || tokens[1].Value == "," || tokens[1].Value == "}" {
				// Shorthand: {foo} means {foo: .foo}
				key = &LiteralNode{Value: tok.Value}
				shorthand = true
				tokens = tokens[1:]
			} else {
				return nil, fmt.Errorf("unexpected token after identifier in object: %s", tokens[1].Value)
			}
		} else if p.isTokenType(tok, "String") {
			// String key
			keyStr := tok.Value[1 : len(tok.Value)-1]
			key = &LiteralNode{Value: unescapeString(keyStr)}
			tokens = tokens[1:]
			if len(tokens) == 0 || tokens[0].Value != ":" {
				return nil, fmt.Errorf("expected : after string key")
			}
			tokens = tokens[1:]
		} else if tok.Value == "(" {
			// Computed key
			// TODO: implement
			return nil, fmt.Errorf("computed keys not yet implemented")
		} else {
			return nil, fmt.Errorf("unexpected token in object key: %s", tok.Value)
		}

		// Parse value (or use shorthand)
		var value ExpressionNode
		if shorthand {
			keyStr := key.(*LiteralNode).Value.(string)
			value = &FieldAccessNode{Field: keyStr, From: &IdentityNode{}}
		} else {
			// Find the extent of the value (until , or end)
			end := 0
			depth := 0
			for i, t := range tokens {
				if t.Value == "[" || t.Value == "{" || t.Value == "(" {
					depth++
				} else if t.Value == "]" || t.Value == "}" || t.Value == ")" {
					depth--
				} else if depth == 0 && t.Value == "," {
					end = i
					break
				}
				end = i + 1
			}
			var err error
			value, _, err = p.parseExpressionTokens(tokens[:end], 0)
			if err != nil {
				return nil, err
			}
			tokens = tokens[end:]
		}

		fields = append(fields, ObjectFieldNode{
			Key:       key,
			Value:     value,
			Shorthand: shorthand,
		})

		// Skip comma if present
		if len(tokens) > 0 && tokens[0].Value == "," {
			tokens = tokens[1:]
		}
	}

	return fields, nil
}

// parseFunctionOrKeyword handles function calls and keywords
func (p *Parser) parseFunctionOrKeyword(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	name := tokens[0].Value
	rest := tokens[1:]

	// Check for keywords
	switch name {
	case "if":
		return p.parseConditional(rest)
	case "try":
		return p.parseTryCatch(rest)
	case "empty":
		return &FunctionCallNode{Name: "empty", Args: nil}, rest, nil
	case "not":
		// "not" can be used as:
		// 1. Prefix: not .foo (negates .foo)
		// 2. Postfix/filter: . | not (negates input)
		// If nothing follows, treat as postfix (zero-arg function that operates on input)
		if len(rest) == 0 || rest[0].Value == "|" || rest[0].Value == ")" || rest[0].Value == "]" || rest[0].Value == "}" || rest[0].Value == "," || rest[0].Value == ";" {
			// Postfix usage - operates on input (identity)
			return &FunctionCallNode{Name: "not", Args: nil}, rest, nil
		}
		// Prefix usage - negates following expression
		expr, rest, err := p.parsePrimary(rest)
		if err != nil {
			return nil, nil, err
		}
		return &UnaryOpNode{Op: "not", Expr: expr}, rest, nil
	}

	// Check for function call with arguments
	if len(rest) > 0 && rest[0].Value == "(" {
		return p.parseFunctionCall(name, rest)
	}

	// Zero-argument function call
	return &FunctionCallNode{Name: name, Args: nil}, rest, nil
}

// parseFunctionCall parses a function call with arguments
func (p *Parser) parseFunctionCall(name string, tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Consume (
	rest := tokens[1:]

	// Find matching )
	depth := 1
	end := 0
	for i, tok := range rest {
		if tok.Value == "(" {
			depth++
		} else if tok.Value == ")" {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if depth != 0 {
		return nil, nil, fmt.Errorf("unmatched parenthesis in function call")
	}

	// Parse arguments (separated by ;)
	argTokens := rest[:end]
	var args []ExpressionNode

	if len(argTokens) > 0 {
		// Split by ; at depth 0
		var current []lexer.Token
		argDepth := 0
		for _, tok := range argTokens {
			if tok.Value == "(" || tok.Value == "[" || tok.Value == "{" {
				argDepth++
			} else if tok.Value == ")" || tok.Value == "]" || tok.Value == "}" {
				argDepth--
			}
			if argDepth == 0 && tok.Value == ";" {
				if len(current) > 0 {
					arg, _, err := p.parseExpressionTokens(current, 0)
					if err != nil {
						return nil, nil, err
					}
					args = append(args, arg)
					current = nil
				}
			} else {
				current = append(current, tok)
			}
		}
		if len(current) > 0 {
			arg, _, err := p.parseExpressionTokens(current, 0)
			if err != nil {
				return nil, nil, err
			}
			args = append(args, arg)
		}
	}

	return &FunctionCallNode{Name: name, Args: args}, rest[end+1:], nil
}

// parseConditional parses if-then-else
func (p *Parser) parseConditional(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// TODO: implement
	return nil, nil, fmt.Errorf("conditionals not yet implemented")
}

// parseTryCatch parses try-catch
func (p *Parser) parseTryCatch(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// TODO: implement
	return nil, nil, fmt.Errorf("try-catch not yet implemented")
}

// getOperatorPrecedence returns the precedence and right-associativity of an operator
func (p *Parser) getOperatorPrecedence(op string) (int, bool) {
	switch op {
	case "|":
		return 1, false
	case ",":
		return 2, false
	case "//":
		return 3, true
	case "or":
		return 4, false
	case "and":
		return 5, false
	case "==", "!=":
		return 6, false
	case "<", ">", "<=", ">=":
		return 7, false
	case "+", "-":
		return 8, false
	case "*", "/", "%":
		return 9, false
	default:
		return -1, false // Not an operator we handle here
	}
}

// buildBinaryNode creates the appropriate node for a binary operator
func (p *Parser) buildBinaryNode(op string, left, right ExpressionNode) ExpressionNode {
	switch op {
	case "|":
		return &PipeNode{Left: left, Right: right}
	case ",":
		// Flatten comma expressions
		if comma, ok := left.(*CommaNode); ok {
			return &CommaNode{Expressions: append(comma.Expressions, right)}
		}
		return &CommaNode{Expressions: []ExpressionNode{left, right}}
	case "//":
		return &AlternativeNode{Left: left, Right: right}
	default:
		return &BinaryOpNode{Op: op, Left: left, Right: right}
	}
}

// isTokenType checks if a token is of a specific type
func (p *Parser) isTokenType(tok lexer.Token, typeName string) bool {
	return p.lexer.Symbols()[typeName] == tok.Type
}

// unescapeString handles escape sequences in strings
func unescapeString(s string) string {
	s = strings.ReplaceAll(s, `\\`, "\x00") // Temp marker
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, "\x00", `\`) // Restore single backslash
	return s
}

// Global default parser instance
var defaultParser = New()

// Parse parses an hq expression using the default parser.
func Parse(expr string) (ExpressionNode, error) {
	return defaultParser.Parse(expr)
}
