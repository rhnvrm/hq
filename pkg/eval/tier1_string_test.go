package eval

import "testing"

// String operation tests
// Tier 1 - Essential (90% of use cases)

var stringCaseScenarios = ScenarioGroup{
	Name:        "string-case",
	Description: "Case conversion functions",
	Scenarios: []Scenario{
		{
			Description: "ascii_downcase",
			Document:    `"Hello World"`,
			Expression:  `ascii_downcase`,
			Expected:    []string{`"hello world"`},
		},
		{
			Description: "ascii_upcase",
			Document:    `"Hello World"`,
			Expression:  `ascii_upcase`,
			Expected:    []string{`"HELLO WORLD"`},
		},
		{
			Description: "ascii_downcase empty",
			Document:    `""`,
			Expression:  `ascii_downcase`,
			Expected:    []string{`""`},
		},
		{
			Description: "ascii_upcase with numbers",
			Document:    `"abc123def"`,
			Expression:  `ascii_upcase`,
			Expected:    []string{`"ABC123DEF"`},
		},
		{
			Description: "ascii_downcase in pipe",
			Document: huml(`
name: "ALICE"
`),
			Expression: `.name | ascii_downcase`,
			Expected:   []string{`"alice"`},
		},
	},
}

var stringTrimScenarios = ScenarioGroup{
	Name:        "string-trim",
	Description: "String trimming functions",
	Scenarios: []Scenario{
		{
			Description: "ltrimstr",
			Document:    `"prefix_value"`,
			Expression:  `ltrimstr("prefix_")`,
			Expected:    []string{`"value"`},
		},
		{
			Description: "ltrimstr not matching",
			Document:    `"value"`,
			Expression:  `ltrimstr("prefix_")`,
			Expected:    []string{`"value"`},
		},
		{
			Description: "rtrimstr",
			Document:    `"value_suffix"`,
			Expression:  `rtrimstr("_suffix")`,
			Expected:    []string{`"value"`},
		},
		{
			Description: "rtrimstr not matching",
			Document:    `"value"`,
			Expression:  `rtrimstr("_suffix")`,
			Expected:    []string{`"value"`},
		},
		{
			Description: "trim whitespace",
			Document:    `"  hello world  "`,
			Expression:  `trim`,
			Expected:    []string{`"hello world"`},
		},
		{
			Description: "trim tabs and newlines",
			Document:    `"\t\nhello\n\t"`,
			Expression:  `trim`,
			Expected:    []string{`"hello"`},
		},
		{
			Description: "trim empty string",
			Document:    `"   "`,
			Expression:  `trim`,
			Expected:    []string{`""`},
		},
	},
}

var splitJoinScenarios = ScenarioGroup{
	Name:        "split-join",
	Description: "split and join functions",
	Scenarios: []Scenario{
		{
			Description: "split by comma",
			Document:    `"a,b,c"`,
			Expression:  `split(",")`,
			Expected:    []string{`["a", "b", "c"]`},
		},
		{
			Description: "split by space",
			Document:    `"hello world"`,
			Expression:  `split(" ")`,
			Expected:    []string{`["hello", "world"]`},
		},
		{
			Description: "split empty string",
			Document:    `""`,
			Expression:  `split(",")`,
			Expected:    []string{`[""]`},
		},
		{
			Description: "split no matches",
			Document:    `"hello"`,
			Expression:  `split(",")`,
			Expected:    []string{`["hello"]`},
		},
		{
			Description: "join with dash",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `join("-")`,
			Expected:   []string{`"a-b-c"`},
		},
		{
			Description: "join with empty string",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `join("")`,
			Expected:   []string{`"abc"`},
		},
		{
			Description: "join empty array",
			Document:    `[]`,
			Expression:  `join(",")`,
			Expected:    []string{`""`},
		},
		{
			Description: "split then join",
			Document:    `"a,b,c"`,
			Expression:  `split(",") | join("-")`,
			Expected:    []string{`"a-b-c"`},
		},
		{
			Description: "join with numbers",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `join(",")`,
			Expected:   []string{`"1,2,3"`},
		},
	},
}

var stringCheckScenarios = ScenarioGroup{
	Name:        "string-check",
	Description: "String checking functions",
	Scenarios: []Scenario{
		{
			Description: "contains true",
			Document:    `"hello world"`,
			Expression:  `contains("world")`,
			Expected:    []string{`true`},
		},
		{
			Description: "contains false",
			Document:    `"hello world"`,
			Expression:  `contains("foo")`,
			Expected:    []string{`false`},
		},
		{
			Description: "contains empty string",
			Document:    `"hello"`,
			Expression:  `contains("")`,
			Expected:    []string{`true`},
		},
		{
			Description: "startswith true",
			Document:    `"hello world"`,
			Expression:  `startswith("hello")`,
			Expected:    []string{`true`},
		},
		{
			Description: "startswith false",
			Document:    `"hello world"`,
			Expression:  `startswith("world")`,
			Expected:    []string{`false`},
		},
		{
			Description: "endswith true",
			Document:    `"hello world"`,
			Expression:  `endswith("world")`,
			Expected:    []string{`true`},
		},
		{
			Description: "endswith false",
			Document:    `"hello world"`,
			Expression:  `endswith("hello")`,
			Expected:    []string{`false`},
		},
		{
			Description: "startswith in select",
			Document: huml(`
- "user_alice"
- "admin_bob"
- "user_carol"
`),
			Expression: `[.[] | select(startswith("user_"))]`,
			Expected:   []string{`["user_alice", "user_carol"]`},
		},
		{
			Description: "endswith in select",
			Document: huml(`
- "config.json"
- "data.yaml"
- "schema.json"
`),
			Expression: `[.[] | select(endswith(".json"))]`,
			Expected:   []string{`["config.json", "schema.json"]`},
		},
	},
}

var stringInterpolationScenarios = ScenarioGroup{
	Name:        "string-interpolation",
	Description: "String interpolation with \\(...)",
	Scenarios: []Scenario{
		{
			Description: "interpolate field",
			Document: huml(`
name: "Alice"
`),
			Expression: `"Hello, \(.name)!"`,
			Expected:   []string{`"Hello, Alice!"`},
		},
		{
			Description: "interpolate multiple fields",
			Document: huml(`
firstName: "Alice"
lastName: "Smith"
`),
			Expression: `"\(.firstName) \(.lastName)"`,
			Expected:   []string{`"Alice Smith"`},
		},
		{
			Description: "interpolate with expression",
			Document: huml(`
x: 10
y: 20
`),
			Expression: `"Sum: \(.x + .y)"`,
			Expected:   []string{`"Sum: 30"`},
		},
		{
			Description: "interpolate number",
			Document: huml(`
count: 42
`),
			Expression: `"Count is \(.count)"`,
			Expected:   []string{`"Count is 42"`},
		},
		{
			Description: "build URL",
			Document: huml(`
host: "api.example.com"
id: 123
`),
			Expression: `"https://\(.host)/users/\(.id)"`,
			Expected:   []string{`"https://api.example.com/users/123"`},
		},
		{
			Description: "nested interpolation",
			Document: huml(`
user:
  name: "Alice"
`),
			Expression: `"User: \(.user.name)"`,
			Expected:   []string{`"User: Alice"`},
		},
	},
}

func TestStringCaseScenarios(t *testing.T) {
	runScenarios(t, stringCaseScenarios)
}

func TestStringTrimScenarios(t *testing.T) {
	runScenarios(t, stringTrimScenarios)
}

func TestSplitJoinScenarios(t *testing.T) {
	runScenarios(t, splitJoinScenarios)
}

func TestStringCheckScenarios(t *testing.T) {
	runScenarios(t, stringCheckScenarios)
}

func TestStringInterpolationScenarios(t *testing.T) {
	runScenarios(t, stringInterpolationScenarios)
}
