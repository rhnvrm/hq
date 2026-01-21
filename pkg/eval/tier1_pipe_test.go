package eval

import "testing"

// Pipe and composition tests
// Tier 1 - Essential (90% of use cases)

var pipeScenarios = ScenarioGroup{
	Name:        "pipe",
	Description: "Pipe operator (|) chains operations",
	Scenarios: []Scenario{
		{
			Description: "pipe identity",
			Document:    `42`,
			Expression:  `. | .`,
			Expected:    []string{`42`},
		},
		{
			Description: "pipe field access",
			Document: huml(`
user:
  name: "Alice"
`),
			Expression: `.user | .name`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "pipe iterator to field",
			Document: huml(`
users:
  - name: "Alice"
  - name: "Bob"
`),
			Expression: `.users[] | .name`,
			Expected:   []string{`"Alice"`, `"Bob"`},
		},
		{
			Description: "multiple pipes",
			Document: huml(`
data:
  users:
    - name: "Alice"
    - name: "Bob"
`),
			Expression: `.data | .users[] | .name`,
			Expected:   []string{`"Alice"`, `"Bob"`},
		},
		{
			Description: "pipe preserves multiple outputs",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `.[] | . + 10`,
			Expected:   []string{`11`, `12`, `13`},
		},
	},
}

var commaScenarios = ScenarioGroup{
	Name:        "comma",
	Description: "Comma operator produces multiple outputs",
	Scenarios: []Scenario{
		{
			Description: "comma produces multiple outputs",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `.name, .age`,
			Expected:   []string{`"Alice"`, `30`},
		},
		{
			Description: "comma with three outputs",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `.a, .b, .c`,
			Expected:   []string{`1`, `2`, `3`},
		},
		{
			Description: "comma in iterator context",
			Document: huml(`
users:
  - name: "Alice"
    age: 30
  - name: "Bob"
    age: 25
`),
			Expression: `.users[] | .name, .age`,
			Expected:   []string{`"Alice"`, `30`, `"Bob"`, `25`},
		},
		{
			Description: "comma with missing field",
			Document: huml(`
name: "Alice"
`),
			Expression: `.name, .age`,
			Expected:   []string{`"Alice"`, `null`},
		},
	},
}

var parenthesesScenarios = ScenarioGroup{
	Name:        "parentheses",
	Description: "Parentheses for grouping expressions",
	Scenarios: []Scenario{
		{
			Description: "parentheses for arithmetic grouping",
			Document: huml(`
a: 2
b: 3
c: 4
`),
			Expression: `(.a + .b) * .c`,
			Expected:   []string{`20`},
		},
		{
			Description: "parentheses change precedence",
			Document: huml(`
a: 2
b: 3
c: 4
`),
			Expression: `.a + (.b * .c)`,
			Expected:   []string{`14`},
		},
		{
			Description: "nested parentheses",
			Document: huml(`
a: 1
b: 2
c: 3
d: 4
`),
			Expression: `((.a + .b) * (.c + .d))`,
			Expected:   []string{`21`},
		},
	},
}

func TestPipeScenarios(t *testing.T) {
	runScenarios(t, pipeScenarios)
}

func TestCommaScenarios(t *testing.T) {
	runScenarios(t, commaScenarios)
}

func TestParenthesesScenarios(t *testing.T) {
	runScenarios(t, parenthesesScenarios)
}
