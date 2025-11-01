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

// TestGenerateParserFunctionExists verifies main parser function is present
func TestGenerateParserFunctionExists(t *testing.T) {
	result := GenerateParser()

	if !strings.Contains(result, "yq_parse()") {
		t.Error("GenerateParser missing yq_parse function")
	}
}

// TestGenerateParserIsNonEmpty verifies output is not empty
func TestGenerateParserIsNonEmpty(t *testing.T) {
	result := GenerateParser()
	if len(result) == 0 {
		t.Error("GenerateParser returned empty string")
	}
}

// TestGenerateParserHandlesPipeOperator verifies pipe operator handling
func TestGenerateParserHandlesPipeOperator(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"Pipe detected",  // Debug comment
		"_before_pipe",   // Variable
		"_after_pipe",    // Variable
		"| ",             // Pipe with space
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser pipe handling missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesAlternativeOperator verifies // operator
func TestGenerateParserHandlesAlternativeOperator(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"//",               // Alternative operator
		"_before_alt",      // Variable
		"_after_alt",       // Variable
		" // ",             // Operator with spaces
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser alternative operator handling missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesArrayIteration verifies .[] support
func TestGenerateParserHandlesArrayIteration(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"\\[\\]",              // Array iteration pattern
		"yq_iterate",          // Iteration function call
		"Iteration handler",   // Debug comment
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser array iteration missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesStringConcatenation verifies + operator
func TestGenerateParserHandlesStringConcatenation(t *testing.T) {
	result := GenerateParser()

	if !strings.Contains(result, " + ") {
		t.Error("Parser string concatenation missing + operator handling")
	}
}

// TestGenerateParserHandlesKeyAccess verifies key access parsing
func TestGenerateParserHandlesKeyAccess(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"yq_key_access",   // Function call
		"_first_token",    // Token variable
		"_remainder",      // Remainder variable
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser key access handling missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesArrayAccess verifies [n] access
func TestGenerateParserHandlesArrayAccess(t *testing.T) {
	result := GenerateParser()

	if !strings.Contains(result, "yq_array_access") {
		t.Error("Parser array access missing yq_array_access function call")
	}
}

// TestGenerateParserHandlesRecursiveDescent verifies .. operator
func TestGenerateParserHandlesRecursiveDescent(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"..",                           // Recursive descent operator
		"yq_recursive_descent_pipe",   // Function call
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser recursive descent handling missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesDepthTracking verifies depth counter
func TestGenerateParserHandlesDepthTracking(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"_yq_parse_depth",      // Depth variable
		"$((_yq_parse_depth",   // Increment operation
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser depth tracking missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesSpecialFunctions verifies built-in functions
func TestGenerateParserHandlesSpecialFunctions(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"length",      // Function name
		"keys",        // Function name
		"to_entries",  // Function name
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser special functions missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesOperators verifies operator functions
func TestGenerateParserHandlesOperators(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"yq_assign",    // Assignment operator
		"yq_update",    // Update operator
		"yq_compare",   // Comparison operator
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser operators handling missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesMapSelect verifies map and select
func TestGenerateParserHandlesMapSelect(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"yq_map",    // Map function
		"yq_select", // Select function
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Parser map/select missing '%s'", test)
		}
	}
}

// TestGenerateParserHandlesTemporaryFiles verifies mktemp usage
func TestGenerateParserHandlesTemporaryFiles(t *testing.T) {
	result := GenerateParser()

	if !strings.Contains(result, "mktemp") {
		t.Error("Parser missing mktemp for temporary file handling")
	}
}

// TestGenerateParserHasErrorHandling verifies error handling
func TestGenerateParserHasErrorHandling(t *testing.T) {
	result := GenerateParser()

	if !strings.Contains(result, "rm -f") {
		t.Error("Parser missing cleanup (rm -f) for temporary files")
	}
}

// TestGenerateParserHandlesQuoteRemoval verifies quote handling
func TestGenerateParserHandlesQuoteRemoval(t *testing.T) {
	result := GenerateParser()

	if !strings.Contains(result, "sed") {
		t.Error("Parser missing sed for quote handling")
	}
}
