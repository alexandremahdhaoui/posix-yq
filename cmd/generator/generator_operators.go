package main

// GenerateOperators returns assignment and mutation operators
func GenerateOperators() string {
	return `
# Assignment operator - set a value
yq_assign() {
    _expr="$1"
    _file="$2"

    # Parse the assignment: .path = value
    _path=$(echo "$_expr" | sed 's/ =.*//')
    _value=$(echo "$_expr" | sed 's/.* = //')

    # Remove leading dot from path
    _path=$(echo "$_path" | sed 's/^\.//')

    # Remove quotes from value if it'\''s a string literal
    _value=$(echo "$_value" | sed 's/^"\(.*\)"$/\1/')

    # Update the file with the new value
    awk -v path="$_path" -v value="$_value" '
    BEGIN {
        found = 0
    }
    {
        # Check if this line matches the key
        if ($0 ~ "^" path ":") {
            # Replace the value
            print path ": " value
            found = 1
        } else {
            print
        }
    }
    END {
        # If key was not found, add it
        if (!found) {
            print path ": " value
        }
    }
    ' "$_file"
}

# Update operator - update a value based on expression
yq_update() {
    _expr="$1"
    _file="$2"

    # Parse the update: .path |= expression
    _path=$(echo "$_expr" | sed 's/ |=.*//')
    _update_expr=$(echo "$_expr" | sed 's/.* |= //')

    # Get current value
    _tmp_current=$(mktemp)
    yq_parse "$_path" "$_file" > "$_tmp_current"

    # Apply the update expression to the current value
    _tmp_updated=$(mktemp)
    yq_parse "$_update_expr" "$_tmp_current" > "$_tmp_updated"
    _new_value=$(cat "$_tmp_updated")

    # Create the assignment expression and apply it
    _assign_expr="$_path = $_new_value"
    yq_assign "$_assign_expr" "$_file"

    rm -f "$_tmp_current" "$_tmp_updated"
}

# Delete function - remove a key from object
yq_del() {
    _path="$1"
    _file="$2"

    # Remove leading dot
    _path=$(echo "$_path" | sed 's/^\.//')

    # Check if nested path
    if echo "$_path" | grep -q '\.'; then
        # Nested key deletion using AWK with path tracking
        # For del(.a.b), we split into parts and track our position in the hierarchy
        _path_parts=$(echo "$_path" | tr '.' '\n')
        _parts_count=$(echo "$_path_parts" | wc -l | tr -d ' ')

        # Use AWK to track path and delete the target
        awk -v path="$_path" -v parts_count="$_parts_count" '
        BEGIN {
            split(path, parts, "\\.")
            skip = 0
            del_indent = -1
            first = 1

            # Track current path depth
            for (i = 1; i <= 100; i++) {
                path_stack[i] = ""
                indent_stack[i] = -1
            }
            stack_depth = 0
        }
        {
            # Calculate indentation
            current_indent = 0
            for (i = 1; i <= length($0); i++) {
                if (substr($0, i, 1) == " ") {
                    current_indent++
                } else {
                    break
                }
            }

            # Get key from current line
            current_key = ""
            if ($0 ~ /^[[:space:]]*[a-zA-Z_][a-zA-Z0-9_]*:/) {
                match($0, /[a-zA-Z_][a-zA-Z0-9_]*/)
                current_key = substr($0, RSTART, RLENGTH)
            }

            # Update stack based on indentation
            if (current_key != "") {
                # Pop stack until we find parent level
                while (stack_depth > 0 && indent_stack[stack_depth] >= current_indent) {
                    stack_depth--
                }

                # Push current key
                stack_depth++
                path_stack[stack_depth] = current_key
                indent_stack[stack_depth] = current_indent

                # Check if current path matches deletion path
                matches = 1
                if (stack_depth == parts_count) {
                    for (i = 1; i <= parts_count; i++) {
                        if (path_stack[i] != parts[i]) {
                            matches = 0
                            break
                        }
                    }
                    if (matches) {
                        skip = 1
                        del_indent = current_indent
                        next
                    }
                }
            }

            # If skipping, check if still in deleted block
            if (skip) {
                if (current_indent > del_indent || $0 ~ /^[[:space:]]*$/) {
                    next
                } else {
                    skip = 0
                }
            }

            # Print the line
            if (first) {
                printf "%s", $0
                first = 0
            } else {
                printf "\n%s", $0
            }
        }
        ' "$_file"
    else
        # Single key deletion
        awk -v key="$_path" '
        BEGIN {
            skip = 0
            key_indent = -1
            first = 1
        }
        {
            # Calculate indentation
            current_indent = 0
            for (i = 1; i <= length($0); i++) {
                if (substr($0, i, 1) == " ") {
                    current_indent++
                } else {
                    break
                }
            }

            # Check if this is the key to delete
            if (!skip && $0 ~ "^[[:space:]]*" key ":") {
                skip = 1
                key_indent = current_indent
                next
            }

            # If skipping, check if we are still in the key block
            if (skip) {
                if (current_indent > key_indent || $0 ~ /^[[:space:]]*$/) {
                    # Still in the deleted key block, skip
                    next
                } else {
                    # Left the block, stop skipping
                    skip = 0
                }
            }

            # Print the line without trailing newline on last line
            if (first) {
                printf "%s", $0
                first = 0
            } else {
                printf "\n%s", $0
            }
        }
        ' "$_file"
    fi
}
`
}
