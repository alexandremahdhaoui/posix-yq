package generator

import (
	"strings"
	"testing"
)

// TestGenerateShellHeader verifies the shell header is generated correctly
func TestGenerateShellHeader(t *testing.T) {
	result := GenerateShellHeader()

	tests := []string{
		"_yq_parse_depth=0",
		"_yq_debug_indent()",
		"DEBUG",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateShellHeader missing '%s'", test)
		}
	}
}

// TestGenerateParser verifies the parser function is generated correctly
func TestGenerateParser(t *testing.T) {
	result := GenerateParser()

	tests := []string{
		"yq_parse()",
		"_query=",
		"yq_iterate",
		"yq_array_access",
		"yq_key_access",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateParser missing '%s'", test)
		}
	}
}

// TestGenerateCoreFunctions verifies core functions are generated
func TestGenerateCoreFunctions(t *testing.T) {
	result := GenerateCoreFunctions()

	tests := []string{
		"yq_unquote()",
		"yq_key_access()",
		"yq_iterate()",
		"yq_array_access()",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateCoreFunctions missing '%s'", test)
		}
	}
}

// TestGenerateAdvancedFunctions verifies advanced functions are generated
func TestGenerateAdvancedFunctions(t *testing.T) {
	result := GenerateAdvancedFunctions()

	tests := []string{
		"yq_map()",
		"yq_select()",
		"yq_compare()",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateAdvancedFunctions missing '%s'", test)
		}
	}
}

// TestGenerateOperators verifies operator functions are generated
func TestGenerateOperators(t *testing.T) {
	result := GenerateOperators()

	tests := []string{
		"yq_assign()",
		"yq_update()",
		"yq_del()",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateOperators missing '%s'", test)
		}
	}
}

// TestGenerateJSON verifies JSON conversion function is generated
func TestGenerateJSON(t *testing.T) {
	result := GenerateJSON()

	tests := []string{
		"yq_yaml_to_json()",
		"Convert YAML output to JSON",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateJSON missing '%s'", test)
		}
	}
}

// TestGenerateEntryPoint verifies the main entry point is generated
func TestGenerateEntryPoint(t *testing.T) {
	result := GenerateEntryPoint()

	tests := []string{
		"_exit_on_null=0",
		"_output_format=",
		"yq_parse",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("GenerateEntryPoint missing '%s'", test)
		}
	}
}

// TestConcatenation verifies all modules can be concatenated without syntax errors
func TestConcatenation(t *testing.T) {
	modules := []string{
		GenerateShellHeader(),
		GenerateParser(),
		GenerateCoreFunctions(),
		GenerateAdvancedFunctions(),
		GenerateOperators(),
		GenerateJSON(),
		GenerateEntryPoint(),
	}

	result := strings.Join(modules, "\n")

	// Check that concatenation produces reasonable output
	if len(result) < 10000 {
		t.Error("Concatenated script too short")
	}

	// Verify all main functions are present in concatenated output
	requiredFunctions := []string{
		"yq_parse",
		"yq_unquote",
		"yq_key_access",
		"yq_iterate",
		"yq_array_access",
		"yq_map",
		"yq_select",
		"yq_compare",
		"yq_assign",
		"yq_del",
		"yq_yaml_to_json",
	}

	for _, fn := range requiredFunctions {
		if !strings.Contains(result, fn+"()") {
			t.Errorf("Concatenated script missing function: %s", fn)
		}
	}
}

// TestOutputFormats verifies all functions return non-empty strings
func TestOutputFormats(t *testing.T) {
	tests := map[string]func() string{
		"GenerateShellHeader":       GenerateShellHeader,
		"GenerateParser":            GenerateParser,
		"GenerateCoreFunctions":     GenerateCoreFunctions,
		"GenerateAdvancedFunctions": GenerateAdvancedFunctions,
		"GenerateOperators":         GenerateOperators,
		"GenerateJSON":              GenerateJSON,
		"GenerateEntryPoint":        GenerateEntryPoint,
	}

	for name, fn := range tests {
		result := fn()
		if len(result) == 0 {
			t.Errorf("%s returned empty string", name)
		}

		// Verify output starts with shell comments or function definitions
		trimmed := strings.TrimSpace(result)
		if !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "_") {
			t.Errorf("%s output doesn't start with comment or variable", name)
		}
	}
}
