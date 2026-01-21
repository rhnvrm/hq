package eval

import "testing"

// Assignment and update tests
// Tier 1 - Essential (90% of use cases)

var assignmentScenarios = ScenarioGroup{
	Name:        "assignment",
	Description: "Assignment operator (=) sets values",
	Scenarios: []Scenario{
		{
			Description: "set simple field",
			Document: huml(`
name: "Alice"
`),
			Expression: `.name = "Bob"`,
			Expected: []string{huml(`
name: "Bob"
`)},
		},
		{
			Description: "set nested field",
			Document: huml(`
user:
  name: "Alice"
`),
			Expression: `.user.name = "Bob"`,
			Expected: []string{huml(`
user:
  name: "Bob"
`)},
		},
		{
			Description: "set new field",
			Document: huml(`
name: "Alice"
`),
			Expression: `.age = 30`,
			Expected: []string{huml(`
name: "Alice"
age: 30
`)},
		},
		{
			Description: "set new nested field creates path",
			Document: huml(`
name: "Alice"
`),
			Expression: `.address.city = "NYC"`,
			Expected: []string{huml(`
name: "Alice"
address:
  city: "NYC"
`)},
		},
		{
			Description: "set array element",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `.[1] = "X"`,
			Expected: []string{huml(`
- "a"
- "X"
- "c"
`)},
		},
		{
			Description: "set field to object",
			Document: huml(`
name: "Alice"
`),
			Expression: `.metadata = {created: "2024-01-01", version: 1}`,
			Expected: []string{huml(`
name: "Alice"
metadata:
  created: "2024-01-01"
  version: 1
`)},
		},
		{
			Description: "set field to array",
			Document: huml(`
name: "Alice"
`),
			Expression: `.tags = ["a", "b", "c"]`,
			Expected: []string{huml(`
name: "Alice"
tags:
  - "a"
  - "b"
  - "c"
`)},
		},
		{
			Description: "set field to null",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `.age = null`,
			Expected: []string{huml(`
name: "Alice"
age: null
`)},
		},
		{
			Description: "set field based on another field",
			Document: huml(`
firstName: "Alice"
lastName: "Smith"
`),
			Expression: `.fullName = .firstName + " " + .lastName`,
			Expected: []string{huml(`
firstName: "Alice"
lastName: "Smith"
fullName: "Alice Smith"
`)},
		},
	},
}

var updateScenarios = ScenarioGroup{
	Name:        "update",
	Description: "Update operator (|=) modifies values relative to current",
	Scenarios: []Scenario{
		{
			Description: "increment value",
			Document: huml(`
count: 10
`),
			Expression: `.count |= . + 1`,
			Expected: []string{huml(`
count: 11
`)},
		},
		{
			Description: "double value",
			Document: huml(`
value: 5
`),
			Expression: `.value |= . * 2`,
			Expected: []string{huml(`
value: 10
`)},
		},
		{
			Description: "uppercase string",
			Document: huml(`
name: "alice"
`),
			Expression: `.name |= ascii_upcase`,
			Expected: []string{huml(`
name: "ALICE"
`)},
		},
		{
			Description: "update all array elements",
			Document: huml(`
values:
  - 1
  - 2
  - 3
`),
			Expression: `.values[] |= . * 10`,
			Expected: []string{huml(`
values:
  - 10
  - 20
  - 30
`)},
		},
		{
			Description: "update nested field",
			Document: huml(`
user:
  score: 100
`),
			Expression: `.user.score |= . + 50`,
			Expected: []string{huml(`
user:
  score: 150
`)},
		},
		{
			Description: "update with conditional",
			Document: huml(`
status: "pending"
`),
			Expression: `.status |= if . == "pending" then "active" else . end`,
			Expected: []string{huml(`
status: "active"
`)},
		},
	},
}

var addAssignScenarios = ScenarioGroup{
	Name:        "add-assign",
	Description: "Add-assign operator (+=) adds to current value",
	Scenarios: []Scenario{
		{
			Description: "add to number",
			Document: huml(`
count: 10
`),
			Expression: `.count += 5`,
			Expected: []string{huml(`
count: 15
`)},
		},
		{
			Description: "append to array",
			Document: huml(`
tags:
  - "a"
  - "b"
`),
			Expression: `.tags += ["c"]`,
			Expected: []string{huml(`
tags:
  - "a"
  - "b"
  - "c"
`)},
		},
		{
			Description: "concatenate string",
			Document: huml(`
name: "Alice"
`),
			Expression: `.name += " Smith"`,
			Expected: []string{huml(`
name: "Alice Smith"
`)},
		},
		{
			Description: "merge objects",
			Document: huml(`
config:
  host: "localhost"
`),
			Expression: `.config += {port: 8080}`,
			Expected: []string{huml(`
config:
  host: "localhost"
  port: 8080
`)},
		},
	},
}

var deleteScenarios = ScenarioGroup{
	Name:        "delete",
	Description: "del() function removes values",
	Scenarios: []Scenario{
		{
			Description: "delete field",
			Document: huml(`
name: "Alice"
password: "secret"
`),
			Expression: `del(.password)`,
			Expected: []string{huml(`
name: "Alice"
`)},
		},
		{
			Description: "delete multiple fields",
			Document: huml(`
name: "Alice"
password: "secret"
token: "abc123"
`),
			Expression: `del(.password, .token)`,
			Expected: []string{huml(`
name: "Alice"
`)},
		},
		{
			Description: "delete nested field",
			Document: huml(`
user:
  name: "Alice"
  internal:
    id: 123
`),
			Expression: `del(.user.internal)`,
			Expected: []string{huml(`
user:
  name: "Alice"
`)},
		},
		{
			Description: "delete array element",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `del(.[1])`,
			Expected: []string{huml(`
- "a"
- "c"
`)},
		},
		{
			Description: "delete matching array elements",
			Document: huml(`
- name: "Alice"
  active: true
- name: "Bob"
  active: false
- name: "Carol"
  active: true
`),
			Expression: `del(.[] | select(.active == false))`,
			Expected: []string{huml(`
- name: "Alice"
  active: true
- name: "Carol"
  active: true
`)},
		},
		{
			Description: "delete non-existent field is no-op",
			Document: huml(`
name: "Alice"
`),
			Expression: `del(.nonexistent)`,
			Expected: []string{huml(`
name: "Alice"
`)},
		},
	},
}

func TestAssignmentScenarios(t *testing.T) {
	runScenarios(t, assignmentScenarios)
}

func TestUpdateScenarios(t *testing.T) {
	runScenarios(t, updateScenarios)
}

func TestAddAssignScenarios(t *testing.T) {
	runScenarios(t, addAssignScenarios)
}

func TestDeleteScenarios(t *testing.T) {
	runScenarios(t, deleteScenarios)
}
