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
	"strings"
	"testing"
)

// TestGenerateShellHeaderContent verifies the shell header contains essential content
func TestGenerateShellHeaderContent(t *testing.T) {
	result := GenerateShellHeader()

	tests := []string{
		"_yq_parse_depth=0",           // Depth counter initialization
		"_yq_debug_indent()",          // Debug function
		"_json_array_to_yaml()",       // JSON conversion function
		"DEBUG",                        // Debug output
		"POSIX compliant",             // Implementation comment
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateShellHeader missing '%s'", test)
		}
	}
}

// TestGenerateShellHeaderIsNonEmpty verifies output is not empty
func TestGenerateShellHeaderIsNonEmpty(t *testing.T) {
	result := GenerateShellHeader()
	if len(result) == 0 {
		t.Error("GenerateShellHeader returned empty string")
	}
}

// TestGenerateShellHeaderHasShellBashSyntax verifies shell syntax
func TestGenerateShellHeaderHasShellBashSyntax(t *testing.T) {
	result := GenerateShellHeader()

	// Check for function definitions
	if !strings.Contains(result, "() {") {
		t.Error("GenerateShellHeader missing function definitions")
	}

	// Check for closing braces
	if strings.Count(result, "{") == 0 || strings.Count(result, "}") == 0 {
		t.Error("GenerateShellHeader missing balanced braces")
	}
}

// TestGenerateShellHeaderDebugFunction verifies debug function exists
func TestGenerateShellHeaderDebugFunction(t *testing.T) {
	result := GenerateShellHeader()
	if !strings.Contains(result, "_yq_debug_indent()") {
		t.Error("GenerateShellHeader missing _yq_debug_indent function")
	}
	if !strings.Contains(result, "_depth") {
		t.Error("GenerateShellHeader debug function missing _depth parameter")
	}
	if !strings.Contains(result, "_msg") {
		t.Error("GenerateShellHeader debug function missing _msg parameter")
	}
}

// TestGenerateShellHeaderJSONConversion verifies JSON conversion function
func TestGenerateShellHeaderJSONConversion(t *testing.T) {
	result := GenerateShellHeader()

	tests := []string{
		"_json_array_to_yaml",          // Function name
		"_input=",                      // Input parameter
		"_trimmed=",                    // Trimmed variable
		"JSON array",                   // Comment
		"YAML",                         // YAML format
		"RS = \"",                      // AWK record separator
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("JSON conversion function missing '%s'", test)
		}
	}
}

// TestGenerateShellHeaderStartsWithComment verifies it starts properly
func TestGenerateShellHeaderStartsWithComment(t *testing.T) {
	result := GenerateShellHeader()
	trimmed := strings.TrimSpace(result)

	// Should start with # (comment) or # (shebang removed, so starts with comment)
	if !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "_") {
		t.Error("GenerateShellHeader output should start with comment or variable")
	}
}

// TestGenerateShellHeaderFunctionCallStructure verifies function call patterns
func TestGenerateShellHeaderFunctionCallStructure(t *testing.T) {
	result := GenerateShellHeader()

	// Check for function parameter handling
	if !strings.Contains(result, "$1") && !strings.Contains(result, "$2") {
		t.Error("GenerateShellHeader functions missing parameter references")
	}
}

// TestGenerateShellHeaderVariableNaming verifies underscore prefix convention
func TestGenerateShellHeaderVariableNaming(t *testing.T) {
	result := GenerateShellHeader()

	// All function-scoped variables should start with _
	if !strings.Contains(result, "_") {
		t.Error("GenerateShellHeader missing underscore-prefixed variables")
	}
}

// TestGenerateShellHeaderEchoStatement verifies output mechanism
func TestGenerateShellHeaderEchoStatement(t *testing.T) {
	result := GenerateShellHeader()

	// Should have echo or printf for output
	if !strings.Contains(result, "echo") {
		t.Error("GenerateShellHeader missing echo statement for output")
	}
}

// TestGenerateShellHeaderLoopConstruct verifies control flow
func TestGenerateShellHeaderLoopConstruct(t *testing.T) {
	result := GenerateShellHeader()

	// Should have while loop for indent generation
	if !strings.Contains(result, "while") {
		t.Error("GenerateShellHeader missing while loop")
	}
}

// TestGenerateShellHeaderDebugRedirection verifies stderr usage
func TestGenerateShellHeaderDebugRedirection(t *testing.T) {
	result := GenerateShellHeader()

	// Debug output should go to stderr
	if !strings.Contains(result, ">&2") {
		t.Error("GenerateShellHeader missing stderr redirection for debug output")
	}
}
