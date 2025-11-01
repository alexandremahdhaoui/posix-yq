package main

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

    # Check if expression starts with .[] by checking first 3 characters
    _sel_expr_start=$(echo "$_sel_expr" | cut -c 1-3)

    # Special handling for select with .[] in condition (e.g., select(.[] == "value"))
    # This should iterate and filter
    if [ "$_sel_expr_start" = ".[]" ] && echo "$_sel_expr" | grep -q ' == \| != \| < \| > \| <= \| >='; then
        # Extract the comparison operator and right side
        if echo "$_sel_expr" | grep -q ' == '; then
            _operator="=="
            _right_side=$(echo "$_sel_expr" | sed 's/.* == //')
        elif echo "$_sel_expr" | grep -q ' != '; then
            _operator="!="
            _right_side=$(echo "$_sel_expr" | sed 's/.* != //')
        else
            # Other operators not implemented yet
            return
        fi

        # Iterate through array and filter
        _tmp_iter=$(mktemp)
        yq_iterate "$_sel_file" > "$_tmp_iter"

        while IFS= read -r _item || [ -n "$_item" ]; do
            _tmp_item=$(mktemp)
            echo "$_item" > "$_tmp_item"

            # Compare this item with right side
            _result=$(yq_compare ". $_operator $_right_side" "$_tmp_item")
            if [ "$_result" = "true" ]; then
                printf "%s\n" "$_item"
            fi
            rm -f "$_tmp_item"
        done < "$_tmp_iter"
        rm -f "$_tmp_iter"
    # Handle select with pipe in condition (e.g., select(.[] | .key == value))
    elif [ "$_sel_expr_start" = ".[]" ] && echo "$_sel_expr" | grep -q ' | '; then
        # Extract parts: .[] | rest
        _rest=$(echo "$_sel_expr" | sed 's/^\.\[\] | //')

        # Iterate through array
        _tmp_iter=$(mktemp)
        yq_iterate "$_sel_file" > "$_tmp_iter"

        _first=1
        while IFS= read -r _item || [ -n "$_item" ]; do
            _tmp_item=$(mktemp)
            echo "$_item" > "$_tmp_item"

            # Evaluate the rest of the expression on this item
            _tmp_result=$(mktemp)
            yq_parse "$_rest" "$_tmp_item" > "$_tmp_result"

            # Check if result evaluates to true (non-empty, not false, not null)
            if [ -s "$_tmp_result" ]; then
                _result_value=$(cat "$_tmp_result")
                if [ "$_result_value" = "true" ] || [ -n "$_result_value" ] && [ "$_result_value" != "false" ] && [ "$_result_value" != "null" ]; then
                    # Item matches, output the whole item with proper formatting
                    if [ $_first -eq 1 ]; then
                        printf "%s" "$_item"
                        _first=0
                    else
                        printf "\n%s" "$_item"
                    fi
                fi
            fi

            rm -f "$_tmp_item" "$_tmp_result"
        done < "$_tmp_iter"
        rm -f "$_tmp_iter"
    # Standard select with comparison
    elif echo "$_sel_expr" | grep -q ' == \| != '; then
        _result=$(yq_compare "$_sel_expr" "$_sel_file")
        if [ "$_result" = "true" ]; then
            cat "$_sel_file"
        fi
    else
        # If no comparison, just check if result is not null/empty
        _result=$(yq_parse "$_sel_expr" "$_sel_file")
        if [ -n "$_result" ] && [ "$_result" != "null" ]; then
            echo "$_result"
        fi
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

    # Evaluate left side
    _left_value=$(yq_parse "$_left" "$_file")

    # Handle null comparisons specially
    if [ "$_right" = "null" ]; then
        _right_value="null"
        # Consider empty as null
        if [ -z "$_left_value" ]; then
            _left_value="null"
        fi
    else
        # Remove quotes from both sides for string comparison
        _left_value=$(echo "$_left_value" | sed 's/^"\(.*\)"$/\1/')
        _right_value=$(echo "$_right" | sed 's/^"\(.*\)"$/\1/')
    fi

    # Perform comparison
    case "$_operator" in
        "==")
            if [ "$_left_value" = "$_right_value" ]; then
                printf "true"
            else
                printf "false"
            fi
            ;;
        "!=")
            if [ "$_left_value" != "$_right_value" ]; then
                printf "true"
            else
                printf "false"
            fi
            ;;
    esac
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
