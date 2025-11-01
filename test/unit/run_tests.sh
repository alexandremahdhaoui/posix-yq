#!/bin/sh
#
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

    # Try to find input file - could be .yaml, .json, or other formats
    INPUT_FILE=""
    if [ -f "${scenario_dir}/input.yaml" ]; then
        INPUT_FILE="${scenario_dir}/input.yaml"
    elif [ -f "${scenario_dir}/input.json" ]; then
        INPUT_FILE="${scenario_dir}/input.json"
    elif [ -f "${scenario_dir}/input.txt" ]; then
        INPUT_FILE="${scenario_dir}/input.txt"
    fi

    # Expected output could be .yaml or .txt (for error messages)
    EXPECTED_OUTPUT=""
    if [ -f "${scenario_dir}/output.yaml" ]; then
        EXPECTED_OUTPUT="${scenario_dir}/output.yaml"
    elif [ -f "${scenario_dir}/output.txt" ]; then
        EXPECTED_OUTPUT="${scenario_dir}/output.txt"
    fi

    if [ ! -f "$COMMAND_FILE" ]; then
        echo "${RED}✗ $SCENARIO_NAME: command.txt not found${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        continue
    fi

    if [ -z "$EXPECTED_OUTPUT" ]; then
        echo "${RED}✗ $SCENARIO_NAME: output file not found (output.yaml or output.txt)${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        continue
    fi

    # Read the command from command.txt
    # The command may contain multiple arguments (e.g., "eval-all '.query'")
    COMMAND=$(cat "$COMMAND_FILE" | tr -d '\n')

    # Run posix-yq and capture output
    # Use a temporary file to capture the actual output
    ACTUAL_OUTPUT=$(mktemp)
    ERROR_OUTPUT=$(mktemp)

    # Build the command - if INPUT_FILE exists, include it; otherwise, run without it
    # This handles commands that work on stdin or expect file-not-found errors
    # Use eval in a subshell to properly handle commands with spaces and special characters
    if [ -n "$INPUT_FILE" ]; then
        timeout 5 sh -c "\"$POSIX_YQ\" \"$COMMAND\" \"$INPUT_FILE\" >\"$ACTUAL_OUTPUT\" 2>\"$ERROR_OUTPUT\""
    else
        timeout 5 sh -c "\"$POSIX_YQ\" \"$COMMAND\" >\"$ACTUAL_OUTPUT\" 2>\"$ERROR_OUTPUT\""
    fi
    EXIT_CODE=$?

    # Check if we expect error output (output.txt) or regular output (output.yaml)
    if [ "${EXPECTED_OUTPUT##*.}" = "txt" ]; then
        # This test expects an error message - compare stderr
        if [ $EXIT_CODE -ne 0 ]; then
            # Command failed as expected, compare error output
            if diff -q "$EXPECTED_OUTPUT" "$ERROR_OUTPUT" >/dev/null 2>&1; then
                echo "${GREEN}✓ $SCENARIO_NAME${NC}"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo "${RED}✗ $SCENARIO_NAME: error output mismatch${NC}"
                echo "  Command: $COMMAND"
                echo "  Expected error:"
                sed 's/^/    /' "$EXPECTED_OUTPUT"
                echo "  Got error:"
                sed 's/^/    /' "$ERROR_OUTPUT"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            echo "${RED}✗ $SCENARIO_NAME: expected command to fail but it succeeded${NC}"
            echo "  Command: $COMMAND"
            echo "  Output:"
            sed 's/^/    /' "$ACTUAL_OUTPUT"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        # Regular test - compare stdout
        if [ $EXIT_CODE -eq 0 ]; then
            # Command executed successfully, compare output
            if diff -q "$EXPECTED_OUTPUT" "$ACTUAL_OUTPUT" >/dev/null 2>&1; then
                echo "${GREEN}✓ $SCENARIO_NAME${NC}"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo "${RED}✗ $SCENARIO_NAME: output mismatch${NC}"
                echo "  Command: $COMMAND"
                echo "  Expected:"
                sed 's/^/    /' "$EXPECTED_OUTPUT"
                echo "  Got:"
                sed 's/^/    /' "$ACTUAL_OUTPUT"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            # Command failed
            echo "${RED}✗ $SCENARIO_NAME: command failed${NC}"
            echo "  Command: $COMMAND"
            echo "  Error:"
            sed 's/^/    /' "$ERROR_OUTPUT"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
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
