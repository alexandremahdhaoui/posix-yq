package generator

import (
	"strings"
	"testing"
)

// TestGenerateEntryPointIsNonEmpty verifies output is not empty
func TestGenerateEntryPointIsNonEmpty(t *testing.T) {
	result := GenerateEntryPoint()
	if len(result) == 0 {
		t.Error("GenerateEntryPoint returned empty string")
	}
}

// TestGenerateEntryPointInitializesVariables verifies variable initialization
func TestGenerateEntryPointInitializesVariables(t *testing.T) {
	result := GenerateEntryPoint()

	tests := []string{
		"_exit_on_null=0",      // Exit on null flag
		"_output_format=",      // Output format
		"_raw_output=0",        // Raw output flag
		"_indent_level=",       // Indentation level
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("EntryPoint initialization missing '%s'", test)
		}
	}
}

// TestGenerateEntryPointHandlesSubcommands verifies subcommand handling
func TestGenerateEntryPointHandlesSubcommands(t *testing.T) {
	result := GenerateEntryPoint()

	tests := []string{
		"\"e\"",  // eval subcommand
		"shift",  // Remove subcommand
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("EntryPoint subcommand handling missing '%s'", test)
		}
	}
}

// TestGenerateEntryPointParsesFlagsExit verifies -e flag parsing
func TestGenerateEntryPointParsesFlagsExit(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "-e") {
		t.Error("EntryPoint missing -e flag support")
	}
	if !strings.Contains(result, "_exit_on_null") {
		t.Error("EntryPoint missing exit on null handling")
	}
}

// TestGenerateEntryPointParsesFlagsRaw verifies -r flag parsing
func TestGenerateEntryPointParsesFlagsRaw(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "-r") || !strings.Contains(result, "--raw-output") {
		t.Error("EntryPoint missing -r/--raw-output flag support")
	}
	if !strings.Contains(result, "_raw_output") {
		t.Error("EntryPoint missing raw output flag handling")
	}
}

// TestGenerateEntryPointParsesFlagsOutput verifies -o flag parsing
func TestGenerateEntryPointParsesFlagsOutput(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "-o=") || !strings.Contains(result, "--output") {
		t.Error("EntryPoint missing -o/--output flag support")
	}
	if !strings.Contains(result, "_output_format") {
		t.Error("EntryPoint missing output format handling")
	}
}

// TestGenerateEntryPointParsesFlagsIndent verifies -I flag parsing
func TestGenerateEntryPointParsesFlagsIndent(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "-I") {
		t.Error("EntryPoint missing -I flag support")
	}
	if !strings.Contains(result, "_indent_level") {
		t.Error("EntryPoint missing indentation level handling")
	}
}

// TestGenerateEntryPointParsesFlagsJSON verifies -j flag parsing
func TestGenerateEntryPointParsesFlagsJSON(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "-j") {
		t.Error("EntryPoint missing -j flag support")
	}
}

// TestGenerateEntryPointHandlesStdin verifies stdin detection
func TestGenerateEntryPointHandlesStdin(t *testing.T) {
	result := GenerateEntryPoint()

	tests := []string{
		"[ ! -t 0 ]",  // stdin check
		"mktemp",      // Create temp file for stdin
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("EntryPoint stdin handling missing '%s'", test)
		}
	}
}

// TestGenerateEntryPointHandlesFileArgument verifies file argument handling
func TestGenerateEntryPointHandlesFileArgument(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "$2") {
		t.Error("EntryPoint not handling file argument")
	}
}

// TestGenerateEntryPointCallsParser verifies parser invocation
func TestGenerateEntryPointCallsParser(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "yq_parse") {
		t.Error("EntryPoint not calling yq_parse")
	}
	if !strings.Contains(result, "_result=") {
		t.Error("EntryPoint not storing parser result")
	}
}

// TestGenerateEntryPointHandlesRawOutput verifies raw output processing
func TestGenerateEntryPointHandlesRawOutput(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "yq_unquote") {
		t.Error("EntryPoint not calling yq_unquote for raw output")
	}
}

// TestGenerateEntryPointHandlesJSONOutput verifies JSON output processing
func TestGenerateEntryPointHandlesJSONOutput(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "yq_yaml_to_json") {
		t.Error("EntryPoint not handling JSON output conversion")
	}
}

// TestGenerateEntryPointPrintsResult verifies output
func TestGenerateEntryPointPrintsResult(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "printf") {
		t.Error("EntryPoint not printing result")
	}
}

// TestGenerateEntryPointHandlesExitCode verifies exit code handling
func TestGenerateEntryPointHandlesExitCode(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "exit") {
		t.Error("EntryPoint not handling exit codes")
	}
}

// TestGenerateEntryPointCleansUpTempFiles verifies cleanup
func TestGenerateEntryPointCleansUpTempFiles(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "rm -f") {
		t.Error("EntryPoint not cleaning up temporary files")
	}
}

// TestGenerateEntryPointHandlesNullCheck verifies null value checking
func TestGenerateEntryPointHandlesNullCheck(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "null") {
		t.Error("EntryPoint missing null value check for -e flag")
	}
}

// TestGenerateEntryPointCaseSwitchForFlags verifies flag parsing structure
func TestGenerateEntryPointCaseSwitchForFlags(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "case") {
		t.Error("EntryPoint not using case statement for flag parsing")
	}
}

// TestGenerateEntryPointWhileLoopForParsing verifies parsing loop
func TestGenerateEntryPointWhileLoopForParsing(t *testing.T) {
	result := GenerateEntryPoint()

	if !strings.Contains(result, "while") {
		t.Error("EntryPoint not using while loop for flag parsing")
	}
}
