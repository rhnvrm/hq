package eval

import "testing"

// Object transformation tests
// Tier 2 - Important (next 8% of use cases)

var toEntriesScenarios = ScenarioGroup{
	Name:        "to_entries",
	Description: "to_entries converts object to array of {key, value}",
	Scenarios: []Scenario{
		{
			Description: "to_entries simple object",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `to_entries`,
			Expected:   []string{`[{"key": "a", "value": 1}, {"key": "b", "value": 2}, {"key": "c", "value": 3}]`},
		},
		{
			Description: "to_entries nested object",
			Document: huml(`
name: "Alice"
config:
  debug: true
`),
			Expression: `to_entries`,
			Expected:   []string{`[{"key": "name", "value": "Alice"}, {"key": "config", "value": {"debug": true}}]`},
		},
		{
			Description: "to_entries empty object",
			Document:    `{}`,
			Expression:  `to_entries`,
			Expected:    []string{`[]`},
		},
		{
			Description: "to_entries array converts to indices",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `to_entries`,
			Expected:   []string{`[{"key": 0, "value": "a"}, {"key": 1, "value": "b"}, {"key": 2, "value": "c"}]`},
		},
	},
}

var fromEntriesScenarios = ScenarioGroup{
	Name:        "from_entries",
	Description: "from_entries converts array of {key, value} to object",
	Scenarios: []Scenario{
		{
			Description: "from_entries simple",
			Document: huml(`
- key: "a"
  value: 1
- key: "b"
  value: 2
`),
			Expression: `from_entries`,
			Expected:   []string{`{"a": 1, "b": 2}`},
		},
		{
			Description: "from_entries with name/value",
			Document: huml(`
- name: "a"
  value: 1
- name: "b"
  value: 2
`),
			Expression: `from_entries`,
			Expected:   []string{`{"a": 1, "b": 2}`},
		},
		{
			Description: "from_entries with k/v",
			Document: huml(`
- k: "a"
  v: 1
- k: "b"
  v: 2
`),
			Expression: `from_entries`,
			Expected:   []string{`{"a": 1, "b": 2}`},
		},
		{
			Description: "from_entries empty",
			Document:    `[]`,
			Expression:  `from_entries`,
			Expected:    []string{`{}`},
		},
	},
}

var withEntriesScenarios = ScenarioGroup{
	Name:        "with_entries",
	Description: "with_entries transforms object via entries",
	Scenarios: []Scenario{
		{
			Description: "with_entries prefix keys",
			Document: huml(`
host: "localhost"
port: 8080
`),
			Expression: `with_entries(.key = "APP_" + .key)`,
			Expected:   []string{`{"APP_host": "localhost", "APP_port": 8080}`},
		},
		{
			Description: "with_entries uppercase keys",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `with_entries(.key |= ascii_upcase)`,
			Expected:   []string{`{"NAME": "Alice", "AGE": 30}`},
		},
		{
			Description: "with_entries transform values",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `with_entries(.value *= 10)`,
			Expected:   []string{`{"a": 10, "b": 20, "c": 30}`},
		},
		{
			Description: "with_entries filter entries",
			Document: huml(`
name: "Alice"
password: "secret"
email: "alice@example.com"
`),
			Expression: `with_entries(select(.key != "password"))`,
			Expected:   []string{`{"name": "Alice", "email": "alice@example.com"}`},
		},
		{
			Description: "with_entries remove null values",
			Document: huml(`
a: 1
b: null
c: 3
`),
			Expression: `with_entries(select(.value != null))`,
			Expected:   []string{`{"a": 1, "c": 3}`},
		},
	},
}

var mapValuesScenarios = ScenarioGroup{
	Name:        "map_values",
	Description: "map_values transforms only values",
	Scenarios: []Scenario{
		{
			Description: "map_values multiply",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `map_values(. * 2)`,
			Expected:   []string{`{"a": 2, "b": 4, "c": 6}`},
		},
		{
			Description: "map_values to string",
			Document: huml(`
x: 10
y: 20
`),
			Expression: `map_values("value: " + tostring)`,
			Expected:   []string{`{"x": "value: 10", "y": "value: 20"}`},
		},
		{
			Description: "map_values nested update",
			Document: huml(`
users:
  alice:
    score: 100
  bob:
    score: 200
`),
			Expression: `.users | map_values(.score += 50)`,
			Expected:   []string{`{"alice": {"score": 150}, "bob": {"score": 250}}`},
		},
	},
}

func TestToEntriesScenarios(t *testing.T) {
	runScenarios(t, toEntriesScenarios)
}

func TestFromEntriesScenarios(t *testing.T) {
	runScenarios(t, fromEntriesScenarios)
}

func TestWithEntriesScenarios(t *testing.T) {
	runScenarios(t, withEntriesScenarios)
}

func TestMapValuesScenarios(t *testing.T) {
	runScenarios(t, mapValuesScenarios)
}
