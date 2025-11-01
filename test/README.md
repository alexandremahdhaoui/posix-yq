# posix-yq Test Suite

## Test Suites

### 1. Unit Tests (`test/unit/`)
**Purpose**: Test individual yq features and operations

**Structure**:
```
test/unit/scenarios/
  01-simple-key/
    command.txt    # The yq query to run
    input.yaml     # Input YAML file
    output.yaml    # Expected output
```

**Running**: `make test-unit-posix-yq`

**Current Status**: **18/33 tests passing (54.5%)**

**Test Runner**: `test/unit/run_tests.sh`

### 2. E2E Tests (`test/e2e/`)
**Purpose**: End-to-end integration testing

**Running**: `make test-e2e`

**Test Runner**: `test/e2e/run_tests.sh`

### 3. Edge-CD Tests (`test/yq-edge-cd-tests.sh`) ⭐ NEW
**Purpose**: Real-world yq usage patterns from production shell scripts

**Running**: `./test/yq-edge-cd-tests.sh ./posix-yq`

**Current Status**: **~2-3/15 categories passing**

**Documentation**: See `test/EDGE_CD_TESTS_STATUS.md` for detailed gap analysis

**Features Tested**:
- ✅ Basic file reading
- ✅ Array iteration
- ✅ Nested path access
- ❌ Stdin input
- ❌ Command-line flags (`-e`, `-r`, `-o`, `-I`)
- ❌ Advanced functions (`to_entries`, `+`, `//`)
- ❌ JSON output
- ❌ Raw output mode

## Running All Tests

```bash
# Run all tests
make test

# Run only unit tests
make test-unit-posix-yq

# Run only E2E tests
make test-e2e

# Run edge-cd real-world tests
./test/yq-edge-cd-tests.sh ./posix-yq

# Compare with real yq
./test/yq-edge-cd-tests.sh $(which yq)
```

## Test Development

### Adding a New Unit Test

1. Create a new directory:
   ```bash
   mkdir test/unit/scenarios/NN-test-name
   ```

2. Create test files:
   ```bash
   # The query to test
   echo '.some.path' > test/unit/scenarios/NN-test-name/command.txt

   # Input YAML
   cat > test/unit/scenarios/NN-test-name/input.yaml <<EOF
   some:
     path: "expected value"
   EOF

   # Expected output
   echo 'expected value' > test/unit/scenarios/NN-test-name/output.yaml
   ```

3. Run tests:
   ```bash
   make test-unit-posix-yq
   ```

### Test Naming Convention

- `01-` through `09-`: Basic operations
- `10-` through `19-`: Advanced operations
- `20-` through `29-`: Mutation operations
- `30-` through `39-`: Error handling and edge cases

## Test Results Overview

### Unit Tests Status

**Passing (18)**:
- Basic key access
- Nested keys
- Array operations (index, slice, iteration)
- Object iteration
- Map function
- Length and keys
- Has function
- Delete operations (including nested)
- Empty input handling

**Failing (15)**:
- Recursive descent (`..`)
- Select with conditions
- Assignment operators
- Multi-document YAML
- Format conversion (YAML ↔ JSON)
- Error handling

### Edge-CD Tests Status

**Critical Missing Features** (would enable 60-70% of tests):
1. Stdin input support
2. `-r` flag (raw output)
3. `-e` flag (error on null)
4. `//` operator (alternative/default values)

**Medium Priority** (would enable 80-85% of tests):
5. String concatenation (`+`)
6. `to_entries` function
7. Array indexing verification

**See**: `test/EDGE_CD_TESTS_STATUS.md` for complete analysis

## Continuous Integration

The test suite is designed to be CI-friendly:

```bash
# Exit code 0 if all tests pass
make test

# Individual test suites also return proper exit codes
make test-unit-posix-yq || echo "Unit tests failed"
make test-e2e || echo "E2E tests failed"
./test/yq-edge-cd-tests.sh ./posix-yq || echo "Edge-CD tests failed"
```

## Test Data

Test fixtures are stored in:
- `test/unit/scenarios/*/` - Unit test data
- `test/fixtures/` - E2E test data (if exists)
- `test/yq-edge-cd-tests.sh` - Generates test data in `.tmp/`

**Note**: The edge-cd test script creates temporary test data in `.tmp/yq-test-data-$$` and cleans up automatically on exit.

## Documentation

- `IMPLEMENTATION_STATUS.md` - Overall implementation progress
- `test/EDGE_CD_TESTS_STATUS.md` - Edge-CD test compatibility analysis
- `test/README.md` - This file

## Contributing

When adding new features:
1. Add unit tests for the feature
2. Verify edge-cd tests if applicable
3. Update `IMPLEMENTATION_STATUS.md`
4. Update `EDGE_CD_TESTS_STATUS.md` if adding edge-cd-relevant features
