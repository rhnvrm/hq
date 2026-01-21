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
	// Use the unified parseExpressionTokens and ignore remaining tokens
	result, rest, err := p.parseExpressionTokens(tokens, minPrec)
	if err != nil {
		return nil, err
	}

	// Note: remaining tokens (rest) may be valid in certain contexts
	// For now, we just return what we have parsed
	_ = rest

	return result, nil
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

		// Special handling for 'as' keyword (variable binding)
		if tokVal == "as" {
			prec, _ := p.getOperatorPrecedence("as")
			if prec < minPrec {
				break
			}

			// Parse: expr as $var | body  OR  expr as {pattern} | body
			rest = rest[1:] // consume 'as'

			if len(rest) == 0 {
				return nil, nil, fmt.Errorf("expected variable after 'as'")
			}

			// Check for variable or destructuring pattern
			if p.isTokenType(rest[0], "Variable") {
				varName := rest[0].Value[1:] // Remove $
				rest = rest[1:]

				// Expect | after variable
				if len(rest) == 0 || rest[0].Value != "|" {
					return nil, nil, fmt.Errorf("expected '|' after variable binding")
				}
				rest = rest[1:] // consume '|'

				// Parse body (rest of expression)
				var body ExpressionNode
				body, rest, err = p.parseExpressionTokens(rest, 0)
				if err != nil {
					return nil, nil, err
				}

				left = &VariableBindNode{
					Expr:    left,
					VarName: varName,
					Body:    body,
				}
			} else if rest[0].Value == "{" {
				// Parse destructuring pattern {key: $var, ...}
				bindings, newRest, err := p.parseDestructurePattern(rest)
				if err != nil {
					return nil, nil, fmt.Errorf("parsing destructure pattern: %w", err)
				}
				rest = newRest

				// Expect | after pattern
				if len(rest) == 0 || rest[0].Value != "|" {
					return nil, nil, fmt.Errorf("expected '|' after destructure pattern")
				}
				rest = rest[1:] // consume '|'

				// Parse body (rest of expression)
				var body ExpressionNode
				body, rest, err = p.parseExpressionTokens(rest, 0)
				if err != nil {
					return nil, nil, err
				}

				left = &DestructureBindNode{
					Expr:     left,
					Bindings: bindings,
					Body:     body,
				}
			} else {
				return nil, nil, fmt.Errorf("expected variable after 'as', got %s", rest[0].Value)
			}
			continue
		}

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

	// Unary minus (negative number or negation)
	case tok.Value == "-":
		rest := tokens[1:]
		if len(rest) == 0 {
			return nil, nil, fmt.Errorf("unexpected end of expression after -")
		}
		// Parse the operand
		operand, rest, err := p.parsePrimary(rest)
		if err != nil {
			return nil, nil, err
		}
		return &UnaryOpNode{Op: "-", Expr: operand}, rest, nil

	// Recursive descent (..)
	case tok.Value == "..":
		return &RecursiveDescentNode{From: nil}, tokens[1:], nil

	// String literal (may contain interpolation)
	case p.isTokenType(tok, "String"):
		// Remove quotes
		s := tok.Value[1 : len(tok.Value)-1]
		// Check for interpolation \(...)
		if strings.Contains(s, `\(`) {
			return p.parseStringInterpolation(s, tokens[1:])
		}
		// Plain string - unescape
		s = unescapeString(s)
		return &LiteralNode{Value: s}, tokens[1:], nil

	// Boolean/null keywords
	case tok.Value == "true":
		return &LiteralNode{Value: true}, tokens[1:], nil
	case tok.Value == "false":
		return &LiteralNode{Value: false}, tokens[1:], nil
	case tok.Value == "null":
		return &LiteralNode{Value: nil}, tokens[1:], nil

	// Variable (may be followed by field access like $u.name)
	case p.isTokenType(tok, "Variable"):
		var node ExpressionNode = &VariableNode{Name: tok.Value[1:]}
		rest := tokens[1:]
		// Check for chained field access
		for len(rest) > 0 && rest[0].Value == "." {
			rest = rest[1:] // consume .
			if len(rest) == 0 {
				return nil, nil, fmt.Errorf("unexpected end of expression after dot")
			}
			if p.isTokenType(rest[0], "Ident") {
				node = &FieldAccessNode{Field: rest[0].Value, From: node}
				rest = rest[1:]
			} else if rest[0].Value == "[" {
				var err error
				node, rest, err = p.parseBracketAccess(node, rest)
				if err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, fmt.Errorf("expected identifier after ., got %s", rest[0].Value)
			}
		}
		return node, rest, nil

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

		// Chained dot access: .user.name (the second . starts another field)
		case tok.Value == ".":
			// Consume the .
			rest = rest[1:]
			if len(rest) == 0 {
				return nil, nil, fmt.Errorf("unexpected end of expression after dot")
			}
			// Next token should be an identifier or [
			nextTok := rest[0]
			if p.isTokenType(nextTok, "Ident") {
				node = &FieldAccessNode{Field: nextTok.Value, From: node}
				rest = rest[1:]
			} else if nextTok.Value == "[" {
				var err error
				node, rest, err = p.parseBracketAccess(node, rest)
				if err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, fmt.Errorf("expected identifier or [ after ., got %s", nextTok.Value)
			}

		// Optional operator: expr?
		case tok.Value == "?":
			node = &OptionalNode{Expr: node}
			rest = rest[1:]

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

	// Check for dynamic index access .[$var] or .[expr]
	// Parse the expression inside brackets
	indexExpr, rest, err := p.parseExpressionTokens(rest, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing bracket expression: %w", err)
	}

	if len(rest) == 0 || rest[0].Value != "]" {
		return nil, nil, fmt.Errorf("expected ] after bracket expression")
	}

	return &DynamicIndexNode{Index: indexExpr, From: from}, rest[1:], nil
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
			// Computed key: {(.expr): value}
			tokens = tokens[1:] // consume (

			// Find matching )
			depth := 1
			end := 0
			for i, t := range tokens {
				if t.Value == "(" {
					depth++
				} else if t.Value == ")" {
					depth--
					if depth == 0 {
						end = i
						break
					}
				}
			}
			if depth != 0 {
				return nil, fmt.Errorf("unmatched parenthesis in computed key")
			}

			// Parse key expression
			keyExpr, _, err := p.parseExpressionTokens(tokens[:end], 0)
			if err != nil {
				return nil, fmt.Errorf("parsing computed key: %w", err)
			}
			key = keyExpr
			tokens = tokens[end+1:] // skip )

			// Expect :
			if len(tokens) == 0 || tokens[0].Value != ":" {
				return nil, fmt.Errorf("expected : after computed key")
			}
			tokens = tokens[1:]
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
	case "reduce":
		return p.parseReduce(rest)
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
// Format: if COND then EXPR [elif COND then EXPR]* [else EXPR] end
func (p *Parser) parseConditional(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Parse condition (tokens start after 'if')
	condTokens, rest := p.extractUntilKeyword(tokens, "then")
	if rest == nil {
		return nil, nil, fmt.Errorf("expected 'then' after if condition")
	}

	// Skip 'then'
	rest = rest[1:]

	cond, _, err := p.parseExpressionTokens(condTokens, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing if condition: %w", err)
	}

	// Parse then branch (until 'elif', 'else', or 'end')
	thenTokens, rest, keyword := p.extractUntilKeywords(rest, []string{"elif", "else", "end"})
	if rest == nil {
		return nil, nil, fmt.Errorf("expected 'elif', 'else', or 'end' after then branch")
	}

	thenExpr, _, err := p.parseExpressionTokens(thenTokens, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing then branch: %w", err)
	}

	// Skip the keyword
	rest = rest[1:]

	var elseExpr ExpressionNode

	switch keyword {
	case "elif":
		// Recursively parse the elif as another if-then-else
		elseExpr, rest, err = p.parseConditional(rest)
		if err != nil {
			return nil, nil, err
		}
	case "else":
		// Parse else branch until 'end'
		elseTokens, rest2 := p.extractUntilKeyword(rest, "end")
		if rest2 == nil {
			return nil, nil, fmt.Errorf("expected 'end' after else branch")
		}

		elseExpr, _, err = p.parseExpressionTokens(elseTokens, 0)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing else branch: %w", err)
		}

		// Skip 'end'
		rest = rest2[1:]
	case "end":
		// No else branch, return identity
		elseExpr = &IdentityNode{}
	}

	return &ConditionalNode{
		Condition: cond,
		Then:      thenExpr,
		Else:      elseExpr,
	}, rest, nil
}

// extractUntilKeyword extracts tokens until a keyword is found (respecting nesting)
func (p *Parser) extractUntilKeyword(tokens []lexer.Token, keyword string) ([]lexer.Token, []lexer.Token) {
	depth := 0
	for i, tok := range tokens {
		// Only match keyword at depth 0 BEFORE adjusting depth
		if depth == 0 && tok.Value == keyword {
			return tokens[:i], tokens[i:]
		}

		// Track nesting depth
		if tok.Value == "if" || tok.Value == "try" {
			depth++
		} else if tok.Value == "end" {
			depth--
		}
		// Parentheses/brackets are handled separately
		if tok.Value == "(" || tok.Value == "[" || tok.Value == "{" {
			depth++
		} else if tok.Value == ")" || tok.Value == "]" || tok.Value == "}" {
			depth--
		}
	}
	return nil, nil
}

// extractUntilKeywords extracts tokens until one of the keywords is found
func (p *Parser) extractUntilKeywords(tokens []lexer.Token, keywords []string) ([]lexer.Token, []lexer.Token, string) {
	depth := 0
	for i, tok := range tokens {
		// Only match keywords at depth 0 BEFORE adjusting depth
		if depth == 0 {
			for _, kw := range keywords {
				if tok.Value == kw {
					return tokens[:i], tokens[i:], kw
				}
			}
		}

		// Track nesting depth
		if tok.Value == "if" || tok.Value == "try" {
			depth++
		} else if tok.Value == "end" {
			depth--
		}
		// Parentheses/brackets are handled separately
		if tok.Value == "(" || tok.Value == "[" || tok.Value == "{" {
			depth++
		} else if tok.Value == ")" || tok.Value == "]" || tok.Value == "}" {
			depth--
		}
	}
	return nil, nil, ""
}

// parseDestructurePattern parses a destructuring pattern like {x: $x, y: $y}
// Returns a map from field name to variable name (without $)
func (p *Parser) parseDestructurePattern(tokens []lexer.Token) (map[string]string, []lexer.Token, error) {
	if len(tokens) == 0 || tokens[0].Value != "{" {
		return nil, nil, fmt.Errorf("expected '{' at start of destructure pattern")
	}
	rest := tokens[1:] // consume '{'

	bindings := make(map[string]string)

	for {
		if len(rest) == 0 {
			return nil, nil, fmt.Errorf("unexpected end of destructure pattern")
		}

		// Check for closing brace
		if rest[0].Value == "}" {
			rest = rest[1:]
			break
		}

		// Parse field name
		var fieldName string
		if p.isTokenType(rest[0], "Ident") {
			fieldName = rest[0].Value
			rest = rest[1:]
		} else if p.isTokenType(rest[0], "String") {
			if len(rest[0].Value) < 2 {
				return nil, nil, fmt.Errorf("invalid string in destructure pattern")
			}
			fieldName = rest[0].Value[1 : len(rest[0].Value)-1]
			rest = rest[1:]
		} else {
			return nil, nil, fmt.Errorf("expected field name in destructure pattern, got %s", rest[0].Value)
		}

		// Expect colon
		if len(rest) == 0 || rest[0].Value != ":" {
			return nil, nil, fmt.Errorf("expected ':' after field name in destructure pattern")
		}
		rest = rest[1:] // consume ':'

		// Expect variable
		if len(rest) == 0 || !p.isTokenType(rest[0], "Variable") {
			return nil, nil, fmt.Errorf("expected variable after ':' in destructure pattern")
		}
		varName := rest[0].Value[1:] // Remove $
		rest = rest[1:]

		bindings[fieldName] = varName

		// Check for comma or closing brace
		if len(rest) > 0 && rest[0].Value == "," {
			rest = rest[1:] // consume ','
		}
	}

	return bindings, rest, nil
}

// parseReduce parses reduce expression
// Format: reduce EXPR as $VAR (INIT; UPDATE)
func (p *Parser) parseReduce(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Parse iterator expression until 'as'
	exprTokens, rest := p.extractUntilKeyword(tokens, "as")
	if rest == nil {
		return nil, nil, fmt.Errorf("expected 'as' in reduce expression")
	}

	// Parse the iterator expression
	expr, _, err := p.parseExpressionTokens(exprTokens, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reduce iterator: %w", err)
	}

	// Skip 'as'
	rest = rest[1:]

	// Expect variable
	if len(rest) == 0 || !p.isTokenType(rest[0], "Variable") {
		return nil, nil, fmt.Errorf("expected variable after 'as' in reduce")
	}
	varName := rest[0].Value[1:] // Remove $
	rest = rest[1:]

	// Expect (
	if len(rest) == 0 || rest[0].Value != "(" {
		return nil, nil, fmt.Errorf("expected '(' after variable in reduce")
	}
	rest = rest[1:]

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
		return nil, nil, fmt.Errorf("unmatched parenthesis in reduce")
	}

	// Split by ; to get init and update
	inner := rest[:end]
	var initTokens, updateTokens []lexer.Token

	semicolonDepth := 0
	splitIdx := -1
	for i, tok := range inner {
		if tok.Value == "(" || tok.Value == "[" || tok.Value == "{" {
			semicolonDepth++
		} else if tok.Value == ")" || tok.Value == "]" || tok.Value == "}" {
			semicolonDepth--
		} else if semicolonDepth == 0 && tok.Value == ";" {
			splitIdx = i
			break
		}
	}

	if splitIdx == -1 {
		return nil, nil, fmt.Errorf("expected ';' in reduce (init; update)")
	}

	initTokens = inner[:splitIdx]
	updateTokens = inner[splitIdx+1:]

	// Parse init
	initExpr, _, err := p.parseExpressionTokens(initTokens, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reduce init: %w", err)
	}

	// Parse update
	updateExpr, _, err := p.parseExpressionTokens(updateTokens, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reduce update: %w", err)
	}

	return &ReduceNode{
		Expr:    expr,
		VarName: varName,
		Init:    initExpr,
		Update:  updateExpr,
	}, rest[end+1:], nil
}

// parseTryCatch parses try-catch
// Format: try EXPR [catch EXPR]
func (p *Parser) parseTryCatch(tokens []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	// Parse try expression - find the extent until 'catch' or end of expression
	// We need to handle nesting properly
	tryTokens, rest, keyword := p.extractUntilKeywords(tokens, []string{"catch"})

	var tryExpr ExpressionNode
	var err error

	if tryTokens == nil {
		// No catch, whole thing is try expression
		tryExpr, rest, err = p.parseExpressionTokens(tokens, 0)
		if err != nil {
			return nil, nil, err
		}
		return &TryCatchNode{
			Try:   tryExpr,
			Catch: nil,
		}, rest, nil
	}

	// Parse try part
	tryExpr, _, err = p.parseExpressionTokens(tryTokens, 0)
	if err != nil {
		return nil, nil, err
	}

	if keyword != "catch" {
		return &TryCatchNode{
			Try:   tryExpr,
			Catch: nil,
		}, rest, nil
	}

	// Skip 'catch'
	rest = rest[1:]

	// Parse catch expression
	catchExpr, rest, err := p.parseExpressionTokens(rest, 0)
	if err != nil {
		return nil, nil, err
	}

	return &TryCatchNode{
		Try:   tryExpr,
		Catch: catchExpr,
	}, rest, nil
}

// getOperatorPrecedence returns the precedence and right-associativity of an operator
func (p *Parser) getOperatorPrecedence(op string) (int, bool) {
	switch op {
	case "=", "|=", "+=", "-=", "*=", "//=":
		return 0, true // Assignment has lowest precedence, right-associative
	case "|":
		return 1, false
	case ",":
		return 2, false
	case "as":
		return 3, true // 'as' binds tighter than comma, looser than pipe
	case "//":
		return 4, true
	case "or":
		return 5, false
	case "and":
		return 6, false
	case "==", "!=":
		return 7, false
	case "<", ">", "<=", ">=":
		return 8, false
	case "+", "-":
		return 9, false
	case "*", "/", "%":
		return 10, false
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
	case "=", "|=", "+=", "-=", "*=", "//=":
		return &AssignNode{Path: left, Op: op, Value: right}
	case "as":
		// For "expr as $var | body", right should be parsed as "var | body"
		// But the way we parse, 'right' is just the variable part
		// We need to handle this specially in parseExpressionTokens
		return &BinaryOpNode{Op: "as", Left: left, Right: right}
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

// parseStringInterpolation parses a string containing \(...) interpolations
func (p *Parser) parseStringInterpolation(s string, rest []lexer.Token) (ExpressionNode, []lexer.Token, error) {
	var parts []StringPart

	for len(s) > 0 {
		// Find next \(
		idx := strings.Index(s, `\(`)
		if idx == -1 {
			// No more interpolations - rest is literal
			if len(s) > 0 {
				parts = append(parts, StringPart{Literal: unescapeString(s)})
			}
			break
		}

		// Add literal part before \(
		if idx > 0 {
			parts = append(parts, StringPart{Literal: unescapeString(s[:idx])})
		}

		// Find matching )
		s = s[idx+2:] // Skip \(
		depth := 1
		end := 0
		for i := 0; i < len(s); i++ {
			if s[i] == '(' {
				depth++
			} else if s[i] == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			} else if s[i] == '\\' && i+1 < len(s) {
				i++ // Skip escaped character
			}
		}
		if depth != 0 {
			return nil, nil, fmt.Errorf("unmatched \\( in string interpolation")
		}

		// Parse the expression inside \(...)
		exprStr := s[:end]
		expr, err := p.Parse(exprStr)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing interpolated expression %q: %w", exprStr, err)
		}
		parts = append(parts, StringPart{Expr: expr})

		// Continue after the )
		s = s[end+1:]
	}

	// If there's only one literal part with no expressions, return plain literal
	if len(parts) == 1 && parts[0].Expr == nil {
		return &LiteralNode{Value: parts[0].Literal}, rest, nil
	}

	return &StringInterpolationNode{Parts: parts}, rest, nil
}

// Global default parser instance
var defaultParser = New()

// Parse parses an hq expression using the default parser.
func Parse(expr string) (ExpressionNode, error) {
	return defaultParser.Parse(expr)
}
