package main

// GenerateJSON returns the JSON output conversion function
func GenerateJSON() string {
	return `
# Convert YAML output to JSON format
# Input: Multi-line YAML formatted as "key: value" pairs from iteration
# Output: One or more JSON objects, one per line
yq_yaml_to_json() {
    _yaml_input="$1"

    # Use AWK to detect object boundaries
    # When we see a key we'\''ve seen before at depth 0, it means a new object is starting
    printf '%s\n' "$_yaml_input" | awk '
    BEGIN {
        first = 1  # First property in current object
        obj_started = 0
    }
    /^[[:space:]]*$/ {
        # Empty line - skip but dont close object (we use key repetition for boundaries)
        next
    }
    /^[^[:space:]]/ {
        # Line starts at column 0, not indented - a top-level property
        if ($0 ~ /^[^:]+:[[:space:]]*/) {
            # Extract the key
            idx = index($0, ":")
            key = substr($0, 1, idx - 1)

            # Check if we'\''ve seen this key in the current object
            # If yes and we'\''ve started an object, close the current one and start new
            if (current_obj_keys[key] == 1 && obj_started == 1) {
                printf "}\n"
                first = 1
                # Reset the object keys tracking for the new object
                for (k in current_obj_keys) delete current_obj_keys[k]
            }

            # Mark this key as seen in the current object
            current_obj_keys[key] = 1

            # Start new object if needed
            if (first == 1) {
                printf "{"
                first = 0
                obj_started = 1
            } else {
                printf ","
            }

            # Extract value (after the colon)
            value = substr($0, idx + 1)

            # Trim whitespace from value
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)

            # Remove quotes from value if present
            if (value ~ /^".*"$/) {
                value = substr(value, 2, length(value) - 2)
            }

            # Escape JSON special characters in value
            gsub(/\\/, "\\\\", value)
            gsub(/"/, "\\\"", value)

            # Output JSON key:value pair
            printf "\"%s\":\"%s\"", key, value
        }
    }
    END {
        # Close final object if one is open
        if (obj_started == 1) {
            printf "}\n"
        }
    }
    '
}
`
}
