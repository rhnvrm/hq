package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	gohuml "github.com/huml-lang/go-huml"
	"gopkg.in/yaml.v3"
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

		// Parse the input document
		// Try JSON first, then YAML (for test convenience), then HUML
		var input any
		if s.Document != "" {
			doc := strings.TrimSpace(s.Document)

			// Try JSON first
			if err := json.Unmarshal([]byte(doc), &input); err != nil {
				// Try YAML as second option (more flexible, supports nested syntax)
				if err2 := yaml.Unmarshal([]byte(doc), &input); err2 != nil {
					// Try HUML as final fallback
					if err3 := gohuml.Unmarshal([]byte(doc), &input); err3 != nil {
						t.Fatalf("failed to parse input document (JSON: %v, YAML: %v, HUML: %v)", err, err2, err3)
					}
				}
			}
		}

		// Evaluate the expression
		results, err := Evaluate(s.Expression, input)

		// Check for expected error
		if s.ExpectedError != "" {
			if err == nil {
				t.Errorf("expected error containing %q, got no error (result: %v)", s.ExpectedError, results)
			} else if !strings.Contains(err.Error(), s.ExpectedError) {
				t.Errorf("expected error containing %q, got %v", s.ExpectedError, err)
			}
			return
		}

		// Check for unexpected error
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Convert results to strings for comparison
		actualStrings := make([]string, len(results))
		for i, result := range results {
			actualStrings[i] = valueToString(result)
		}

		// Compare results
		if !compareResultStrings(t, s.Expected, actualStrings) {
			t.Errorf("result mismatch\nexpression: %s\nexpected: %v\ngot: %v", s.Expression, s.Expected, actualStrings)
		}
	})
}

// valueToString converts a value to its JSON string representation for comparison.
func valueToString(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		// Quote strings
		b, _ := json.Marshal(val)
		return string(b)
	case float64:
		// Format numbers nicely (no trailing .0 for integers)
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []any, map[string]any:
		// Use JSON for complex types
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
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

// compareResultStrings performs semantic comparison of expected vs actual results.
func compareResultStrings(t *testing.T, expected []string, actual []string) bool {
	t.Helper()

	if len(expected) != len(actual) {
		t.Logf("length mismatch: expected %d, got %d", len(expected), len(actual))
		return false
	}

	for i := range expected {
		if !compareStringValues(t, expected[i], actual[i]) {
			t.Logf("value mismatch at index %d: expected %q, got %q", i, expected[i], actual[i])
			return false
		}
	}
	return true
}

// compareStringValues compares two string values semantically.
// Handles scalars (strings, numbers, booleans, null) and complex types (objects, arrays).
func compareStringValues(t *testing.T, expected, actual string) bool {
	t.Helper()

	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	// Try to parse as complex types first (objects/arrays)
	// If both parse successfully, compare semantically
	expectedVal, expectedErr := parseStringValue(expected)
	actualVal, actualErr := parseStringValue(actual)

	if expectedErr == nil && actualErr == nil {
		// Both are parseable - compare semantically
		return reflect.DeepEqual(normalizeValue(expectedVal), normalizeValue(actualVal))
	}

	// Fall back to string comparison for scalars
	return expected == actual
}

// parseStringValue attempts to parse a string as JSON or HUML.
// Returns the parsed value or an error if not parseable.
func parseStringValue(s string) (any, error) {
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

	// Try parsing as JSON first (handles objects, arrays, strings, numbers)
	if strings.HasPrefix(s, "[") || strings.HasPrefix(s, "{") || strings.HasPrefix(s, "\"") {
		var val any
		if err := json.Unmarshal([]byte(s), &val); err == nil {
			return val, nil
		}
	}

	// Try parsing as HUML
	var val any
	if err := gohuml.Unmarshal([]byte(s), &val); err == nil {
		return val, nil
	}

	// Try bare number
	var num float64
	if _, err := fmt.Sscanf(s, "%f", &num); err == nil {
		return num, nil
	}

	// Return as string
	return s, nil
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
