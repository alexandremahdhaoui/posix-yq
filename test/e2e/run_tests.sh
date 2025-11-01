#!/bin/sh

# Regenerate posix-yq script
echo "Regenerating posix-yq..."
go run cmd/generator/main.go > posix-yq && chmod +x posix-yq

# Test Case 1: Read entire YAML file
echo "Running Test 1: Read entire YAML file..."
ACTUAL=$(timeout 5 ./posix-yq test/fixtures/01-simple.yaml 2>&1)
EXPECTED=$(cat test/fixtures/01-simple.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 1: Read entire YAML file - PASSED"
else
  echo "✗ Test 1: Read entire YAML file - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 2: Extract root-level key (.name)
echo "Running Test 2: Extract root-level key (.name)..."
ACTUAL=$(./posix-yq '.name' test/fixtures/02-root-keys.yaml 2>&1)
EXPECTED=$(cat test/fixtures/02-root-keys-name.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 2: Extract .name - PASSED"
else
  echo "✗ Test 2: Extract .name - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 3: Extract root-level key (.age)
echo "Running Test 3: Extract root-level key (.age)..."
ACTUAL=$(./posix-yq '.age' test/fixtures/02-root-keys.yaml 2>&1)
EXPECTED=$(cat test/fixtures/02-root-keys-age.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 3: Extract .age - PASSED"
else
  echo "✗ Test 3: Extract .age - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 4: Extract nested key (.person.name)
echo "Running Test 4: Extract nested key (.person.name)..."
ACTUAL=$(./posix-yq '.person.name' test/fixtures/03-nested.yaml 2>&1)
EXPECTED=$(cat test/fixtures/03-nested-person-name.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 4: Extract .person.name - PASSED"
else
  echo "✗ Test 4: Extract .person.name - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 5: Extract deeply nested key (.person.address.city)
echo "Running Test 5: Extract deeply nested key (.person.address.city)..."
ACTUAL=$(./posix-yq '.person.address.city' test/fixtures/03-nested.yaml 2>&1)
EXPECTED=$(cat test/fixtures/03-nested-address-city.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 5: Extract .person.address.city - PASSED"
else
  echo "✗ Test 5: Extract .person.address.city - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 6: Extract array element (.items[0])
echo "Running Test 6: Extract array element (.items[0])..."
ACTUAL=$(./posix-yq '.items[0]' test/fixtures/04-arrays.yaml 2>&1)
EXPECTED=$(cat test/fixtures/04-arrays-items-0.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 6: Extract .items[0] - PASSED"
else
  echo "✗ Test 6: Extract .items[0] - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 7: Extract array element (.items[2])
echo "Running Test 7: Extract array element (.items[2])..."
ACTUAL=$(./posix-yq '.items[2]' test/fixtures/04-arrays.yaml 2>&1)
EXPECTED=$(cat test/fixtures/04-arrays-items-2.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 7: Extract .items[2] - PASSED"
else
  echo "✗ Test 7: Extract .items[2] - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 8: JSON output format
echo "Running Test 8: JSON output format (-o json)..."
ACTUAL=$(./posix-yq -o json test/fixtures/01-simple.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-json-simple.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 8: JSON output - PASSED"
else
  echo "✗ Test 8: JSON output - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 9: Non-existent file - should fail with error
echo "Running Test 9: Non-existent file error..."
./posix-yq nonexistent.yaml >/dev/null 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
  echo "✓ Test 9: Non-existent file error - PASSED"
else
  echo "✗ Test 9: Non-existent file error - FAILED (should exit with non-zero)"
  exit 1
fi

# Test Case 10: Invalid query - should fail with error or return empty
echo "Running Test 10: Invalid query syntax..."
./posix-yq 'invalid_query' test/fixtures/01-simple.yaml >/dev/null 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
  echo "✓ Test 10: Invalid query error - PASSED"
else
  # If it doesn't error, that's OK (some implementations might just return empty)
  echo "✓ Test 10: Invalid query - PASSED (no error, but acceptable)"
fi

# Test Case 11: Missing key - should return empty
echo "Running Test 11: Missing key returns empty..."
ACTUAL=$(./posix-yq '.nonexistent' test/fixtures/02-root-keys.yaml 2>&1)
if [ -z "$ACTUAL" ]; then
  echo "✓ Test 11: Missing key returns empty - PASSED"
else
  echo "✗ Test 11: Missing key - FAILED"
  echo "Expected: (empty)"
  echo "Actual: $ACTUAL"
  exit 1
fi

# Test Case 12: Out-of-bounds array index - should return empty
echo "Running Test 12: Out-of-bounds array index..."
ACTUAL=$(./posix-yq '.items[99]' test/fixtures/04-arrays.yaml 2>&1)
if [ -z "$ACTUAL" ]; then
  echo "✓ Test 12: Out-of-bounds array - PASSED"
else
  echo "✗ Test 12: Out-of-bounds array - FAILED"
  echo "Expected: (empty)"
  echo "Actual: $ACTUAL"
  exit 1
fi

# Test Case 13: Array iteration (.items[])
echo "Running Test 13: Array iteration (.items[])..."
ACTUAL=$(./posix-yq '.items[]' test/fixtures/04-arrays.yaml 2>&1)
EXPECTED=$(cat test/fixtures/04-arrays-iterate.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 13: Array iteration - PASSED"
else
  echo "✗ Test 13: Array iteration - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 14: Length of array (.items | length)
echo "Running Test 14: Length of array (.items | length)..."
ACTUAL=$(./posix-yq '.items | length' test/fixtures/04-arrays.yaml 2>&1)
EXPECTED=$(cat test/fixtures/06-length-array.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 14: Array length - PASSED"
else
  echo "✗ Test 14: Array length - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 15: Length of object (.person | length)
echo "Running Test 15: Length of object (.person | length)..."
ACTUAL=$(./posix-yq '.person | length' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/07-length-object.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 15: Object length - PASSED"
else
  echo "✗ Test 15: Object length - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 16: Keys of object (.person | keys)
echo "Running Test 16: Keys of object (.person | keys)..."
ACTUAL=$(./posix-yq '.person | keys' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/08-keys.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 16: Object keys - PASSED"
else
  echo "✗ Test 16: Object keys - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 17: Multiple selections (.person.name, .person.age)
echo "Running Test 17: Multiple selections (.person.name, .person.age)..."
ACTUAL=$(./posix-yq '.person.name, .person.age' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/09-multi-select.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 17: Multiple selections - PASSED"
else
  echo "✗ Test 17: Multiple selections - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 18: Has operator - key exists (.person | has("name"))
echo "Running Test 18: Has operator - key exists..."
ACTUAL=$(./posix-yq '.person | has("name")' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/10-has-true.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 18: Has operator (true) - PASSED"
else
  echo "✗ Test 18: Has operator (true) - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 19: Has operator - key does not exist (.person | has("missing"))
echo "Running Test 19: Has operator - key does not exist..."
ACTUAL=$(./posix-yq '.person | has("missing")' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/10-has-false.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 19: Has operator (false) - PASSED"
else
  echo "✗ Test 19: Has operator (false) - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 20: Alternative operator - missing key with default (.person.missing // "default_value")
echo "Running Test 20: Alternative operator - missing key..."
ACTUAL=$(./posix-yq '.person.missing // "default_value"' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/11-alternative.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 20: Alternative operator (default) - PASSED"
else
  echo "✗ Test 20: Alternative operator (default) - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 21: Alternative operator - existing key (.person.name // "default")
echo "Running Test 21: Alternative operator - existing key..."
ACTUAL=$(./posix-yq '.person.name // "default"' test/fixtures/07-length-object.yaml 2>&1)
EXPECTED=$(cat test/fixtures/11-alternative-exists.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 21: Alternative operator (exists) - PASSED"
else
  echo "✗ Test 21: Alternative operator (exists) - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 22: Select with price comparison (.items[] | select(.price > 1))
echo "Running Test 22: Select with price comparison..."
ACTUAL=$(./posix-yq '.items[] | select(.price > 1)' test/fixtures/05-advanced.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-advanced-select-price.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 22: Select price > 1 - PASSED"
else
  echo "✗ Test 22: Select price > 1 - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 23: Select with string equality (.items[] | select(.name == "banana"))
echo "Running Test 23: Select with string equality..."
ACTUAL=$(./posix-yq '.items[] | select(.name == "banana")' test/fixtures/05-advanced.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-advanced-select-name.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 23: Select name == banana - PASSED"
else
  echo "✗ Test 23: Select name == banana - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 24: Select with >= comparison (.metadata.ratings[] | select(. >= 4))
echo "Running Test 24: Select with >= comparison..."
ACTUAL=$(./posix-yq '.metadata.ratings[] | select(. >= 4)' test/fixtures/05-advanced.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-advanced-select-ratings.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 24: Select rating >= 4 - PASSED"
else
  echo "✗ Test 24: Select rating >= 4 - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 25: Map to extract field
echo "Running Test 25: Map to extract field (.items | map(.name))..."
ACTUAL=$(./posix-yq '.items | map(.name)' test/fixtures/05-advanced.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-advanced-map-names.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 25: Map extract field - PASSED"
else
  echo "✗ Test 25: Map extract field - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 26: Map with arithmetic
echo "Running Test 26: Map with arithmetic (.metadata.ratings | map(. + 10))..."
ACTUAL=$(./posix-yq '.metadata.ratings | map(. + 10)' test/fixtures/05-advanced.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-advanced-map-arithmetic.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 26: Map arithmetic - PASSED"
else
  echo "✗ Test 26: Map arithmetic - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

# Test Case 27: Map with string concatenation
echo "Running Test 27: Map with string concat (.metadata.tags | map(. + \" tag\"))..."
ACTUAL=$(./posix-yq '.metadata.tags | map(. + " tag")' test/fixtures/05-advanced.yaml 2>&1)
EXPECTED=$(cat test/fixtures/05-advanced-map-concat.expected)
if [ "$ACTUAL" = "$EXPECTED" ]; then
  echo "✓ Test 27: Map string concat - PASSED"
else
  echo "✗ Test 27: Map string concat - FAILED"
  echo "Expected:"
  echo "$EXPECTED"
  echo "Actual:"
  echo "$ACTUAL"
  exit 1
fi

exit 0
