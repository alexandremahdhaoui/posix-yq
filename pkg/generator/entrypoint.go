package generator

// GenerateEntryPoint returns the main entry point script
func GenerateEntryPoint() string {
	return `
# Main entry point
_exit_on_null=0
_output_format="yaml"
_raw_output=0
_indent_level=2

# Skip yq subcommand if present (e.g., "yq e -o=j" has 'e' as subcommand)
if [ "$1" = "e" ] || [ "$1" = "eval" ] || [ "$1" = "select" ] || [ "$1" = "empty" ]; then
    shift
fi

# Parse flags
while [ $# -gt 0 ]; do
    case "$1" in
        -e)
            _exit_on_null=1
            shift
            ;;
        -r|--raw-output)
            _raw_output=1
            shift
            ;;
        -o|--output)
            _output_format="$2"
            shift 2
            ;;
        -o=*)
            _output_format="${1#-o=}"
            shift
            ;;
        -I|--indent)
            _indent_level="$2"
            shift 2
            ;;
        -I=*)
            _indent_level="${1#-I=}"
            shift
            ;;
        -j|--json)
            _output_format="json"
            shift
            ;;
        --raw-input)
            # Placeholder for raw input mode
            shift
            ;;
        -*)
            # Unknown flag, but treat as query for backwards compatibility
            break
            ;;
        *)
            break
            ;;
    esac
done

# First positional argument is the query
QUERY="$1"
FILE="$2"

# If no file provided, read from stdin
if [ -z "$FILE" ]; then
    # Check if stdin has content
    if [ ! -t 0 ]; then
        # stdin is available
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: Reading from stdin"
        FILE=$(mktemp)
        _stdin_content=$(cat)
        _converted_content=$(_json_array_to_yaml "$_stdin_content")
        printf '%s' "$_converted_content" > "$FILE"
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: Stdin written to $FILE"
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: File size: $(wc -c < $FILE)"
        _cleanup_file="$FILE"
    elif [ -z "$QUERY" ]; then
        # No query and no file - error
        >&2 echo "Error: No input file and no query provided"
        exit 1
    else
        # Only query provided, assume it'\''s actually the file
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: Only query provided, using as file"
        FILE="$QUERY"
        QUERY="."
    fi
fi

# Execute the query
_result=$(yq_parse "$QUERY" "$FILE")
_exit_code=$?

# Cleanup temporary file if created
if [ -n "$_cleanup_file" ]; then
    rm -f "$_cleanup_file"
fi

# Handle -r (raw output) flag
if [ $_raw_output -eq 1 ]; then
    # Raw output mode - output values without quotes
    _result=$(yq_unquote "$_result")
else
    # Normal mode - unquote only for simple string values
    # For structured data (arrays/objects), keep YAML formatting
    _result=$(yq_unquote "$_result")
fi

# Handle JSON output format (must be before cleanup to preserve object separators)
if [ "$_output_format" = "json" ] || [ "$_output_format" = "j" ]; then
    # Convert YAML output to JSON
    # The function handles grouping of multi-line blocks separated by blank lines
    _result=$(yq_yaml_to_json "$_result")
else
    # Clean up result: remove blank line separators from array iteration
    # The iteration uses blank lines as separators, but we only want actual content
    while printf '%s' "$_result" | grep -q '^[[:space:]]*$'; do
        _result=$(printf '%s\n' "$_result" | grep -v '^[[:space:]]*$')
    done
fi

# Output result (preserve newlines from multiline results)
printf '%s\n' "$_result"

# Handle -e flag: exit with code 5 if result is empty or null
# MUST output the result BEFORE checking exit condition
if [ $_exit_on_null -eq 1 ]; then
    if [ -z "$_result" ] || [ "$_result" = "null" ]; then
        exit 5
    fi
fi

exit $_exit_code
`
}
