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

// GenerateParser returns the yq_parse recursive parser function
func GenerateParser() string {
	return `
# Parse and execute yq query recursively
yq_parse() {
    _query="$1"
    _file="$2"

    # Increment depth for this call
    _yq_parse_depth=$((_yq_parse_depth + 1))
    [ -n "$POSIX_YQ_DEBUG" ] && _yq_debug_indent "$_yq_parse_depth" "yq_parse called with query='$_query'"

    # Check for recursive descent operator BEFORE removing leading dot
    if [ "$_query" = ".." ]; then
        yq_recursive_descent "$_file"
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

    # Check for functions with arguments BEFORE pipes (so pipes inside function args aren't split)
    # Pattern: functionname(args)
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
                        # Add a separator line between iterations
                        if [ "$_item_idx" -gt 1 ]; then
                            echo ""
                        fi
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

    # Check for arithmetic operators (-, *, /) - these are definitely not concatenation
    if echo "$_query" | grep -q ' - \| \* \| / '; then
        yq_arithmetic "$_query" "$_file"
        _yq_parse_depth=$((_yq_parse_depth - 1))
        return
    fi

    # Check for addition/concatenation with +
    # If it's a simple expr + number, treat as arithmetic
    # Otherwise, treat as string concatenation
    if echo "$_query" | grep -q ' + '; then
        # Check if there's only one + and right side is a number (arithmetic case)
        if ! echo "$_query" | grep -q '.* + .* + '; then
            _test_left=$(echo "$_query" | sed 's/ +.*//')
            _test_right=$(echo "$_query" | sed 's/.* + //')

            # If right side is just a number, might be arithmetic
            if echo "$_test_right" | grep -q '^[0-9]\+$'; then
                yq_arithmetic "$_query" "$_file"
                _yq_parse_depth=$((_yq_parse_depth - 1))
                return
            fi
        fi

        # Otherwise, split by + and evaluate each part, then concatenate
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
`
}
