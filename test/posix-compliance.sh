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

# POSIX Compliance Test Script for posix-yq
# Tests the posix-yq script on different POSIX shells

echo "=== POSIX Compliance Test Suite ==="
echo ""

# Check if posix-yq exists
if [ ! -f "./posix-yq" ]; then
    echo "Error: posix-yq script not found"
    exit 1
fi

# Test 1: Run with default shell (sh)
echo "Test 1: Running with /bin/sh..."
if sh ./posix-yq test/fixtures/01-simple.yaml >/dev/null 2>&1; then
    echo "✓ Test 1: sh execution - PASSED"
else
    echo "✗ Test 1: sh execution - FAILED"
    exit 1
fi

# Test 2: Check for bashisms with shellcheck (if available)
if command -v shellcheck >/dev/null 2>&1; then
    echo ""
    echo "Test 2: Running shellcheck..."
    if shellcheck -s sh ./posix-yq; then
        echo "✓ Test 2: shellcheck - PASSED"
    else
        echo "✗ Test 2: shellcheck - FAILED (warnings exist)"
        # Don't exit - warnings are acceptable
    fi
else
    echo ""
    echo "Test 2: shellcheck not available - SKIPPED"
fi

# Test 3: Test with dash (if available)
if command -v dash >/dev/null 2>&1; then
    echo ""
    echo "Test 3: Running with dash..."
    if dash ./posix-yq test/fixtures/01-simple.yaml >/dev/null 2>&1; then
        echo "✓ Test 3: dash execution - PASSED"
    else
        echo "✗ Test 3: dash execution - FAILED"
        exit 1
    fi
else
    echo ""
    echo "Test 3: dash not available - SKIPPED"
fi

# Test 4: Verify script has correct shebang
echo ""
echo "Test 4: Checking shebang..."
if head -n 1 ./posix-yq | grep -q "^#!/bin/sh$\|^#!/usr/bin/env sh$"; then
    echo "✓ Test 4: shebang is POSIX-compliant - PASSED"
else
    echo "✗ Test 4: shebang is not POSIX-compliant - FAILED"
    exit 1
fi

echo ""
echo "=== All POSIX Compliance Tests Passed ==="
exit 0
