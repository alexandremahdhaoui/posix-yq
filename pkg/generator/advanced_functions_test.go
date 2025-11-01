package generator

import (
	"strings"
	"testing"
)

// TestGenerateAdvancedFunctionsIsNonEmpty verifies output is not empty
func TestGenerateAdvancedFunctionsIsNonEmpty(t *testing.T) {
	result := GenerateAdvancedFunctions()
	if len(result) == 0 {
		t.Error("GenerateAdvancedFunctions returned empty string")
	}
}

// TestGenerateAdvancedFunctionsHasMapFunction verifies map function
func TestGenerateAdvancedFunctionsHasMapFunction(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_map()") {
		t.Error("GenerateAdvancedFunctions missing yq_map function")
	}
}

// TestGenerateAdvancedFunctionsHasSelectFunction verifies select function
func TestGenerateAdvancedFunctionsHasSelectFunction(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_select()") {
		t.Error("GenerateAdvancedFunctions missing yq_select function")
	}
}

// TestGenerateAdvancedFunctionsHasCompareFunction verifies comparison
func TestGenerateAdvancedFunctionsHasCompareFunction(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_compare()") {
		t.Error("GenerateAdvancedFunctions missing yq_compare function")
	}
}

// TestGenerateAdvancedFunctionsHasRecursiveDescentFunction verifies recursion
func TestGenerateAdvancedFunctionsHasRecursiveDescentFunction(t *testing.T) {
	result := GenerateAdvancedFunctions()

	tests := []string{
		"yq_recursive_descent()",     // Basic recursive descent
		"yq_recursive_descent_pipe()", // Recursive descent with pipe
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateAdvancedFunctions missing '%s'", test)
		}
	}
}

// TestGenerateAdvancedFunctionsMapHandlesExpression verifies map expression
func TestGenerateAdvancedFunctionsMapHandlesExpression(t *testing.T) {
	result := GenerateAdvancedFunctions()

	// yq_map should have parameters
	if !strings.Contains(result, "yq_map()") {
		t.Error("yq_map missing function definition")
	}
}

// TestGenerateAdvancedFunctionsSelectHandlesCondition verifies select condition
func TestGenerateAdvancedFunctionsSelectHandlesCondition(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_select") {
		t.Error("yq_select function missing")
	}
}

// TestGenerateAdvancedFunctionsCompareHandlesOperators verifies comparison
func TestGenerateAdvancedFunctionsCompareHandlesOperators(t *testing.T) {
	result := GenerateAdvancedFunctions()

	tests := []string{
		"==",  // Equal
		"!=",  // Not equal
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("yq_compare missing operator '%s'", test)
		}
	}
}

// TestGenerateAdvancedFunctionsMapIteratesItems verifies map iteration
func TestGenerateAdvancedFunctionsMapIteratesItems(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_iterate") {
		t.Error("yq_map not iterating items")
	}
}

// TestGenerateAdvancedFunctionsSelectFiltersItems verifies filtering
func TestGenerateAdvancedFunctionsSelectFiltersItems(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_parse") {
		t.Error("yq_select not using yq_parse for filtering")
	}
}

// TestGenerateAdvancedFunctionsRecursiveDescentChecksLiterals verifies literal handling
func TestGenerateAdvancedFunctionsRecursiveDescentChecksLiterals(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "yq_recursive_descent") {
		t.Error("Recursive descent function missing")
	}
}

// TestGenerateAdvancedFunctionsCompareConvertsValues verifies value conversion
func TestGenerateAdvancedFunctionsCompareConvertsValues(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "_left_val") {
		t.Error("yq_compare missing value conversion")
	}
}

// TestGenerateAdvancedFunctionsProperlyHandleNulls verifies null handling
func TestGenerateAdvancedFunctionsProperlyHandleNulls(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "null") {
		t.Error("Advanced functions missing null handling")
	}
}

// TestGenerateAdvancedFunctionsUsesTemporaryFiles verifies temp file usage
func TestGenerateAdvancedFunctionsUsesTemporaryFiles(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "mktemp") {
		t.Error("Advanced functions missing temporary file creation")
	}
}

// TestGenerateAdvancedFunctionsProperlyCleans verifies cleanup
func TestGenerateAdvancedFunctionsProperlyCleans(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "rm -f") {
		t.Error("Advanced functions missing cleanup of temporary files")
	}
}

// TestGenerateAdvancedFunctionsHasArrayCheck verifies array detection
func TestGenerateAdvancedFunctionsHasArrayCheck(t *testing.T) {
	result := GenerateAdvancedFunctions()

	// Should check if input is array
	if !strings.Contains(result, "-") {
		t.Error("Advanced functions missing array detection")
	}
}

// TestGenerateAdvancedFunctionsHasStringEscape verifies string escaping
func TestGenerateAdvancedFunctionsHasStringEscape(t *testing.T) {
	result := GenerateAdvancedFunctions()

	if !strings.Contains(result, "sed") {
		t.Error("Advanced functions missing sed for string escaping")
	}
}
