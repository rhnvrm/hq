package eval

import (
	"fmt"
	"regexp"
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
			return nil, fmt.Errorf("cannot get first element of empty array")
		}
		results = append(results, types.NewCandidateNode(arr[0]))
	}

	return results, nil
}

// evalFirstExpr evaluates an expression and returns the first result.
func evalFirstExpr(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate the expression
	results, err := evaluate(expr, ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("cannot get first element of empty sequence")
	}

	return []*types.CandidateNode{results[0]}, nil
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
			return nil, fmt.Errorf("cannot get last element of empty array")
		}
		results = append(results, types.NewCandidateNode(arr[len(arr)-1]))
	}

	return results, nil
}

// evalLastExpr evaluates an expression and returns the last result.
func evalLastExpr(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate the expression
	results, err := evaluate(expr, ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("cannot get last element of empty sequence")
	}

	return []*types.CandidateNode{results[len(results)-1]}, nil
}

// evalToEntries converts an object to an array of {key, value} entries.
func evalToEntries(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		switch v := node.Value.(type) {
		case map[string]any:
			entries := make([]any, 0, len(v))
			// Get keys sorted for consistent output
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				entry := map[string]any{
					"key":   k,
					"value": v[k],
				}
				entries = append(entries, entry)
			}
			results = append(results, types.NewCandidateNode(entries))
		case []any:
			entries := make([]any, len(v))
			for i, val := range v {
				entry := map[string]any{
					"key":   float64(i),
					"value": val,
				}
				entries[i] = entry
			}
			results = append(results, types.NewCandidateNode(entries))
		default:
			return nil, fmt.Errorf("to_entries requires object or array input, got %T", node.Value)
		}
	}

	return results, nil
}

// evalFromEntries converts an array of entries to an object.
func evalFromEntries(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("from_entries requires array input, got %T", node.Value)
		}

		obj := make(map[string]any)
		for _, elem := range arr {
			entry, ok := elem.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("from_entries: entry must be an object")
			}

			// Support both {key, value} and {name, value} and {k, v}
			var key string
			var value any
			var keyFound, valueFound bool

			for k, v := range entry {
				switch k {
				case "key", "name", "k":
					if s, ok := v.(string); ok {
						key = s
						keyFound = true
					} else if n, ok := toNumber(v); ok {
						key = fmt.Sprintf("%v", int64(n))
						keyFound = true
					}
				case "value", "v":
					value = v
					valueFound = true
				}
			}

			if !keyFound || !valueFound {
				return nil, fmt.Errorf("from_entries: entry must have key/value")
			}

			obj[key] = value
		}

		results = append(results, types.NewCandidateNode(obj))
	}

	return results, nil
}

// evalWithEntries applies an expression to each entry and rebuilds the object.
func evalWithEntries(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		// Convert to entries
		entriesCtx := ctx.Clone()
		entriesCtx.SetMatchingNodes([]*types.CandidateNode{node})
		entries, err := evalToEntries(entriesCtx)
		if err != nil {
			return nil, err
		}
		if len(entries) == 0 {
			continue
		}

		// Apply expression to each entry
		entriesArr := entries[0].Value.([]any)
		transformedEntries := make([]any, 0, len(entriesArr))

		for _, entry := range entriesArr {
			entryCtx := ctx.Clone()
			entryCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(entry)})

			transformed, err := evaluate(expr, entryCtx)
			if err != nil {
				return nil, err
			}

			for _, t := range transformed {
				transformedEntries = append(transformedEntries, t.Value)
			}
		}

		// Convert back from entries
		fromCtx := ctx.Clone()
		fromCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(transformedEntries)})
		result, err := evalFromEntries(fromCtx)
		if err != nil {
			return nil, err
		}

		results = append(results, result...)
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

// evalContains checks if a value contains another (deep containment).
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
		results = append(results, types.NewCandidateNode(deepContains(node.Value, arg)))
	}

	return results, nil
}

// deepContains checks if a contains b (recursively for objects/arrays).
func deepContains(a, b any) bool {
	// String containment
	if as, aok := a.(string); aok {
		if bs, bok := b.(string); bok {
			return strings.Contains(as, bs)
		}
		return false
	}

	// Array containment - b must be subset of a
	if ba, bok := b.([]any); bok {
		aa, aok := a.([]any)
		if !aok {
			return false
		}
		// Every element in b must be contained in some element of a
		for _, belem := range ba {
			found := false
			for _, aelem := range aa {
				if deepContains(aelem, belem) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	// Object containment - all keys in b must exist in a with contained values
	if bm, bok := b.(map[string]any); bok {
		am, aok := a.(map[string]any)
		if !aok {
			return false
		}
		for k, bv := range bm {
			av, exists := am[k]
			if !exists {
				return false
			}
			if !deepContains(av, bv) {
				return false
			}
		}
		return true
	}

	// For scalars, use equality
	return equals(a, b)
}

// evalInside checks if a value is inside another (inverse of contains).
func evalInside(argExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate argument
	argResults, err := evaluate(argExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(argResults) == 0 {
		return nil, fmt.Errorf("inside: argument produced no value")
	}
	arg := argResults[0].Value

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		// inside is the inverse of contains: (a | inside(b)) == (b | contains(a))
		results = append(results, types.NewCandidateNode(deepContains(arg, node.Value)))
	}

	return results, nil
}

// evalTypeFilter filters values by type (numbers, strings, etc.).
func evalTypeFilter(ctx *types.Context, typeName string) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		match := false
		switch typeName {
		case "number":
			_, match = toNumber(node.Value)
		case "string":
			_, match = node.Value.(string)
		case "boolean":
			_, match = node.Value.(bool)
		case "null":
			match = node.Value == nil
		case "array":
			_, match = node.Value.([]any)
		case "object":
			_, match = node.Value.(map[string]any)
		}
		if match {
			results = append(results, node)
		}
	}

	return results, nil
}

// evalTest tests if a string matches a regex pattern.
func evalTest(patternExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate pattern
	patternResults, err := evaluate(patternExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(patternResults) == 0 {
		return nil, fmt.Errorf("test: pattern produced no value")
	}
	pattern, ok := patternResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("test: pattern must be a string, got %T", patternResults[0].Value)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("test: invalid regex: %w", err)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("test: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(re.MatchString(s)))
	}

	return results, nil
}

// evalMatch returns match information for a regex.
func evalMatch(patternExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate pattern
	patternResults, err := evaluate(patternExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(patternResults) == 0 {
		return nil, fmt.Errorf("match: pattern produced no value")
	}
	pattern, ok := patternResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("match: pattern must be a string, got %T", patternResults[0].Value)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("match: invalid regex: %w", err)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("match: input must be a string, got %T", node.Value)
		}

		match := re.FindStringSubmatchIndex(s)
		if match == nil {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		// Build match object
		captures := buildCaptures(re, s, match)
		if captures == nil {
			captures = []any{} // Empty array, not null
		}
		matchObj := map[string]any{
			"offset":   float64(match[0]),
			"length":   float64(match[1] - match[0]),
			"string":   s[match[0]:match[1]],
			"captures": captures,
		}

		results = append(results, types.NewCandidateNode(matchObj))
	}

	return results, nil
}

// buildCaptures builds the captures array from match indices.
func buildCaptures(re *regexp.Regexp, s string, match []int) []any {
	names := re.SubexpNames()
	var captures []any

	for i := 1; i < len(match)/2; i++ {
		start, end := match[i*2], match[i*2+1]
		if start == -1 {
			captures = append(captures, map[string]any{
				"offset": float64(-1),
				"length": float64(0),
				"string": nil,
				"name":   names[i],
			})
		} else {
			captures = append(captures, map[string]any{
				"offset": float64(start),
				"length": float64(end - start),
				"string": s[start:end],
				"name":   names[i],
			})
		}
	}

	return captures
}

// evalCapture extracts named capture groups.
func evalCapture(patternExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate pattern
	patternResults, err := evaluate(patternExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(patternResults) == 0 {
		return nil, fmt.Errorf("capture: pattern produced no value")
	}
	pattern, ok := patternResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("capture: pattern must be a string, got %T", patternResults[0].Value)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("capture: invalid regex: %w", err)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("capture: input must be a string, got %T", node.Value)
		}

		match := re.FindStringSubmatch(s)
		if match == nil {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		// Build capture object with named groups
		captureObj := make(map[string]any)
		names := re.SubexpNames()
		for i, name := range names {
			if name != "" && i < len(match) {
				captureObj[name] = match[i]
			}
		}

		results = append(results, types.NewCandidateNode(captureObj))
	}

	return results, nil
}

// evalSub replaces first match of regex.
func evalSub(patternExpr, replacementExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate pattern
	patternResults, err := evaluate(patternExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(patternResults) == 0 {
		return nil, fmt.Errorf("sub: pattern produced no value")
	}
	pattern, ok := patternResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("sub: pattern must be a string, got %T", patternResults[0].Value)
	}

	// Evaluate replacement
	replacementResults, err := evaluate(replacementExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(replacementResults) == 0 {
		return nil, fmt.Errorf("sub: replacement produced no value")
	}
	replacement, ok := replacementResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("sub: replacement must be a string, got %T", replacementResults[0].Value)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("sub: invalid regex: %w", err)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("sub: input must be a string, got %T", node.Value)
		}

		// Replace first match only
		loc := re.FindStringIndex(s)
		if loc == nil {
			results = append(results, types.NewCandidateNode(s))
		} else {
			result := s[:loc[0]] + replacement + s[loc[1]:]
			results = append(results, types.NewCandidateNode(result))
		}
	}

	return results, nil
}

// evalGsub replaces all matches of regex.
func evalGsub(patternExpr, replacementExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate pattern
	patternResults, err := evaluate(patternExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(patternResults) == 0 {
		return nil, fmt.Errorf("gsub: pattern produced no value")
	}
	pattern, ok := patternResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("gsub: pattern must be a string, got %T", patternResults[0].Value)
	}

	// Evaluate replacement
	replacementResults, err := evaluate(replacementExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(replacementResults) == 0 {
		return nil, fmt.Errorf("gsub: replacement produced no value")
	}
	replacement, ok := replacementResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("gsub: replacement must be a string, got %T", replacementResults[0].Value)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("gsub: invalid regex: %w", err)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("gsub: input must be a string, got %T", node.Value)
		}

		result := re.ReplaceAllString(s, replacement)
		results = append(results, types.NewCandidateNode(result))
	}

	return results, nil
}

// evalGroupBy groups array elements by a key expression.
func evalGroupBy(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("group_by requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode([]any{}))
			continue
		}

		// Group elements by key
		groups := make(map[string][]any)
		var keyOrder []string

		for _, elem := range arr {
			// Evaluate key expression
			elemCtx := ctx.Clone()
			elemCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(elem)})

			keyResults, err := evaluate(expr, elemCtx)
			if err != nil {
				return nil, err
			}
			if len(keyResults) == 0 {
				continue
			}

			// Convert key to string for grouping
			keyStr := fmt.Sprintf("%v", keyResults[0].Value)
			if _, exists := groups[keyStr]; !exists {
				keyOrder = append(keyOrder, keyStr)
			}
			groups[keyStr] = append(groups[keyStr], elem)
		}

		// Build result array preserving order
		grouped := make([]any, 0, len(groups))
		for _, key := range keyOrder {
			grouped = append(grouped, groups[key])
		}

		results = append(results, types.NewCandidateNode(grouped))
	}

	return results, nil
}

// evalMapValues transforms only the values of an object (keeps keys).
func evalMapValues(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		obj, ok := node.Value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("map_values requires object input, got %T", node.Value)
		}

		result := make(map[string]any)
		for k, v := range obj {
			// Evaluate expression with value as input
			valCtx := ctx.Clone()
			valCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(v)})

			valResults, err := evaluate(expr, valCtx)
			if err != nil {
				return nil, err
			}
			if len(valResults) > 0 {
				result[k] = valResults[0].Value
			}
		}

		results = append(results, types.NewCandidateNode(result))
	}

	return results, nil
}

// evalToString converts a value to string.
func evalToString(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		var str string
		switch v := node.Value.(type) {
		case string:
			str = v
		case float64:
			if v == float64(int64(v)) {
				str = fmt.Sprintf("%d", int64(v))
			} else {
				str = fmt.Sprintf("%v", v)
			}
		case int:
			str = fmt.Sprintf("%d", v)
		case int64:
			str = fmt.Sprintf("%d", v)
		case bool:
			str = fmt.Sprintf("%v", v)
		case nil:
			str = "null"
		default:
			str = fmt.Sprintf("%v", v)
		}
		results = append(results, types.NewCandidateNode(str))
	}

	return results, nil
}

// evalToNumber converts a value to number.
func evalToNumber(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		switch v := node.Value.(type) {
		case float64:
			results = append(results, types.NewCandidateNode(v))
		case int:
			results = append(results, types.NewCandidateNode(float64(v)))
		case int64:
			results = append(results, types.NewCandidateNode(float64(v)))
		case string:
			// Try to parse as number
			var f float64
			_, err := fmt.Sscanf(v, "%f", &f)
			if err != nil {
				return nil, fmt.Errorf("cannot convert %q to number", v)
			}
			results = append(results, types.NewCandidateNode(f))
		default:
			return nil, fmt.Errorf("cannot convert %T to number", node.Value)
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

// evalLtrimstr removes a prefix from a string.
func evalLtrimstr(prefixExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate prefix
	prefixResults, err := evaluate(prefixExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(prefixResults) == 0 {
		return nil, fmt.Errorf("ltrimstr: prefix produced no value")
	}
	prefix, ok := prefixResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("ltrimstr: prefix must be a string, got %T", prefixResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("ltrimstr: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.TrimPrefix(s, prefix)))
	}

	return results, nil
}

// evalRtrimstr removes a suffix from a string.
func evalRtrimstr(suffixExpr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	// Evaluate suffix
	suffixResults, err := evaluate(suffixExpr, ctx)
	if err != nil {
		return nil, err
	}
	if len(suffixResults) == 0 {
		return nil, fmt.Errorf("rtrimstr: suffix produced no value")
	}
	suffix, ok := suffixResults[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("rtrimstr: suffix must be a string, got %T", suffixResults[0].Value)
	}

	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("rtrimstr: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.TrimSuffix(s, suffix)))
	}

	return results, nil
}

// evalTrim trims whitespace from both ends of a string.
func evalTrim(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		s, ok := node.Value.(string)
		if !ok {
			return nil, fmt.Errorf("trim: input must be a string, got %T", node.Value)
		}

		results = append(results, types.NewCandidateNode(strings.TrimSpace(s)))
	}

	return results, nil
}

// evalMin returns the minimum element of an array.
func evalMin(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("min requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		minVal := arr[0]
		for _, elem := range arr[1:] {
			if compareValues(elem, minVal) < 0 {
				minVal = elem
			}
		}

		results = append(results, types.NewCandidateNode(minVal))
	}

	return results, nil
}

// evalMax returns the maximum element of an array.
func evalMax(ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("max requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		maxVal := arr[0]
		for _, elem := range arr[1:] {
			if compareValues(elem, maxVal) > 0 {
				maxVal = elem
			}
		}

		results = append(results, types.NewCandidateNode(maxVal))
	}

	return results, nil
}

// evalMinBy returns the element with the minimum value for a given expression.
func evalMinBy(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("min_by requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		// Find element with minimum key
		var minElem any
		var minKey any

		for _, elem := range arr {
			// Evaluate key expression
			elemCtx := ctx.Clone()
			elemCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(elem)})

			keyResults, err := evaluate(expr, elemCtx)
			if err != nil {
				return nil, err
			}
			if len(keyResults) == 0 {
				continue
			}

			key := keyResults[0].Value
			if minKey == nil || compareValues(key, minKey) < 0 {
				minElem = elem
				minKey = key
			}
		}

		results = append(results, types.NewCandidateNode(minElem))
	}

	return results, nil
}

// evalMaxBy returns the element with the maximum value for a given expression.
func evalMaxBy(expr parser.ExpressionNode, ctx *types.Context) ([]*types.CandidateNode, error) {
	var results []*types.CandidateNode

	for _, node := range ctx.MatchingNodes {
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("max_by requires array input, got %T", node.Value)
		}

		if len(arr) == 0 {
			results = append(results, types.NewCandidateNode(nil))
			continue
		}

		// Find element with maximum key
		var maxElem any
		var maxKey any

		for _, elem := range arr {
			// Evaluate key expression
			elemCtx := ctx.Clone()
			elemCtx.SetMatchingNodes([]*types.CandidateNode{types.NewCandidateNode(elem)})

			keyResults, err := evaluate(expr, elemCtx)
			if err != nil {
				return nil, err
			}
			if len(keyResults) == 0 {
				continue
			}

			key := keyResults[0].Value
			if maxKey == nil || compareValues(key, maxKey) > 0 {
				maxElem = elem
				maxKey = key
			}
		}

		results = append(results, types.NewCandidateNode(maxElem))
	}

	return results, nil
}
