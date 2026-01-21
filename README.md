# hq - HUML Query Processor

A command-line processor for [HUML](https://huml.io) (Human-Oriented Markup Language) with jq-compatible syntax.

```bash
# Query a HUML config file
cat config.huml
# database::
#   host: "localhost"
#   port: 5432
# features:: "logging", "metrics"

hq '.database.host' config.huml
# "localhost"

hq '.features[]' config.huml
# "logging"
# "metrics"
```

## What is HUML?

HUML is a strict, human-readable configuration format. It looks like YAML but avoids its complexity and ambiguity. Key features:

- `::` suffix marks complex types (objects/arrays)
- Inline lists: `ports:: 80, 443`
- Inline dicts: `props:: host: "localhost", port: 8080`
- Multi-line strings with `"""`
- Comments with `#`

```huml
# HUML config example
server::
  host: "0.0.0.0"
  port: 8080

database::
  url: "postgres://localhost/mydb"
  pool_size: 10

users::
  - ::
    name: "Alice"
    role: "admin"
  - ::
    name: "Bob"
    role: "user"
```

hq lets you query, filter, and transform HUML data using familiar jq syntax. It also accepts JSON and YAML input.

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/rhnvrm/hq/releases/latest):

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | x86_64 | `hq_VERSION_linux_x86_64.tar.gz` |
| Linux | arm64 | `hq_VERSION_linux_arm64.tar.gz` |
| macOS | Apple Silicon | `hq_VERSION_darwin_arm64.tar.gz` |
| macOS | Intel | `hq_VERSION_darwin_x86_64.tar.gz` |
| Windows | x86_64 | `hq_VERSION_windows_x86_64.zip` |
| Windows | arm64 | `hq_VERSION_windows_arm64.zip` |

```bash
# Example: Linux x86_64
tar xzf hq_*_linux_x86_64.tar.gz
sudo mv hq /usr/local/bin/
```

### From Source

Requires Go 1.21+:

```bash
go install github.com/rhnvrm/hq/cmd/hq@latest
```

## Usage

```bash
hq [options] <expression> [file...]

# Query HUML files
hq '.server.port' config.huml

# Read from stdin
cat config.huml | hq '.database.url'

# Output as different formats
hq '.' config.huml              # HUML output (default)
hq -o json '.' config.huml      # JSON output
hq -o yaml '.' config.huml      # YAML output
hq -r '.server.host' config.huml # Raw string (no quotes)
```

## Examples

### Querying Config Files

```bash
# config.huml:
# app::
#   name: "myservice"
#   version: "1.2.3"
# database::
#   primary::
#     host: "db1.example.com"
#   replica::
#     host: "db2.example.com"

# Get nested value
hq '.database.primary.host' config.huml
# "db1.example.com"

# Get multiple values
hq '.database | .primary.host, .replica.host' config.huml
# "db1.example.com"
# "db2.example.com"
```

### Working with Arrays

```bash
# users.huml:
# users::
#   - ::
#     name: "Alice"
#     role: "admin"
#     active: true
#   - ::
#     name: "Bob"
#     role: "user"
#     active: false
#   - ::
#     name: "Carol"
#     role: "user"
#     active: true

# List all names
hq '.users[].name' users.huml
# "Alice"
# "Bob"
# "Carol"

# Filter active users
hq '.users[] | select(.active) | .name' users.huml
# "Alice"
# "Carol"

# Count by role
hq '.users | group_by(.role) | map({role: .[0].role, count: length})' users.huml
# [{role: "admin", count: 1}, {role: "user", count: 2}]
```

### Transforming Data

```bash
# Transform HUML structure
hq '.users | map({username: .name, is_admin: (.role == "admin")})' users.huml

# Add/modify fields
hq '.users[].environment = "production"' users.huml

# Delete fields
hq 'del(.users[].active)' users.huml
```

### Converting Formats

```bash
# HUML to JSON
hq -o json '.' config.huml > config.json

# JSON to HUML  
cat config.json | hq '.' > config.huml

# HUML to YAML
hq -o yaml '.' config.huml > config.yaml
```

## Features

Full jq-compatible expression language:

- **Navigation**: `.`, `.foo`, `.[]`, `.[n]`, `.[n:m]`, `..`
- **Operators**: `|`, `,`, `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `and`, `or`, `not`
- **Conditionals**: `if-then-else`, `//`, `try-catch`, `?`
- **Variables**: `.x as $v | ...`, destructuring `{x: $x, y: $y}`
- **Functions**: `select`, `map`, `sort`, `unique`, `group_by`, `keys`, `values`, `length`, `type`, `has`, `in`, `contains`, `split`, `join`, `test`, `match`, `sub`, `gsub`, and more
- **Construction**: `{...}`, `[...]`, string interpolation `"Hello \(.name)"`
- **Assignment**: `=`, `|=`, `+=`, `-=`, `*=`, `del()`
- **Advanced**: `reduce`, `path`, `getpath`, `setpath`, `to_entries`, `from_entries`, `with_entries`

See [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md) for comprehensive examples.

## Related Projects

- [HUML Specification](https://huml.io) - Official HUML language specification
- [go-huml](https://github.com/huml-lang/go-huml) - Go library for HUML parsing
- [jq](https://github.com/jqlang/jq) - JSON processor (syntax inspiration)

## License

MIT
