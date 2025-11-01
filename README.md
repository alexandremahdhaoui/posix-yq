# posix-yq

A POSIX-compliant shell script for parsing and querying YAML files, generated from Go. Works on any POSIX-compliant shell without external dependencies.

## Features

### Basic Operations
- **Read entire YAML files**: Output complete YAML content
- **Extract root-level keys**: Query top-level fields like `.name`
- **Navigate nested structures**: Access nested fields like `.person.address.city`
- **Array indexing**: Access array elements like `.items[0]`
- **Array iteration**: Iterate over all elements like `.items[]`
- **JSON output**: Convert YAML to JSON with `-o json` flag

### Advanced Operations
- **Pipe operator**: Chain operations like `.items | length`
- **Length operator**: Count array/object elements like `.items | length`
- **Keys operator**: List all object keys like `.person | keys`
- **Multiple selections**: Query multiple fields like `.name, .age`
- **Has operator**: Check key existence like `.person | has("name")`
- **Alternative operator**: Provide defaults like `.missing // "default"`

### Technical
- **POSIX compliant**: Runs on sh, dash, bash, and any POSIX shell
- **Zero dependencies**: Uses only standard POSIX commands (awk, sed, cat, grep)

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/alexandremahdhaoui/posix-yq.git
cd posix-yq

# Build the generator
make build

# Generate the posix-yq script
make generate

# The posix-yq script is now ready to use
./posix-yq --help
```

## Usage

### Basic Usage

```bash
# Read entire YAML file
./posix-yq file.yaml

# Extract root-level key
./posix-yq .name file.yaml

# Extract nested key
./posix-yq .person.address.city file.yaml

# Access array element by index
./posix-yq .items[0] file.yaml

# Iterate over array elements
./posix-yq '.items[]' file.yaml

# Convert to JSON
./posix-yq -o json file.yaml
```

### Advanced Usage

```bash
# Get length of array
./posix-yq '.items | length' file.yaml

# Get length of object (count keys)
./posix-yq '.person | length' file.yaml

# List all keys in an object
./posix-yq '.person | keys' file.yaml

# Select multiple fields
./posix-yq '.name, .age' file.yaml

# Check if key exists
./posix-yq '.person | has("email")' file.yaml

# Provide default value for missing keys
./posix-yq '.missing // "default value"' file.yaml

# Chain operations with pipe
./posix-yq '.items | length' file.yaml
```

### Examples

Given a YAML file `person.yaml`:
```yaml
name: Alice
age: 25
address:
  city: Paris
  country: France
hobbies:
  - reading
  - cycling
  - cooking
```

#### Extract specific values:

```bash
# Get name
$ ./posix-yq .name person.yaml
Alice

# Get nested city
$ ./posix-yq .address.city person.yaml
Paris

# Get first hobby
$ ./posix-yq .hobbies[0] person.yaml
reading

# List all hobbies
$ ./posix-yq '.hobbies[]' person.yaml
reading
cycling
cooking

# Count hobbies
$ ./posix-yq '.hobbies | length' person.yaml
3

# List all keys in address
$ ./posix-yq '.address | keys' person.yaml
city
country

# Get multiple values
$ ./posix-yq '.name, .age' person.yaml
Alice
25

# Check if key exists
$ ./posix-yq '.address | has("city")' person.yaml
true

# Use default value for missing key
$ ./posix-yq '.email // "no-email@example.com"' person.yaml
no-email@example.com
```

#### Convert to JSON:

```bash
$ ./posix-yq -o json person.yaml
{"name":"Alice","age":25,"address":{"city":"Paris","country":"France"},"hobbies":["reading","cycling","cooking"]}
```

## Development

### Requirements

- Go 1.24+ (for building the generator)
- POSIX-compliant shell (sh, dash, bash, etc.)
- Make (optional, for convenience)

### Project Structure

```
posix-yq/
├── cmd/generator/              # Go generator orchestrator
│   └── main.go                # Concatenates all modules
├── pkg/generator/              # Generator modules
│   ├── shell_header.go        # Shell initialization & debug utilities
│   ├── parser.go              # Main recursive parser (yq_parse)
│   ├── core_functions.go      # Key access, iteration, array operations
│   ├── advanced_functions.go  # Map, select, recursion, comparison
│   ├── operators.go           # Assignment, update, delete operators
│   ├── json.go                # JSON output conversion
│   ├── entrypoint.go          # Main entry point and flag parsing
│   ├── test_helper.go         # Test utilities
│   └── *_test.go              # Unit tests for each module
├── test/                       # Test infrastructure
│   ├── e2e/                   # End-to-end tests
│   │   └── run_tests.sh       # E2E test runner
│   ├── unit/                  # Unit test scenarios
│   │   ├── scenarios/         # 35+ test scenarios
│   │   └── run_tests.sh       # Unit test runner
│   ├── fixtures/              # Test YAML files
│   ├── yq-edge-cd-tests.sh    # Real-world usage tests
│   └── posix-compliance.sh    # POSIX compliance tests
├── posix-yq                    # Generated POSIX shell script
├── Makefile                    # Build automation
├── README.md                   # This file
├── CLAUDE.md                   # Instructions for Claude Code
├── GEMINI.md                   # Instructions for Gemini
└── posix-yq-prompt.md         # Comprehensive developer guide
```

### Building

```bash
# Build the Go generator binary
make build

# Generate the posix-yq shell script
make generate

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests (unit + E2E)
make test

# Run only Go unit tests
make test-unit-generator

# Run only E2E tests
make test-e2e

# Run POSIX compliance tests
./test/posix-compliance.sh
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the Go generator binary into `./build/` |
| `make generate` | Generate the `posix-yq` shell script |
| `make clean` | Remove build artifacts |
| `make test-unit-generator` | Run Go generator unit tests |
| `make test-unit-posix-yq` | Run posix-yq script unit tests against test scenarios |
| `make test-unit` | Run all unit tests (edge-cd + scenarios) |
| `make test-e2e` | Run E2E tests (depends on test-unit) |
| `make test` | Run all tests (unit + E2E) |
| `make help` | Show available targets |

## How It Works

posix-yq is a two-part system:

1. **Go Generator**: A Go program that generates POSIX-compliant shell code
2. **Generated Script**: The `posix-yq` shell script that performs YAML parsing

The generator creates shell code that uses standard POSIX utilities:
- `awk` for pattern matching and text processing
- `sed` for text transformation
- `cat` for file reading
- Standard shell constructs (`if`, `while`, etc.)

This approach ensures:
- **Portability**: Works on any POSIX system
- **No dependencies**: No external tools required
- **Auditability**: The generated script is human-readable
- **Performance**: Compiled Go generator produces optimized shell code

## Testing Strategy

The project follows strict Test-Driven Development (TDD):

1. **Unit Tests**: Go unit tests for each generator function
2. **E2E Tests**: Shell script tests for the generated posix-yq script
3. **POSIX Compliance Tests**: Verification on multiple shells (sh, dash)

All tests must pass before any feature is considered complete.

## Supported yq Features

✅ **Implemented**:
- Basic selection (`.key`)
- Nested selection (`.key.nested.deep`)
- Array indexing (`.items[0]`)
- Array iteration (`.items[]`)
- Pipe operator (`|`)
- Length operator (`.items | length`)
- Keys operator (`.person | keys`)
- Multiple selections (`.name, .age`)
- Has operator (`.person | has("key")`)
- Alternative operator (`.missing // "default"`)
- JSON output (`-o json`)

❌ **Not Yet Implemented** (may be added in future versions):
- Select/filter operators (`.items[] | select(. == "value")`)
- String operators (`upcase`, `downcase`, `split`, `join`)
- Math operators (`+`, `-`, `*`, `/`)
- Comparison operators (`==`, `!=`, `>`, `<`)
- Boolean operators (`and`, `or`, `not`)
- Recursive descent (`..key`)
- Map operator (`.items | map(expr)`)
- Sort operators (`sort`, `sort_by`)
- Group by, reduce, unique, flatten
- In-place editing (`-i` flag)
- Multiple document support
- YAML anchors and aliases
- Input formats other than YAML (XML, JSON, CSV, TOML)

## Contributing

Contributions are welcome! Please:

1. Follow the TDD approach (tests first)
2. Ensure all tests pass with `make test`
3. Verify POSIX compliance with `./test/posix-compliance.sh`
4. Update documentation as needed

## License

[Add your license here]

## Author

Alexandre Mahdhaoui

## Acknowledgments

Inspired by `yq` (https://github.com/mikefarah/yq) and the need for a portable YAML query tool.
