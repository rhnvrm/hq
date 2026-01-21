package eval

import "testing"

// Object and array construction tests
// Tier 1 - Essential (90% of use cases)

var objectConstructionScenarios = ScenarioGroup{
	Name:        "object-construction",
	Description: "Construct objects with {...}",
	Scenarios: []Scenario{
		{
			Description: "simple object literal",
			Document:    `null`,
			Expression:  `{a: 1, b: 2}`,
			Expected:    []string{`{"a": 1, "b": 2}`},
		},
		{
			Description: "object with string values",
			Document:    `null`,
			Expression:  `{name: "Alice", city: "NYC"}`,
			Expected:    []string{`{"name": "Alice", "city": "NYC"}`},
		},
		{
			Description: "object from document fields",
			Document: huml(`
user:
  firstName: "Alice"
  lastName: "Smith"
  age: 30
  email: "alice@example.com"
`),
			Expression: `{name: .user.firstName, email: .user.email}`,
			Expected:   []string{`{"name": "Alice", "email": "alice@example.com"}`},
		},
		{
			Description: "object shorthand syntax",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `{name, age}`,
			Expected:   []string{`{"name": "Alice", "age": 30}`},
		},
		{
			Description: "object with computed key",
			Document: huml(`
key: "mykey"
value: 42
`),
			Expression: `{(.key): .value}`,
			Expected:   []string{`{"mykey": 42}`},
		},
		{
			Description: "nested object construction",
			Document: huml(`
name: "Alice"
city: "NYC"
`),
			Expression: `{user: {name: .name, location: {city: .city}}}`,
			Expected:   []string{`{"user": {"name": "Alice", "location": {"city": "NYC"}}}`},
		},
		{
			Description: "object in iterator context",
			Document: huml(`
users:
  - id: 1
    name: "Alice"
    email: "alice@example.com"
  - id: 2
    name: "Bob"
    email: "bob@example.com"
`),
			Expression: `.users[] | {id, name}`,
			Expected:   []string{`{"id": 1, "name": "Alice"}`, `{"id": 2, "name": "Bob"}`},
		},
		{
			Description: "empty object",
			Document:    `null`,
			Expression:  `{}`,
			Expected:    []string{`{}`},
		},
	},
}

var arrayConstructionScenarios = ScenarioGroup{
	Name:        "array-construction",
	Description: "Construct arrays with [...]",
	Scenarios: []Scenario{
		{
			Description: "simple array literal",
			Document:    `null`,
			Expression:  `[1, 2, 3]`,
			Expected:    []string{`[1, 2, 3]`},
		},
		{
			Description: "array with mixed types",
			Document:    `null`,
			Expression:  `["a", 1, true, null]`,
			Expected:    []string{`["a", 1, true, null]`},
		},
		{
			Description: "array from document fields",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `[.a, .b, .c]`,
			Expected:   []string{`[1, 2, 3]`},
		},
		{
			Description: "collect iterator into array",
			Document: huml(`
- name: "Alice"
- name: "Bob"
- name: "Carol"
`),
			Expression: `[.[].name]`,
			Expected:   []string{`["Alice", "Bob", "Carol"]`},
		},
		{
			Description: "array with expression",
			Document: huml(`
x: 10
`),
			Expression: `[.x, .x * 2, .x * 3]`,
			Expected:   []string{`[10, 20, 30]`},
		},
		{
			Description: "nested array",
			Document:    `null`,
			Expression:  `[[1, 2], [3, 4]]`,
			Expected:    []string{`[[1, 2], [3, 4]]`},
		},
		{
			Description: "empty array",
			Document:    `null`,
			Expression:  `[]`,
			Expected:    []string{`[]`},
		},
		{
			Description: "array with filtered values",
			Document: huml(`
- value: 10
  include: true
- value: 20
  include: false
- value: 30
  include: true
`),
			Expression: `[.[] | select(.include) | .value]`,
			Expected:   []string{`[10, 30]`},
		},
	},
}

func TestObjectConstructionScenarios(t *testing.T) {
	runScenarios(t, objectConstructionScenarios)
}

func TestArrayConstructionScenarios(t *testing.T) {
	runScenarios(t, arrayConstructionScenarios)
}
