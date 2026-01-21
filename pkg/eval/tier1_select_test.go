package eval

import "testing"

// Select and filtering tests
// Tier 1 - Essential (90% of use cases)

var selectScenarios = ScenarioGroup{
	Name:        "select",
	Description: "Filter values based on boolean conditions",
	Scenarios: []Scenario{
		{
			Description: "select with equality",
			Document: huml(`
- name: "Alice"
  active: true
- name: "Bob"
  active: false
- name: "Carol"
  active: true
`),
			Expression: `.[] | select(.active == true)`,
			Expected: []string{huml(`
name: "Alice"
active: true
`), huml(`
name: "Carol"
active: true
`)},
		},
		{
			Description: "select with numeric comparison",
			Document: huml(`
- name: "Alice"
  age: 30
- name: "Bob"
  age: 17
- name: "Carol"
  age: 25
`),
			Expression: `.[] | select(.age > 18)`,
			Expected: []string{huml(`
name: "Alice"
age: 30
`), huml(`
name: "Carol"
age: 25
`)},
		},
		{
			Description: "select with greater or equal",
			Document: huml(`
- price: 50
- price: 100
- price: 150
`),
			Expression: `.[] | select(.price >= 100)`,
			Expected: []string{huml(`
price: 100
`), huml(`
price: 150
`)},
		},
		{
			Description: "select with less than",
			Document: huml(`
- score: 80
- score: 60
- score: 90
`),
			Expression: `.[] | select(.score < 70)`,
			Expected: []string{huml(`
score: 60
`)},
		},
		{
			Description: "select with inequality",
			Document: huml(`
- status: "active"
- status: "deleted"
- status: "pending"
`),
			Expression: `.[] | select(.status != "deleted")`,
			Expected: []string{huml(`
status: "active"
`), huml(`
status: "pending"
`)},
		},
		{
			Description: "select with and",
			Document: huml(`
- name: "Alice"
  age: 30
  active: true
- name: "Bob"
  age: 25
  active: false
- name: "Carol"
  age: 35
  active: true
`),
			Expression: `.[] | select(.age > 28 and .active == true)`,
			Expected: []string{huml(`
name: "Alice"
age: 30
active: true
`), huml(`
name: "Carol"
age: 35
active: true
`)},
		},
		{
			Description: "select with or",
			Document: huml(`
- role: "admin"
  active: false
- role: "user"
  active: true
- role: "guest"
  active: false
`),
			Expression: `.[] | select(.role == "admin" or .active == true)`,
			Expected: []string{huml(`
role: "admin"
active: false
`), huml(`
role: "user"
active: true
`)},
		},
		{
			Description: "select with not",
			Document: huml(`
- enabled: true
- enabled: false
- enabled: true
`),
			Expression: `.[] | select(.enabled | not)`,
			Expected: []string{huml(`
enabled: false
`)},
		},
		{
			Description: "select none matching",
			Document: huml(`
- value: 1
- value: 2
- value: 3
`),
			Expression: `.[] | select(.value > 100)`,
			Expected:   []string{},
		},
		{
			Description: "select all matching",
			Document: huml(`
- value: 100
- value: 200
- value: 300
`),
			Expression: `.[] | select(.value > 50)`,
			Expected: []string{huml(`
value: 100
`), huml(`
value: 200
`), huml(`
value: 300
`)},
		},
		{
			Description: "select with string comparison",
			Document: huml(`
- name: "alice"
- name: "bob"
- name: "carol"
`),
			Expression: `.[] | select(.name == "bob")`,
			Expected: []string{huml(`
name: "bob"
`)},
		},
		{
			Description: "select with null check",
			Document: huml(`
- name: "Alice"
  email: "alice@example.com"
- name: "Bob"
- name: "Carol"
  email: "carol@example.com"
`),
			Expression: `.[] | select(.email != null)`,
			Expected: []string{huml(`
name: "Alice"
email: "alice@example.com"
`), huml(`
name: "Carol"
email: "carol@example.com"
`)},
		},
	},
}

var comparisonScenarios = ScenarioGroup{
	Name:        "comparison",
	Description: "Comparison operators (==, !=, <, >, <=, >=)",
	Scenarios: []Scenario{
		{
			Description: "equality with string",
			Document:    `"hello"`,
			Expression:  `. == "hello"`,
			Expected:    []string{`true`},
		},
		{
			Description: "equality with number",
			Document:    `42`,
			Expression:  `. == 42`,
			Expected:    []string{`true`},
		},
		{
			Description: "equality false",
			Document:    `42`,
			Expression:  `. == 43`,
			Expected:    []string{`false`},
		},
		{
			Description: "inequality true",
			Document:    `42`,
			Expression:  `. != 43`,
			Expected:    []string{`true`},
		},
		{
			Description: "inequality false",
			Document:    `42`,
			Expression:  `. != 42`,
			Expected:    []string{`false`},
		},
		{
			Description: "less than true",
			Document:    `5`,
			Expression:  `. < 10`,
			Expected:    []string{`true`},
		},
		{
			Description: "less than false",
			Document:    `10`,
			Expression:  `. < 5`,
			Expected:    []string{`false`},
		},
		{
			Description: "greater than true",
			Document:    `10`,
			Expression:  `. > 5`,
			Expected:    []string{`true`},
		},
		{
			Description: "greater than false",
			Document:    `5`,
			Expression:  `. > 10`,
			Expected:    []string{`false`},
		},
		{
			Description: "less than or equal - equal case",
			Document:    `10`,
			Expression:  `. <= 10`,
			Expected:    []string{`true`},
		},
		{
			Description: "less than or equal - less case",
			Document:    `5`,
			Expression:  `. <= 10`,
			Expected:    []string{`true`},
		},
		{
			Description: "greater than or equal - equal case",
			Document:    `10`,
			Expression:  `. >= 10`,
			Expected:    []string{`true`},
		},
		{
			Description: "greater than or equal - greater case",
			Document:    `15`,
			Expression:  `. >= 10`,
			Expected:    []string{`true`},
		},
		{
			Description: "string comparison less than",
			Document:    `"apple"`,
			Expression:  `. < "banana"`,
			Expected:    []string{`true`},
		},
		{
			Description: "null equality",
			Document:    `null`,
			Expression:  `. == null`,
			Expected:    []string{`true`},
		},
		{
			Description: "boolean equality",
			Document:    `true`,
			Expression:  `. == true`,
			Expected:    []string{`true`},
		},
	},
}

var booleanScenarios = ScenarioGroup{
	Name:        "boolean",
	Description: "Boolean operators (and, or, not)",
	Scenarios: []Scenario{
		{
			Description: "and - both true",
			Document:    `null`,
			Expression:  `true and true`,
			Expected:    []string{`true`},
		},
		{
			Description: "and - one false",
			Document:    `null`,
			Expression:  `true and false`,
			Expected:    []string{`false`},
		},
		{
			Description: "or - both false",
			Document:    `null`,
			Expression:  `false or false`,
			Expected:    []string{`false`},
		},
		{
			Description: "or - one true",
			Document:    `null`,
			Expression:  `false or true`,
			Expected:    []string{`true`},
		},
		{
			Description: "not - true",
			Document:    `true`,
			Expression:  `. | not`,
			Expected:    []string{`false`},
		},
		{
			Description: "not - false",
			Document:    `false`,
			Expression:  `. | not`,
			Expected:    []string{`true`},
		},
		{
			Description: "not - null is falsy",
			Document:    `null`,
			Expression:  `. | not`,
			Expected:    []string{`true`},
		},
		{
			Description: "complex boolean expression",
			Document: huml(`
a: true
b: false
c: true
`),
			Expression: `(.a and .c) or .b`,
			Expected:   []string{`true`},
		},
		{
			Description: "boolean with comparison",
			Document: huml(`
x: 10
y: 20
`),
			Expression: `.x > 5 and .y < 30`,
			Expected:   []string{`true`},
		},
	},
}

func TestSelectScenarios(t *testing.T) {
	runScenarios(t, selectScenarios)
}

func TestComparisonScenarios(t *testing.T) {
	runScenarios(t, comparisonScenarios)
}

func TestBooleanScenarios(t *testing.T) {
	runScenarios(t, booleanScenarios)
}
