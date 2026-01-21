package eval

import "testing"

// Real HUML input tests
// These tests verify that hq correctly parses real HUML syntax with ::
// as defined in the HUML specification (https://huml.io/specifications/v0-2-0/)

// realHUML is a helper that marks the string as real HUML syntax
// (as opposed to YAML-like syntax that the harness also accepts)
func realHUML(s string) string {
	return s
}

var humlBasicScenarios = ScenarioGroup{
	Name:        "huml-basic",
	Description: "Basic HUML syntax with :: for complex types",
	Scenarios: []Scenario{
		{
			Description: "nested dict with :: syntax",
			Document: realHUML(`server::
  host: "localhost"
  port: 8080`),
			Expression: `.server.host`,
			Expected:   []string{`"localhost"`},
		},
		{
			Description: "nested dict port access",
			Document: realHUML(`server::
  host: "localhost"
  port: 8080`),
			Expression: `.server.port`,
			Expected:   []string{`8080`},
		},
		{
			Description: "inline list with :: syntax",
			Document:    realHUML(`ports:: 80, 443, 8080`),
			Expression:  `.ports`,
			Expected:    []string{`[80, 443, 8080]`},
		},
		{
			Description: "inline list iteration",
			Document:    realHUML(`ports:: 80, 443, 8080`),
			Expression:  `.ports[]`,
			Expected:    []string{`80`, `443`, `8080`},
		},
		{
			Description: "inline dict with :: syntax",
			Document:    realHUML(`config:: host: "localhost", port: 8080`),
			Expression:  `.config.host`,
			Expected:    []string{`"localhost"`},
		},
		{
			Description: "multiline list with :: syntax",
			Document: realHUML(`tags::
  - "web"
  - "api"
  - "v2"`),
			Expression: `.tags[]`,
			Expected:   []string{`"web"`, `"api"`, `"v2"`},
		},
		{
			Description: "empty list with :: syntax",
			Document:    realHUML(`items:: []`),
			Expression:  `.items`,
			Expected:    []string{`[]`},
		},
		{
			Description: "empty dict with :: syntax",
			Document:    realHUML(`config:: {}`),
			Expression:  `.config`,
			Expected:    []string{`{}`},
		},
	},
}

var humlNestedScenarios = ScenarioGroup{
	Name:        "huml-nested",
	Description: "Nested HUML structures with :: syntax",
	Scenarios: []Scenario{
		{
			Description: "deeply nested dicts",
			Document: realHUML(`app::
  database::
    primary::
      host: "db1.example.com"
      port: 5432
    replica::
      host: "db2.example.com"
      port: 5432`),
			Expression: `.app.database.primary.host`,
			Expected:   []string{`"db1.example.com"`},
		},
		{
			Description: "nested dict replica access",
			Document: realHUML(`app::
  database::
    primary::
      host: "db1.example.com"
      port: 5432
    replica::
      host: "db2.example.com"
      port: 5432`),
			Expression: `.app.database.replica.host`,
			Expected:   []string{`"db2.example.com"`},
		},
		{
			Description: "list of dicts with :: syntax",
			Document: realHUML(`users::
  - ::
    name: "Alice"
    role: "admin"
  - ::
    name: "Bob"
    role: "user"`),
			Expression: `.users[].name`,
			Expected:   []string{`"Alice"`, `"Bob"`},
		},
		{
			Description: "list of dicts filter",
			Document: realHUML(`users::
  - ::
    name: "Alice"
    role: "admin"
    active: true
  - ::
    name: "Bob"
    role: "user"
    active: false
  - ::
    name: "Carol"
    role: "user"
    active: true`),
			Expression: `.users[] | select(.active) | .name`,
			Expected:   []string{`"Alice"`, `"Carol"`},
		},
		{
			Description: "list of dicts with nested objects",
			Document: realHUML(`servers::
  - ::
    id: "node-1"
    config::
      cpu: 4
      ram: 16
  - ::
    id: "node-2"
    config::
      cpu: 8
      ram: 32`),
			Expression: `.servers[].config.cpu`,
			Expected:   []string{`4`, `8`},
		},
	},
}

var humlInlineScenarios = ScenarioGroup{
	Name:        "huml-inline",
	Description: "Inline HUML list and dict syntax",
	Scenarios: []Scenario{
		{
			Description: "inline list of strings",
			Document:    realHUML(`tags:: "web", "api", "backend"`),
			Expression:  `.tags | length`,
			Expected:    []string{`3`},
		},
		{
			Description: "inline list first element",
			Document:    realHUML(`colors:: "red", "green", "blue"`),
			Expression:  `.colors[0]`,
			Expected:    []string{`"red"`},
		},
		{
			Description: "inline list last element",
			Document:    realHUML(`colors:: "red", "green", "blue"`),
			Expression:  `.colors[-1]`,
			Expected:    []string{`"blue"`},
		},
		{
			Description: "inline dict multiple keys",
			Document:    realHUML(`point:: x: 10, y: 20, z: 30`),
			Expression:  `.point | .x + .y + .z`,
			Expected:    []string{`60`},
		},
		{
			Description: "inline dict keys",
			Document:    realHUML(`point:: x: 10, y: 20`),
			Expression:  `.point | keys`,
			Expected:    []string{`["x", "y"]`},
		},
		{
			Description: "mixed inline and multiline",
			Document: realHUML(`server::
  host: "localhost"
  ports:: 80, 443`),
			Expression: `.server.ports[]`,
			Expected:   []string{`80`, `443`},
		},
	},
}

var humlMixedTypesScenarios = ScenarioGroup{
	Name:        "huml-mixed-types",
	Description: "HUML with mixed scalar and complex types",
	Scenarios: []Scenario{
		{
			Description: "boolean values in HUML",
			Document: realHUML(`config::
  debug: true
  production: false`),
			Expression: `.config.debug`,
			Expected:   []string{`true`},
		},
		{
			Description: "null value in HUML",
			Document: realHUML(`user::
  name: "Alice"
  email: null`),
			Expression: `.user.email`,
			Expected:   []string{`null`},
		},
		{
			Description: "numeric types",
			Document: realHUML(`values::
  integer: 42
  float: 3.14
  negative: -100`),
			Expression: `.values.float`,
			Expected:   []string{`3.14`},
		},
		{
			Description: "complex nested structure",
			Document: realHUML(`application::
  name: "myapp"
  version: "1.0.0"
  features::
    - "auth"
    - "logging"
  database::
    host: "localhost"
    port: 5432
    options:: ssl: true, timeout: 30`),
			Expression: `.application.database.options.ssl`,
			Expected:   []string{`true`},
		},
		{
			Description: "features list length",
			Document: realHUML(`application::
  name: "myapp"
  features::
    - "auth"
    - "logging"`),
			Expression: `.application.features | length`,
			Expected:   []string{`2`},
		},
	},
}

var humlTransformScenarios = ScenarioGroup{
	Name:        "huml-transform",
	Description: "Transforming real HUML data",
	Scenarios: []Scenario{
		{
			Description: "map over HUML list",
			Document:    realHUML(`numbers:: 1, 2, 3, 4, 5`),
			Expression:  `.numbers | map(. * 2)`,
			Expected:    []string{`[2, 4, 6, 8, 10]`},
		},
		{
			Description: "filter HUML list",
			Document:    realHUML(`numbers:: 1, 2, 3, 4, 5`),
			Expression:  `[.numbers[] | select(. > 3)]`,
			Expected:    []string{`[4, 5]`},
		},
		{
			Description: "transform list of dicts",
			Document: realHUML(`users::
  - ::
    name: "Alice"
    age: 30
  - ::
    name: "Bob"
    age: 25`),
			Expression: `.users | map({username: .name, adult: (.age >= 18)})`,
			Expected:   []string{`[{"username": "Alice", "adult": true}, {"username": "Bob", "adult": true}]`},
		},
		{
			Description: "reduce HUML list",
			Document:    realHUML(`prices:: 10.5, 20.0, 15.5`),
			Expression:  `reduce .prices[] as $p (0; . + $p)`,
			Expected:    []string{`46`},
		},
		{
			Description: "group by on HUML data",
			Document: realHUML(`items::
  - ::
    category: "fruit"
    name: "apple"
  - ::
    category: "vegetable"
    name: "carrot"
  - ::
    category: "fruit"
    name: "banana"`),
			Expression: `.items | group_by(.category) | length`,
			Expected:   []string{`2`},
		},
	},
}

var humlCommentsScenarios = ScenarioGroup{
	Name:        "huml-comments",
	Description: "HUML with comments",
	Scenarios: []Scenario{
		{
			Description: "HUML with inline comment",
			Document: realHUML(`server::
  host: "localhost" # Primary server
  port: 8080`),
			Expression: `.server.host`,
			Expected:   []string{`"localhost"`},
		},
		{
			Description: "HUML with comment-only line",
			Document: realHUML(`# Configuration file
app::
  name: "myservice"
  # Database settings
  db::
    host: "localhost"`),
			Expression: `.app.db.host`,
			Expected:   []string{`"localhost"`},
		},
	},
}

func TestHUMLBasicScenarios(t *testing.T) {
	runScenarios(t, humlBasicScenarios)
}

func TestHUMLNestedScenarios(t *testing.T) {
	runScenarios(t, humlNestedScenarios)
}

func TestHUMLInlineScenarios(t *testing.T) {
	runScenarios(t, humlInlineScenarios)
}

func TestHUMLMixedTypesScenarios(t *testing.T) {
	runScenarios(t, humlMixedTypesScenarios)
}

func TestHUMLTransformScenarios(t *testing.T) {
	runScenarios(t, humlTransformScenarios)
}

func TestHUMLCommentsScenarios(t *testing.T) {
	runScenarios(t, humlCommentsScenarios)
}
