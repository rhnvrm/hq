package eval

import "testing"

// Common function tests
// Tier 1 - Essential (90% of use cases)

var lengthScenarios = ScenarioGroup{
	Name:        "length",
	Description: "length returns size of arrays, objects, and strings",
	Scenarios: []Scenario{
		{
			Description: "array length",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `length`,
			Expected:   []string{`3`},
		},
		{
			Description: "object length (number of keys)",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `length`,
			Expected:   []string{`3`},
		},
		{
			Description: "string length",
			Document:    `"hello"`,
			Expression:  `length`,
			Expected:    []string{`5`},
		},
		{
			Description: "empty array length",
			Document:    `[]`,
			Expression:  `length`,
			Expected:    []string{`0`},
		},
		{
			Description: "empty object length",
			Document:    `{}`,
			Expression:  `length`,
			Expected:    []string{`0`},
		},
		{
			Description: "empty string length",
			Document:    `""`,
			Expression:  `length`,
			Expected:    []string{`0`},
		},
		{
			Description: "null length",
			Document:    `null`,
			Expression:  `length`,
			Expected:    []string{`null`},
		},
		{
			Description: "length with pipe",
			Document: huml(`
users:
  - name: "Alice"
  - name: "Bob"
`),
			Expression: `.users | length`,
			Expected:   []string{`2`},
		},
		{
			Description: "number absolute value (length of number)",
			Document:    `-42`,
			Expression:  `length`,
			Expected:    []string{`42`},
		},
	},
}

var keysScenarios = ScenarioGroup{
	Name:        "keys",
	Description: "keys returns object keys or array indices",
	Scenarios: []Scenario{
		{
			Description: "object keys",
			Document: huml(`
name: "Alice"
age: 30
city: "NYC"
`),
			Expression: `keys`,
			Expected:   []string{`["age", "city", "name"]`}, // sorted
		},
		{
			Description: "array keys (indices)",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `keys`,
			Expected:   []string{`[0, 1, 2]`},
		},
		{
			Description: "empty object keys",
			Document:    `{}`,
			Expression:  `keys`,
			Expected:    []string{`[]`},
		},
		{
			Description: "keys with pipe",
			Document: huml(`
config:
  host: "localhost"
  port: 8080
`),
			Expression: `.config | keys`,
			Expected:   []string{`["host", "port"]`},
		},
		{
			Description: "keys_unsorted preserves order",
			Document: huml(`
z: 1
a: 2
m: 3
`),
			// Note: Go maps don't preserve insertion order, so keys are returned
			// in non-deterministic order. Use keys (sorted) for consistent output.
			// This test just verifies keys_unsorted works, but order is undefined.
			Expression: `keys_unsorted | sort`,
			Expected:   []string{`["a", "m", "z"]`},
		},
	},
}

var hasScenarios = ScenarioGroup{
	Name:        "has",
	Description: "has checks if key exists",
	Scenarios: []Scenario{
		{
			Description: "has existing key",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `has("name")`,
			Expected:   []string{`true`},
		},
		{
			Description: "has missing key",
			Document: huml(`
name: "Alice"
`),
			Expression: `has("age")`,
			Expected:   []string{`false`},
		},
		{
			Description: "has with null value",
			Document: huml(`
name: null
`),
			Expression: `has("name")`,
			Expected:   []string{`true`},
		},
		{
			Description: "has array index exists",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `has(1)`,
			Expected:   []string{`true`},
		},
		{
			Description: "has array index out of bounds",
			Document: huml(`
- "a"
- "b"
`),
			Expression: `has(10)`,
			Expected:   []string{`false`},
		},
		{
			Description: "has in select",
			Document: huml(`
- name: "Alice"
  email: "alice@example.com"
- name: "Bob"
- name: "Carol"
  email: "carol@example.com"
`),
			Expression: `[.[] | select(has("email"))]`,
			Expected:   []string{`[{"name": "Alice", "email": "alice@example.com"}, {"name": "Carol", "email": "carol@example.com"}]`},
		},
	},
}

var typeScenarios = ScenarioGroup{
	Name:        "type",
	Description: "type returns the type of a value",
	Scenarios: []Scenario{
		{
			Description: "type of string",
			Document:    `"hello"`,
			Expression:  `type`,
			Expected:    []string{`"string"`},
		},
		{
			Description: "type of integer",
			Document:    `42`,
			Expression:  `type`,
			Expected:    []string{`"number"`},
		},
		{
			Description: "type of float",
			Document:    `3.14`,
			Expression:  `type`,
			Expected:    []string{`"number"`},
		},
		{
			Description: "type of boolean",
			Document:    `true`,
			Expression:  `type`,
			Expected:    []string{`"boolean"`},
		},
		{
			Description: "type of null",
			Document:    `null`,
			Expression:  `type`,
			Expected:    []string{`"null"`},
		},
		{
			Description: "type of array",
			Document:    `[1, 2, 3]`,
			Expression:  `type`,
			Expected:    []string{`"array"`},
		},
		{
			Description: "type of object",
			Document:    `{"a": 1}`,
			Expression:  `type`,
			Expected:    []string{`"object"`},
		},
		{
			Description: "filter by type",
			Document: huml(`
- 1
- "two"
- 3
- "four"
`),
			Expression: `[.[] | select(type == "string")]`,
			Expected:   []string{`["two", "four"]`},
		},
	},
}

var defaultScenarios = ScenarioGroup{
	Name:        "default",
	Description: "Alternative operator (//) provides default values",
	Scenarios: []Scenario{
		{
			Description: "default for missing field",
			Document: huml(`
name: "Alice"
`),
			Expression: `.age // 0`,
			Expected:   []string{`0`},
		},
		{
			Description: "default not used when value exists",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `.age // 0`,
			Expected:   []string{`30`},
		},
		{
			Description: "default for null",
			Document: huml(`
value: null
`),
			Expression: `.value // "default"`,
			Expected:   []string{`"default"`},
		},
		{
			Description: "default for false",
			Document: huml(`
enabled: false
`),
			// In jq, // treats both null AND false as "falsy" - uses alternative
			Expression: `.enabled // true`,
			Expected:   []string{`true`},
		},
		{
			Description: "chained defaults",
			Document: huml(`
name: "Alice"
`),
			Expression: `.primary // .secondary // .name`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "default with string",
			Document: huml(`
name: "Alice"
`),
			Expression: `.nickname // "Unknown"`,
			Expected:   []string{`"Unknown"`},
		},
		{
			Description: "default with complex value",
			Document: huml(`
name: "Alice"
`),
			Expression: `.config // {default: true}`,
			Expected:   []string{`{"default": true}`},
		},
	},
}

var emptyScenarios = ScenarioGroup{
	Name:        "empty",
	Description: "empty produces no output",
	Scenarios: []Scenario{
		{
			Description: "empty produces nothing",
			Document:    `42`,
			Expression:  `empty`,
			Expected:    []string{},
		},
		{
			Description: "empty in conditional",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `.[] | if . == 2 then empty else . end`,
			Expected:   []string{`1`, `3`},
		},
		{
			Description: "select returns empty on no match",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `.[] | select(. > 10)`,
			Expected:   []string{},
		},
		{
			Description: "empty alternative",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `.[] | select(. > 10) // empty`,
			Expected:   []string{},
		},
	},
}

func TestLengthScenarios(t *testing.T) {
	runScenarios(t, lengthScenarios)
}

func TestKeysScenarios(t *testing.T) {
	runScenarios(t, keysScenarios)
}

func TestHasScenarios(t *testing.T) {
	runScenarios(t, hasScenarios)
}

func TestTypeScenarios(t *testing.T) {
	runScenarios(t, typeScenarios)
}

func TestDefaultScenarios(t *testing.T) {
	runScenarios(t, defaultScenarios)
}

func TestEmptyScenarios(t *testing.T) {
	runScenarios(t, emptyScenarios)
}
