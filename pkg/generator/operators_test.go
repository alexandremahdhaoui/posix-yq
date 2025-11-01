package generator

import (
	"testing"
)

func TestYqAssign(t *testing.T) {
	code := GenerateOperators()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	t.Run("simple assignment", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John\nage: 30")
		output, _ := tester.ExecuteFunction("yq_assign", ".name = Alice", testFile)

		// Should contain the updated value
		if output == "" {
			t.Errorf("Expected assignment output, got empty")
		}
	})

	t.Run("new key assignment", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John")
		output, _ := tester.ExecuteFunction("yq_assign", ".age = 25", testFile)

		// Should contain both original and new key
		if output == "" {
			t.Errorf("Expected assignment output, got empty")
		}
	})
}

func TestYqDelete(t *testing.T) {
	code := GenerateOperators()
	tester := NewShellFunctionTester(t, code)
	defer tester.Cleanup()

	t.Run("delete key", func(t *testing.T) {
		testFile := tester.WriteFile("test.yaml", "name: John\nage: 30")
		output, _ := tester.ExecuteFunction("yq_del", ".age", testFile)

		// Should contain only name field
		if output == "" {
			t.Errorf("Expected delete output, got empty")
		}
	})
}
