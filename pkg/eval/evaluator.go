package eval

import (
	"fmt"

	"github.com/huml-lang/hq/pkg/parser"
	"github.com/huml-lang/hq/pkg/types"
)

// Evaluate evaluates an hq expression against input data.
// Returns a slice of results (multiple outputs for iterators/commas).
func Evaluate(expr string, input any) ([]any, error) {
	// Parse the expression
	ast, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Create evaluation context
	ctx := types.NewContext(input)

	// Evaluate the AST
	results, err := evaluate(ast, ctx)
	if err != nil {
		return nil, err
	}

	// Extract values from CandidateNodes
	values := make([]any, len(results))
	for i, node := range results {
		values[i] = node.Value
	}

	return values, nil
}

// evaluate recursively evaluates an AST node.
func evaluate(node parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	switch n := node.(type) {
	case *parser.IdentityNode:
		return evalIdentity(ctx)

	case *parser.LiteralNode:
		return evalLiteral(n, ctx)

	case *parser.FieldAccessNode:
		return evalFieldAccess(n, ctx)

	case *parser.IndexAccessNode:
		return evalIndexAccess(n, ctx)

	case *parser.SliceNode:
		return evalSlice(n, ctx)

	case *parser.IteratorNode:
		return evalIterator(n, ctx)

	case *parser.PipeNode:
		return evalPipe(n, ctx)

	case *parser.CommaNode:
		return evalComma(n, ctx)

	case *parser.BinaryOpNode:
		return evalBinaryOp(n, ctx)

	case *parser.UnaryOpNode:
		return evalUnaryOp(n, ctx)

	case *parser.FunctionCallNode:
		return evalFunctionCall(n, ctx)

	case *parser.ArrayConstructNode:
		return evalArrayConstruct(n, ctx)

	case *parser.ObjectConstructNode:
		return evalObjectConstruct(n, ctx)

	case *parser.VariableNode:
		return evalVariable(n, ctx)

	case *parser.AlternativeNode:
		return evalAlternative(n, ctx)

	case *parser.ConditionalNode:
		return evalConditional(n, ctx)

	case *parser.VariableBindNode:
		return evalVariableBind(n, ctx)

	case *parser.RecursiveDescentNode:
		return evalRecursiveDescent(n, ctx)

	case *parser.OptionalNode:
		return evalOptional(n, ctx)

	case *parser.TryCatchNode:
		return evalTryCatch(n, ctx)

	default:
		return nil, fmt.Errorf("unimplemented expression type: %T", node)
	}
}

// evalIdentity returns the current matching nodes unchanged.
func evalIdentity(ctx *types.Context) ([]*types.CandidateNode, error) {
	return ctx.MatchingNodes, nil
}

// evalLiteral returns a literal value.
func evalLiteral(n *parser.LiteralNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	return []*types.CandidateNode{types.NewCandidateNode(n.Value)}, nil
}

// evalFieldAccess evaluates field access (.foo or .["key"]).
func evalFieldAccess(n *parser.FieldAccessNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// First evaluate what we're accessing from
	var sources []*types.CandidateNode
	var err error

	if n.From != nil {
		sources, err = evaluate(n.From, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		sources = ctx.MatchingNodes
	}

	// Access the field from each source
	var results []*types.CandidateNode
	for _, source := range sources {
		value := accessField(source.Value, n.Field)
		results = append(results, source.WithPath(n.Field))
		results[len(results)-1].Value = value
	}

	return results, nil
}

// accessField gets a field from a value, returning null for missing/invalid.
func accessField(value any, field string) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		if val, ok := v[field]; ok {
			return val
		}
		return nil
	default:
		return nil
	}
}

// evalIndexAccess evaluates array index access (.[n]).
func evalIndexAccess(n *parser.IndexAccessNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// First evaluate what we're accessing from
	var sources []*types.CandidateNode
	var err error

	if n.From != nil {
		sources, err = evaluate(n.From, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		sources = ctx.MatchingNodes
	}

	// Access the index from each source
	var results []*types.CandidateNode
	for _, source := range sources {
		value := accessIndex(source.Value, n.Index)
		newNode := source.WithPath(n.Index)
		newNode.Value = value
		results = append(results, newNode)
	}

	return results, nil
}

// accessIndex gets an element from an array, returning null for out-of-bounds.
func accessIndex(value any, index int) any {
	if value == nil {
		return nil
	}

	arr, ok := value.([]any)
	if !ok {
		return nil
	}

	// Handle negative indices
	if index < 0 {
		index = len(arr) + index
	}

	if index < 0 || index >= len(arr) {
		return nil
	}

	return arr[index]
}

// evalSlice evaluates array/string slicing (.[start:end]).
func evalSlice(n *parser.SliceNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// First evaluate what we're slicing
	var sources []*types.CandidateNode
	var err error

	if n.From != nil {
		sources, err = evaluate(n.From, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		sources = ctx.MatchingNodes
	}

	var results []*types.CandidateNode
	for _, source := range sources {
		value := sliceValue(source.Value, n.Start, n.End)
		results = append(results, types.NewCandidateNode(value))
	}

	return results, nil
}

// sliceValue slices an array or string.
func sliceValue(value any, start, end *int) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []any:
		length := len(v)
		s, e := resolveSliceBounds(start, end, length)
		if s >= e || s >= length || e < 0 {
			return []any{}
		}
		if s < 0 {
			s = 0
		}
		if e > length {
			e = length
		}
		return v[s:e]

	case string:
		length := len(v)
		s, e := resolveSliceBounds(start, end, length)
		if s >= e || s >= length || e < 0 {
			return ""
		}
		if s < 0 {
			s = 0
		}
		if e > length {
			e = length
		}
		return v[s:e]

	default:
		return nil
	}
}

// resolveSliceBounds resolves optional start/end bounds.
func resolveSliceBounds(start, end *int, length int) (int, int) {
	s := 0
	e := length

	if start != nil {
		s = *start
		if s < 0 {
			s = length + s
		}
	}

	if end != nil {
		e = *end
		if e < 0 {
			e = length + e
		}
	}

	return s, e
}

// evalIterator evaluates the iterator operator (.[]).
func evalIterator(n *parser.IteratorNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// First evaluate what we're iterating
	var sources []*types.CandidateNode
	var err error

	if n.From != nil {
		sources, err = evaluate(n.From, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		sources = ctx.MatchingNodes
	}

	// Iterate each source
	var results []*types.CandidateNode
	for _, source := range sources {
		items, err := iterateValue(source)
		if err != nil {
			return nil, err
		}
		results = append(results, items...)
	}

	return results, nil
}

// iterateValue returns all elements of an array or values of an object.
func iterateValue(node *types.CandidateNode) ([]*types.CandidateNode, error) {
	if node.Value == nil {
		return nil, fmt.Errorf("cannot iterate over null")
	}

	switch v := node.Value.(type) {
	case []any:
		results := make([]*types.CandidateNode, len(v))
		for i, elem := range v {
			newNode := node.WithPath(i)
			newNode.Value = elem
			results[i] = newNode
		}
		return results, nil

	case map[string]any:
		results := make([]*types.CandidateNode, 0, len(v))
		for k, val := range v {
			newNode := node.WithPath(k)
			newNode.Value = val
			results = append(results, newNode)
		}
		return results, nil

	default:
		return nil, fmt.Errorf("cannot iterate over %T", node.Value)
	}
}

// evalPipe evaluates the pipe operator (left | right).
func evalPipe(n *parser.PipeNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate left side
	leftResults, err := evaluate(n.Left, ctx)
	if err != nil {
		return nil, err
	}

	// For each left result, evaluate right side and collect
	var results []*types.CandidateNode
	for _, leftNode := range leftResults {
		// Create new context with this single node
		newCtx := ctx.Clone()
		newCtx.SetMatchingNodes([]*types.CandidateNode{leftNode})

		// Evaluate right side
		rightResults, err := evaluate(n.Right, newCtx)
		if err != nil {
			return nil, err
		}

		results = append(results, rightResults...)
	}

	return results, nil
}

// evalComma evaluates the comma operator (a, b).
func evalComma(n *parser.CommaNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, expr := range n.Expressions {
		exprResults, err := evaluate(expr, ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, exprResults...)
	}

	return results, nil
}

// evalBinaryOp evaluates binary operators (+, -, *, /, ==, etc.).
func evalBinaryOp(n *parser.BinaryOpNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate both sides
	leftResults, err := evaluate(n.Left, ctx)
	if err != nil {
		return nil, err
	}

	rightResults, err := evaluate(n.Right, ctx)
	if err != nil {
		return nil, err
	}

	// For simplicity, use first result from each side
	// (Full implementation would handle multiple outputs)
	if len(leftResults) == 0 || len(rightResults) == 0 {
		return nil, fmt.Errorf("empty operand for %s", n.Op)
	}

	left := leftResults[0].Value
	right := rightResults[0].Value

	result, err := applyBinaryOp(n.Op, left, right)
	if err != nil {
		return nil, err
	}

	return []*types.CandidateNode{types.NewCandidateNode(result)}, nil
}

// applyBinaryOp applies a binary operator to two values.
func applyBinaryOp(op string, left, right any) (any, error) {
	switch op {
	case "+":
		return add(left, right)
	case "-":
		return subtract(left, right)
	case "*":
		return multiply(left, right)
	case "/":
		return divide(left, right)
	case "%":
		return modulo(left, right)
	case "==":
		return equals(left, right), nil
	case "!=":
		return !equals(left, right), nil
	case "<":
		return lessThan(left, right)
	case ">":
		return greaterThan(left, right)
	case "<=":
		lt, err := lessThan(left, right)
		if err != nil {
			return nil, err
		}
		return lt || equals(left, right), nil
	case ">=":
		gt, err := greaterThan(left, right)
		if err != nil {
			return nil, err
		}
		return gt || equals(left, right), nil
	case "and":
		return isTruthy(left) && isTruthy(right), nil
	case "or":
		return isTruthy(left) || isTruthy(right), nil
	default:
		return nil, fmt.Errorf("unknown operator: %s", op)
	}
}

// add handles addition of numbers and string concatenation.
func add(left, right any) (any, error) {
	// null is identity for addition (jq compatibility)
	if left == nil {
		return right, nil
	}
	if right == nil {
		return left, nil
	}

	// String concatenation
	if ls, ok := left.(string); ok {
		if rs, ok := right.(string); ok {
			return ls + rs, nil
		}
	}

	// Numeric addition
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln + rn, nil
	}

	// Array concatenation
	if la, ok := left.([]any); ok {
		if ra, ok := right.([]any); ok {
			result := make([]any, len(la)+len(ra))
			copy(result, la)
			copy(result[len(la):], ra)
			return result, nil
		}
	}

	// Object merge
	if lm, ok := left.(map[string]any); ok {
		if rm, ok := right.(map[string]any); ok {
			result := make(map[string]any)
			for k, v := range lm {
				result[k] = v
			}
			for k, v := range rm {
				result[k] = v
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("cannot add %T and %T", left, right)
}

func subtract(left, right any) (any, error) {
	// Numeric subtraction
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln - rn, nil
	}

	// Array subtraction: remove elements that match
	if la, ok := left.([]any); ok {
		if ra, ok := right.([]any); ok {
			result := make([]any, 0)
			for _, lv := range la {
				found := false
				for _, rv := range ra {
					if equals(lv, rv) {
						found = true
						break
					}
				}
				if !found {
					result = append(result, lv)
				}
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("cannot subtract %T from %T", right, left)
}

func multiply(left, right any) (any, error) {
	// Numeric multiplication
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln * rn, nil
	}

	// String repetition: "ab" * 3 = "ababab"
	if ls, ok := left.(string); ok {
		if rn, rok := toNumber(right); rok {
			n := int(rn)
			if n <= 0 {
				return "", nil
			}
			result := ""
			for i := 0; i < n; i++ {
				result += ls
			}
			return result, nil
		}
	}

	// Object deep merge
	if lm, ok := left.(map[string]any); ok {
		if rm, ok := right.(map[string]any); ok {
			return deepMerge(lm, rm), nil
		}
	}

	return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
}

// deepMerge recursively merges two objects.
func deepMerge(base, overlay map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		if baseV, exists := result[k]; exists {
			// If both are objects, recursively merge
			if baseObj, ok := baseV.(map[string]any); ok {
				if overlayObj, ok := v.(map[string]any); ok {
					result[k] = deepMerge(baseObj, overlayObj)
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}

func divide(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		if rn == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return ln / rn, nil
	}
	return nil, fmt.Errorf("cannot divide %T by %T", left, right)
}

func modulo(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		if rn == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return float64(int64(ln) % int64(rn)), nil
	}
	return nil, fmt.Errorf("cannot modulo %T by %T", left, right)
}

func lessThan(left, right any) (bool, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln < rn, nil
	}
	if ls, ok := left.(string); ok {
		if rs, ok := right.(string); ok {
			return ls < rs, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func greaterThan(left, right any) (bool, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln > rn, nil
	}
	if ls, ok := left.(string); ok {
		if rs, ok := right.(string); ok {
			return ls > rs, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

// toNumber converts a value to float64 if possible.
func toNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// equals checks if two values are equal.
func equals(left, right any) bool {
	// Handle nil
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	// Same type comparisons
	switch l := left.(type) {
	case float64:
		if r, ok := toNumber(right); ok {
			return l == r
		}
	case int:
		if r, ok := toNumber(right); ok {
			return float64(l) == r
		}
	case int64:
		if r, ok := toNumber(right); ok {
			return float64(l) == r
		}
	case string:
		if r, ok := right.(string); ok {
			return l == r
		}
	case bool:
		if r, ok := right.(bool); ok {
			return l == r
		}
	}

	return false
}

// isTruthy checks if a value is truthy (jq semantics: null and false are falsy).
func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

// evalOptional evaluates the optional operator (?).
// It suppresses errors and returns empty instead of errors/null.
func evalOptional(n *parser.OptionalNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	results, err := evaluate(n.Expr, ctx)
	if err != nil {
		// Suppress errors - return empty
		return []*types.CandidateNode{}, nil
	}

	// Filter out null values
	var filtered []*types.CandidateNode
	for _, r := range results {
		if r.Value != nil {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

// evalTryCatch evaluates try-catch for error handling.
func evalTryCatch(n *parser.TryCatchNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Try to evaluate the try expression
	results, err := evaluate(n.Try, ctx)
	if err == nil {
		// Success - return results
		return results, nil
	}

	// Error occurred - evaluate catch if present
	if n.Catch != nil {
		return evaluate(n.Catch, ctx)
	}

	// No catch - return empty
	return []*types.CandidateNode{}, nil
}

// evalRecursiveDescent evaluates the recursive descent operator (..).
// It returns all values in the input, recursively descending into arrays and objects.
func evalRecursiveDescent(n *parser.RecursiveDescentNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		// Add the current value
		results = append(results, node)
		// Recursively add all nested values
		results = append(results, collectAllValues(node.Value)...)
	}

	return results, nil
}

// collectAllValues recursively collects all values from arrays and objects.
func collectAllValues(v any) []*types.CandidateNode {
	var results []*types.CandidateNode

	switch val := v.(type) {
	case []any:
		for _, elem := range val {
			results = append(results, types.NewCandidateNode(elem))
			results = append(results, collectAllValues(elem)...)
		}
	case map[string]any:
		for _, elem := range val {
			results = append(results, types.NewCandidateNode(elem))
			results = append(results, collectAllValues(elem)...)
		}
	}

	return results
}

// evalVariableBind evaluates variable binding (expr as $var | body).
func evalVariableBind(n *parser.VariableBindNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		// Evaluate the expression to bind
		exprCtx := ctx.Clone()
		exprCtx.SetMatchingNodes([]*types.CandidateNode{node})

		exprResults, err := evaluate(n.Expr, exprCtx)
		if err != nil {
			return nil, err
		}

		// For each result from the expression, bind to variable and evaluate body
		for _, exprResult := range exprResults {
			// Create new context with variable bound
			bodyCtx := ctx.Clone()
			bodyCtx.SetMatchingNodes([]*types.CandidateNode{node})
			bodyCtx.Variables[n.VarName] = exprResult.Value

			bodyResults, err := evaluate(n.Body, bodyCtx)
			if err != nil {
				return nil, err
			}

			results = append(results, bodyResults...)
		}
	}

	return results, nil
}

// evalConditional evaluates if-then-else.
func evalConditional(n *parser.ConditionalNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		// Evaluate condition with this node as input
		condCtx := ctx.Clone()
		condCtx.SetMatchingNodes([]*types.CandidateNode{node})

		condResults, err := evaluate(n.Condition, condCtx)
		if err != nil {
			return nil, err
		}

		// Check if condition is truthy
		var branch parser.ExpressionNode
		if len(condResults) > 0 && isTruthy(condResults[0].Value) {
			branch = n.Then
		} else {
			branch = n.Else
		}

		// Evaluate the chosen branch
		branchCtx := ctx.Clone()
		branchCtx.SetMatchingNodes([]*types.CandidateNode{node})

		branchResults, err := evaluate(branch, branchCtx)
		if err != nil {
			return nil, err
		}

		results = append(results, branchResults...)
	}

	return results, nil
}

// evalUnaryOp evaluates unary operators (not, -).
func evalUnaryOp(n *parser.UnaryOpNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	results, err := evaluate(n.Expr, ctx)
	if err != nil {
		return nil, err
	}

	var outputs []*types.CandidateNode
	for _, node := range results {
		var result any
		switch n.Op {
		case "not":
			result = !isTruthy(node.Value)
		case "-":
			if num, ok := toNumber(node.Value); ok {
				result = -num
			} else {
				return nil, fmt.Errorf("cannot negate %T", node.Value)
			}
		default:
			return nil, fmt.Errorf("unknown unary operator: %s", n.Op)
		}
		outputs = append(outputs, types.NewCandidateNode(result))
	}

	return outputs, nil
}

// evalFunctionCall evaluates a function call.
func evalFunctionCall(n *parser.FunctionCallNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	switch n.Name {
	case "length":
		return evalLength(ctx)
	case "keys":
		return evalKeys(ctx)
	case "values":
		return evalValues(ctx)
	case "type":
		return evalType(ctx)
	case "empty":
		return []*types.CandidateNode{}, nil
	case "not":
		// Postfix not - negates the input values
		var results []*types.CandidateNode
		for _, node := range ctx.MatchingNodes {
			results = append(results, types.NewCandidateNode(!isTruthy(node.Value)))
		}
		return results, nil
	case "select":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("select requires 1 argument")
		}
		return evalSelect(n.Args[0], ctx)
	case "map":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("map requires 1 argument")
		}
		return evalMap(n.Args[0], ctx)
	case "add":
		return evalAdd(ctx)
	case "first":
		if len(n.Args) == 1 {
			return evalFirstExpr(n.Args[0], ctx)
		}
		return evalFirst(ctx)
	case "last":
		if len(n.Args) == 1 {
			return evalLastExpr(n.Args[0], ctx)
		}
		return evalLast(ctx)
	case "reverse":
		return evalReverse(ctx)
	case "sort":
		return evalSort(ctx)
	case "sort_by":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("sort_by requires 1 argument")
		}
		return evalSortBy(n.Args[0], ctx)
	case "unique":
		return evalUnique(ctx)
	case "unique_by":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("unique_by requires 1 argument")
		}
		return evalUniqueBy(n.Args[0], ctx)
	case "flatten":
		depth := 1
		if len(n.Args) > 0 {
			// TODO: evaluate depth argument
		}
		return evalFlatten(ctx, depth)
	case "has":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("has requires 1 argument")
		}
		return evalHas(n.Args[0], ctx)
	case "contains":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("contains requires 1 argument")
		}
		return evalContains(n.Args[0], ctx)
	case "inside":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("inside requires 1 argument")
		}
		return evalInside(n.Args[0], ctx)
	case "numbers":
		return evalTypeFilter(ctx, "number")
	case "strings":
		return evalTypeFilter(ctx, "string")
	case "booleans":
		return evalTypeFilter(ctx, "boolean")
	case "nulls":
		return evalTypeFilter(ctx, "null")
	case "arrays":
		return evalTypeFilter(ctx, "array")
	case "objects":
		return evalTypeFilter(ctx, "object")
	case "test":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("test requires 1 argument")
		}
		return evalTest(n.Args[0], ctx)
	case "match":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("match requires 1 argument")
		}
		return evalMatch(n.Args[0], ctx)
	case "capture":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("capture requires 1 argument")
		}
		return evalCapture(n.Args[0], ctx)
	case "sub":
		if len(n.Args) != 2 {
			return nil, fmt.Errorf("sub requires 2 arguments")
		}
		return evalSub(n.Args[0], n.Args[1], ctx)
	case "gsub":
		if len(n.Args) != 2 {
			return nil, fmt.Errorf("gsub requires 2 arguments")
		}
		return evalGsub(n.Args[0], n.Args[1], ctx)
	case "error":
		if len(n.Args) == 0 {
			return nil, fmt.Errorf("error")
		}
		if len(n.Args) == 1 {
			msgResults, err := evaluate(n.Args[0], ctx)
			if err != nil {
				return nil, err
			}
			if len(msgResults) > 0 {
				if msg, ok := msgResults[0].Value.(string); ok {
					return nil, fmt.Errorf("%s", msg)
				}
			}
			return nil, fmt.Errorf("error")
		}
		return nil, fmt.Errorf("error takes 0 or 1 argument")
	case "group_by":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("group_by requires 1 argument")
		}
		return evalGroupBy(n.Args[0], ctx)
	case "map_values":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("map_values requires 1 argument")
		}
		return evalMapValues(n.Args[0], ctx)
	case "tostring":
		return evalToString(ctx)
	case "tonumber":
		return evalToNumber(ctx)
	case "split":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("split requires 1 argument")
		}
		return evalSplit(n.Args[0], ctx)
	case "join":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("join requires 1 argument")
		}
		return evalJoin(n.Args[0], ctx)
	case "ascii_downcase":
		return evalAsciiDowncase(ctx)
	case "ascii_upcase":
		return evalAsciiUpcase(ctx)
	case "startswith":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("startswith requires 1 argument")
		}
		return evalStartsWith(n.Args[0], ctx)
	case "endswith":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("endswith requires 1 argument")
		}
		return evalEndsWith(n.Args[0], ctx)
	case "ltrimstr":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("ltrimstr requires 1 argument")
		}
		return evalLtrimstr(n.Args[0], ctx)
	case "rtrimstr":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("rtrimstr requires 1 argument")
		}
		return evalRtrimstr(n.Args[0], ctx)
	case "trim":
		return evalTrim(ctx)
	case "min":
		return evalMin(ctx)
	case "max":
		return evalMax(ctx)
	case "min_by":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("min_by requires 1 argument")
		}
		return evalMinBy(n.Args[0], ctx)
	case "max_by":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("max_by requires 1 argument")
		}
		return evalMaxBy(n.Args[0], ctx)
	case "to_entries":
		return evalToEntries(ctx)
	case "from_entries":
		return evalFromEntries(ctx)
	case "with_entries":
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("with_entries requires 1 argument")
		}
		return evalWithEntries(n.Args[0], ctx)
	default:
		return nil, fmt.Errorf("unknown function: %s", n.Name)
	}
}

// evalArrayConstruct evaluates array construction [...].
func evalArrayConstruct(n *parser.ArrayConstructNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	if n.Elements == nil {
		return []*types.CandidateNode{types.NewCandidateNode([]any{})}, nil
	}

	elements, err := evaluate(n.Elements, ctx)
	if err != nil {
		return nil, err
	}

	arr := make([]any, len(elements))
	for i, elem := range elements {
		arr[i] = elem.Value
	}

	return []*types.CandidateNode{types.NewCandidateNode(arr)}, nil
}

// evalObjectConstruct evaluates object construction {...}.
func evalObjectConstruct(n *parser.ObjectConstructNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	if len(n.Fields) == 0 {
		return []*types.CandidateNode{types.NewCandidateNode(map[string]any{})}, nil
	}

	obj := make(map[string]any)

	for _, field := range n.Fields {
		// Evaluate key
		keyResults, err := evaluate(field.Key, ctx)
		if err != nil {
			return nil, err
		}
		if len(keyResults) == 0 {
			continue
		}
		key, ok := keyResults[0].Value.(string)
		if !ok {
			return nil, fmt.Errorf("object key must be a string, got %T", keyResults[0].Value)
		}

		// Evaluate value
		valueResults, err := evaluate(field.Value, ctx)
		if err != nil {
			return nil, err
		}
		if len(valueResults) == 0 {
			obj[key] = nil
		} else {
			obj[key] = valueResults[0].Value
		}
	}

	return []*types.CandidateNode{types.NewCandidateNode(obj)}, nil
}

// evalVariable evaluates a variable reference.
func evalVariable(n *parser.VariableNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	val, ok := ctx.GetVariable(n.Name)
	if !ok {
		return nil, fmt.Errorf("undefined variable: $%s", n.Name)
	}
	return []*types.CandidateNode{types.NewCandidateNode(val)}, nil
}

// evalAlternative evaluates the alternative operator (//).
func evalAlternative(n *parser.AlternativeNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	leftResults, err := evaluate(n.Left, ctx)
	if err == nil && len(leftResults) > 0 {
		// Check if result is not null/false
		for _, result := range leftResults {
			if result.Value != nil && result.Value != false {
				return leftResults, nil
			}
		}
	}

	// Fall back to right side
	return evaluate(n.Right, ctx)
}
