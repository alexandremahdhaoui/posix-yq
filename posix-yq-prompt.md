# POSIX-YQ: Comprehensive Developer Guide for AI Coding Agents

## Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Understanding the Generator](#understanding-the-generator)
4. [Generated Script Structure](#generated-script-structure)
5. [How POSIX-YQ Works Internally](#how-posix-yq-works-internally)
6. [Variable Scoping & Depth Tracking](#variable-scoping--depth-tracking)
7. [How to Extend Features](#how-to-extend-features)
8. [How to Fix Bugs](#how-to-fix-bugs)
9. [How to Test](#how-to-test)
10. [Critical Implementation Details](#critical-implementation-details)
11. [Common Pitfalls & Solutions](#common-pitfalls--solutions)

---

## Project Overview

### What is POSIX-YQ?

POSIX-YQ is a POSIX-compliant shell implementation of `yq`, the YAML/JSON query processor. It:
- Runs in pure POSIX shell (sh/dash) without external dependencies like jq or yq
- Supports complex nested queries with multiple pipes and operators
- Handles both YAML and JSON input formats
- Implements recursive descent and complex path traversal
- Provides array/object iteration, filtering, and transformation

### Why It Matters

Traditional `yq` may not be available in minimal containers or restricted environments. POSIX-YQ fills this gap by providing YAML query functionality in any POSIX-compliant shell.

### Directory Structure

```
posix-yq/
├── cmd/generator/
│   ├── main.go                   # Main orchestrator - concatenates all modules
│   ├── shell_header.go           # Shell header and debug utilities
│   ├── parser.go                 # yq_parse recursive parser function
│   ├── core_functions.go         # Key access, iteration, array operations
│   ├── advanced_functions.go     # Map, select, recursion, comparison
│   ├── operators.go              # Assignment, update, delete operators
│   ├── json.go                   # JSON output conversion
│   ├── entrypoint.go             # Main entry point and flag parsing
│   └── main_test.go              # Go unit tests for generator modules
├── posix-yq                       # Generated executable shell script (output)
├── Makefile                       # Build and test orchestration
├── CLAUDE.md                      # Project instructions (follow these!)
├── posix-yq-prompt.md            # This documentation file
├── test/
│   ├── yq-edge-cd-tests.sh       # Main test suite (15 different test scenarios)
│   ├── fixtures/                  # Test data files
│   ├── unit/                      # Unit test scenarios
│   └── e2e/                       # End-to-end tests
└── build/                         # Build artifacts directory
```

### Modular Architecture (Refactored)

The generator has been refactored into modular, testable components:

- **main.go**: Orchestrates the module concatenation in correct order
- **shell_header.go**: POSIX shell initialization, depth tracking, debug utilities
- **parser.go**: The main recursive parser with pipe, alternative, and concatenation operators
- **core_functions.go**: String unquoting, key extraction, array iteration, length, keys operations
- **advanced_functions.go**: Map, select, comparison, and recursive descent functionality
- **operators.go**: Assignment (=), update (|=), and delete (del) operators
- **json.go**: YAML-to-JSON conversion for `-o=j` output format
- **entrypoint.go**: Flag parsing, stdin detection, main execution flow

Each module is independently testable via Go unit tests (`main_test.go`) that verify:
- Generated functions are present and complete
- Concatenation produces valid output
- No obvious shell syntax errors
- All required functionality is included

This modular approach allows for easier:
- **Unit testing**: Each function generator can be tested independently
- **Maintenance**: Changes to one function type don't affect others
- **Feature addition**: New functions can be added as new modules
- **Code review**: Smaller, focused files are easier to review

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────┐
│  cmd/generator/main.go (Go program)    │
│  - String builder with shell functions │
│  - Outputs raw shell script text       │
└──────────────┬──────────────────────────┘
               │ make generate
               ▼
        ┌──────────────┐
        │  posix-yq    │ (executable shell script)
        └──────┬───────┘
               │ stdin/args
               ▼
       ┌───────────────────┐
       │   yq_parse()      │ (recursive parser)
       ├───────────────────┤
       │ Helper Functions: │
       ├───────────────────┤
       │ • yq_key_access   │
       │ • yq_iterate      │
       │ • yq_unquote      │
       │ • yq_length       │
       │ • yq_keys         │
       │ • yq_to_entries   │
       │ • yq_recursive... │
       └────────┬──────────┘
                │
                ▼
         YAML/JSON Output
```

### Key Principle: Generator Pattern

The entire POSIX shell script is **generated from Go code**, not hand-written. This allows:
- Type-safe string building
- Complex nested string generation
- Consistent formatting
- Easy feature addition via Go templates

**Important**: Never edit `posix-yq` directly! Always edit `cmd/generator/main.go`.

---

## Understanding the Generator

### Modular Architecture

The generator consists of multiple Go files, each responsible for generating a specific part of the shell script:

1. **main.go**: Orchestrates module concatenation
   - Calls all generator functions in the correct order
   - Prints shell shebang
   - Concatenates all parts into final script

2. **shell_header.go**: Generates shell initialization
   - Shebang and header comments
   - Depth tracking variable `_yq_parse_depth`
   - Debug utilities (`_yq_debug_indent`)

3. **parser.go**: Generates the main parser
   - `yq_parse()` recursive function
   - Pipe operator handling
   - Alternative operator (`//`) handling
   - String concatenation (`+`) operator
   - Function calls and operator detection

4. **core_functions.go**: Generates core utilities
   - `yq_unquote()`: String unquoting
   - `yq_key_access()`: YAML key extraction
   - `yq_iterate()`: Array/object iteration
   - `yq_array_access()`: Array indexing and slicing
   - `yq_length()`, `yq_keys()`, `yq_to_entries()`, `yq_has()`

5. **advanced_functions.go**: Generates advanced functionality
   - `yq_map()`: Apply expression to array elements
   - `yq_select()`: Filter based on conditions
   - `yq_compare()`: Comparison operations
   - `yq_recursive_descent()`: Tree traversal
   - `yq_recursive_descent_pipe()`: Tree traversal with piping

6. **operators.go**: Generates mutation operators
   - `yq_assign()`: Assignment operator (`=`)
   - `yq_update()`: Update operator (`|=`)
   - `yq_del()`: Delete operator

7. **json.go**: Generates JSON conversion
   - `yq_yaml_to_json()`: YAML to JSON formatter for `-o=j`

8. **entrypoint.go**: Generates main entry point
   - Flag parsing (`-e`, `-r`, `-o`, `-I`, `-j`)
   - Stdin detection and reading
   - Query execution orchestration
   - Output formatting and cleanup

### File: `cmd/generator/main.go` (Refactored)

The main function now orchestrates all generator modules:

```go
package main

import "fmt"

func main() {
    fmt.Println("#!/bin/sh")
    fmt.Println()

    fmt.Print(GenerateShellHeader())
    fmt.Println()

    fmt.Print(GenerateParser())
    fmt.Println()

    fmt.Print(GenerateCoreFunctions())
    fmt.Println()

    // ... more modules
}
```

#### Key Characteristics

- **Modular functions**: Each module returns a string of shell script code
- **Backtick strings**: Used for shell code in Go (allows `$` without escaping)
- **String concatenation**: main.go concatenates modules with `fmt.Print()`
- **Order matters**: Modules must be concatenated in correct dependency order

### How to Modify the Generator

1. **Add a new shell function**:
   - Decide which logical module it belongs to
   - Add it to the appropriate module file in pkg/generator/
   - Regenerate with `make build generate`

2. **Modify existing function**:
   - Find the function in its corresponding module file in pkg/generator/
   - Edit the return string
   - Regenerate with `make build generate`

3. **Add new flags**:
   - Modify the flag parsing section in `entrypoint.go`
   - Add necessary variable handling
   - Test with `make build generate && make test-unit`

4. **Create new module**:
   - Create new file in pkg/generator/: `<feature>.go`
   - Implement `func Generate<Feature>() string`
   - Call the function from cmd/generator/main.go in appropriate order
   - Add unit tests in pkg/generator/*_test.go
   - Verify with `make build generate && go test ./pkg/generator`

### Unit Tests

The generator includes Go unit tests (`cmd/generator/main_test.go`) that verify:

```bash
go test ./cmd/generator -v

# Tests verify:
# - Each module returns non-empty strings
# - All required functions are generated
# - Concatenated output is complete
# - No obvious shell syntax errors
# - All dependencies are present
```

Run tests with:
```bash
make test-unit-generator
# or
cd cmd/generator && go test -v
```

#### Build Process

```bash
make build generate
# 1. Compiles cmd/generator/main.go into ./build/generator
# 2. Runs ./build/generator and pipes output to ./posix-yq
# 3. Makes ./posix-yq executable
```

---

## Generated Script Structure

### Script Sections (in order)

#### 1. Shebang & Initialization
```sh
#!/bin/sh
_yq_parse_depth=0              # Global depth counter for recursion tracking
```

#### 2. Helper Functions
- `_yq_debug_indent()` - Indented debug output with depth markers
- `yq_unquote()` - Remove YAML string quotes
- `yq_parse()` - Main recursive parser (SEE NEXT SECTION)
- `yq_key_access()` - Extract value for a YAML key
- `yq_iterate()` - Iterate over array/object elements
- `yq_length()`, `yq_keys()`, `yq_to_entries()` - YAML utilities
- `yq_recursive_descent_pipe()` - Handle `..` operator
- Various other helper functions (assign, update, compare, delete)

#### 3. Main Entry Point
```sh
# Parse command-line flags (-e, -r, -o, etc.)
# Detect stdin vs file input
# Execute query
# Clean up and format output
```

---

## How POSIX-YQ Works Internally

### The Recursive Parser: `yq_parse()`

This is the **heart of the system**. It's called recursively to break down queries into smaller pieces.

#### Function Signature
```sh
yq_parse() {
    _query="$1"      # The YQ query (e.g., ".foo.bar")
    _file="$2"       # Path to YAML/JSON file containing data

    _yq_parse_depth=$((_yq_parse_depth + 1))  # Track nesting level
    # ... process query ...
    _yq_parse_depth=$((_yq_parse_depth - 1))  # Decrement before return
}
```

#### Processing Order (Priority)

The parser checks these patterns **in order** and returns after matching:

1. **Recursive descent (..)**: `if [ "$_query" = ".." ]`
2. **Parentheses with/without pipe**: `if echo "$_query" | grep -q '^([^)]*)'`
3. **Alternative operator (//)**: `if echo "$_query" | grep -q ' // '`
4. **Pipe operator (|)**: `if echo "$_query" | grep -q ' | '`
   - **Special case: Left side ends with .[]**: Iteration handler
5. **String concatenation (+)**: `if echo "$_query" | grep -q ' + '`
6. **String literal**: `if [ "${_query#\"}" != "$_query" ]`
7. **Function calls**: `if echo "$_query" | grep -q '^[a-zA-Z_][a-zA-Z0-9_]*(';`
8. **Assignment/Update operators**: `=`, `|=`
9. **Comparison operators**: `==`, `!=`, etc.
10. **Key access**: `.foo`, `.foo.bar`, etc.

**Critical**: Once a pattern matches, the function returns. The **order matters**.

### Example: Query Execution Flow

Query: `(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value`

```
1. yq_parse(full_query, file)
   └─ Matches PARENTHESES pattern
      └─ yq_parse("(.extraEnvs // [])", file)
         └─ Matches ALTERNATIVE OPERATOR
            └─ yq_parse(".extraEnvs", file)
            └─ yq_parse("[]", file)  [only if extraEnvs is null]
   └─ Continues with PIPE OPERATOR
      └─ Left side: "."
      └─ Right side: ".[] | to_entries | ..."
         └─ Matches ITERATION HANDLER (because left side is ".")
            └─ yq_parse(".[]", file)  [extract all items]
            └─ For each item: yq_parse("to_entries | ...", item_file)
               └─ Matches PIPE again
                  └─ yq_parse("to_entries", item)
                  └─ yq_parse(".[] | .key + ...", transformed_item)
                     └─ Matches ITERATION HANDLER again
```

### Depth Tracking System

#### What is Depth?

Depth is a global counter that tracks how deeply nested we are in recursive function calls. It serves two purposes:

1. **Debug output formatting**: Indent debug messages by depth level
2. **Variable scoping workaround**: Create unique variable names per depth level

#### How It Works

```sh
# At function entry
_yq_parse_depth=$((_yq_parse_depth + 1))

# Use depth for unique variable names (CRITICAL!)
_saved_iter_base_${_yq_parse_depth}="$_iter_tmp_items"

# At function exit (MUST HAPPEN ON ALL RETURN PATHS!)
_yq_parse_depth=$((_yq_parse_depth - 1))
return
```

#### Why This Matters

In POSIX shell, **all variables are global**. When a nested iteration creates `_saved_iter_state`, it overwrites the outer iteration's variable:

```sh
# Depth 3 iteration: _saved_iter_state="/tmp/aaa"
# Depth 5 iteration: _iter_state=$(mktemp)  # Creates "/tmp/bbb"
#                    _saved_iter_state="$_iter_state"  # OVERWRITES depth 3's value!
# Depth 3 loop iteration 2: tries to read from "/tmp/bbb" instead of "/tmp/aaa"
# Result: FILE NOT FOUND error
```

**Solution**: Use depth-specific variable names:
```sh
eval "_saved_iter_base_${_yq_parse_depth}='$_iter_tmp_items'"
eval "_current_iter_base=\$_saved_iter_base_${_yq_parse_depth}"
```

#### Critical Rule: Every Return Path Must Decrement Depth

**MUST** have `_yq_parse_depth=$((_yq_parse_depth - 1))` before:
- Every `return` statement
- End of function before closing brace `}`

Forgetting this causes depth to become misaligned, leading to:
- Iteration handlers executing at wrong depths
- Wrong state being read from depth-specific variables
- "FILE NOT FOUND" errors in subsequent iterations

### State File System

#### Problem
YAML processing often needs to store intermediate results (parsed values, iteration states, etc.). Using shell variables for complex data is error-prone.

#### Solution: Temporary Files as State

```sh
# Create unique temp files for this depth level
_iter_state=$(mktemp)                 # Base state file path
_iter_tmp_items=$(mktemp)             # File containing item paths

# Store state in separate files to avoid sed delimiter issues
printf "%s" "$_iter_tmp_items" > "$_iter_state.base"
printf "%s" "$_num_items" > "$_iter_state.num"
printf "%s" "$_after_pipe" > "$_iter_state.query"

# Save to depth-specific variable (so nested calls don't overwrite)
eval "_saved_iter_base_${_yq_parse_depth}='$_iter_state'"

# Later, restore state in loop
eval "_current_iter_base=\$_saved_iter_base_${_yq_parse_depth}"
_state_iter_tmp_items=$(cat "$_current_iter_base.base")
```

#### AWK Item Splitting

When iterating over arrays/objects, items are stored in separate files:

```
/tmp/tmp.xxx       (directory-like prefix, actually just a string)
/tmp/tmp.xxx.1     (first item)
/tmp/tmp.xxx.2     (second item)
/tmp/tmp.xxx.3     (third item)
```

Items are separated by **blank lines** in the input, and AWK splits them:

```awk
if ($0 ~ /^[[:space:]]*$/) {
    # Blank line found - save current item
    print item > (tmpbase "." item_count)
    item_count++
}
```

### Key Access: `yq_key_access()`

Extracts a value for a specific key from YAML using AWK.

**Logic:**
1. Search for line matching `key:` pattern
2. If value is inline (`key: value`), extract and return
3. If value is multi-line, collect all indented lines below the key
4. If key not found, return `null` (in END block)

**Important**: Handles both inline and block values:
```yaml
# Inline
foo: "bar"        # Returns: "bar"

# Block
foo:
  - item1
  - item2         # Returns entire block
```

### Iteration Handler: Array/Object Loop

When a pipe has `.[]` on the left side, special handling applies.

**Steps:**
1. Execute left side (`.[]`) to get all items
2. Split output by blank lines into separate files (using AWK)
3. **Store iteration state in depth-specific variables** (critical!)
4. Loop through each item file:
   - Execute right side of pipe with that item as input
5. Clean up temporary files

**Blank Line Cleanup:**
The iteration uses blank lines as separators internally, but final output shouldn't have them:
```sh
# Remove consecutive blank lines from final result
_result=$(printf '%s\n' "$_result" | grep -v '^[[:space:]]*$')
```

---

## Variable Scoping & Depth Tracking

### The POSIX Shell Limitation

POSIX shell has **no concept of local variables**. All variables are global:

```sh
outer_func() {
    _var="outer value"
    inner_func
    echo "$_var"  # Prints "inner value" if inner_func changed it!
}

inner_func() {
    _var="inner value"  # This modifies the GLOBAL _var
}
```

### Our Solution: Depth-Specific Variable Names

Instead of trying to make variables local, we create unique variable names using the depth counter:

```sh
# At depth 3
_saved_iter_base_3="/tmp/aaa"

# At depth 5
_saved_iter_base_5="/tmp/bbb"

# Both can coexist without conflict!
```

### Implementation Pattern

**When saving state:**
```sh
eval "_saved_iter_base_${_yq_parse_depth}='$_iter_tmp_items'"
```

**When retrieving state:**
```sh
eval "_current_iter_base=\$_saved_iter_base_${_yq_parse_depth}"
_value=$(cat "$_current_iter_base.1")
```

### Why `eval` is Necessary

We need to create variable names dynamically:
```sh
_saved_iter_base_3="value"     # Literal syntax - won't work
eval "_saved_iter_base_3='value'"  # eval - WILL work

# Later, when we don't know which depth we are:
depth=3
eval "_var=\$_saved_iter_base_${depth}"  # Retrieves _saved_iter_base_3
```

---

## How to Extend Features

### Adding a New Operator/Query Type

#### Example: Adding a New Operator `%%` (hypothetical)

**Step 1: Understand the Query Pattern**
```
Query: ".foo %% .bar"
Meaning: Apply some operation between foo and bar
```

**Step 2: Add Pattern Detection to yq_parse()**

In `cmd/generator/main.go`, find the `yq_parse()` function around line 30, and add your check (in order priority):

```go
    # Check for new %% operator (insert in correct priority order)
    if echo "$_query" | grep -q ' %% '; then
        [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "New %% operator detected"

        _left=$(echo "$_query" | sed 's/ %%.*//')
        _right=$(echo "$_query" | sed 's/^[^%]* %% //')

        # Process both sides
        yq_parse "$_left" "$_file" > /tmp/left_result
        yq_parse "$_right" "$_file" > /tmp/right_result

        # Combine results
        # ... your logic ...

        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi
```

**Step 3: Add Unit Tests**

In `test/yq-edge-cd-tests.sh`, add a test case:

```bash
echo "=== Test: New %% Operator ==="
expected="expected_value"
actual=$("${YQ_BIN}" '.foo %% .bar' "${TEST_DIR}/test.yaml")
assert_equals "new operator test" "${expected}" "${actual}"
```

**Step 4: Rebuild and Test**

```bash
make build generate
./test/yq-edge-cd-tests.sh ./posix-yq
```

### Adding a New Command-Line Flag

#### Example: Adding `-s` (slurp) Flag

**Step 1: Add Flag Parsing**

Find the main entry point (around line 1307 in generator):

```go
    -s|--slurp)
        _slurp=1
        shift
        ;;
```

**Step 2: Implement Flag Logic**

```go
    if [ $_slurp -eq 1 ]; then
        # Read entire input as array
        # Process as array instead of line-by-line
    fi
```

**Step 3: Add Tests**

```bash
echo "Test -s flag"
actual=$("${YQ_BIN}" -s '.[]' "${TEST_DIR}/multi.yaml")
```

### Adding a New Helper Function

#### Example: Adding `yq_split()` Function

**Step 1: Implement Function**

Add to generator before main entry point:

```go
# Split string by delimiter
yq_split() {
    _string="$1"
    _delimiter="$2"

    printf '%s\n' "$_string" | sed "s/${_delimiter}/\n/g"
}
```

**Step 2: Use in yq_parse()**

```go
    if echo "$_query" | grep -q 'split('; then
        _arg=$(echo "$_query" | sed "s/split(\(.*\))/\1/")
        yq_split "$input" "$_arg"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi
```

---

## How to Fix Bugs

### Bug Diagnosis Process

#### Step 1: Reproduce with Minimal Example

```bash
# Create minimal test case
cat > /tmp/test_bug.yaml << 'EOF'
key: value
EOF

# Run with debug output
POSIX_YQ_DEBUG=1 ./posix-yq '.key' /tmp/test_bug.yaml
```

#### Step 2: Enable Debug Output

The generated script supports `POSIX_YQ_DEBUG=1`:

```bash
POSIX_YQ_DEBUG=1 ./posix-yq '.foo.bar' file.yaml 2>&1 | head -50
```

Output format:
```
DEBUG[1]  yq_parse called with query='.foo.bar'
DEBUG[2]    yq_parse called with query='bar'
DEBUG[3]      Found key 'bar' in file
```

The number in brackets is the **depth level**. Use this to trace the call stack.

#### Step 3: Check Depth Tracking

If you see "depth jumping" or out-of-sequence numbers, there's a **missing depth decrement**.

**Symptoms:**
```
DEBUG[1]
DEBUG[2]
DEBUG[5]  <-- jumped from 2 to 5, missed 3 and 4
DEBUG[3]  <-- depth went backwards!
```

**Cure**: Find the functions that returned without decrementing and add `_yq_parse_depth=$((_yq_parse_depth - 1))`.

### Common Bug Patterns

#### Pattern 1: FILE NOT FOUND in Iteration

**Symptom**:
```
Iteration: item #2 - FILE NOT FOUND: /tmp/tmp.xyz.2
```

**Cause**: Variable overwrite in nested iterations

**Solution**: Check that iteration state uses depth-specific variables:
```sh
# WRONG:
_saved_iter_state="$_iter_state"

# RIGHT:
eval "_saved_iter_base_${_yq_parse_depth}='$_iter_tmp_items'"
```

#### Pattern 2: Extra Blank Lines in Output

**Symptom**:
```
apt-get

install

-y
```

Expected:
```
apt-get
install
-y
```

**Cause**: Array iteration uses blank lines as separators

**Solution**: Add blank line cleanup before output:
```sh
_result=$(printf '%s\n' "$_result" | grep -v '^[[:space:]]*$')
```

#### Pattern 3: Wrong Values Returned

**Symptom**: Query returns `null` instead of actual value

**Cause**:
1. Missing depth increment/decrement
2. Query pattern matching in wrong order
3. State file not being saved/restored properly

**Solution**:
1. Add `make build generate` to get fresh copy
2. Check `yq_key_access()` END block has `if (!found) print "null"`
3. Verify pattern check order in `yq_parse()`

#### Pattern 4: Syntax Errors in Generated Script

**Symptom**:
```
./posix-yq: line 150: unexpected operator
```

**Cause**: Problem in generator output

**Solution**:
1. Check the actual generated script: `sed -n '148,152p' posix-yq`
2. Look for unclosed quotes or parentheses in generator
3. Check if string escaping is correct (backtick strings vs. quoted strings)

### Debugging Workflow

```bash
# 1. Create minimal test case
cat > /tmp/test.yaml << 'EOF'
foo: bar
EOF

# 2. Run with debug
POSIX_YQ_DEBUG=1 ./posix-yq '.foo' /tmp/test.yaml 2>&1 > /tmp/debug.log

# 3. Analyze output
cat /tmp/debug.log

# 4. Check depth progression
grep "DEBUG\[" /tmp/debug.log

# 5. If depth is wrong, find missing decrements
grep -n "_yq_parse_depth\|return" posix-yq | head -30

# 6. Modify generator
vim cmd/generator/main.go

# 7. Rebuild and test
make build generate
POSIX_YQ_DEBUG=1 ./posix-yq '.foo' /tmp/test.yaml 2>&1
```

---

## How to Test

### Test Structure

```
test/
├── yq-edge-cd-tests.sh      # Main test suite (15 tests, ~300+ assertions)
├── unit/
│   └── run_tests.sh         # Unit test runner
├── e2e/
│   └── run_tests.sh         # End-to-end tests
└── fixtures/
    ├── 01-simple.yaml
    ├── 01-complex.yaml
    └── ...
```

### Running Tests

```bash
# All tests
make test

# Just unit tests
make test-unit

# Just edge-cd tests (most important)
cd test && bash yq-edge-cd-tests.sh ../posix-yq

# Single test with output
./posix-yq '.key' file.yaml
```

### Understanding Test Output

```
=== Test 1: Extract key=value from extraEnvs ===
✓ extraEnvs key=value extraction
✓ extraEnvs with missing field (default to empty)

=== Test 2: Read YAML path from stdin with -e ===
✗ stdin read with -e flag
  Expected: value123
  Got:
```

Green checkmark (✓) = PASS
Red X (✗) = FAIL

### Writing New Tests

Add to `test/yq-edge-cd-tests.sh`:

```bash
echo "=== Test: Your Test Name ==="

# Create test data
cat >"${TEST_DIR}/your-test.yaml" <<'EOF'
data:
  - item1
  - item2
EOF

# Define expected output
expected="item1
item2"

# Run query
actual=$("${YQ_BIN}" '.data[]' "${TEST_DIR}/your-test.yaml")

# Assert
assert_equals "your test description" "${expected}" "${actual}"
```

### Test Coverage

Current test areas:
1. **Complex nested queries** - Multiple pipes and operators
2. **Stdin handling** - `-e` and `-r` flags with pipes
3. **Array/object iteration** - `.[]` operator on nested structures
4. **Missing paths** - Return `null` for nonexistent keys
5. **Type handling** - Strings, numbers, booleans
6. **Operators** - `//` (alternative), `+` (concatenation)

### Key Test Commands

```bash
# Test complex nested iteration (was failing, now passing)
./posix-yq '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value' file.yaml

# Test stdin
echo "foo: bar" | ./posix-yq '.foo'

# Test with flags
./posix-yq -e -r '.path' file.yaml

# Test missing keys
./posix-yq '.missing.path' file.yaml   # Should output: null
```

---

## Critical Implementation Details

### String Handling in Shell

#### Quoted vs. Unquoted Strings

YAML quotes strings:
```yaml
key: "value"          # Quoted string
key: value            # Unquoted string (both are values)
```

Our `yq_unquote()` removes the YAML quotes:
```sh
yq_unquote '"value"'  # Returns: value
yq_unquote 'null'     # Returns: null (unchanged)
yq_unquote 'true'     # Returns: true (unchanged)
```

**Critical**: Don't unquote special values like `null`, `true`, `false`.

#### Escaping in sed

SED uses delimiters. We use spaces as delimiter to avoid issues:
```sh
# WRONG - fails if string contains "/"
sed 's/old/new/g'

# RIGHT - works with any character
sed 's old new g'
```

### Array vs. Object Iteration

Both use `yq_iterate()` which checks:
1. If first line starts with `-` → Array (YAML format: `- item`)
2. Otherwise → Object (YAML format: `key: value`)

```yaml
# Array
- item1
- item2

# Object
key1: value1
key2: value2
```

Both are iterated the same way for `.[]`.

### Temp File Management

All temp files are created with `mktemp`:
```sh
_tmp=$(mktemp)
# ... use _tmp ...
rm -f "$_tmp"
```

**Important**: Always clean up with `rm -f` at function end, even on error paths.

#### State File Cleanup

For iterations, we create numbered files:
```sh
/tmp/tmp.xyz.1
/tmp/tmp.xyz.2
/tmp/tmp.xyz.3
```

These must be removed after processing:
```sh
rm -f "$_saved_iter_base.$_item_idx"
```

### AWK Multiline Processing

AWK processes files line-by-line, but we need to handle multiline YAML blocks.

**Solution**: Use blank lines as item separators:
```yaml
item1_line1
item1_line2

item2_line1
item2_line2
```

Then AWK splits on blank lines and re-joins:
```awk
if ($0 ~ /^[[:space:]]*$/) {
    # Save previous item and start new one
    print previous_item > (base "." count)
    count++
    previous_item = ""
}
```

---

## Common Pitfalls & Solutions

### Pitfall 1: Editing posix-yq Directly

**Symptom**: Changes work once, then disappear after `make generate`

**Problem**: `posix-yq` is generated, not source-controlled

**Solution**: Always edit `cmd/generator/main.go`, never edit `posix-yq`

```bash
# WRONG
vim posix-yq

# RIGHT
vim cmd/generator/main.go
make build generate
```

### Pitfall 2: Missing Depth Decrements

**Symptom**: Iteration fails with "FILE NOT FOUND", depth numbers don't match

**Problem**: Forgot to decrement depth before returning

**Solution**: Add `_yq_parse_depth=$((_yq_parse_depth - 1))` before every `return` and at end of function

```sh
# WRONG
yq_parse() {
    _yq_parse_depth=$((_yq_parse_depth + 1))
    # ... code ...
    if some_condition; then
        return  # Missing depth decrement!
    fi
    _yq_parse_depth=$((_yq_parse_depth - 1))
    return
}

# RIGHT
yq_parse() {
    _yq_parse_depth=$((_yq_parse_depth + 1))
    # ... code ...
    if some_condition; then
        _yq_parse_depth=$((_yq_parse_depth - 1))  # Added!
        return
    fi
    _yq_parse_depth=$((_yq_parse_depth - 1))
    return
}
```

### Pitfall 3: Variable Overwriting in Nested Calls

**Symptom**: Same code works at depth 1 but fails at depth 2+

**Problem**: Nested calls are overwriting parent's variables

**Solution**: Use depth-specific variable names
```sh
# WRONG
_saved_state="$value"

# RIGHT
eval "_saved_state_${_yq_parse_depth}='$value'"
```

### Pitfall 4: AWK Syntax Errors

**Symptom**: `awk: syntax error` at runtime

**Problem**: Complex AWK code with syntax errors

**Solution**:
1. Test AWK script separately
2. Use single quotes to avoid shell expansion
3. Check for unmatched brackets/braces in AWK code

```sh
# WRONG - shell expands variables inside backticks
awk "BEGIN { print $HOME }"

# RIGHT - AWK gets the variable name
awk 'BEGIN { print "hello" }'
```

### Pitfall 5: Blank Line Issues

**Symptom**: Extra newlines in output, test assertions fail

**Problem**: Array iteration uses blank lines as separators

**Solution**: Clean up blank lines before output
```sh
_result=$(printf '%s\n' "$_result" | grep -v '^[[:space:]]*$')
```

### Pitfall 6: Infinite Loops in sed/awk

**Symptom**: Script hangs, no output, high CPU

**Problem**: Regex or logic creates infinite loop

**Solution**:
1. Test regex separately: `echo "test" | sed 's/pattern/replace/'`
2. Add safeguards: `while [ $count -lt 100 ]; do ... done`
3. Debug with simple cases first

### Pitfall 7: Quote/Escape Hell

**Symptom**: Query with special characters fails or produces wrong output

**Problem**: Insufficient escaping of quotes and special chars

**Solution**:
- Use double quotes for shell variables: `"$_var"`
- Use single quotes for literals: `'string'`
- Escape quotes in strings: `\"` or `\'`
- Test incrementally with `POSIX_YQ_DEBUG=1`

---

## Typical AI Agent Workflow for This Repo

### When Adding a Feature

```
1. Read this document (you're here!)
2. Identify which code section needs modification
   - New operator? → Add to yq_parse() pattern checks
   - New flag? → Add to flag parsing in main entry point
   - New utility? → Create helper function
3. Create minimal test case
4. Modify cmd/generator/main.go
5. Run: make build generate
6. Test: POSIX_YQ_DEBUG=1 ./posix-yq 'query' file.yaml
7. Run: cd test && bash yq-edge-cd-tests.sh ../posix-yq
8. If depth issues appear, verify depth increments/decrements
9. Commit with clear message
```

### When Debugging a Failing Test

```
1. Isolate the failing test case
2. Create /tmp/test.yaml with minimal data
3. Run with debug: POSIX_YQ_DEBUG=1 ./posix-yq 'query' /tmp/test.yaml 2>&1
4. Check depth progression: look for jumps or backwards numbers
5. If depth issue: add missing _yq_parse_depth decrements
6. If logic issue: trace through yq_parse() function matching steps
7. If blank line issue: add cleanup grep before output
8. If state issue: verify depth-specific variables are used
9. Rebuild and test again
10. Once working, commit fix
```

### When Tests Keep Failing

```
1. Don't edit posix-yq directly - it will be overwritten!
2. Check cmd/generator/main.go has your changes
3. Run make build generate to refresh posix-yq
4. Check syntax: try running the shell script with set -x
5. Use POSIX_YQ_DEBUG=1 to see execution flow
6. Verify each return path has depth decrement
7. Verify state-saving uses depth-specific variables
8. Check that new patterns are added in correct order in yq_parse()
```

---

## Quick Reference: Critical Code Locations

### Main Parser
- **File**: `cmd/generator/main.go`
- **Function**: `yq_parse()` (starts around line 30)
- **Pattern checks**: Lines 38-305+ (order matters!)

### Helper Functions (in generator)
- **Key access**: `yq_key_access()` around line 495
- **Iteration**: `yq_iterate()` around line 570
- **Unquoting**: `yq_unquote()` around line 465
- **Length/Keys**: `yq_length()`, `yq_keys()` around lines 690-750

### Main Entry Point
- **Flag parsing**: Lines 1307-1337
- **Stdin handling**: Lines 1344-1350
- **Output formatting**: Lines 1366-1400

### Tests
- **Main test file**: `test/yq-edge-cd-tests.sh`
- **Test 1**: Line 83 (complex nested queries)
- **Test 2**: Line 107 (stdin with -e)
- **Test 3**: Line 121 (null handling)
- **Test 8+**: Array iteration tests

---

## Absolute Rules

These are non-negotiable:

1. **NEVER edit posix-yq directly**. It's generated. Edit `cmd/generator/main.go`
2. **EVERY return path MUST decrement depth**. Use grep to find all `return` statements
3. **Use depth-specific variables for state**. Pattern: `eval "_var_${_yq_parse_depth}='value'"`
4. **Clean up temp files**. Pattern: `rm -f "$_tmpfile"`
5. **Test after every change**. Run `make build generate` then test
6. **Match pattern order**. Higher priority patterns must come first in yq_parse()
7. **Use POSIX-compliant shell**. No bash-isms like `[[`, `(( ))`, etc.
8. **Debug with POSIX_YQ_DEBUG=1**. Provides execution trace
9. **Verify depth numbers**. They should increment/decrement sequentially
10. **Keep it simple**. Complex solutions are harder to maintain

---

## Summary

POSIX-YQ is a recursive YAML parser written in POSIX shell, generated from Go code. The key insights are:

1. **Generator Pattern**: Go builds the shell script, not hand-written
2. **Recursive Parser**: `yq_parse()` breaks queries into patterns, processes each
3. **Depth Tracking**: Workaround for lack of local variables in shell
4. **State Files**: Store complex data in temp files instead of variables
5. **Pattern Matching**: Query patterns are checked in priority order
6. **Blank Lines**: Used internally as separators, must be cleaned before output

With this understanding, you can extend, debug, and test this implementation effectively.

Good luck, AI agent! The codebase is ready for your improvements.
