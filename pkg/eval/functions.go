package eval

import (
	"fmt"
	"sort"
	"strings"

	"github.com/huml-lang/hq/pkg/parser"
	"github.com/huml-lang/hq/pkg/types"
)

// evalLength returns the length of an array, string, or object.
func evalLength(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		var length int

		switch v := node.Value.(type) {
		case []any:
			length = len(v)
		case string:
			length = len(v)
		case map[string]any:
			length = len(v)
		case nil:
			length = 0
		default:
			return nil, fmt.Errorf("cannot get length of %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(float64(length)))
	}

	return results, nil
}

// evalKeys returns the keys of an object.
func evalKeys(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		switch v := node.Value.(type) {
		case map[string]any:
			keys := make([]any, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			// Sort keys for consistent output
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].(string) < keys[j].(string)
			})
			results = append(results, types.NewCandidateNode(keys))
		case []any:
			// For arrays, return indices
			keys := make([]any, len(v))
			for i := range v {
				keys[i] = float64(i)
			}
			results = append(results, types.NewCandidateNode(keys))
		default:
			return nil, fmt.Errorf("cannot get keys of %T", node.Value)
		}
	}

	return results, nil
}

// evalValues returns the values of an object or array.
func evalValues(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		switch v := node.Value.(type) {
		case map[string]any:
			values := make([]any, 0, len(v))
			// Get keys sorted for consistent output
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				values = append(values, v[k])
			}
			results = append(results, types.NewCandidateNode(values))
		case []any:
			// For arrays, values are the elements
			results = append(results, types.NewCandidateNode(v))
		default:
			return nil, fmt.Errorf("cannot get values of %T", node.Value)
		}
	}

	return results, nil
}

// evalType returns the type name of a value.
func evalType(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		var typeName string

		switch node.Value.(type) {
		case nil:
			typeName = "null"
		case bool:
			typeName = "boolean"
		case float64, int, int64:
			typeName = "number"
		case string:
			typeName = "string"
		case []any:
			typeName = "array"
		case map[string]any:
			typeName = "object"
		default:
			typeName = "unknown"
		}

		results = append(results, types.NewCandidateNode(typeName))
	}

	return results, nil
}

// evalSelect filters values where the condition is truthy.
func evalSelect(condition parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		// Evaluate condition with this node as input
		condCtx := ctx.Clone()
		condCtx.SetMatchingNodes([]*types.CandidateNode{node})

		condResults, err := evaluate(condition, condCtx)
		if err != nil {
			return nil, err
		}

		// If condition is truthy, keep this node
		if len(condResults) > 0 && isTruthy(condResults[0].Value) {
			results = append(results, node)
		}
	}

	return results, nil
}

// evalMap applies an expression to each element of an array.
func evalMap(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("map requires array input, got %T", node.Value)
		}

		mapped := make([]any, 0, len(arr))
		for _, elem := range arr {
			elemCtx := ctx.Clone()
			elemCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(elem)})

			elemResults, err := evaluate(expr, elemCtx)
			if err != nil {
				return nil, err
			}

			for _, r := range elemResults {
				mapped = append(mapped, r.Value)
			}
		}

		results = append(results, types.NewCandidateNode(mapped))
	}

	return results, nil
}

// evalAdd sums an array of numbers or concatenates strings/arrays.
func evalAdd(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("add requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		// Determine type from first element
		switch arr[0].(type) {
		case float64, int, int64:
			sum := 0.0
			for _, elem := range arr {
				if n, ok := toNumber(elem); ok {
					sum += n
				} else {
					return nil, fmt.Errorf("add: cannot add %T", elem)
				}
			}
			results = append(results, types.NewCandidateNode(sum))

		case string:
			var builder strings.Builder
			for _, elem := range arr {
				if s, ok := elem.(string); ok {
					builder.WriteString(s)
				} else {
					return nil, fmt.Errorf("add: cannot concatenate %T", elem)
				}
			}
			results = append(results, types.NewCandidateNode(builder.String()))

		case []any:
			var result []any
			for _, elem := range arr {
				if a, ok := elem.([]any); ok {
					result = append(result, a...)
				} else {
					return nil, fmt.Errorf("add: cannot concatenate %T", elem)
				}
			}
			results = append(results, types.NewCandidateNode(result))

		default:
			return nil, fmt.Errorf("add: unsupported type %T", arr[0])
		}
	}

	return results, nil
}

// evalFirst returns the first element of an array.
func evalFirst(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("first requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
		} else {
			results = append(results, types.NewCandidateNode(arr[0]))
		}
	}

	return results, nil
}

// evalLast returns the last element of an array.
func evalLast(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("last requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
		} else {
			results = append(results, types.NewCandidateNode(arr[len(arr)-1]))
		}
	}

	return results, nil
}

// evalReverse reverses an array.
func evalReverse(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("reverse requires array input, got %T", node.Value)
		}

		reversed := make([]any, len(arr))
		for i, v := range arr {
			reversed[len(arr)-1-i] = v
		}

		results = append(results, types.NewCandidateNode(reversed))
	}

	return results, nil
}

// evalSort sorts an array.
func evalSort(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("sort requires array input, got %T", node.Value)
		}

		// Copy array
		sorted := make([]any, len(arr))
		copy(sorted, arr)

		// Sort based on type of first element
		sort.Slice(sorted, func(i, j int) bool {
			return compareValues(sorted[i], sorted[j]) < 0
		})

		results = append(results, types.NewCandidateNode(sorted))
	}

	return results, nil
}

// compareValues compares two values for sorting.
func compareValues(a, b any) int {
	// Nulls first
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Numbers
	if an, aok := toNumber(a); aok {
		if bn, bok := toNumber(b); bok {
			if an < bn {
				return -1
			}
			if an > bn {
				return 1
			}
			return 0
		}
	}

	// Strings
	if as, aok := a.(string); aok {
		if bs, bok := b.(string); bok {
			if as < bs {
				return -1
			}
			if as > bs {
				return 1
			}
			return 0
		}
	}

	// Booleans (false < true)
	if ab, aok := a.(bool); aok {
		if bb, bok := b.(bool); bok {
			if !ab && bb {
				return -1
			}
			if ab && !bb {
				return 1
			}
			return 0
		}
	}

	return 0
}

// evalUnique removes duplicate elements from an array.
func evalUnique(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("unique requires array input, got %T", node.Value)
		}

		seen := make(map[string]bool)
		var unique []any

		for _, elem := range arr {
			key := fmt.Sprintf("%v", elem)
			if !seen[key] {
				seen[key] = true
				unique = append(unique, elem)
			}
		}

		results = append(results, types.NewCandidateNode(unique))
	}

	return results, nil
}

// evalFlatten flattens nested arrays.
func evalFlatten(ctx *types.Context, depth int) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("flatten requires array input, got %T", node.Value)
		}

		flattened := flattenArray(arr, depth)
		results = append(results, types.NewCandidateNode(flattened))
	}

	return results, nil
}

// flattenArray recursively flattens an array.
func flattenArray(arr []any, depth int) []any {
	var result []any

	for _, elem := range arr {
		if inner, ok := elem.([]any); ok && depth > 0 {
			result = append(result, flattenArray(inner, depth-1)...)
		} else {
			result = append(result, elem)
		}
	}

	return result
}

// evalHas checks if an object has a key.
func evalHas(keyExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate key expression
	keyResults, err := evaluate(keyExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(keyResults) == 0 {
		return nil, fmt.Errorf("has: key expression produced no value")
	}
	key, ok := keyResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("has: key must be a string, got %T", keyResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		switch v := node.Value.(type) {
		case map[string]any:
			_, exists := v[key]
			results = append(results, types.NewCandidateNode(exists))
		default:
			results = append(results, types.NewCandidateNode(false))
		}
	}

	return results, nil
}

// evalContains checks if a value contains another.
func evalContains(argExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate argument
	argResults, err := evaluate(argExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(argResults) == 0 {
		return nil, fmt.Errorf("contains: argument produced no value")
	}
	arg := argResults[0].Value

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		switch v := node.Value.(type) {
		case string:
			if s, ok := arg.(string); ok {
				results = append(results, types.NewCandidateNode(strings.Contains(v, s)))
			} else {
				results = append(results, types.NewCandidateNode(false))
			}
		case []any:
			found := false
			for _, elem := range v {
				if equals(elem, arg) {
					found = true
					break
				}
			}
			results = append(results, types.NewCandidateNode(found))
		default:
			results = append(results, types.NewCandidateNode(equals(node.Value, arg)))
		}
	}

	return results, nil
}

// evalSplit splits a string by a delimiter.
func evalSplit(delimExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate delimiter
	delimResults, err := evaluate(delimExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(delimResults) == 0 {
		return nil, fmt.Errorf("split: delimiter produced no value")
	}
	delim, ok := delimResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("split: delimiter must be a string, got %T", delimResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("split: input must be a string, got %T", node.Value)
		}

		parts := strings.Split(s, delim)
		arr := make([]any, len(parts))
		for i, p := range parts {
			arr[i] = p
		}

		results = append(results, types.NewCandidateNode(arr))
	}

	return results, nil
}

// evalJoin joins an array of strings.
func evalJoin(delimExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate delimiter
	delimResults, err := evaluate(delimExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(delimResults) == 0 {
		return nil, fmt.Errorf("join: delimiter produced no value")
	}
	delim, ok := delimResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("join: delimiter must be a string, got %T", delimResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("join: input must be an array, got %T", node.Value)
		}

		parts := make([]string, len(arr))
		for i, elem := range arr {
			if s, ok := elem.(string); ok {
				parts[i] = s
			} else {
				parts[i] = fmt.Sprintf("%v", elem)
			}
		}

		results = append(results, types.NewCandidateNode(strings.Join(parts, delim)))
	}

	return results, nil
}

// evalAsciiDowncase converts a string to lowercase.
func evalAsciiDowncase(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("ascii_downcase: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.ToLower(s)))
	}

	return results, nil
}

// evalAsciiUpcase converts a string to uppercase.
func evalAsciiUpcase(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("ascii_upcase: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.ToUpper(s)))
	}

	return results, nil
}

// evalStartsWith checks if a string starts with a prefix.
func evalStartsWith(prefixExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate prefix
	prefixResults, err := evaluate(prefixExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(prefixResults) == 0 {
		return nil, fmt.Errorf("startswith: prefix produced no value")
	}
	prefix, ok := prefixResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("startswith: prefix must be a string, got %T", prefixResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("startswith: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.HasPrefix(s, prefix)))
	}

	return results, nil
}

// evalEndsWith checks if a string ends with a suffix.
func evalEndsWith(suffixExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate suffix
	suffixResults, err := evaluate(suffixExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(suffixResults) == 0 {
		return nil, fmt.Errorf("endswith: suffix produced no value")
	}
	suffix, ok := suffixResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("endswith: suffix must be a string, got %T", suffixResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("endswith: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.HasSuffix(s, suffix)))
	}

	return results, nil
}
