package main

import (
	"fmt"
)

func main() {
	fmt.Println("#!/bin/sh")
	fmt.Println()
	fmt.Println(`
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

# Parse and execute yq query recursively
yq_parse() {
    _query="$1"
    _file="$2"

    # Increment depth for this call
    _yq_parse_depth=$((_yq_parse_depth + 1))
    [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "yq_parse called with query='$_query'"

    # Check for recursive descent operator BEFORE removing leading dot
    if [ "$_query" = ".." ]; then
        # Recursive descent not fully implemented yet - return empty for now
        # TODO: Fix variable scoping issues in recursive functions
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for parentheses - but only if they wrap the entire expression
    # Need to handle cases like (.foo) | .bar differently from just (.foo)
    if echo "$_query" | grep -q '^([^)]*)'$; then
        # Just parentheses, no pipe after
        _inner=$(echo "$_query" | sed 's/^(//' | sed 's/)$//')
        yq_parse "$_inner" "$_file"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    elif echo "$_query" | grep -q '^([^)]*) | '; then
        # Parentheses followed by pipe
        _paren_part=$(echo "$_query" | sed 's/^\(([^)]*)\).*/\1/')
        _after_paren=$(echo "$_query" | sed 's/^[^)]*) //')

        # Process the parenthesized part first
        _inner=$(echo "$_paren_part" | sed 's/^(//' | sed 's/)$//')
        _tmp_paren=$(mktemp)
        yq_parse "$_inner" "$_file" > "$_tmp_paren"

        # Then process what comes after (including the pipe) with the result
        yq_parse "$_after_paren" "$_tmp_paren"
        rm -f "$_tmp_paren"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for alternative operator // (e.g., .key // default)
    # Must check before pipe operator since it contains //
    if echo "$_query" | grep -q ' // '; then
        _before_alt=$(echo "$_query" | sed 's/ \/\/.*//')
        _after_alt=$(echo "$_query" | sed 's/.* \/\/ //')

        # Try to evaluate the first part
        _tmp_alt=$(mktemp)
        yq_parse "$_before_alt" "$_file" > "$_tmp_alt" 2>/dev/null

        # Check if result is empty, null, or doesn't exist
        _alt_result=""
        if [ -s "$_tmp_alt" ]; then
            _alt_result=$(cat "$_tmp_alt")
        fi

        if [ -z "$_alt_result" ] || [ "$_alt_result" = "null" ] || [ "$_alt_result" = "" ]; then
            # Use alternative value
            rm -f "$_tmp_alt"

            # Check if alternative is a literal value
            if [ "$_after_alt" = "[]" ]; then
                # Empty array - return nothing
                return
            elif [ "$_after_alt" = "null" ]; then
                printf "null"
            else
                # Try to parse as expression
                yq_parse "$_after_alt" "$_file"
            fi
        else
            # Use first value
            printf "%s" "$_alt_result"
            rm -f "$_tmp_alt"
        fi
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for pipe operator (split on | and process sequentially)
    if echo "$_query" | grep -q ' | \|^| '; then
        # Handle case where query starts with pipe
        if echo "$_query" | grep -q '^| '; then
            _before_pipe="."
            _after_pipe=$(echo "$_query" | sed 's/^| //')
        else
            _before_pipe=$(echo "$_query" | sed 's/ |.*//')
            _after_pipe=$(echo "$_query" | sed 's/^[^|]* | //')
        fi

        [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Pipe detected - Before: '$_before_pipe' After: '$_after_pipe'"

        # Special handling for recursive descent
        if [ "$_before_pipe" = ".." ]; then
            yq_recursive_descent_pipe "$_file" "$_after_pipe"
            _yq_parse_depth=$((_yq_parse_depth - 1))
            return
        fi

        # Check if left side ends with .[] (iteration)
        # If so, we need to apply right side to each item separately
        if echo "$_before_pipe" | grep -q '\[\]$'; then
            [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration handler: processing items with remainder: '$_after_pipe'"
            # Create unique state files for this iteration level
            # Each level gets its own state file that won't be clobbered by nested calls
            _iter_state=$(mktemp)

            # Process left side to get items
            _tmp_pipe=$(mktemp)
            yq_parse "$_before_pipe" "$_file" > "$_tmp_pipe"

            # Debug: Show what's in tmp_pipe
            [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: tmp_pipe content after $_before_pipe:" && >&2 cat "$_tmp_pipe" && >&2 echo "DEBUG: ---"

            # Check if result is multi-line (multiple items)
            if [ -s "$_tmp_pipe" ]; then
                # Process each item separately
                # We need to detect item boundaries for multi-line items
                _iter_tmp_items=$(mktemp)
                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: about to split AWK input, first 10 lines:" && >&2 head -10 "$_tmp_pipe" | >&2 sed "s/^/  /"
                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: AWK input has $(wc -l < $_tmp_pipe) lines total"
                awk -v tmpbase="$_iter_tmp_items" '
                BEGIN {
                    item = ""
                    item_count = 0
                }
                {
                    # Check for blank line (item separator)
                    if ($0 ~ /^[[:space:]]*$/) {
                        # Output current item if we have one
                        if (item != "") {
                            item_count++
                            outfile = tmpbase "." item_count
                            print item > outfile
                            close(outfile)
                            item = ""
                        }
                        next
                    }

                    # Add line to current item
                    if (item == "") {
                        item = $0
                    } else {
                        item = item "\n" $0
                    }
                }
                END {
                    # Output last item
                    if (item != "") {
                        item_count++
                        outfile = tmpbase "." item_count
                        print item > outfile
                        close(outfile)
                    }
                    print item_count
                }
                ' "$_tmp_pipe" > "$_iter_tmp_items"

                # Read number of items
                _num_items=$(cat "$_iter_tmp_items")

                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: AWK split result - num_items=$_num_items"
                if [ "$_num_items" -gt 0 ]; then
                    [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: first item file size:" && >&2 wc -c < "$_iter_tmp_items.1" 2>/dev/null | >&2 sed "s/^/  /"
                fi

                # Store iteration state in file to prevent variable scoping issues in nested calls
                # Use separate files for each state variable to avoid sed delimiter issues
                printf "%s" "$_iter_tmp_items" > "$_iter_state.base"
                printf "%s" "$_num_items" > "$_iter_state.num"
                printf "%s" "$_after_pipe" > "$_iter_state.query"

                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: found $_num_items items to process"
                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: base path: $_iter_tmp_items"

                # Store iteration state in DEPTH-SPECIFIC variables to prevent nesting issues
                # All variables in POSIX shell are global, so we must use unique names per depth
                eval "_saved_iter_base_${_yq_parse_depth}='$_iter_tmp_items'"
                eval "_saved_iter_query_${_yq_parse_depth}='$_after_pipe'"
                eval "_saved_iter_count_${_yq_parse_depth}='$_num_items'"

                # Use local loop counter variable unique to this depth
                _loop_iter_var="_loop_idx_${_yq_parse_depth}"
                eval "$_loop_iter_var=0"

                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: saved base path: $_iter_tmp_items, count: $_num_items"
                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: seq command will produce: $(seq 1 $_num_items | tr '\n' ' ')"
                _loop_idx=0
                for _item_idx in $(seq 1 $_num_items); do
                    # Read from depth-specific variables
                    eval "_current_iter_base=\$_saved_iter_base_${_yq_parse_depth}"
                    eval "_current_iter_query=\$_saved_iter_query_${_yq_parse_depth}"
                    eval "_current_iter_count=\$_saved_iter_count_${_yq_parse_depth}"

                    _loop_idx=$((_loop_idx + 1))
                    [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: loop iteration $_loop_idx (processing item $_item_idx)"

                    if [ -f "$_current_iter_base.$_item_idx" ]; then
                        [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: item #$_item_idx - calling yq_parse with query: '$_current_iter_query'"

                        # Process the item (output goes directly to stdout)
                        yq_parse "$_current_iter_query" "$_current_iter_base.$_item_idx"

                        [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: item #$_item_idx - yq_parse returned"
                        rm -f "$_current_iter_base.$_item_idx"
                    else
                        [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: item #$_item_idx - FILE NOT FOUND: $_current_iter_base.$_item_idx"
                    fi
                    [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: continuing loop, next item will be $((_item_idx + 1))"
                done
                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: loop exited after processing $_loop_idx items"
                [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "Iteration: loop completed for all $_num_items items"

                rm -f "$_iter_state"* "$_iter_tmp_items"
            fi
            rm -f "$_tmp_pipe"
            _yq_parse_depth=$((_yq_parse_depth - 1))
            return
        fi

        # Standard pipe processing (no iteration)
        # Save variables before recursive calls to prevent scoping issues
        # Use unique variable names to avoid conflicts with nested pipe handlers
        _std_before_pipe="$_before_pipe"
        _std_after_pipe="$_after_pipe"

        _tmp_pipe=$(mktemp)
        yq_parse "$_std_before_pipe" "$_file" > "$_tmp_pipe"

        # Process second part with result from first part
        yq_parse "$_std_after_pipe" "$_tmp_pipe"
        rm -f "$_tmp_pipe"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for string concatenation with + (e.g., .key + "=" + .value)
    if echo "$_query" | grep -q ' + '; then
        # Split by + and evaluate each part, then concatenate
        _result=""
        _remaining="$_query"

        while echo "$_remaining" | grep -q ' + '; do
            _part=$(echo "$_remaining" | sed 's/ +.*//')
            _remaining=$(echo "$_remaining" | sed 's/^[^+]* + //')

            # Evaluate this part
            if echo "$_part" | grep -q '^".*"$'; then
                # String literal
                _part_value=$(echo "$_part" | sed 's/^"\(.*\)"$/\1/')
            else
                # Expression - evaluate it
                _tmp_concat=$(mktemp)
                yq_parse "$_part" "$_file" > "$_tmp_concat"
                _part_value=$(cat "$_tmp_concat")
                # Remove quotes if present
                _part_value=$(echo "$_part_value" | sed 's/^"\(.*\)"$/\1/')
                rm -f "$_tmp_concat"
            fi

            _result="${_result}${_part_value}"
        done

        # Process last part
        if echo "$_remaining" | grep -q '^".*"$'; then
            # String literal
            _part_value=$(echo "$_remaining" | sed 's/^"\(.*\)"$/\1/')
        else
            # Expression
            _tmp_concat=$(mktemp)
            yq_parse "$_remaining" "$_file" > "$_tmp_concat"
            _part_value=$(cat "$_tmp_concat")
            # Remove quotes if present
            _part_value=$(echo "$_part_value" | sed 's/^"\(.*\)"$/\1/')
            rm -f "$_tmp_concat"
        fi
        _result="${_result}${_part_value}"

        echo "$_result"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Remove leading dot
    _query=$(echo "$_query" | sed 's/^\.//')

    # Handle empty query (identity)
    if [ -z "$_query" ]; then
        cat "$_file"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for functions without arguments
    case "$_query" in
        "length")
            yq_length "$_file"
            _yq_parse_depth=$((_yq_parse_depth - 1))
            return
            ;;
        "keys")
            yq_keys "$_file"
            _yq_parse_depth=$((_yq_parse_depth - 1))
            return
            ;;
        "to_entries")
            yq_to_entries "$_file"
            _yq_parse_depth=$((_yq_parse_depth - 1))
            return
            ;;
    esac

    # Check for assignment operator (=)
    if echo "$_query" | grep -q ' = ' && ! echo "$_query" | grep -q ' == '; then
        yq_assign "$_query" "$_file"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for update operator (|=)
    if echo "$_query" | grep -q ' |= '; then
        yq_update "$_query" "$_file"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for comparison operators (==, !=, etc.)
    if echo "$_query" | grep -q ' == \| != '; then
        yq_compare "$_query" "$_file"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for functions with arguments
    if echo "$_query" | grep -q '^[a-zA-Z_][a-zA-Z0-9_]*('; then
        _func_name=$(echo "$_query" | sed 's/(.*//')
        _func_args=$(echo "$_query" | sed 's/^[^(]*//' | sed 's/^(//' | sed 's/)$//')

        case "$_func_name" in
            "has")
                # Remove quotes from argument
                _key=$(echo "$_func_args" | sed 's/^"\(.*\)"$/\1/')
                yq_has "$_key" "$_file"
                return
                ;;
            "map")
                yq_map "$_func_args" "$_file"
                return
                ;;
            "select")
                yq_select "$_func_args" "$_file"
                return
                ;;
            "del")
                yq_del "$_func_args" "$_file"
                return
                ;;
        esac
    fi

    # Parse first token and remainder
    _first_token=""
    _remainder=""

    # Check for array iteration .[]
    if echo "$_query" | grep -q '^\[\]'; then
        _first_token="[]"
        _remainder=$(echo "$_query" | sed 's/^\[\]//')
    # Check for array index .[n] or .[-n]
    elif echo "$_query" | grep -q '^\[-\?[0-9][0-9]*\]'; then
        _first_token=$(echo "$_query" | sed 's/^\(\[-\?[0-9][0-9]*\]\).*/\1/')
        _remainder=$(echo "$_query" | sed 's/^\[-\?[0-9][0-9]*\]//')
    # Check for array slice .[n:m]
    elif echo "$_query" | grep -q '^\[[0-9]*:[0-9]*\]'; then
        _first_token=$(echo "$_query" | sed 's/^\(\[[0-9]*:[0-9]*\]\).*/\1/')
        _remainder=$(echo "$_query" | sed 's/^\[[0-9]*:[0-9]*\]//')
    # Check for key access
    elif echo "$_query" | grep -q '^[a-zA-Z_][a-zA-Z0-9_]*'; then
        _first_token=$(echo "$_query" | sed 's/^\([a-zA-Z_][a-zA-Z0-9_]*\).*/\1/')
        _remainder=$(echo "$_query" | sed 's/^[a-zA-Z_][a-zA-Z0-9_]*//')
    else
        # Unknown token, return empty
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Remove leading dot from remainder
    _remainder=$(echo "$_remainder" | sed 's/^\.//')

    # Apply the first token operation
    _tmp_result=$(mktemp)

    case "$_first_token" in
        "[]")
            # Array/object iteration
            yq_iterate "$_file" > "$_tmp_result"
            ;;
        \[*\])
            # Array access (index or slice)
            yq_array_access "$_first_token" "$_file" > "$_tmp_result"
            ;;
        *)
            # Key access
            yq_key_access "$_first_token" "$_file" > "$_tmp_result"
            ;;
    esac

    # If there's a remainder, recursively process each result
    if [ -n "$_remainder" ]; then
        # Check if tmp_result has content
        if [ -s "$_tmp_result" ]; then
            # For iteration results, process each item separately
            if [ "$_first_token" = "[]" ]; then
                _line_num=0
                while IFS= read -r _line || [ -n "$_line" ]; do
                    _line_num=$((_line_num + 1))
                    _item_file=$(mktemp)
                    echo "$_line" > "$_item_file"
                    yq_parse "$_remainder" "$_item_file"
                    rm -f "$_item_file"
                done < "$_tmp_result"
            else
                yq_parse "$_remainder" "$_tmp_result"
            fi
        fi
    else
        cat "$_tmp_result"
    fi

    rm -f "$_tmp_result"
    _yq_parse_depth=$((_yq_parse_depth - 1))
}

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
            # Positive index
            awk -v target="$_idx" '
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
    # Check if it's an object (has keys with colons)
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
    # Check if it's an array
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
    # Check if it's an object (has keys with colons)
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
    # Check if it's an array
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

# Assignment operator - set a value
yq_assign() {
    _expr="$1"
    _file="$2"

    # Parse the assignment: .path = value
    _path=$(echo "$_expr" | sed 's/ =.*//')
    _value=$(echo "$_expr" | sed 's/.* = //')

    # Remove leading dot from path
    _path=$(echo "$_path" | sed 's/^\.//')

    # Remove quotes from value if it's a string literal
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

# Main entry point
_exit_on_null=0
_output_format="yaml"
_raw_output=0

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
        -j|--json|--raw-input)
            # Placeholder for other flags - just skip them for now
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
        cat > "$FILE"
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: Stdin written to $FILE"
        [ -n "$POSIX_YQ_DEBUG" ] && >&2 echo "DEBUG: File size: $(wc -c < $FILE)"
        _cleanup_file="$FILE"
    elif [ -z "$QUERY" ]; then
        # No query and no file - error
        >&2 echo "Error: No input file and no query provided"
        exit 1
    else
        # Only query provided, assume it's actually the file
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

# Clean up result: remove blank line separators from array iteration
# The iteration uses blank lines as separators, but we only want actual content
while printf '%s' "$_result" | grep -q '^[[:space:]]*$'; do
    _result=$(printf '%s\n' "$_result" | grep -v '^[[:space:]]*$')
done

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
`)
}