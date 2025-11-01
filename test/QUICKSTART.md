# Edge-CD Test Suite Quickstart

## What Was Added

A comprehensive test suite (`test/yq-edge-cd-tests.sh`) that validates posix-yq against real-world yq usage patterns found in production edge-cd shell scripts.

## Quick Test

```bash
# Test with posix-yq
./test/yq-edge-cd-tests.sh ./posix-yq

# Compare with real yq
./test/yq-edge-cd-tests.sh $(which yq)
```

## Current Results

**~2-3 out of 15 test categories passing**

### What Works ✅
- Basic file reading
- Nested path access
- Array iteration
- Null handling

### What's Missing ❌
- Stdin input (40% of tests need this)
- `-r` flag (raw output)
- `-e` flag (error on null)
- `//` operator (defaults)
- `to_entries` function
- String concatenation with `+`

## Implementation Roadmap

### Phase 1: Critical (5 hours) → 50% tests passing
1. **Stdin input**: Read YAML from pipe
2. **-r flag**: Remove quotes from string output
3. **-e flag**: Exit with error on null

### Phase 2: High Priority (10-15 hours) → 70% tests passing
4. **// operator**: Default values (`.field // "default"`)
5. **+ operator**: String concatenation
6. **to_entries**: Convert objects to key-value pairs

### Phase 3: Medium Priority (10-15 hours) → 85% tests passing
7. Array indexing verification
8. More advanced functions

### Phase 4: Low Priority (10-12 hours) → 95% tests passing
9. JSON output (`-o=j`)
10. Indentation control (`-I`)

## Documentation

- `test/README.md` - Complete test suite overview
- `test/EDGE_CD_TESTS_STATUS.md` - Detailed gap analysis
- `test/yq-edge-cd-tests.sh` - The test script itself

## Example Output

When you run the tests, you'll see:

```
=== Test 1: Extract key=value from extraEnvs ===
✗ extraEnvs key=value extraction
  Expected: FOO=bar
            BAZ=qux
  Got:      

✓ extraEnvs with missing field (default to empty)

=== Test 2: Read YAML path from stdin with -e ===
... (test hangs waiting for stdin support)
```

## Next Steps

1. Review `test/EDGE_CD_TESTS_STATUS.md` for detailed analysis
2. Prioritize features based on test coverage impact
3. Implement high-impact features first (stdin, -r, -e flags)
4. Re-run tests to track progress

## Contributing

When implementing features:
1. Reference the edge-cd test that needs it
2. Check if the feature also helps unit tests
3. Update both documentation files
4. Run all three test suites to verify

---

**Created**: 2025-11-01  
**Status**: Edge-CD test suite fully integrated  
**Next**: Implement Phase 1 features for 50% test coverage
