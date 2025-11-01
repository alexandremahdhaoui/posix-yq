// Copyright 2025 Alexandre Mahdhaoui
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package generator

import (
	"testing"
)

func TestYqUnquote(t *testing.T) {
	code := GenerateCoreFunctions()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "quoted string",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "unquoted string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "null value",
			input:    "null",
			expected: "null",
		},
		{
			name:     "boolean true",
			input:    "true",
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: "false",
		},
		{
			name:     "number",
			input:    "42",
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester.ExecuteFunctionExpect(tt.expected, "yq_unquote", tt.input)
		})
	}
}

func TestYqKeyAccess(t *testing.T) {
	code := GenerateCoreFunctions()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	// Test simple key access
	t.Run("simple key", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John\nage: 30")
		tester.ExecuteFunctionExpect("John", "yq_key_access", "name", testFile)
	})

	// Test missing key returns null
	t.Run("missing key", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John")
		tester.ExecuteFunctionExpect("null", "yq_key_access", "missing", testFile)
	})

	// Test numeric value
	t.Run("numeric value", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "count: 42")
		tester.ExecuteFunctionExpect("42", "yq_key_access", "count", testFile)
	})
}

func TestYqArrayAccess(t *testing.T) {
	code := GenerateCoreFunctions()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	// Test array index access
	t.Run("first element", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "- apple\n- banana\n- cherry")
		tester.ExecuteFunctionExpect("apple", "yq_array_access", "[0]", testFile)
	})

	// Test negative index
	t.Run("last element with negative index", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "- apple\n- banana\n- cherry")
		tester.ExecuteFunctionExpect("cherry", "yq_array_access", "[-1]", testFile)
	})
}

func TestYqIterate(t *testing.T) {
	code := GenerateCoreFunctions()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	// Test array iteration
	t.Run("array iteration", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "- apple\n- banana\n- cherry")
		output, _ := tester.ExecuteFunction("yq_iterate", testFile)

		// Output should contain all elements (may have blank line separators)
		if output != "apple\n\nbanana\n\ncherry\n" {
			t.Errorf("Expected array elements, got: %q", output)
		}
	})

	// Test object iteration
	t.Run("object iteration", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John\nage: 30")
		output, _ := tester.ExecuteFunction("yq_iterate", testFile)

		// Output should contain both values
		if output == "" {
			t.Errorf("Expected object values, got empty output")
		}
	})
}
