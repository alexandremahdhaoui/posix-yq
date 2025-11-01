# Edge-CD Test Suite Status

## Overview

This document tracks compatibility with the edge-cd yq test suite (`test/yq-edge-cd-tests.sh`).

The edge-cd tests verify real-world yq usage patterns found in production shell scripts.

## Test Suite: 15 test categories

### Current Status: **Estimated 2-3/15 categories passing**

Most tests require features not yet implemented in posix-yq.

## Feature Gap Analysis

### Missing Critical Features

#### 1. Command-Line Flags ❌
- `-e` flag: Exit with error on null/missing values
- `-r` flag: Raw output (no quotes around strings)
- `-o=j` flag: JSON output format
- `-I=0` flag: Indentation control
- `e` command: Evaluate expression (shorthand)

**Impact**: ~90% of edge-cd tests use these flags

#### 2. Stdin Input ❌
- Reading YAML from stdin via pipe
- Pattern: `echo "$yaml" | yq '.path'`
- Pattern: `printf '%s\n' "$content" | yq -e '.path'`

**Impact**: ~40% of edge-cd tests use stdin

#### 3. Advanced Functions ❌
- `to_entries` - Convert object to key-value pairs array
- String concatenation with `+` operator
- Alternative operator `//` for default values

**Impact**: Critical for test 1 (extraEnvs key=value extraction)

#### 4. Output Formatting ❌
- Raw string output (no quotes)
- JSON output format
- Compact JSON (zero indentation)

**Impact**: Tests expect unquoted string values

## Detailed Test Results

### Test 1: Extract key=value from extraEnvs ❌
**Command**: `yq '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value'`

**Missing Features**:
- Alternative operator `//`
- `to_entries` function
- String concatenation with `+`

**Priority**: High (common pattern)

### Test 2-3: Read from stdin ❌
**Command**: `printf '%s\n' "${content}" | yq -e ".path"`

**Missing Features**:
- Stdin input support
- `-e` flag

**Priority**: High (very common pattern)

### Test 4-5: Read from file with/without -e ⚠️
**Command**: `yq -e '.config.setting' file.yaml`

**Current Status**:
- ✅ Basic path reading works
- ❌ `-e` flag not recognized
- ⚠️ Output has quotes: `"enabled"` instead of `enabled`

**Missing Features**:
- `-e` flag
- `-r` flag for raw output

**Priority**: Medium (works without flags, needs formatting)

### Test 6: Read entire file with raw output ❌
**Command**: `yq -e -r '.' file.yaml`

**Missing Features**:
- `-e` flag
- `-r` flag
- Identity operator `.` for entire file

**Priority**: Medium

### Test 7: Boolean values ⚠️
**Command**: `yq -e '.packageManager.autoUpgrade' file.yaml`

**Current Status**:
- ✅ Can read boolean values
- ❌ `-e` flag not implemented
- ❌ Error handling for missing values

**Priority**: Medium

### Test 8-9: Array iteration ✅
**Command**: `yq -e -r '.array[]' file.yaml`

**Current Status**:
- ✅ Array iteration works: `.array[]`
- ❌ `-e` and `-r` flags not implemented

**Priority**: Low (core feature works)

### Test 10: JSON output ❌
**Command**: `yq e -o=j -I=0 '.files[]' file.yaml`

**Missing Features**:
- `e` command
- `-o=j` flag (JSON output)
- `-I=0` flag (indentation)

**Priority**: Low (format conversion)

### Test 11: Read from JSON stdin ❌
**Command**: `echo "${json}" | yq -e -r '.[]'`

**Missing Features**:
- JSON input support
- Stdin input
- `-e` and `-r` flags

**Priority**: Low (JSON support)

### Test 12: Service manager commands ⚠️
**Command**: `yq -e -r '.commands.restart[]' file.yaml`

**Current Status**:
- ✅ Path navigation works
- ✅ Array iteration works
- ❌ `-e` and `-r` flags

**Priority**: Low (works without flags)

### Test 13: Empty array handling ✅
**Command**: `yq '.emptyArray[]' file.yaml`

**Current Status**:
- ✅ Empty arrays return empty output
- ✅ Null values handled

**Priority**: Low (works)

### Test 14: Complex nested structures ⚠️
**Command**: `yq -e '.spec.files[0].restartServices[]' file.yaml`

**Current Status**:
- ✅ Deep nesting works
- ⚠️ Array index `[0]` needs verification
- ❌ `-e` flag

**Priority**: Medium

### Test 15: Error handling with -e ❌
**Command**: `yq -e '.nonexistent.path' file.yaml`

**Missing Features**:
- `-e` flag for error on null
- Proper exit codes

**Priority**: High (error handling)

## Implementation Priority

### Phase 1: High Priority (Enable majority of tests)
1. **Stdin Input Support**
   - Read YAML from stdin when no file specified
   - Detect if stdin has data
   - Effort: 2-4 hours

2. **Raw Output Mode (-r flag)**
   - Remove quotes from string outputs
   - Keep quotes only for YAML output mode
   - Effort: 1-2 hours

3. **Error on Null (-e flag)**
   - Exit with code 1 on null/missing values
   - Skip exit on valid false/0 values
   - Effort: 2-3 hours

4. **Alternative Operator (//)**
   - Provide default value if null
   - Pattern: `.field // "default"`
   - Effort: 3-4 hours

### Phase 2: Medium Priority
5. **String Concatenation (+)**
   - Concatenate strings
   - Pattern: `.key + "=" + .value`
   - Effort: 2-3 hours

6. **to_entries Function**
   - Convert `{key: value}` to `[{key: "key", value: "value"}]`
   - Effort: 4-6 hours

7. **Array Indexing ([0])**
   - Already implemented, needs verification
   - Effort: 1 hour

### Phase 3: Low Priority
8. **JSON Output (-o=j)**
   - Convert YAML to JSON
   - Effort: 6-8 hours

9. **Indentation Control (-I)**
   - Control output indentation
   - Effort: 2-3 hours

10. **eval command (e)**
    - Shorthand for eval
    - Effort: 1 hour

## Estimated Total Effort

- **Phase 1 (Critical)**: ~10-15 hours → Would enable ~60-70% of tests
- **Phase 2 (Medium)**: ~10-15 hours → Would enable ~80-85% of tests
- **Phase 3 (Low)**: ~10-12 hours → Would enable ~95% of tests

**Total**: ~30-42 hours for full edge-cd compatibility

## Quick Wins

Features that could be implemented quickly for immediate test improvements:

1. **Flag Parsing**: Add `-e`, `-r` flag recognition (~2 hours)
2. **Raw Output**: Remove quotes in output (~1 hour)
3. **Stdin Support**: Read from stdin when no file given (~2 hours)

**Total Quick Wins**: ~5 hours → Would enable ~40-50% of tests

## Recommendations

1. **Immediate**: Implement stdin support + raw output mode
   - This alone would make 40-50% of tests runnable

2. **Short-term**: Add `-e` flag and `//` operator
   - This would enable proper error handling and defaults

3. **Long-term**: Add `to_entries`, `+` operator, JSON output
   - These enable advanced data transformation patterns

## Notes

- Array iteration `.[]` already works ✅
- Nested path access works ✅
- Basic YAML reading works ✅
- The foundation is solid, mainly need flag/format support
