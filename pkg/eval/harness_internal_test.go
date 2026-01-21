package eval

import (
	"testing"
)

// Test the semantic comparison logic itself
func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		want     bool
	}{
		// Scalar comparisons
		{
			name:     "identical strings",
			expected: `"hello"`,
			actual:   `"hello"`,
			want:     true,
		},
		{
			name:     "different strings",
			expected: `"hello"`,
			actual:   `"world"`,
			want:     false,
		},
		{
			name:     "identical numbers",
			expected: `42`,
			actual:   `42`,
			want:     true,
		},
		{
			name:     "int vs float (whole number)",
			expected: `42`,
			actual:   `42.0`,
			want:     true,
		},
		{
			name:     "different numbers",
			expected: `42`,
			actual:   `43`,
			want:     false,
		},
		{
			name:     "boolean true",
			expected: `true`,
			actual:   `true`,
			want:     true,
		},
		{
			name:     "boolean false",
			expected: `false`,
			actual:   `false`,
			want:     true,
		},
		{
			name:     "null values",
			expected: `null`,
			actual:   `null`,
			want:     true,
		},

		// Object comparisons (JSON format)
		{
			name:     "identical objects",
			expected: `{"a": 1, "b": 2}`,
			actual:   `{"a": 1, "b": 2}`,
			want:     true,
		},
		{
			name:     "objects different key order",
			expected: `{"a": 1, "b": 2}`,
			actual:   `{"b": 2, "a": 1}`,
			want:     true,
		},
		{
			name:     "nested objects",
			expected: `{"user": {"name": "Alice", "age": 30}}`,
			actual:   `{"user": {"age": 30, "name": "Alice"}}`,
			want:     true,
		},
		{
			name:     "different objects",
			expected: `{"a": 1}`,
			actual:   `{"a": 2}`,
			want:     false,
		},

		// Array comparisons
		{
			name:     "identical arrays",
			expected: `[1, 2, 3]`,
			actual:   `[1, 2, 3]`,
			want:     true,
		},
		{
			name:     "arrays different order",
			expected: `[1, 2, 3]`,
			actual:   `[3, 2, 1]`,
			want:     false, // Order matters for arrays
		},
		{
			name:     "array of objects",
			expected: `[{"id": 1}, {"id": 2}]`,
			actual:   `[{"id": 1}, {"id": 2}]`,
			want:     true,
		},

		// Mixed numeric types
		{
			name:     "array with int vs float",
			expected: `[1, 2, 3]`,
			actual:   `[1.0, 2.0, 3.0]`,
			want:     true,
		},
		{
			name:     "object with int vs float",
			expected: `{"count": 42}`,
			actual:   `{"count": 42.0}`,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareValues(t, tt.expected, tt.actual)
			if got != tt.want {
				t.Errorf("compareValues(%q, %q) = %v, want %v", tt.expected, tt.actual, got, tt.want)
			}
		})
	}
}

func TestCompareResults(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
		want     bool
	}{
		{
			name:     "single matching value",
			expected: []string{`"hello"`},
			actual:   []string{`"hello"`},
			want:     true,
		},
		{
			name:     "multiple matching values",
			expected: []string{`"a"`, `"b"`, `"c"`},
			actual:   []string{`"a"`, `"b"`, `"c"`},
			want:     true,
		},
		{
			name:     "different lengths",
			expected: []string{`"a"`, `"b"`},
			actual:   []string{`"a"`},
			want:     false,
		},
		{
			name:     "empty results",
			expected: []string{},
			actual:   []string{},
			want:     true,
		},
		{
			name:     "mixed types matching",
			expected: []string{`42`, `{"a": 1}`, `[1, 2]`},
			actual:   []string{`42.0`, `{"a": 1}`, `[1, 2]`},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareResults(t, tt.expected, tt.actual)
			if got != tt.want {
				t.Errorf("compareResults(%v, %v) = %v, want %v", tt.expected, tt.actual, got, tt.want)
			}
		})
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{
			name:  "float64 whole number becomes int64",
			input: float64(42),
			want:  int64(42),
		},
		{
			name:  "float64 with decimal stays float64",
			input: float64(3.14),
			want:  float64(3.14),
		},
		{
			name:  "int becomes int64",
			input: int(42),
			want:  int64(42),
		},
		{
			name:  "string unchanged",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "bool unchanged",
			input: true,
			want:  true,
		},
		{
			name:  "nil unchanged",
			input: nil,
			want:  nil,
		},
		{
			name:  "nested map normalized",
			input: map[string]any{"count": float64(42)},
			want:  map[string]any{"count": int64(42)},
		},
		{
			name:  "slice normalized",
			input: []any{float64(1), float64(2), float64(3)},
			want:  []any{int64(1), int64(2), int64(3)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeValue(tt.input)
			if !deepEqual(got, tt.want) {
				t.Errorf("normalizeValue(%v) = %v (%T), want %v (%T)", tt.input, got, got, tt.want, tt.want)
			}
		})
	}
}

// deepEqual handles comparison with proper type checking
func deepEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !deepEqual(v, bv[k]) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
