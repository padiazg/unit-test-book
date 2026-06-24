# Chapter 28: JSON Format Verification

## Description

Test JSON serialization and deserialization with `json.Marshal`, `json.Unmarshal`, `json.MarshalIndent`, and `assert.JSONEq`. Verify field presence, omitempty behavior, indentation, round-trip correctness, and structural validation of nested JSON objects.

## Code

```go
type Order struct {
	ID     string  `json:"id"`
	Items  []Item  `json:"items"`
	Total  float64 `json:"total"`
	Status string  `json:"status"`
}

func FormatOrder(order Order) (string, error) {
	data, err := json.MarshalIndent(order, "", "  ")
	return string(data), err
}

func ValidateOrderJSON(jsonStr string) error {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	required := []string{"id", "user_id", "items", "total", "status"}
	for _, field := range required {
		if _, ok := raw[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}
```

## Test

```go
func TestFormatOrder(t *testing.T) {
	order := Order{ID: "ord_1", Items: []Item{{Name: "Widget", Quantity: 2, Price: 9.99}}, Total: 19.98}
	jsonStr, err := FormatOrder(order)
	require.NoError(t, err)
	assert.True(t, strings.Contains(jsonStr, `"id": "ord_1"`))
	assert.True(t, strings.HasPrefix(jsonStr, "{"))
}

func TestJSONEqAssert(t *testing.T) {
	expected := `{"id":"1","user_id":"u1","items":[],"total":0,"status":"","created_at":"0001-01-01T00:00:00Z"}`
	order, err := ParseOrder(`{"id":"1","user_id":"u1","items":[]}`)
	require.NoError(t, err)
	actual, _ := FormatOrder(*order)
	assert.JSONEq(t, expected, actual)
}

func TestValidateOrderJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		err := ValidateOrderJSON(`{"id":"1","user_id":"u1","items":[],"total":10,"status":"ok"}`)
		assert.NoError(t, err)
	})
	t.Run("missing field", func(t *testing.T) {
		err := ValidateOrderJSON(`{"id":"1"}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required")
	})
}

func TestFormatOrder_RoundTrip(t *testing.T) {
	original := Order{ID: "ord_1", Items: []Item{{Name: "Gadget", Quantity: 1, Price: 49.99}}, Status: "shipped"}
	jsonStr, _ := FormatOrder(original)
	parsed, _ := ParseOrder(jsonStr)
	assert.Equal(t, original.ID, parsed.ID)
	assert.Len(t, parsed.Items, 1)
}

func TestContainsJSONKey(t *testing.T) {
	assert.True(t, ContainsJSONKey(`{"found":true}`, "found"))
	assert.False(t, ContainsJSONKey(`{"found":true}`, "missing"))
}

func TestNormalizeJSON(t *testing.T) {
	n, _ := NormalizeJSON(`{  "b" : 2 , "a" : 1 }`)
	assert.Equal(t, `{"a":1,"b":2}`, n)
}

func TestCompactJSON(t *testing.T) {
	result := CompactJSON("{\n\t\"name\": \"test\"\n}")
	assert.Equal(t, `{"name":"test"}`, result)
}
```

## Testing Approach

JSON format verification:

1. **`assert.JSONEq`** — compares two JSON strings semantically, ignoring key order and whitespace. Use it instead of `assert.Equal` with raw strings to avoid brittle ordering dependencies.
2. **`json.Valid` and `json.Unmarshal` into `map[string]interface{}`** — `IsJSON()` uses `json.Valid` for quick syntax checking. `ValidateOrderJSON` unmarshals into a generic map to check field presence without defining a struct.
3. **Round-trip testing** — create a struct → marshal → unmarshal → compare original and parsed. Verifies that all fields survive serialization without data loss.
4. **`MarshalIndent` for human readability** — `FormatOrder` produces indented JSON. Tests verify the prefix (`{`) and indentation. `CompactJSON` shows the reverse: minification.
