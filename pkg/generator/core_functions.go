package generator

// GenerateCoreFunctions returns core YAML manipulation functions
func GenerateCoreFunctions() string {
	return `
# Unquote YAML strings (but not null values)
yq_unquote() {
    _value="$1"
    # Don't unquote null or other special values
    if [ "$_value" = "null" ] || [ "$_value" = "true" ] || [ "$_value" = "false" ]; then
        echo "$_value"
        return
    fi
    # Check if the value is a quoted string
    if [ "${_value#\"}" != "$_value" ]; then
        # Remove double quotes if they surround the entire value
        _trimmed="${_value%\"}"
        _trimmed="${_trimmed#\"}"
        # Make sure we actually had quotes at both ends
        if [ "${#_value}" -gt 2 ] && [ "${_value%\"}" != "$_value" ]; then
            _value="$_trimmed"
            # Unescape special characters
            _value=$(echo "$_value" | sed 's/\\"/"/g; s/\\\\/\\/g')
        fi
    elif [ "${_value#\'}" != "$_value" ]; then
        # Remove single quotes if they surround the entire value
        _trimmed="${_value%\'}"
        _trimmed="${_trimmed#\'}"
        if [ "${#_value}" -gt 2 ] && [ "${_value%\'}" != "$_value" ]; then
            _value="$_trimmed"
        fi
    fi
    echo "$_value"
}

# Extract value for a key
yq_key_access() {
    _key="$1"
    _file="$2"

    awk -v key="$_key" '
    BEGIN {
        found = 0
        key_indent = -1
        block_indent = -1
        in_block = 0
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

        if (found && in_block) {
            # We are printing the block
            if (block_indent == -1) {
                # First line after the key
                if (current_indent > key_indent || $0 ~ /^[[:space:]]*$/) {
                    if ($0 !~ /^[[:space:]]*$/) {
                        block_indent = current_indent
                        # Remove the block indentation
                        sub("^" sprintf("%*s", block_indent, ""), "")
                        print
                    }
                } else {
                    exit
                }
            } else {
                # Subsequent lines
                if (current_indent >= block_indent || $0 ~ /^[[:space:]]*$/) {
                    # Remove the block indentation
                    if ($0 !~ /^[[:space:]]*$/) {
                        sub("^" sprintf("%*s", block_indent, ""), "")
                    }
                    print
                } else {
                    exit
                }
            }
        } else if ($0 ~ "^[[:space:]]*" key ":") {
            # Found the key
            found = 1
            key_indent = current_indent

            # Check if value is on same line
            if ($0 ~ "^[[:space:]]*" key ": ") {
                # Inline value
                sub("^[[:space:]]*" key ": ", "")
                print
                exit
            } else {
                # Block value
                in_block = 1
            }
        }
    }
    END {
        # If key was never found, print null
        if (!found) {
            print "null"
        }
    }
    ' "$_file"
}

# Iterate over array or object elements
yq_iterate() {
    _file="$1"

    # Try as array first
    if head -n 1 "$_file" | grep -q '^-'; then
        # Array iteration - handle multi-line array elements
        awk '
        BEGIN {
            in_item = 0
            item_indent = -1
        }
        /^-/ {
            # New array item
            if (in_item) {
                # Print TWO newlines before new item to create blank line separator
                printf "\n\n"
            }
            in_item = 1
            # Remove "- " prefix and print
            sub(/^- /, "")
            # If line has content after "- ", print it
            if (length($0) > 0) {
                printf "%s", $0
                item_indent = 2  # Content started on same line
            } else {
                item_indent = -1  # Content on next line
            }
            next
        }
        in_item && /^[[:space:]]/ {
            # Continuation of array item (indented line)
            # Calculate indentation
            indent = 0
            for (i = 1; i <= length($0); i++) {
                if (substr($0, i, 1) == " ") {
                    indent++
                } else {
                    break
                }
            }

            # If this is the first indented line after "- ", set the base indent
            if (item_indent == -1) {
                item_indent = indent
            }

            # Remove the base indentation
            if (indent >= item_indent) {
                for (i = 1; i <= item_indent; i++) {
                    sub(/^ /, "")
                }
                printf "\n%s", $0
            }
        }
        ' "$_file"

        # Add final newline if we printed anything
        if [ -s "$_file" ]; then
            printf "\n"
        fi
    else
        # Object iteration (return values)
        awk '
        /^[a-zA-Z_]/ {
            if ($0 ~ /:/) {
                sub(/^[^:]*: */, "")
                print
            }
        }
        ' "$_file"
    fi
}

# Access array by index or slice
yq_array_access() {
    _spec="$1"
    _file="$2"

    # Extract index/slice from brackets
    _inner=$(echo "$_spec" | sed 's/\[\(.*\)\]/\1/')

    # Check if it's a slice (contains :)
    if echo "$_inner" | grep -q ':'; then
        # Slice operation
        _start=$(echo "$_inner" | cut -d: -f1)
        _end=$(echo "$_inner" | cut -d: -f2)

        # Default values
        [ -z "$_start" ] && _start=0
        [ -z "$_end" ] && _end=9999999

        awk -v start="$_start" -v end="$_end" '
        BEGIN {
            idx = 0
        }
        /^-/ {
            if (idx >= start && idx < end) {
                print
            }
            idx++
            if (idx >= end) exit
        }
        ' "$_file"
    else
        # Single index access
        _idx="$_inner"

        # Handle negative indices
        if echo "$_idx" | grep -q '^-'; then
            # Negative index - count from end
            _pos_idx=$(echo "$_idx" | sed 's/^-//')

            # Count total array elements
            _total=$(grep -c '^-' "$_file")
            _target_idx=$((_total - _pos_idx))

            awk -v target="$_target_idx" '
            BEGIN {
                idx = 0
            }
            /^-/ {
                if (idx == target) {
                    sub(/^- /, "")
                    print
                    exit
                }
                idx++
            }
            ' "$_file"
        else
            # Positive index - print the entire object/item at this index
            awk -v target="$_idx" '
            BEGIN {
                idx = 0
                found = 0
                in_item = 0
            }
            /^-/ {
                # Start of a new array item
                if (found == 1) {
                    # We were printing a previous item, stop now
                    exit
                }
                if (idx == target) {
                    # Found the target index
                    found = 1
                    in_item = 1
                    sub(/^- /, "")
                    print
                } else {
                    idx++
                }
                next
            }
            /^[^ -]/ && found == 1 {
                # Non-indented line after finding target = next item
                exit
            }
            found == 1 && in_item == 1 {
                # We'"'"'re in the target item, print all lines
                print
            }
            ' "$_file"
        fi
    fi
}

# Get length of array or object
yq_length() {
    _file="$1"

    # Check if it's an array (starts with -)
    if head -n 1 "$_file" | grep -q '^-'; then
        # Count array elements
        _count=$(grep -c '^-' "$_file")
    else
        # Count object keys
        _count=$(grep -c '^[a-zA-Z_][a-zA-Z0-9_]*:' "$_file")
    fi
    printf "%s" "$_count"
}

# Get keys of an object
yq_keys() {
    _file="$1"

    awk '
    BEGIN {
        first = 1
    }
    /^[a-zA-Z_][a-zA-Z0-9_]*:/ {
        match($0, /^[a-zA-Z_][a-zA-Z0-9_]*/)
        key = substr($0, 1, RLENGTH)
        if (first) {
            printf "- %s", key
            first = 0
        } else {
            printf "\n- %s", key
        }
    }
    ' "$_file"
}

# Convert object to entries (array of {key: k, value: v})
yq_to_entries() {
    _file="$1"

    # Parse the object and output as array of entries
    awk '
    BEGIN {
        first = 1
    }
    /^[a-zA-Z_][a-zA-Z0-9_]*:/ {
        # Extract key
        match($0, /^[a-zA-Z_][a-zA-Z0-9_]*/)
        key = substr($0, 1, RLENGTH)

        # Extract value (everything after ": ")
        value_start = index($0, ": ") + 2
        value = substr($0, value_start)

        # Output as entry object
        if (first) {
            printf "- key: %s\n  value: %s", key, value
            first = 0
        } else {
            printf "\n- key: %s\n  value: %s", key, value
        }
    }
    ' "$_file"

    # Add trailing newline if we printed anything
    if [ -s "$_file" ]; then
        printf "\n"
    fi
}

# Check if object has a key
yq_has() {
    _key="$1"
    _file="$2"

    if grep -q "^$_key:" "$_file" || grep -q "^[[:space:]]*$_key:" "$_file"; then
        printf "true"
    else
        printf "false"
    fi
}
`
}
