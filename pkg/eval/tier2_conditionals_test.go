package eval

import "testing"

// Conditional and variable tests
// Tier 2 - Important (next 8% of use cases)

var conditionalScenarios = ScenarioGroup{
	Name:        "if-then-else",
	Description: "if-then-else conditional expressions",
	Scenarios: []Scenario{
		{
			Description: "if true then",
			Document:    `5`,
			Expression:  `if . > 3 then "big" else "small" end`,
			Expected:    []string{`"big"`},
		},
		{
			Description: "if false else",
			Document:    `2`,
			Expression:  `if . > 3 then "big" else "small" end`,
			Expected:    []string{`"small"`},
		},
		{
			Description: "if without else",
			Document:    `5`,
			Expression:  `if . > 3 then "big" else . end`,
			Expected:    []string{`"big"`},
		},
		{
			Description: "if in map",
			Document: huml(`
- 1
- 5
- 2
- 8
`),
			Expression: `map(if . > 3 then "big" else "small" end)`,
			Expected:   []string{`["small", "big", "small", "big"]`},
		},
		{
			Description: "elif chain",
			Document: huml(`
score: 85
`),
			Expression: `if .score >= 90 then "A" elif .score >= 80 then "B" elif .score >= 70 then "C" else "F" end`,
			Expected:   []string{`"B"`},
		},
		{
			Description: "if with complex condition",
			Document: huml(`
status: "active"
role: "admin"
`),
			Expression: `if .status == "active" and .role == "admin" then "full access" else "limited" end`,
			Expected:   []string{`"full access"`},
		},
		{
			Description: "nested if",
			Document: huml(`
a: true
b: false
`),
			Expression: `if .a then (if .b then "both" else "only a" end) else "neither" end`,
			Expected:   []string{`"only a"`},
		},
		{
			Description: "if with null check",
			Document: huml(`
name: "Alice"
`),
			Expression: `if .email then .email else "no email" end`,
			Expected:   []string{`"no email"`},
		},
	},
}

var variableScenarios = ScenarioGroup{
	Name:        "variables",
	Description: "Variable binding with 'as $var'",
	Scenarios: []Scenario{
		{
			Description: "bind and use variable",
			Document: huml(`
multiplier: 10
value: 5
`),
			Expression: `.multiplier as $m | .value * $m`,
			Expected:   []string{`50`},
		},
		{
			Description: "variable in object construction",
			Document: huml(`
users:
  - name: "Alice"
    id: 1
  - name: "Bob"
    id: 2
`),
			Expression: `.users[] as $u | {id: $u.id, greeting: "Hello, \($u.name)"}`,
			Expected:   []string{`{"id": 1, "greeting": "Hello, Alice"}`, `{"id": 2, "greeting": "Hello, Bob"}`},
		},
		{
			Description: "multiple variables",
			Document: huml(`
a: 10
b: 20
`),
			Expression: `.a as $x | .b as $y | $x + $y`,
			Expected:   []string{`30`},
		},
		{
			Description: "variable preserves value",
			Document: huml(`
items:
  - 1
  - 2
  - 3
total: 100
`),
			Expression: `.total as $t | [.items[] | . / $t * 100]`,
			Expected:   []string{`[1, 2, 3]`},
		},
		{
			Description: "destructuring bind",
			Document: huml(`
point:
  x: 10
  y: 20
`),
			Expression: `.point as {x: $x, y: $y} | $x + $y`,
			Expected:   []string{`30`},
		},
		{
			Description: "variable in select",
			Document: huml(`
threshold: 50
values:
  - 30
  - 60
  - 40
  - 80
`),
			Expression: `.threshold as $t | [.values[] | select(. > $t)]`,
			Expected:   []string{`[60, 80]`},
		},
	},
}

var recursiveDescentScenarios = ScenarioGroup{
	Name:        "recursive-descent",
	Description: "Recursive descent operator (..)",
	Scenarios: []Scenario{
		{
			Description: "find all values",
			Document: huml(`
a: 1
b:
  c: 2
  d:
    e: 3
`),
			Expression: `[.. | numbers]`,
			Expected:   []string{`[1, 2, 3]`},
		},
		{
			Description: "find all strings",
			Document: huml(`
name: "Alice"
data:
  label: "test"
  nested:
    title: "hello"
`),
			// Keys are traversed alphabetically (data < name), so data's strings come first
			Expression: `[.. | strings]`,
			Expected:   []string{`["test", "hello", "Alice"]`},
		},
		{
			Description: "find all id fields",
			Document: huml(`
id: 1
children:
  - id: 2
    data:
      id: 3
  - id: 4
`),
			Expression: `[.. | .id? // empty]`,
			Expected:   []string{`[1, 2, 3, 4]`},
		},
		{
			Description: "recursive select",
			Document: huml(`
type: "folder"
name: "root"
children:
  - type: "file"
    name: "a.txt"
  - type: "folder"
    name: "sub"
    children:
      - type: "file"
        name: "b.txt"
`),
			Expression: `[.. | select(.type? == "file") | .name]`,
			Expected:   []string{`["a.txt", "b.txt"]`},
		},
	},
}

var reduceScenarios = ScenarioGroup{
	Name:        "reduce",
	Description: "reduce for aggregation",
	Scenarios: []Scenario{
		{
			Description: "reduce sum",
			Document: huml(`
- 1
- 2
- 3
- 4
- 5
`),
			Expression: `reduce .[] as $x (0; . + $x)`,
			Expected:   []string{`15`},
		},
		{
			Description: "reduce product",
			Document: huml(`
- 1
- 2
- 3
- 4
`),
			Expression: `reduce .[] as $x (1; . * $x)`,
			Expected:   []string{`24`},
		},
		{
			Description: "reduce build object",
			Document: huml(`
- key: "a"
  value: 1
- key: "b"
  value: 2
`),
			Expression: `reduce .[] as $item ({}; .[$item.key] = $item.value)`,
			Expected:   []string{`{"a": 1, "b": 2}`},
		},
		{
			Description: "reduce index by id",
			Document: huml(`
- id: 1
  name: "Alice"
- id: 2
  name: "Bob"
`),
			Expression: `reduce .[] as $item ({}; .[$item.id | tostring] = $item)`,
			Expected:   []string{`{"1": {"id": 1, "name": "Alice"}, "2": {"id": 2, "name": "Bob"}}`},
		},
		{
			Description: "reduce count",
			Document: huml(`
- "a"
- "b"
- "a"
- "c"
- "a"
`),
			Expression: `reduce .[] as $x ({}; .[$x] = (.[$x] // 0) + 1)`,
			Expected:   []string{`{"a": 3, "b": 1, "c": 1}`},
		},
	},
}

func TestConditionalScenarios(t *testing.T) {
	runScenarios(t, conditionalScenarios)
}

func TestVariableScenarios(t *testing.T) {
	runScenarios(t, variableScenarios)
}

func TestRecursiveDescentScenarios(t *testing.T) {
	runScenarios(t, recursiveDescentScenarios)
}

func TestReduceScenarios(t *testing.T) {
	runScenarios(t, reduceScenarios)
}
