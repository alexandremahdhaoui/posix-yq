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

// GenerateAdvancedFunctions returns advanced manipulation functions
func GenerateAdvancedFunctions() string {
	return `
# Map function - apply expression to each array element
yq_map() {
    _expr="$1"
    _file="$2"

    # Create temporary file with grep results to avoid subshell issues
    _tmp_grep=$(mktemp)
    grep '^-' "$_file" > "$_tmp_grep"

    # Iterate over array elements
    _first=1
    while IFS= read -r _line || [ -n "$_line" ]; do
        _value=$(echo "$_line" | sed 's/^- //')

        # Apply expression to value
        _result=""

        # Handle simple arithmetic expressions like ". * 2"
        if echo "$_expr" | grep -q '^\. \* [0-9]*$'; then
            _multiplier=$(echo "$_expr" | sed 's/^\. \* //')
            _result=$((_value * _multiplier))
        # Handle simple addition like ". + 1"
        elif echo "$_expr" | grep -q '^\. + [0-9]*$'; then
            _addend=$(echo "$_expr" | sed 's/^\. + //')
            _result=$((_value + _addend))
        # Handle key access like ".name"
        elif echo "$_expr" | grep -q '^\.[a-zA-Z_]'; then
            _tmp_item=$(mktemp)
            echo "$_value" > "$_tmp_item"
            _result=$(yq_parse "$_expr" "$_tmp_item")
            rm -f "$_tmp_item"
        else
            _result="$_value"
        fi

        if [ "$_first" -eq 1 ]; then
            printf "%s" "- $_result"
            _first=0
        else
            printf "\n%s" "- $_result"
        fi
    done < "$_tmp_grep"

    # Add trailing newline if we printed anything
    if [ "$_first" -eq 0 ]; then
        printf "\n"
    fi

    rm -f "$_tmp_grep"
}

# Select function - filter elements based on condition
yq_select() {
    _sel_expr="$1"
    _sel_file="$2"

    [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG[select]: expr='$_sel_expr' file='$_sel_file'"

    # For now, a simple approximation:
    # if the condition contains "==", evaluate it and check for true
    # This is a simplified implementation for basic select() support

    # Evaluate the expression
    _sel_result=$(yq_parse "$_sel_expr" "$_sel_file" 2>/dev/null) || _sel_result=""

    [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG[select]: result='$_sel_result'"

    # Check if result contains any "true" value
    if printf '%s' "$_sel_result" | grep -q "true"; then
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG[select]: found 'true' in result, outputting input"
        # Output the input unchanged
        cat "$_sel_file"
    else
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG[select]: did not find 'true' in result"
    fi
}

# Comparison function - evaluate comparison expressions
yq_compare() {
    _expr="$1"
    _file="$2"

    # Determine operator
    if echo "$_expr" | grep -q ' == '; then
        _operator="=="
        _left=$(echo "$_expr" | sed 's/ ==.*//')
        _right=$(echo "$_expr" | sed 's/.* == //')
    elif echo "$_expr" | grep -q ' != '; then
        _operator="!="
        _left=$(echo "$_expr" | sed 's/ !=.*//')
        _right=$(echo "$_expr" | sed 's/.* != //')
    else
        printf "false"
        return
    fi

    # Evaluate left side (may produce multiple values)
    _tmp_left=$(mktemp)
    yq_parse "$_left" "$_file" > "$_tmp_left" 2>/dev/null || true

    # Process null comparisons
    if [ "$_right" = "null" ]; then
        _right_value="null"
    else
        # Remove quotes from right side for string comparison
        _right_value=$(echo "$_right" | sed 's/^"\(.*\)"$/\1/')
    fi

    # Compare each left value with right value
    _first=1
    while IFS= read -r _left_value || [ -n "$_left_value" ]; do
        # Skip empty lines
        [ -z "$_left_value" ] && continue

        # Handle null comparisons
        if [ "$_right" = "null" ]; then
            if [ -z "$_left_value" ]; then
                _left_value="null"
            fi
        else
            # Remove quotes from left side
            _left_value=$(echo "$_left_value" | sed 's/^"\(.*\)"$/\1/')
        fi

        # Perform comparison
        case "$_operator" in
            "==")
                if [ "$_left_value" = "$_right_value" ]; then
                    if [ $_first -eq 1 ]; then
                        printf "true"
                        _first=0
                    else
                        printf "\ntrue"
                    fi
                else
                    if [ $_first -eq 1 ]; then
                        printf "false"
                        _first=0
                    else
                        printf "\nfalse"
                    fi
                fi
                ;;
            "!=")
                if [ "$_left_value" != "$_right_value" ]; then
                    if [ $_first -eq 1 ]; then
                        printf "true"
                        _first=0
                    else
                        printf "\ntrue"
                    fi
                else
                    if [ $_first -eq 1 ]; then
                        printf "false"
                        _first=0
                    else
                        printf "\nfalse"
                    fi
                fi
                ;;
        esac
    done < "$_tmp_left"

    rm -f "$_tmp_left"
}

# Recursive descent with pipe - apply expression to each node
yq_recursive_descent_pipe() {
    _rdp_file="$1"
    _rdp_pipe_expr="$2"

    # Process current node through pipe
    _rdp_tmp_node=$(mktemp)
    cat "$_rdp_file" > "$_rdp_tmp_node"
    yq_parse "$_rdp_pipe_expr" "$_rdp_tmp_node"
    rm -f "$_rdp_tmp_node"

    # Recursively descend into all child nodes
    # Check if it'\''s an object (has keys with colons)
    if grep -q '^[a-zA-Z_][a-zA-Z0-9_]*:' "$_rdp_file" 2>/dev/null; then
        # For each TOP-LEVEL key in the object (indent = 0)
        _rdp_tmp_keys=$(mktemp)
        awk '/^[a-zA-Z_][a-zA-Z0-9_]*:/ {
            match($0, /^[a-zA-Z_][a-zA-Z0-9_]*/)
            print substr($0, RSTART, RLENGTH)
        }' "$_rdp_file" > "$_rdp_tmp_keys" 2>/dev/null

        # Read all keys into a space-separated list to avoid fd conflicts
        _rdp_keys_list=$(cat "$_rdp_tmp_keys" | tr '\n' ' ')
        rm -f "$_rdp_tmp_keys"

        # Process each key
        for _rdp_key in $_rdp_keys_list; do
            _rdp_tmp_child=$(mktemp)
            yq_key_access "$_rdp_key" "$_rdp_file" > "$_rdp_tmp_child" 2>/dev/null
            if [ -s "$_rdp_tmp_child" ]; then
                yq_recursive_descent_pipe "$_rdp_tmp_child" "$_rdp_pipe_expr"
            fi
            rm -f "$_rdp_tmp_child"
        done
    # Check if it'\''s an array
    elif grep -q '^-' "$_rdp_file" 2>/dev/null; then
        # For each element in the array
        _rdp_tmp_arr=$(mktemp)
        grep '^-' "$_rdp_file" > "$_rdp_tmp_arr"

        while IFS= read -r _rdp_line || [ -n "$_rdp_line" ]; do
            _rdp_value=$(echo "$_rdp_line" | sed 's/^- //')
            _rdp_tmp_elem=$(mktemp)
            echo "$_rdp_value" > "$_rdp_tmp_elem"
            yq_recursive_descent_pipe "$_rdp_tmp_elem" "$_rdp_pipe_expr"
            rm -f "$_rdp_tmp_elem"
        done < "$_rdp_tmp_arr"
        rm -f "$_rdp_tmp_arr"
    fi
}

# Recursive descent - output all nodes in tree
yq_recursive_descent() {
    _rd_file="$1"

    # Output the current node
    cat "$_rd_file"

    # Check if current node is an object (has top-level keys)
    _rd_keys=$(grep -E '^[a-zA-Z_][a-zA-Z0-9_]*:' "$_rd_file" 2>/dev/null | sed 's/:.*$//')

    if [ -n "$_rd_keys" ]; then
        # Write keys to a temp file to process without pipe (avoids subshell issues)
        _rd_keys_file=$(mktemp)
        printf "%s\n" "$_rd_keys" > "$_rd_keys_file"

        # Process each key using file redirection instead of pipe
        exec 3< "$_rd_keys_file"
        while IFS= read -r _rd_key <&3 || [ -n "$_rd_key" ]; do
            [ -z "$_rd_key" ] && continue

            _rd_tmp=$(mktemp)
            yq_key_access "$_rd_key" "$_rd_file" > "$_rd_tmp" 2>/dev/null

            if [ -s "$_rd_tmp" ]; then
                yq_recursive_descent "$_rd_tmp"
            fi

            rm -f "$_rd_tmp"
        done
        exec 3<&-

        rm -f "$_rd_keys_file"
    elif grep -q '^-' "$_rd_file" 2>/dev/null; then
        # Current node is an array - process each element
        _rd_arr_file=$(mktemp)
        grep '^-' "$_rd_file" > "$_rd_arr_file"

        exec 4< "$_rd_arr_file"
        while IFS= read -r _rd_line <&4 || [ -n "$_rd_line" ]; do
            [ -z "$_rd_line" ] && continue

            _rd_value=$(printf "%s" "$_rd_line" | sed 's/^- //')
            _rd_tmp=$(mktemp)
            printf "%s" "$_rd_value" > "$_rd_tmp"

            if [ -s "$_rd_tmp" ]; then
                yq_recursive_descent "$_rd_tmp"
            fi

            rm -f "$_rd_tmp"
        done
        exec 4<&-

        rm -f "$_rd_arr_file"
    fi
}

# Arithmetic operations - handles +, -, *, / operators
yq_arithmetic() {
    _expr="$1"
    _file="$2"

    # Determine operator
    _operator=""
    _left=""
    _right=""

    if echo "$_expr" | grep -q ' + '; then
        _operator="+"
        _left=$(echo "$_expr" | sed 's/ +.*//')
        _right=$(echo "$_expr" | sed 's/.* + //')
    elif echo "$_expr" | grep -q ' - '; then
        _operator="-"
        _left=$(echo "$_expr" | sed 's/ -.*//')
        _right=$(echo "$_expr" | sed 's/.* - //')
    elif echo "$_expr" | grep -q ' \* '; then
        _operator="*"
        _left=$(echo "$_expr" | sed 's/ \*.*//')
        _right=$(echo "$_expr" | sed 's/.* \* //')
    elif echo "$_expr" | grep -q ' / '; then
        _operator="/"
        _left=$(echo "$_expr" | sed 's/ \/.*//')
        _right=$(echo "$_expr" | sed 's/.* \/ //')
    else
        echo "null"
        return
    fi

    # Evaluate both sides
    _tmp_left=$(mktemp)
    _tmp_right=$(mktemp)

    yq_parse "$_left" "$_file" > "$_tmp_left" 2>/dev/null || true

    # For right side, check if it's a literal number or expression
    if echo "$_right" | grep -q '^[0-9]\+$'; then
        echo "$_right" > "$_tmp_right"
    else
        yq_parse "$_right" "$_file" > "$_tmp_right" 2>/dev/null || true
    fi

    # Read left and right values
    _left_val=$(cat "$_tmp_left" 2>/dev/null)
    _right_val=$(cat "$_tmp_right" 2>/dev/null)

    # Check if both are numeric
    if echo "$_left_val" | grep -q '^-\?[0-9]\+\(\\.[0-9]\+\)\?$' && \
       echo "$_right_val" | grep -q '^-\?[0-9]\+\(\\.[0-9]\+\)\?$'; then
        # Numeric operation - use awk to avoid POSIX shell arithmetic issues
        case "$_operator" in
            "+")
                echo "$_left_val" | awk -v r="$_right_val" '{printf "%d", $1 + r}'
                ;;
            "-")
                echo "$_left_val" | awk -v r="$_right_val" '{printf "%d", $1 - r}'
                ;;
            "*")
                echo "$_left_val" | awk -v r="$_right_val" '{printf "%d", $1 * r}'
                ;;
            "/")
                echo "$_left_val" | awk -v r="$_right_val" '{printf "%d", int($1 / r)}'
                ;;
        esac
    else
        # String concatenation (only for +)
        if [ "$_operator" = "+" ]; then
            printf '%s%s' "$_left_val" "$_right_val"
        else
            echo "null"
        fi
    fi

    rm -f "$_tmp_left" "$_tmp_right"
}
`
}
