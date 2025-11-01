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

    # Evaluate the expression and check if any output is truthy
    # select() passes through the input if the expression produces any truthy value
    _tmp_result=$(mktemp)
    yq_parse "$_sel_expr" "$_sel_file" > "$_tmp_result" 2>/dev/null || true

    # Check if result contains any truthy values
    _has_truthy=0
    while IFS= read -r _line || [ -n "$_line" ]; do
        # Skip empty lines
        [ -z "$_line" ] && continue
        # Check if line is truthy (not false, not null, not empty)
        if [ "$_line" != "false" ] && [ "$_line" != "null" ]; then
            _has_truthy=1
            break
        fi
    done < "$_tmp_result"

    rm -f "$_tmp_result"

    # If we found any truthy value, output the original input
    if [ "$_has_truthy" -eq 1 ]; then
        cat "$_sel_file"
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
        }' "$_rdp_file" > "$_rdp_tmp_keys"

        # Read all keys into a space-separated list to avoid fd conflicts
        _rdp_keys_list=$(cat "$_rdp_tmp_keys" | tr '\n' ' ')
        rm -f "$_rdp_tmp_keys"

        # Process each key
        for _rdp_key in $_rdp_keys_list; do
            _rdp_tmp_child=$(mktemp)
            yq_key_access "$_rdp_key" "$_rdp_file" > "$_rdp_tmp_child"
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

    # Recursively descend into all child nodes
    # Check if it'\''s an object (has keys with colons)
    if grep -q '^[a-zA-Z_][a-zA-Z0-9_]*:' "$_rd_file" 2>/dev/null; then
        # For each TOP-LEVEL key in the object (indent = 0)
        _rd_tmp_keys=$(mktemp)
        awk '/^[a-zA-Z_][a-zA-Z0-9_]*:/ {
            match($0, /^[a-zA-Z_][a-zA-Z0-9_]*/)
            print substr($0, RSTART, RLENGTH)
        }' "$_rd_file" > "$_rd_tmp_keys"

        # Read all keys into a space-separated list to avoid fd conflicts
        _rd_keys_list=$(cat "$_rd_tmp_keys" | tr '\n' ' ')
        rm -f "$_rd_tmp_keys"

        # Process each key
        for _rd_key in $_rd_keys_list; do
            _rd_tmp_child=$(mktemp)
            yq_key_access "$_rd_key" "$_rd_file" > "$_rd_tmp_child"
            if [ -s "$_rd_tmp_child" ]; then
                yq_recursive_descent "$_rd_tmp_child"
            fi
            rm -f "$_rd_tmp_child"
        done
    # Check if it'\''s an array
    elif grep -q '^-' "$_rd_file" 2>/dev/null; then
        # For each element in the array
        _rd_tmp_arr=$(mktemp)
        grep '^-' "$_rd_file" > "$_rd_tmp_arr"

        while IFS= read -r _rd_line || [ -n "$_rd_line" ]; do
            _rd_value=$(echo "$_rd_line" | sed 's/^- //')
            _rd_tmp_elem=$(mktemp)
            echo "$_rd_value" > "$_rd_tmp_elem"
            yq_recursive_descent "$_rd_tmp_elem"
            rm -f "$_rd_tmp_elem"
        done < "$_rd_tmp_arr"
        rm -f "$_rd_tmp_arr"
    fi
}
`
}
