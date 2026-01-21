package eval

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/knadh/koanf/parsers/json"
)

// testScenario runs a single scenario test.
// This is the core test harness function that all operator tests use.
func testScenario(t *testing.T, s *Scenario) {
	t.Helper()

	name := s.Description
	if name == "" {
		name = s.Expression
	}

	t.Run(name, func(t *testing.T) {
		// Set environment variables if specified
		for k, v := range s.EnvVars {
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}

		// TODO: Implementation
		// 1. Parse the expression using the parser package
		// expr, err := parser.Parse(s.Expression)
		//
		// 2. Parse the input document(s) using go-huml
		// doc, err := huml.Unmarshal([]byte(s.Document))
		//
		// 3. Evaluate the expression
		// results, err := eval.Evaluate(expr, doc)
		//
		// 4. Compare results with expected using semantic comparison
		// if s.ExpectedError != "" {
		//     if err == nil || !strings.Contains(err.Error(), s.ExpectedError) {
		//         t.Errorf("expected error containing %q, got %v", s.ExpectedError, err)
		//     }
		// } else {
		//     if err != nil {
		//         t.Fatalf("unexpected error: %v", err)
		//     }
		//     if !compareResults(t, s.Expected, results) {
		//         t.Errorf("result mismatch\nexpected: %v\ngot: %v", s.Expected, results)
		//     }
		// }

		// For now, mark as skipped until implementation exists
		if s.ExpectedError != "" {
			t.Skipf("TODO: implement expression evaluation (expecting error: %s)", s.ExpectedError)
		} else {
			t.Skipf("TODO: implement expression evaluation (expecting: %v)", s.Expected)
		}
	})
}

// runScenarios runs all scenarios in a group.
func runScenarios(t *testing.T, group ScenarioGroup) {
	t.Helper()
	for _, scenario := range group.Scenarios {
		testScenario(t, &scenario)
	}
}

// Helper function to create multi-line HUML documents in tests
func huml(s string) string {
	return strings.TrimPrefix(s, "\n")
}

// compareResults performs semantic comparison of expected vs actual results.
// It uses koanf to normalize different formats (JSON, HUML) to a common representation.
// This allows tests to use either JSON or HUML syntax in Expected values.
func compareResults(t *testing.T, expected []string, actual []string) bool {
	t.Helper()

	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		if !compareValues(t, expected[i], actual[i]) {
			return false
		}
	}
	return true
}

// compareValues compares two values semantically.
// Handles scalars (strings, numbers, booleans, null) and complex types (objects, arrays).
func compareValues(t *testing.T, expected, actual string) bool {
	t.Helper()

	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	// Try to parse as complex types first (objects/arrays)
	// If both parse successfully, compare semantically
	expectedVal, expectedErr := parseValue(expected)
	actualVal, actualErr := parseValue(actual)

	if expectedErr == nil && actualErr == nil {
		// Both are parseable - compare semantically
		return reflect.DeepEqual(normalizeValue(expectedVal), normalizeValue(actualVal))
	}

	// Fall back to string comparison for scalars
	return expected == actual
}

// parseValue attempts to parse a string as JSON (which covers most cases).
// Returns the parsed value or an error if not parseable.
func parseValue(s string) (any, error) {
	s = strings.TrimSpace(s)

	// Handle simple scalars that aren't valid JSON on their own
	if s == "null" {
		return nil, nil
	}
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	// Try parsing as JSON (handles objects, arrays, strings, numbers)
	parser := json.Parser()

	// Arrays need special handling - wrap to parse via koanf's map-based parser
	if strings.HasPrefix(s, "[") {
		wrapped := `{"_":` + s + `}`
		data, err := parser.Unmarshal([]byte(wrapped))
		if err != nil {
			return nil, err
		}
		return data["_"], nil
	}

	// Objects can be parsed directly
	if strings.HasPrefix(s, "{") {
		data, err := parser.Unmarshal([]byte(s))
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// Bare values (strings, numbers) - wrap to parse
	wrapped := `{"_":` + s + `}`
	data, err := parser.Unmarshal([]byte(wrapped))
	if err != nil {
		return nil, err
	}
	return data["_"], nil
}

// normalizeValue converts numeric types for consistent comparison.
// JSON unmarshals numbers as float64, but HUML may use int64.
func normalizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, v := range val {
			result[k] = normalizeValue(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = normalizeValue(v)
		}
		return result
	case float64:
		// Convert whole numbers to int for comparison
		if val == float64(int64(val)) {
			return int64(val)
		}
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	default:
		return val
	}
}
