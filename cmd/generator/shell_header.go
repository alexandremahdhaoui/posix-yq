package main

// GenerateShellHeader returns the shell script header and debug utilities
func GenerateShellHeader() string {
	return `
# Main yq script - POSIX compliant implementation

# Initialize depth counter for debug output
_yq_parse_depth=0

# Helper function to indent debug output based on call depth
_yq_debug_indent() {
    _depth="$1"
    _msg="$2"
    _indent=""
    _i=0
    while [ $_i -lt $_depth ]; do
        _indent="$_indent  "
        _i=$((_i + 1))
    done
    >&2 echo "DEBUG[$_depth]$_indent$_msg"
}
`
}
