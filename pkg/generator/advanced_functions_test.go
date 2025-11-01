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

func TestYqCompare(t *testing.T) {
	// yq_compare depends on yq_parse
	tester := NewShellFunctionTesterWithDeps(t,
		GenerateShellHeader(),
		GenerateCoreFunctions(),
		GenerateParser(),
		GenerateAdvancedFunctions(),
	)
	defer tester.Cleanup()

	tests := []struct {
		name        string
		expr        string
		fileContent string
		expected    string
	}{
		{
			name:        "simple equality true",
			expr:        ". == 5",
			fileContent: "5",
			expected:    "true",
		},
		{
			name:        "simple equality false",
			expr:        ". == 5",
			fileContent: "3",
			expected:    "false",
		},
		{
			name:        "string equality",
			expr:        `. == "hello"`,
			fileContent: `"hello"`,
			expected:    "true",
		},
		{
			name:        "not equal operator",
			expr:        ". != 5",
			fileContent: "3",
			expected:    "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tester.WriteFile("test.yaml", tt.fileContent)
			tester.ExecuteFunctionExpect(tt.expected, "yq_compare", tt.expr, testFile)
		})
	}
}

func TestYqSelect(t *testing.T) {
	// yq_select depends on yq_parse
	tester := NewShellFunctionTesterWithDeps(t,
		GenerateShellHeader(),
		GenerateCoreFunctions(),
		GenerateParser(),
		GenerateAdvancedFunctions(),
	)
	defer tester.Cleanup()

	t.Run("select with truthy condition", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "apple")
		// Using == "apple" which should evaluate to "true"
		output, _ := tester.ExecuteFunction("yq_select", `. == "apple"`, testFile)

		// Should return the input since condition is truthy
		if output == "" {
			t.Errorf("Expected select output, got empty")
		}

		// Check that the output is the input file content
		if !contains(output, "apple") {
			t.Errorf("Expected output to contain 'apple', got: %s", output)
		}
	})
}

func TestYqMap(t *testing.T) {
	code := GenerateAdvancedFunctions()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	t.Run("map arithmetic", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "- 1\n- 2\n- 3")
		output, _ := tester.ExecuteFunction("yq_map", ". * 2", testFile)

		// Should contain mapped values
		if output != "- 2\n- 4\n- 6\n" {
			t.Errorf("Expected mapped values, got: %q", output)
		}
	})
}

func TestYqKeys(t *testing.T) {
	code := GenerateAdvancedFunctions()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	t.Run("object keys", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John\nage: 30\ncity: NYC")
		output, _ := tester.ExecuteFunction("yq_keys", testFile)

		// Should contain keys
		if output == "" {
			t.Errorf("Expected keys output, got empty")
		}
	})
}
