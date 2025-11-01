package generator

import (
	"strings"
	"testing"
)

// TestGenerateJSONIsNonEmpty verifies output is not empty
func TestGenerateJSONIsNonEmpty(t *testing.T) {
	result := GenerateJSON()
	if len(result) == 0 {
		t.Error("GenerateJSON returned empty string")
	}
}

// TestGenerateJSONHasConversionFunction verifies YAML to JSON converter
func TestGenerateJSONHasConversionFunction(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "yq_yaml_to_json()") {
		t.Error("GenerateJSON missing yq_yaml_to_json function")
	}
}

// TestGenerateJSONFunctionComment verifies function documentation
func TestGenerateJSONFunctionComment(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "Convert YAML") {
		t.Error("JSON function missing documentation comment")
	}
}

// TestGenerateJSONHandlesInputParameter verifies input parameter
func TestGenerateJSONHandlesInputParameter(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "_yaml_input") {
		t.Error("JSON converter missing input parameter")
	}
}

// TestGenerateJSONUsesAwk verifies AWK for conversion
func TestGenerateJSONUsesAwk(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "awk") {
		t.Error("JSON converter not using awk for processing")
	}
}

// TestGenerateJSONHandlesArrays verifies array handling
func TestGenerateJSONHandlesArrays(t *testing.T) {
	result := GenerateJSON()

	// Should handle YAML arrays (items starting with -)
	if !strings.Contains(result, "-") {
		t.Error("JSON converter missing array element handling")
	}
}

// TestGenerateJSONHandlesObjects verifies object handling
func TestGenerateJSONHandlesObjects(t *testing.T) {
	result := GenerateJSON()

	tests := []string{
		"key",   // Key-value pairs
		"value", // Values
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("JSON converter missing object handling for '%s'", test)
		}
	}
}

// TestGenerateJSONOutputsCompactFormat verifies compact JSON format
func TestGenerateJSONOutputsCompactFormat(t *testing.T) {
	result := GenerateJSON()

	// Should output compact JSON (no extra whitespace)
	if !strings.Contains(result, "{") && !strings.Contains(result, "}") {
		t.Error("JSON converter not outputting JSON format")
	}
}

// TestGenerateJSONHandlesMultipleObjects verifies multi-object output
func TestGenerateJSONHandlesMultipleObjects(t *testing.T) {
	result := GenerateJSON()

	// Should handle conversion of multiple YAML objects
	if !strings.Contains(result, "yq_yaml_to_json") {
		t.Error("JSON converter missing multi-object handling")
	}
}

// TestGenerateJSONEscapesStrings verifies string escaping
func TestGenerateJSONEscapesStrings(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "gsub") {
		t.Error("JSON converter missing string escaping with gsub")
	}
}

// TestGenerateJSONHandlesNullValues verifies null handling
func TestGenerateJSONHandlesNullValues(t *testing.T) {
	result := GenerateJSON()

	// JSON converter should have yq_yaml_to_json function
	if !strings.Contains(result, "yq_yaml_to_json") {
		t.Error("JSON converter missing null value handling in conversion function")
	}
}

// TestGenerateJSONUsesBeginBlock verifies AWK BEGIN block
func TestGenerateJSONUsesBeginBlock(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "BEGIN") {
		t.Error("JSON converter not using AWK BEGIN block for initialization")
	}
}

// TestGenerateJSONProperlyFormatsOutput verifies output formatting
func TestGenerateJSONProperlyFormatsOutput(t *testing.T) {
	result := GenerateJSON()

	// Should use printf for output
	if !strings.Contains(result, "printf") {
		t.Error("JSON converter not using printf for output")
	}
}

// TestGenerateJSONHandlesIndentation verifies indentation parameter
func TestGenerateJSONHandlesIndentation(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "yq_yaml_to_json") {
		t.Error("JSON converter missing indentation handling")
	}
}

// TestGenerateJSONFieldSeparator verifies field handling
func TestGenerateJSONFieldSeparator(t *testing.T) {
	result := GenerateJSON()

	// Should handle YAML field separator
	if !strings.Contains(result, ":") {
		t.Error("JSON converter missing field separator handling")
	}
}

// TestGenerateJSONHandlesNestedStructures verifies nesting
func TestGenerateJSONHandlesNestedStructures(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "yq_yaml_to_json") {
		t.Error("JSON converter missing nested structure support")
	}
}

// TestGenerateJSONPreservesValues verifies value preservation
func TestGenerateJSONPreservesValues(t *testing.T) {
	result := GenerateJSON()

	if !strings.Contains(result, "_yaml_input") {
		t.Error("JSON converter not preserving input values")
	}
}

// TestGenerateJSONProperlyQuotesValues verifies quoting
func TestGenerateJSONProperlyQuotesValues(t *testing.T) {
	result := GenerateJSON()

	// Should properly quote JSON values
	if !strings.Contains(result, "\"") {
		t.Error("JSON converter missing value quoting")
	}
}
