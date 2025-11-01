// Copyright 2025 Alexandre Mahdhaoui
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package generator

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

# Convert JSON arrays to YAML format
# This allows JSON input to be processed by the YAML parser
_json_array_to_yaml() {
    _input="$1"

    # Check if it looks like a JSON array (starts with [)
    _trimmed=$(printf '%s' "$_input" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    if [ "${_trimmed#[}" != "$_trimmed" ]; then
        # It'"'"'s a JSON array - convert to YAML
        # Remove leading [ and trailing ], then convert to YAML array format
        printf '%s' "$_trimmed" | sed 's/^\[//;s/\]//' | awk '
        BEGIN { RS = "\"" }
        NF > 0 && !/^[[:space:]]*,?[[:space:]]*$/ && !/^\[/ {
            gsub(/^[[:space:]]*/, "")
            gsub(/[[:space:]]*$/, "")
            gsub(/^,/, "")
            gsub(/,/, "")
            if ($0 != "" && NR > 1) {
                printf "- %s\n", $0
            }
        }
        '
    else
        # Not a JSON array, return as-is
        printf '%s' "$_input"
    fi
}
`
}
