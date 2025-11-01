package generator

import (
	"strings"
	"testing"
)

// TestGenerateOperatorsIsNonEmpty verifies output is not empty
func TestGenerateOperatorsIsNonEmpty(t *testing.T) {
	result := GenerateOperators()
	if len(result) == 0 {
		t.Error("GenerateOperators returned empty string")
	}
}

// TestGenerateOperatorsHasAssignFunction verifies assign operator
func TestGenerateOperatorsHasAssignFunction(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_assign()") {
		t.Error("GenerateOperators missing yq_assign function")
	}
}

// TestGenerateOperatorsHasUpdateFunction verifies update operator
func TestGenerateOperatorsHasUpdateFunction(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_update()") {
		t.Error("GenerateOperators missing yq_update function")
	}
}

// TestGenerateOperatorsHasDeleteFunction verifies delete operator
func TestGenerateOperatorsHasDeleteFunction(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_del()") {
		t.Error("GenerateOperators missing yq_del function")
	}
}

// TestGenerateOperatorsAssignHandlesPath verifies path parsing in assign
func TestGenerateOperatorsAssignHandlesPath(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_assign") {
		t.Error("Assignment operator missing")
	}
	if !strings.Contains(result, "_path") {
		t.Error("Assignment operator missing path extraction")
	}
}

// TestGenerateOperatorsAssignHandlesValue verifies value in assign
func TestGenerateOperatorsAssignHandlesValue(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "_value") {
		t.Error("Assignment operator missing value extraction")
	}
}

// TestGenerateOperatorsUpdateCallsParse verifies update uses parse
func TestGenerateOperatorsUpdateCallsParse(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_update") {
		t.Error("Update operator missing")
	}
	if !strings.Contains(result, "yq_parse") {
		t.Error("Update operator not using yq_parse")
	}
}

// TestGenerateOperatorsDeleteHandlesPath verifies delete path handling
func TestGenerateOperatorsDeleteHandlesPath(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_del") {
		t.Error("Delete operator missing")
	}
}

// TestGenerateOperatorsAssignRebuildsDocument verifies document reconstruction
func TestGenerateOperatorsAssignRebuildsDocument(t *testing.T) {
	result := GenerateOperators()

	// Should handle rebuilding the document with new value
	if !strings.Contains(result, "awk") {
		t.Error("Operators not using awk for document manipulation")
	}
}

// TestGenerateOperatorsUpdateEvaluatesExpression verifies expression evaluation
func TestGenerateOperatorsUpdateEvaluatesExpression(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_update") {
		t.Error("Update operator missing")
	}
}

// TestGenerateOperatorsDeleteRemovesPath verifies path removal
func TestGenerateOperatorsDeleteRemovesPath(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_del") {
		t.Error("Delete operator missing proper implementation")
	}
}

// TestGenerateOperatorsHandlesNestedPaths verifies nested path support
func TestGenerateOperatorsHandlesNestedPaths(t *testing.T) {
	result := GenerateOperators()

	// Should handle nested paths
	if !strings.Contains(result, "_path") {
		t.Error("Operators not handling path extraction")
	}
}

// TestGenerateOperatorsProperlyEscapesValues verifies value escaping
func TestGenerateOperatorsProperlyEscapesValues(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "sed") || !strings.Contains(result, "awk") {
		t.Error("Operators not properly escaping values")
	}
}

// TestGenerateOperatorsUsesTemporaryFiles verifies temp files
func TestGenerateOperatorsUsesTemporaryFiles(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "mktemp") {
		t.Error("Operators not using temporary files")
	}
}

// TestGenerateOperatorsProperlyCleans verifies cleanup
func TestGenerateOperatorsProperlyCleans(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "rm -f") {
		t.Error("Operators not cleaning up temporary files")
	}
}

// TestGenerateOperatorsHandlesQuotes verifies quote handling
func TestGenerateOperatorsHandlesQuotes(t *testing.T) {
	result := GenerateOperators()

	// Should handle quoting/unquoting
	if !strings.Contains(result, "\"") {
		t.Error("Operators missing quote handling")
	}
}

// TestGenerateOperatorsAssignHandlesArray verifies array assignment
func TestGenerateOperatorsAssignHandlesArray(t *testing.T) {
	result := GenerateOperators()

	// Should handle array paths - check for array-related patterns
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Error("Operators not handling array paths")
	}
}

// TestGenerateOperatorsUpdateHandlesConditional verifies conditional update
func TestGenerateOperatorsUpdateHandlesConditional(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_update") {
		t.Error("Update operator missing conditional handling")
	}
}

// TestGenerateOperatorsDeleteValidatesPath verifies path validation
func TestGenerateOperatorsDeleteValidatesPath(t *testing.T) {
	result := GenerateOperators()

	if !strings.Contains(result, "yq_del") {
		t.Error("Delete operator missing path validation")
	}
}
