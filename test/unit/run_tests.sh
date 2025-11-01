# Copyright 2025 Alexandre Mahdhaoui
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


#!/bin/sh

# Unit test runner for posix-yq
# This script tests the generated posix-yq script against test scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Get the root directory of the project
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
POSIX_YQ="${POSIX_YQ:-${PROJECT_ROOT}/posix-yq}"
SCENARIOS_DIR="${PROJECT_ROOT}/test/unit/scenarios"

# Check if posix-yq exists
if [ ! -f "$POSIX_YQ" ]; then
    echo "${RED}Error: posix-yq script not found at $POSIX_YQ${NC}"
    echo "Please run 'make generate' first to create the posix-yq script"
    exit 1
fi

# Check if posix-yq is executable
if [ ! -x "$POSIX_YQ" ]; then
    echo "${RED}Error: posix-yq script is not executable${NC}"
    echo "Please run 'chmod +x $POSIX_YQ'"
    exit 1
fi

echo "Running posix-yq unit tests..."
echo "Using posix-yq at: $POSIX_YQ"
echo "Scenarios directory: $SCENARIOS_DIR"
echo ""

# Find all scenario directories (directories with names like 01-*, 02-*, etc.)
for scenario_dir in "$SCENARIOS_DIR"/[0-9][0-9]-*; do
    # Skip if not a directory
    if [ ! -d "$scenario_dir" ]; then
        continue
    fi

    SCENARIO_NAME=$(basename "$scenario_dir")
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Check if required files exist
    COMMAND_FILE="${scenario_dir}/command.txt"
    INPUT_FILE="${scenario_dir}/input.yaml"
    EXPECTED_OUTPUT="${scenario_dir}/output.yaml"

    if [ ! -f "$COMMAND_FILE" ]; then
        echo "${RED}✗ $SCENARIO_NAME: command.txt not found${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        continue
    fi

    if [ ! -f "$INPUT_FILE" ]; then
        echo "${RED}✗ $SCENARIO_NAME: input.yaml not found${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        continue
    fi

    if [ ! -f "$EXPECTED_OUTPUT" ]; then
        echo "${RED}✗ $SCENARIO_NAME: output.yaml not found${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        continue
    fi

    # Read the command (query) from command.txt
    # Remove any trailing newlines or whitespace
    QUERY=$(cat "$COMMAND_FILE" | tr -d '\n')

    # Run posix-yq and capture output
    # Use a temporary file to capture the actual output
    ACTUAL_OUTPUT=$(mktemp)
    ERROR_OUTPUT=$(mktemp)

    # Execute the command with a timeout (5 seconds per test)
    if timeout 5 "$POSIX_YQ" "$QUERY" "$INPUT_FILE" >"$ACTUAL_OUTPUT" 2>"$ERROR_OUTPUT"; then
        # Command executed successfully, compare output
        if diff -q "$EXPECTED_OUTPUT" "$ACTUAL_OUTPUT" >/dev/null 2>&1; then
            echo "${GREEN}✓ $SCENARIO_NAME${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo "${RED}✗ $SCENARIO_NAME: output mismatch${NC}"
            echo "  Query: $QUERY"
            echo "  Expected:"
            sed 's/^/    /' "$EXPECTED_OUTPUT"
            echo "  Got:"
            sed 's/^/    /' "$ACTUAL_OUTPUT"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        # Command failed
        echo "${RED}✗ $SCENARIO_NAME: command failed${NC}"
        echo "  Query: $QUERY"
        echo "  Error:"
        sed 's/^/    /' "$ERROR_OUTPUT"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    # Clean up temporary files
    rm -f "$ACTUAL_OUTPUT" "$ERROR_OUTPUT"
done

# Print summary
echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo "Total tests:  $TOTAL_TESTS"
echo "${GREEN}Passed tests: $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo "${RED}Failed tests: $FAILED_TESTS${NC}"
else
    echo "Failed tests: $FAILED_TESTS"
fi
echo "========================================"

# Exit with error if any tests failed
if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
fi

echo ""
echo "${GREEN}All tests passed!${NC}"
exit 0
