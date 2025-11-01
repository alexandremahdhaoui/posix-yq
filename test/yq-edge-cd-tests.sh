#!/bin/sh
# yq Test Script
# Tests all yq command patterns found in edge-cd shell scripts
# Usage: ./yq-test-script <path-to-yq-binary>

set -e

YQ_BIN="${1}"

if [ -z "${YQ_BIN}" ]; then
    echo "Error: YQ_BIN path required as first argument"
    echo "Usage: $0 <path-to-yq-binary>"
    exit 1
fi

if [ ! -x "${YQ_BIN}" ]; then
    echo "Error: ${YQ_BIN} is not executable or does not exist"
    exit 1
fi

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

TESTS_PASSED=0
TESTS_FAILED=0
TEST_DIR=".tmp/yq-test-data-$$"

# Cleanup function
cleanup() {
    rm -rf "${TEST_DIR}"
}
trap cleanup EXIT

# Create test directory
mkdir -p "${TEST_DIR}"

# Test helper functions
test_pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    printf "${GREEN}✓${NC} %s\n" "$1"
}

test_fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    printf "${RED}✗${NC} %s\n" "$1"
    printf "  Expected: %s\n" "$2"
    printf "  Got:      %s\n" "$3"
}

assert_equals() {
    test_name="$1"
    expected="$2"
    actual="$3"

    if [ "${expected}" = "${actual}" ]; then
        test_pass "${test_name}"
    else
        test_fail "${test_name}" "${expected}" "${actual}"
    fi
}

assert_contains() {
    test_name="$1"
    expected_substring="$2"
    actual="$3"

    case "${actual}" in
        *"${expected_substring}"*)
            test_pass "${test_name}"
            ;;
        *)
            test_fail "${test_name}" "contains '${expected_substring}'" "${actual}"
            ;;
    esac
}

# ============================================================
# TEST 1: Extract key=value pairs from extraEnvs
# Command: yq '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value' file.yaml
# ============================================================
echo "=== Test 1: Extract key=value from extraEnvs ==="

cat >"${TEST_DIR}/extraenvs.yaml" <<'EOF'
extraEnvs:
  - FOO: "bar"
  - BAZ: "qux"
  - NUM: "123"
EOF

expected="FOO=bar
BAZ=qux
NUM=123"
actual=$("${YQ_BIN}" '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value' "${TEST_DIR}/extraenvs.yaml")
assert_equals "extraEnvs key=value extraction" "${expected}" "${actual}"

# Test with empty extraEnvs
cat >"${TEST_DIR}/no-extraenvs.yaml" <<'EOF'
other: value
EOF

expected=""
actual=$("${YQ_BIN}" '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value' "${TEST_DIR}/no-extraenvs.yaml" || true)
assert_equals "extraEnvs with missing field (default to empty)" "${expected}" "${actual}"

# ============================================================
# TEST 2: Read YAML path from stdin with -e (error on null)
# Command: printf '%s\n' "${content}" | yq -e ".path"
# ============================================================
echo ""
echo "=== Test 2: Read YAML path from stdin with -e ==="

yaml_content='foo:
  bar: "value123"'

expected="value123"
actual=$(printf '%s\n' "${yaml_content}" | "${YQ_BIN}" -e '.foo.bar')
assert_equals "stdin read with -e flag" "${expected}" "${actual}"

# ============================================================
# TEST 3: Read YAML path from stdin without -e (optional)
# Command: printf '%s\n' "${content}" | yq ".path"
# ============================================================
echo ""
echo "=== Test 3: Read YAML path from stdin without -e ==="

yaml_content='foo:
  bar: "value456"'

expected="value456"
actual=$(printf '%s\n' "${yaml_content}" | "${YQ_BIN}" '.foo.bar')
assert_equals "stdin read without -e flag" "${expected}" "${actual}"

# Test null handling
yaml_content='foo:
  bar: "value"'

expected="null"
actual=$(printf '%s\n' "${yaml_content}" | "${YQ_BIN}" '.missing.path')
assert_equals "stdin read null path returns 'null'" "${expected}" "${actual}"

# ============================================================
# TEST 4: Read YAML path from file with -e
# Command: yq -e ".path" file.yaml
# ============================================================
echo ""
echo "=== Test 4: Read YAML path from file with -e ==="

cat >"${TEST_DIR}/test.yaml" <<'EOF'
config:
  setting: "enabled"
  value: 42
EOF

expected="enabled"
actual=$("${YQ_BIN}" -e '.config.setting' "${TEST_DIR}/test.yaml")
assert_equals "file read with -e flag" "${expected}" "${actual}"

expected="42"
actual=$("${YQ_BIN}" -e '.config.value' "${TEST_DIR}/test.yaml")
assert_equals "file read numeric value" "${expected}" "${actual}"

# ============================================================
# TEST 5: Read YAML path from file without -e (optional)
# Command: yq ".path" file.yaml
# ============================================================
echo ""
echo "=== Test 5: Read YAML path from file without -e ==="

expected="enabled"
actual=$("${YQ_BIN}" '.config.setting' "${TEST_DIR}/test.yaml")
assert_equals "file read without -e flag" "${expected}" "${actual}"

expected="null"
actual=$("${YQ_BIN}" '.nonexistent' "${TEST_DIR}/test.yaml")
assert_equals "file read null path returns 'null'" "${expected}" "${actual}"

# ============================================================
# TEST 6: Read entire YAML file with raw output
# Command: yq -e -r '.' file.yaml
# ============================================================
echo ""
echo "=== Test 6: Read entire file with raw output ==="

cat >"${TEST_DIR}/package-config.yaml" <<'EOF'
name: "test-package"
version: "1.0.0"
install:
  - apt-get
  - install
  - -y
update:
  - apt-get
  - update
upgrade:
  - apt-get
  - upgrade
  - -y
EOF

actual=$("${YQ_BIN}" -e -r '.' "${TEST_DIR}/package-config.yaml")
assert_contains "read entire file" "test-package" "${actual}"
assert_contains "read entire file has version" "1.0.0" "${actual}"

# ============================================================
# TEST 7: Read boolean value
# Command: yq -e '.packageManager.autoUpgrade' file.yaml
# ============================================================
echo ""
echo "=== Test 7: Read boolean values ==="

cat >"${TEST_DIR}/config-spec.yaml" <<'EOF'
packageManager:
  autoUpgrade: true
  requiredPackages:
    - git
    - curl
    - vim
EOF

expected="true"
actual=$("${YQ_BIN}" -e '.packageManager.autoUpgrade' "${TEST_DIR}/config-spec.yaml")
assert_equals "read boolean true" "${expected}" "${actual}"

cat >"${TEST_DIR}/config-spec-false.yaml" <<'EOF'
packageManager:
  autoUpgrade: false
EOF

expected="false"
# Without -e flag, false values are read correctly without exit code 1
actual=$("${YQ_BIN}" '.packageManager.autoUpgrade' "${TEST_DIR}/config-spec-false.yaml")
assert_equals "read boolean false" "${expected}" "${actual}"

# Test the actual pattern used in code: || echo "false" for missing fields
# Note: yq -e outputs "null" for missing fields, then || echo adds "false"
# Result is "null\nfalse" but in practice, code checks != "true" so it works
cat >"${TEST_DIR}/config-spec-missing.yaml" <<'EOF'
packageManager:
  name: "apt"
EOF

expected="null
false"
actual=$("${YQ_BIN}" -e '.packageManager.autoUpgrade' "${TEST_DIR}/config-spec-missing.yaml" 2>/dev/null || echo "false")
assert_equals "read missing boolean with default (null + false)" "${expected}" "${actual}"

# Alternative: suppress null output by checking for it
actual=$("${YQ_BIN}" '.packageManager.autoUpgrade' "${TEST_DIR}/config-spec-missing.yaml" 2>/dev/null)
if [ "${actual}" = "null" ]; then
    actual="false"
fi
expected="false"
assert_equals "read missing boolean clean (null -> false)" "${expected}" "${actual}"

# ============================================================
# TEST 8: Read array elements and output each on newline
# Command: yq -e -r '.array[]'
# ============================================================
echo ""
echo "=== Test 8: Read array elements ==="

expected="apt-get
update"
actual=$("${YQ_BIN}" -e -r '.update[]' "${TEST_DIR}/package-config.yaml")
assert_equals "read update array" "${expected}" "${actual}"

expected="apt-get
install
-y"
actual=$("${YQ_BIN}" -e -r '.install[]' "${TEST_DIR}/package-config.yaml")
assert_equals "read install array" "${expected}" "${actual}"

# ============================================================
# TEST 9: Read nested array (packageManager.requiredPackages)
# Command: yq -e '.packageManager.requiredPackages[]' file.yaml
# ============================================================
echo ""
echo "=== Test 9: Read nested array elements ==="

expected="git
curl
vim"
actual=$("${YQ_BIN}" -e '.packageManager.requiredPackages[]' "${TEST_DIR}/config-spec.yaml")
assert_equals "read required packages array" "${expected}" "${actual}"

# Test with tr '\n' ' ' to join with spaces (common pattern)
expected="git curl vim "
actual=$("${YQ_BIN}" -e '.packageManager.requiredPackages[]' "${TEST_DIR}/config-spec.yaml" | tr '\n' ' ')
assert_equals "read packages and join with space" "${expected}" "${actual}"

# ============================================================
# TEST 10: Output JSON format with specific indentation
# Command: yq e -o=j -I=0 '.files[]' file.yaml
# ============================================================
echo ""
echo "=== Test 10: JSON output with zero indentation ==="

cat >"${TEST_DIR}/files-config.yaml" <<'EOF'
files:
  - source: "/src/nginx.conf"
    dest: "/etc/nginx/nginx.conf"
    mode: "644"
    restartServices:
      - nginx
  - source: "/src/app.conf"
    dest: "/etc/app/app.conf"
    mode: "600"
    restartServices:
      - app-service
EOF

actual=$("${YQ_BIN}" e -o=j -I=0 '.files[]' "${TEST_DIR}/files-config.yaml")
# Just verify it's valid JSON format (compact)
assert_contains "JSON output format" '"source":"/src/nginx.conf"' "${actual}"
assert_contains "JSON output has dest" '"dest":"/etc/nginx/nginx.conf"' "${actual}"
assert_contains "JSON output has mode" '"mode":"644"' "${actual}"

# Verify it outputs multiple JSON objects (one per line)
line_count=$(echo "${actual}" | grep -c . || echo 0)
expected="2"
assert_equals "JSON output line count" "${expected}" "${line_count}"

# ============================================================
# TEST 11: Read and transform service restart array
# Command: echo "${json}" | yq -e -r '.[]'
# ============================================================
echo ""
echo "=== Test 11: Read service array from JSON ==="

services_json='["nginx","apache2","app-service"]'
expected="nginx
apache2
app-service"
actual=$(echo "${services_json}" | "${YQ_BIN}" -e -r '.[]')
assert_equals "read services from JSON array" "${expected}" "${actual}"

# ============================================================
# TEST 12: Service manager commands with array expansion
# Command: yq -e -r '.commands.restart[]' | sed | tr
# ============================================================
echo ""
echo "=== Test 12: Service manager command arrays ==="

cat >"${TEST_DIR}/svc-mgr.yaml" <<'EOF'
commands:
  restart:
    - systemctl
    - restart
    - __SERVICE_NAME__
  enable:
    - systemctl
    - enable
    - __SERVICE_NAME__
EOF

expected="systemctl
restart
__SERVICE_NAME__"
actual=$("${YQ_BIN}" -e -r '.commands.restart[]' "${TEST_DIR}/svc-mgr.yaml")
assert_equals "read restart command array" "${expected}" "${actual}"

# Test with transformation (sed + tr as used in code)
expected="systemctl restart nginx "
actual=$("${YQ_BIN}" -e -r '.commands.restart[]' "${TEST_DIR}/svc-mgr.yaml" | sed -e 's/__SERVICE_NAME__/nginx/g' | tr '\n' ' ')
assert_equals "transform restart command with sed and tr" "${expected}" "${actual}"

expected="systemctl
enable
__SERVICE_NAME__"
actual=$("${YQ_BIN}" -e -r '.commands.enable[]' "${TEST_DIR}/svc-mgr.yaml")
assert_equals "read enable command array" "${expected}" "${actual}"

# ============================================================
# TEST 13: Empty array handling
# ============================================================
echo ""
echo "=== Test 13: Empty array handling ==="

cat >"${TEST_DIR}/empty-arrays.yaml" <<'EOF'
emptyArray: []
nullValue: null
missingField: ~
EOF

expected=""
actual=$("${YQ_BIN}" '.emptyArray[]' "${TEST_DIR}/empty-arrays.yaml" || true)
assert_equals "empty array returns empty" "${expected}" "${actual}"

expected="null"
actual=$("${YQ_BIN}" '.nullValue' "${TEST_DIR}/empty-arrays.yaml")
assert_equals "null value returns 'null'" "${expected}" "${actual}"

# ============================================================
# TEST 14: Complex nested structure
# ============================================================
echo ""
echo "=== Test 14: Complex nested structure ==="

cat >"${TEST_DIR}/complex.yaml" <<'EOF'
spec:
  packageManager:
    name: "apt"
    autoUpgrade: true
    requiredPackages:
      - git
      - curl
  files:
    - source: "/src/file1"
      dest: "/dest/file1"
      mode: "644"
      restartServices:
        - service1
        - service2
extraEnvs:
  - FOO: "bar"
  - BAZ: "qux"
EOF

expected="apt"
actual=$("${YQ_BIN}" -e '.spec.packageManager.name' "${TEST_DIR}/complex.yaml")
assert_equals "nested packageManager name" "${expected}" "${actual}"

expected="service1
service2"
actual=$("${YQ_BIN}" -e '.spec.files[0].restartServices[]' "${TEST_DIR}/complex.yaml")
assert_equals "nested restartServices array" "${expected}" "${actual}"

# ============================================================
# TEST 15: Error handling with -e flag
# ============================================================
echo ""
echo "=== Test 15: Error handling with -e flag ==="

# This should fail with exit code != 0
if "${YQ_BIN}" -e '.nonexistent.deeply.nested.path' "${TEST_DIR}/complex.yaml" 2>/dev/null; then
    test_fail "yq -e should fail on missing path" "non-zero exit code" "zero exit code"
else
    test_pass "yq -e fails correctly on missing path"
fi

# ============================================================
# SUMMARY
# ============================================================
echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
printf "${GREEN}Passed:${NC} %d\n" "${TESTS_PASSED}"
printf "${RED}Failed:${NC} %d\n" "${TESTS_FAILED}"
echo "========================================"

if [ "${TESTS_FAILED}" -gt 0 ]; then
    echo "Some tests failed!"
    exit 1
else
    echo "All tests passed!"
    exit 0
fi
