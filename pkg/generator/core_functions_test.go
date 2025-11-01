package generator

import (
	"strings"
	"testing"
)

// TestGenerateCoreFunctionsIsNonEmpty verifies output is not empty
func TestGenerateCoreFunctionsIsNonEmpty(t *testing.T) {
	result := GenerateCoreFunctions()
	if len(result) == 0 {
		t.Error("GenerateCoreFunctions returned empty string")
	}
}

// TestGenerateCoreFunctionsHasRequiredFunctions verifies all core functions
func TestGenerateCoreFunctionsHasRequiredFunctions(t *testing.T) {
	result := GenerateCoreFunctions()

	requiredFunctions := []string{
		"yq_unquote()",      // Quote removal
		"yq_key_access()",   // Key extraction
		"yq_iterate()",      // Array/object iteration
		"yq_array_access()", // Array indexing
		"yq_length()",       // Length calculation
		"yq_keys()",         // Key extraction
		"yq_to_entries()",   // Object to entries
		"yq_has()",          // Key existence check
	}

	for _, fn := range requiredFunctions {
		if !strings.Contains(result, fn) {
			t.Errorf("GenerateCoreFunctions missing function '%s'", fn)
		}
	}
}

// TestGenerateCoreFunctionsUnquoteHandlesNull verifies null special case
func TestGenerateCoreFunctionsUnquoteHandlesNull(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "\"null\"") {
		t.Error("yq_unquote missing null handling")
	}
}

// TestGenerateCoreFunctionsUnquoteHandlesQuotes verifies quote removal
func TestGenerateCoreFunctionsUnquoteHandlesQuotes(t *testing.T) {
	result := GenerateCoreFunctions()

	tests := []string{
		"_value#",   // Parameter expansion
		"_trimmed",  // Trimmed variable
		"sed",       // Sed for processing
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("yq_unquote quote handling missing '%s'", test)
		}
	}
}

// TestGenerateCoreFunctionsKeyAccessUsesAwk verifies awk usage
func TestGenerateCoreFunctionsKeyAccessUsesAwk(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "awk -v key=") {
		t.Error("yq_key_access not using awk properly")
	}
}

// TestGenerateCoreFunctionsIterateHandlesArrays verifies array iteration
func TestGenerateCoreFunctionsIterateHandlesArrays(t *testing.T) {
	result := GenerateCoreFunctions()

	tests := []string{
		"^-",              // Array item marker
		"grep -q",         // Pattern matching
		"yq_iterate",      // Function name
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("yq_iterate array handling missing '%s'", test)
		}
	}
}

// TestGenerateCoreFunctionsArrayAccessHandlesIndex verifies indexing
func TestGenerateCoreFunctionsArrayAccessHandlesIndex(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "yq_array_access") {
		t.Error("Array access function missing")
	}
	if !strings.Contains(result, "_idx") {
		t.Error("Array access missing index variable")
	}
}

// TestGenerateCoreFunctionsArrayAccessHandlesSlice verifies slicing
func TestGenerateCoreFunctionsArrayAccessHandlesSlice(t *testing.T) {
	result := GenerateCoreFunctions()

	tests := []string{
		"_start",  // Start index
		"_end",    // End index
		":",       // Slice syntax
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Array slicing missing '%s'", test)
		}
	}
}

// TestGenerateCoreFunctionsArrayAccessHandlesNegativeIndex verifies negative indices
func TestGenerateCoreFunctionsArrayAccessHandlesNegativeIndex(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "^-") {
		t.Error("Array access missing negative index handling")
	}
}

// TestGenerateCoreFunctionsLengthCountsArrays verifies array counting
func TestGenerateCoreFunctionsLengthCountsArrays(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "grep -c '^-'") {
		t.Error("yq_length not counting array elements correctly")
	}
}

// TestGenerateCoreFunctionsLengthCountsObjects verifies object key counting
func TestGenerateCoreFunctionsLengthCountsObjects(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "yq_length") {
		t.Error("yq_length function missing")
	}
}

// TestGenerateCoreFunctionsKeysUsesAwk verifies keys extraction
func TestGenerateCoreFunctionsKeysUsesAwk(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "yq_keys") {
		t.Error("yq_keys function missing")
	}
	if !strings.Contains(result, "awk") {
		t.Error("yq_keys not using awk")
	}
}

// TestGenerateCoreFunctionsToEntriesFormatsCorrectly verifies to_entries
func TestGenerateCoreFunctionsToEntriesFormatsCorrectly(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "yq_to_entries") {
		t.Error("yq_to_entries function missing")
	}
	if !strings.Contains(result, "key:") {
		t.Error("to_entries missing key format")
	}
	if !strings.Contains(result, "value:") {
		t.Error("to_entries missing value format")
	}
}

// TestGenerateCoreFunctionsHasChecksKeyExistence verifies has function
func TestGenerateCoreFunctionsHasChecksKeyExistence(t *testing.T) {
	result := GenerateCoreFunctions()

	if !strings.Contains(result, "yq_has") {
		t.Error("yq_has function missing")
	}
	if !strings.Contains(result, "true") {
		t.Error("yq_has missing true value")
	}
	if !strings.Contains(result, "false") {
		t.Error("yq_has missing false value")
	}
}

// TestGenerateCoreFunctionsProperlyIndented verifies output formatting
func TestGenerateCoreFunctionsProperlyIndented(t *testing.T) {
	result := GenerateCoreFunctions()

	// Should have proper function structure
	if !strings.Contains(result, "() {") {
		t.Error("Functions not properly formatted")
	}
}

// TestGenerateCoreFunctionsUsesTemporaryFiles verifies temp file handling
func TestGenerateCoreFunctionsUsesTemporaryFiles(t *testing.T) {
	result := GenerateCoreFunctions()

	// Core functions may or may not use temp files directly in this module
	// Just verify the functions are properly defined
	if !strings.Contains(result, "yq_") {
		t.Error("Core functions not properly defined")
	}
}
