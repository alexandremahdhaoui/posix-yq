# Phase 3 Implementation Plan - Critical yq Features

## Test File Location
All tests use: `test/fixtures/05-advanced.yaml`

## Feature 1: Select/Filter Operator

### Syntax
```bash
.items[] | select(.price > 1)
.items[] | select(.name == "banana")
.metadata.ratings[] | select(. >= 4)
```

### Expected Output (from real yq)
```bash
# .items[] | select(.price > 1)
name: apple
price: 1.50
name: cherry
price: 2.00

# .items[] | select(.name == "banana")
name: banana
price: 0.75

# .metadata.ratings[] | select(. >= 4)
4
5
5
4
```

### Implementation Tasks
1. Parse `select(condition)` syntax
2. Extract comparison: left operator right
3. Support operators: `>`, `<`, `>=`, `<=`, `==`, `!=`
4. Generate AWK code to filter
5. Output only matching elements

### Test Cases
- Test 22: `.items[] | select(.price > 1)`
- Test 23: `.items[] | select(.name == "banana")`
- Test 24: `.metadata.ratings[] | select(. >= 4)`

---

## Feature 2: Map Operator

### Syntax
```bash
.items | map(.name)
.metadata.ratings | map(. + 10)
```

### Expected Output
```bash
# .items | map(.name)
- apple
- banana
- cherry

# .metadata.ratings | map(. + 10)
[14, 15, 13, 15, 14]
```

### Implementation Tasks
1. Parse `map(expression)` syntax
2. Extract field or expression
3. Apply to each array element
4. Return transformed array

---

## Feature 3: String Operators

### Contains
```bash
yq '.name | contains("Jo")' → true
```

### Split
```bash
yq '.name | split(" ")' → ["John", "Doe"]
```

### Join
```bash
yq '.metadata.tags | join(", ")' → "fruit, healthy, organic"
```

---

## Feature 4: Arithmetic Operators

### Syntax
```bash
.age + 5        → 35
.price * 2      → 3.00
.score / 10     → 9.55
```

### Implementation
Use AWK for float arithmetic

---

## Feature 5: Sort and Unique

### Sort
```bash
yq '.metadata.ratings | sort' → [3, 4, 4, 5, 5]
```

### Unique
```bash
yq '.metadata.ratings | unique' → [4, 5, 3]
```

---

## Feature 6: Min/Max

```bash
yq '.metadata.ratings | max' → 5
yq '.metadata.ratings | min' → 3
```

---

## Implementation Order

1. **Select + Comparisons** (CRITICAL) - 8-10 hours
2. **String operators** (HIGH) - 4-5 hours
3. **Map** (HIGH) - 5-6 hours
4. **Arithmetic** (MEDIUM) - 4-5 hours
5. **Sort/Unique/Min/Max** (MEDIUM) - 6-8 hours

**Total: ~30-35 hours**
