package eval

import "testing"

// Path operation tests
// Tier 2 - Important (next 8% of use cases)

var pathScenarios = ScenarioGroup{
	Name:        "path",
	Description: "path returns path to value",
	Scenarios: []Scenario{
		{
			Description: "path to field",
			Document: huml(`
a:
  b:
    c: 1
`),
			Expression: `path(.a.b.c)`,
			Expected:   []string{`["a", "b", "c"]`},
		},
		{
			Description: "path to array element",
			Document: huml(`
items:
  - "a"
  - "b"
  - "c"
`),
			Expression: `path(.items[1])`,
			Expected:   []string{`["items", 1]`},
		},
		{
			Description: "paths to all leaves",
			Document: huml(`
a: 1
b:
  c: 2
  d: 3
`),
			Expression: `[paths(scalars)]`,
			Expected:   []string{`[["a"], ["b", "c"], ["b", "d"]]`},
		},
		{
			Description: "paths to all arrays",
			Document: huml(`
items:
  - 1
  - 2
nested:
  data:
    - 3
`),
			Expression: `[paths(type == "array")]`,
			Expected:   []string{`[["items"], ["nested", "data"]]`},
		},
	},
}

var getpathScenarios = ScenarioGroup{
	Name:        "getpath",
	Description: "getpath retrieves value at path",
	Scenarios: []Scenario{
		{
			Description: "getpath simple",
			Document: huml(`
a:
  b:
    c: 42
`),
			Expression: `getpath(["a", "b", "c"])`,
			Expected:   []string{`42`},
		},
		{
			Description: "getpath array index",
			Document: huml(`
items:
  - "first"
  - "second"
  - "third"
`),
			Expression: `getpath(["items", 1])`,
			Expected:   []string{`"second"`},
		},
		{
			Description: "getpath nonexistent",
			Document: huml(`
a: 1
`),
			Expression: `getpath(["x", "y", "z"])`,
			Expected:   []string{`null`},
		},
		{
			Description: "getpath root",
			Document:    `42`,
			Expression:  `getpath([])`,
			Expected:    []string{`42`},
		},
	},
}

var setpathScenarios = ScenarioGroup{
	Name:        "setpath",
	Description: "setpath sets value at path",
	Scenarios: []Scenario{
		{
			Description: "setpath existing",
			Document: huml(`
a:
  b: 1
`),
			Expression: `setpath(["a", "b"]; 42)`,
			Expected: []string{huml(`
a:
  b: 42
`)},
		},
		{
			Description: "setpath creates path",
			Document: huml(`
a: 1
`),
			Expression: `setpath(["x", "y", "z"]; "value")`,
			Expected: []string{huml(`
a: 1
x:
  y:
    z: "value"
`)},
		},
		{
			Description: "setpath array index",
			Document: huml(`
items:
  - "a"
  - "b"
  - "c"
`),
			Expression: `setpath(["items", 1]; "X")`,
			Expected: []string{huml(`
items:
  - "a"
  - "X"
  - "c"
`)},
		},
	},
}

var delpathsScenarios = ScenarioGroup{
	Name:        "delpaths",
	Description: "delpaths removes values at paths",
	Scenarios: []Scenario{
		{
			Description: "delpaths single path",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `delpaths([["b"]])`,
			Expected: []string{huml(`
a: 1
c: 3
`)},
		},
		{
			Description: "delpaths multiple paths",
			Document: huml(`
a: 1
b: 2
c: 3
d: 4
`),
			Expression: `delpaths([["a"], ["c"]])`,
			Expected: []string{huml(`
b: 2
d: 4
`)},
		},
		{
			Description: "delpaths nested",
			Document: huml(`
user:
  name: "Alice"
  password: "secret"
  email: "alice@example.com"
`),
			Expression: `delpaths([["user", "password"]])`,
			Expected: []string{huml(`
user:
  name: "Alice"
  email: "alice@example.com"
`)},
		},
	},
}

var containsInsideScenarios = ScenarioGroup{
	Name:        "contains-inside",
	Description: "contains and inside for subset checking",
	Scenarios: []Scenario{
		{
			Description: "array contains subset",
			Document: huml(`
- 1
- 2
- 3
- 4
- 5
`),
			Expression: `contains([2, 4])`,
			Expected:   []string{`true`},
		},
		{
			Description: "array not contains",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `contains([5])`,
			Expected:   []string{`false`},
		},
		{
			Description: "object contains subset",
			Document: huml(`
name: "Alice"
age: 30
city: "NYC"
`),
			Expression: `contains({name: "Alice"})`,
			Expected:   []string{`true`},
		},
		{
			Description: "object contains nested",
			Document: huml(`
user:
  name: "Alice"
  address:
    city: "NYC"
`),
			Expression: `contains({user: {address: {city: "NYC"}}})`,
			Expected:   []string{`true`},
		},
		{
			Description: "inside array",
			Document: huml(`
- 2
- 4
`),
			Expression: `inside([1, 2, 3, 4, 5])`,
			Expected:   []string{`true`},
		},
		{
			Description: "inside object",
			Document: huml(`
name: "Alice"
`),
			Expression: `inside({name: "Alice", age: 30})`,
			Expected:   []string{`true`},
		},
	},
}

func TestPathScenarios(t *testing.T) {
	runScenarios(t, pathScenarios)
}

func TestGetpathScenarios(t *testing.T) {
	runScenarios(t, getpathScenarios)
}

func TestSetpathScenarios(t *testing.T) {
	runScenarios(t, setpathScenarios)
}

func TestDelpathsScenarios(t *testing.T) {
	runScenarios(t, delpathsScenarios)
}

func TestContainsInsideScenarios(t *testing.T) {
	runScenarios(t, containsInsideScenarios)
}
