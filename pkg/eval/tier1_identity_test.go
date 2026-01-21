package eval

import "testing"

// Identity and Navigation tests
// Tier 1 - Essential (90% of use cases)

var identityScenarios = ScenarioGroup{
	Name:        "identity",
	Description: "Identity operator (.) returns input unchanged",
	Scenarios: []Scenario{
		{
			Description: "identity returns scalar",
			Document:    `"hello"`,
			Expression:  `.`,
			Expected:    []string{`"hello"`},
		},
		{
			Description: "identity returns number",
			Document:    `42`,
			Expression:  `.`,
			Expected:    []string{`42`},
		},
		{
			Description: "identity returns boolean",
			Document:    `true`,
			Expression:  `.`,
			Expected:    []string{`true`},
		},
		{
			Description: "identity returns null",
			Document:    `null`,
			Expression:  `.`,
			Expected:    []string{`null`},
		},
		{
			Description: "identity returns object",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `.`,
			Expected: []string{huml(`
name: "Alice"
age: 30
`)},
		},
		{
			Description: "identity returns array",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `.`,
			Expected: []string{huml(`
- 1
- 2
- 3
`)},
		},
	},
}

var fieldAccessScenarios = ScenarioGroup{
	Name:        "field-access",
	Description: "Field access with dot notation",
	Scenarios: []Scenario{
		{
			Description: "access simple field",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `.name`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "access numeric field",
			Document: huml(`
name: "Alice"
age: 30
`),
			Expression: `.age`,
			Expected:   []string{`30`},
		},
		{
			Description: "access nested field",
			Document: huml(`
user:
  name: "Alice"
  address:
    city: "NYC"
`),
			Expression: `.user.name`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "access deeply nested field",
			Document: huml(`
user:
  name: "Alice"
  address:
    city: "NYC"
`),
			Expression: `.user.address.city`,
			Expected:   []string{`"NYC"`},
		},
		{
			Description: "access missing field returns null",
			Document: huml(`
name: "Alice"
`),
			Expression: `.age`,
			Expected:   []string{`null`},
		},
		{
			Description: "access nested missing field returns null",
			Document: huml(`
user:
  name: "Alice"
`),
			Expression: `.user.address.city`,
			Expected:   []string{`null`},
		},
		{
			Description: "bracket notation for simple key",
			Document: huml(`
name: "Alice"
`),
			Expression: `.["name"]`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "bracket notation for key with spaces",
			Document: huml(`
"first name": "Alice"
`),
			Expression: `.["first name"]`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "bracket notation for numeric key",
			Document: huml(`
"123": "value"
`),
			Expression: `.["123"]`,
			Expected:   []string{`"value"`},
		},
	},
}

var arrayAccessScenarios = ScenarioGroup{
	Name:        "array-access",
	Description: "Array indexing and iteration",
	Scenarios: []Scenario{
		{
			Description: "access first element",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `.[0]`,
			Expected:   []string{`"a"`},
		},
		{
			Description: "access middle element",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `.[1]`,
			Expected:   []string{`"b"`},
		},
		{
			Description: "access last element with negative index",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `.[-1]`,
			Expected:   []string{`"c"`},
		},
		{
			Description: "access second-to-last with negative index",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `.[-2]`,
			Expected:   []string{`"b"`},
		},
		{
			Description: "out of bounds returns null",
			Document: huml(`
- "a"
- "b"
`),
			Expression: `.[10]`,
			Expected:   []string{`null`},
		},
		{
			Description: "negative out of bounds returns null",
			Document: huml(`
- "a"
- "b"
`),
			Expression: `.[-10]`,
			Expected:   []string{`null`},
		},
		{
			Description: "array element field access",
			Document: huml(`
- name: "Alice"
- name: "Bob"
`),
			Expression: `.[0].name`,
			Expected:   []string{`"Alice"`},
		},
	},
}

var sliceScenarios = ScenarioGroup{
	Name:        "slice",
	Description: "Array and string slicing",
	Scenarios: []Scenario{
		{
			Description: "slice middle elements",
			Document: huml(`
- "a"
- "b"
- "c"
- "d"
- "e"
`),
			Expression: `.[1:4]`,
			Expected: []string{huml(`
- "b"
- "c"
- "d"
`)},
		},
		{
			Description: "slice from beginning",
			Document: huml(`
- "a"
- "b"
- "c"
- "d"
`),
			Expression: `.[:2]`,
			Expected: []string{huml(`
- "a"
- "b"
`)},
		},
		{
			Description: "slice to end",
			Document: huml(`
- "a"
- "b"
- "c"
- "d"
`),
			Expression: `.[2:]`,
			Expected: []string{huml(`
- "c"
- "d"
`)},
		},
		{
			Description: "slice with negative end",
			Document: huml(`
- "a"
- "b"
- "c"
- "d"
`),
			Expression: `.[1:-1]`,
			Expected: []string{huml(`
- "b"
- "c"
`)},
		},
		{
			Description: "slice last two elements",
			Document: huml(`
- "a"
- "b"
- "c"
- "d"
`),
			Expression: `.[-2:]`,
			Expected: []string{huml(`
- "c"
- "d"
`)},
		},
		{
			Description: "slice string",
			Document:    `"hello world"`,
			Expression:  `.[0:5]`,
			Expected:    []string{`"hello"`},
		},
	},
}

var iteratorScenarios = ScenarioGroup{
	Name:        "iterator",
	Description: "Iterator/splat operator (.[]) returns all values",
	Scenarios: []Scenario{
		{
			Description: "iterate array values",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `.[]`,
			Expected:   []string{`"a"`, `"b"`, `"c"`},
		},
		{
			Description: "iterate object values",
			Document: huml(`
a: 1
b: 2
c: 3
`),
			Expression: `.[]`,
			Expected:   []string{`1`, `2`, `3`},
		},
		{
			Description: "iterate nested array",
			Document: huml(`
users:
  - name: "Alice"
  - name: "Bob"
`),
			Expression: `.users[]`,
			Expected: []string{huml(`
name: "Alice"
`), huml(`
name: "Bob"
`)},
		},
		{
			Description: "iterate and access field",
			Document: huml(`
users:
  - name: "Alice"
  - name: "Bob"
`),
			Expression: `.users[].name`,
			Expected:   []string{`"Alice"`, `"Bob"`},
		},
		{
			Description: "iterate empty array",
			Document:    `[]`,
			Expression:  `.[]`,
			Expected:    []string{},
		},
		{
			Description: "iterate empty object",
			Document:    `{}`,
			Expression:  `.[]`,
			Expected:    []string{},
		},
	},
}

func TestIdentityScenarios(t *testing.T) {
	runScenarios(t, identityScenarios)
}

func TestFieldAccessScenarios(t *testing.T) {
	runScenarios(t, fieldAccessScenarios)
}

func TestArrayAccessScenarios(t *testing.T) {
	runScenarios(t, arrayAccessScenarios)
}

func TestSliceScenarios(t *testing.T) {
	runScenarios(t, sliceScenarios)
}

func TestIteratorScenarios(t *testing.T) {
	runScenarios(t, iteratorScenarios)
}
