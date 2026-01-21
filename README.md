# hq - HUML Query Processor

A lightweight, portable command-line HUML processor with jq-compatible syntax.

```bash
# Extract a field
hq '.name' config.huml

# Filter an array
hq '.users[] | select(.active == true)' data.huml

# Transform data
hq '.items | map({id, name})' data.huml

# Output as JSON
hq -o json '.' config.huml
```

## Status

**Work in Progress** - TDD test harness complete (400+ test scenarios), implementation pending.

```bash
just progress
# Expression Tests: 397 pending
# Implementation: 0%
```

## Features (Planned)

### Tier 1 - Essential (90% of use cases)
- Navigation: `.`, `.foo`, `.[]`, `.[n]`, `.[n:m]`
- Pipe and composition: `|`, `,`, `()`
- Filtering: `select`, comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Boolean: `and`, `or`, `not`
- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Functions: `length`, `keys`, `has`, `type`, `map`, `sort`, `unique`, `group_by`
- Assignment: `=`, `|=`, `+=`, `del`
- String ops: `split`, `join`, `ascii_downcase`, `ascii_upcase`, `contains`, `startswith`, `endswith`
- Construction: `{...}`, `[...]`
- Default: `//`

### Tier 2 - Important (next 8% of use cases)
- Regex: `test`, `match`, `capture`, `sub`, `gsub`
- Object transforms: `to_entries`, `from_entries`, `with_entries`
- Conditionals: `if-then-else`
- Variables: `as $var`, `--arg`
- Recursive descent: `..`
- Error handling: `try-catch`, `?`
- Path operations: `path`, `getpath`, `setpath`

## Installation

```bash
# From source (once implemented)
go install github.com/huml-lang/hq@latest
```

## Usage

```bash
# Basic usage
hq [options] <expression> [file...]

# Read from stdin
cat config.huml | hq '.database.host'

# Read from file
hq '.users[] | .name' users.huml

# Output formats
hq -o json '.' config.huml    # JSON output
hq -o yaml '.' config.huml    # YAML output
hq -r '.name' config.huml     # Raw string (no quotes)

# In-place editing
hq -i '.version = "2.0.0"' config.huml

# Null input (for generating data)
hq -n '{name: "test", values: [1,2,3]}'
```

## Development

### Prerequisites

- Go 1.21+
- [just](https://github.com/casey/just) (command runner)

### Quick Start

```bash
# Clone
git clone https://github.com/huml-lang/hq.git
cd hq

# Run tests
just test

# Check progress
just progress

# Run specific tests
just test-tier1              # Essential features
just test-tier2              # Important features
just test-pattern Select     # Specific operator
```

### Project Structure

```
hq/
├── cmd/                     # CLI entry point (urfave/cli v3)
│   └── cli_test.go          # CLI integration tests
├── pkg/
│   ├── eval/                # Expression evaluation engine
│   │   ├── scenario.go      # Test scenario definitions
│   │   ├── harness_test.go  # Test harness with semantic comparison
│   │   ├── tier1_*_test.go  # Essential feature tests (~250 scenarios)
│   │   └── tier2_*_test.go  # Important feature tests (~100 scenarios)
│   ├── parser/              # Expression parser (Participle-based)
│   └── types/               # Data model (wraps go-huml)
├── testdata/                # Sample HUML files for testing
├── scripts/                 # Development scripts
└── justfile                 # Development commands
```

### Test-Driven Development

This project follows TDD - all tests are written first and skip until implemented:

```bash
# See current progress
just progress

# Output:
# Expression Tests:
#   Total:     397
#   Passing:   0
#   Pending:   397
#
# Implementation Progress: 0%
# [----------------------------------------] 0%
```

As you implement features, tests automatically start passing.

### Available Commands

```bash
just --list          # Show all commands
just test            # Run all tests
just test-verbose    # Verbose test output
just progress        # Show implementation progress
just build           # Build binary
just fmt             # Format code
just lint            # Run linter
just loc             # Count lines of code
```

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Expression syntax | jq-compatible | User familiarity, easier adoption |
| Parser | Participle | Proven approach (used by yq), good errors |
| Data model | Wrap go-huml | We own go-huml, can fix issues upstream |
| CLI framework | urfave/cli v3 | Minimal deps, clean API |
| Test comparison | Semantic (koanf) | Format-agnostic, handles JSON/HUML/int/float |

## Related Projects

- [go-huml](https://github.com/huml-lang/go-huml) - Go library for HUML parsing/encoding
- [jq](https://github.com/jqlang/jq) - JSON processor (inspiration)
- [yq](https://github.com/mikefarah/yq) - YAML/JSON/XML processor (architecture reference)

## License

MIT

## Contributing

Contributions welcome! The TDD setup makes it easy to add new features:

1. Find the appropriate test file (`tier1_*_test.go` or `tier2_*_test.go`)
2. Add test scenarios (they'll skip until implemented)
3. Implement the feature
4. Watch tests pass!
