package generator

import "testing"

func TestRecursiveDescent(t *testing.T) {
	// Get the functions
	coreCode := GenerateCoreFunctions()
	advCode := GenerateAdvancedFunctions()
	
	// Combine them
	shellCode := coreCode + "\n" + advCode
	
	// Create test input
	input := `a:
  b:
    c: 1
d:
  c: 2`
	
	// Execute recursive descent
	tester := NewShellFunctionTester(t, shellCode)
	defer tester.Cleanup()
	
	inputFile := tester.WriteFile("input.yaml", input)
	result, _ := tester.ExecuteFunction("yq_recursive_descent", inputFile)
	
	t.Logf("Recursive descent output:\n%s", result)
	
	// Check that we got output from both a and d
	if !contains(result, "a:") {
		t.Errorf("Missing 'a:' key in output:\n%s", result)
	}
	if !contains(result, "d:") {
		t.Errorf("Missing 'd:' key in output:\n%s", result)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
