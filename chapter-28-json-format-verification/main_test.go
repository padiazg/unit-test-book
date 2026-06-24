package json_format_verification

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatOrder(t *testing.T) {
	order := Order{
		ID:     "ord_1",
		UserID: "usr_1",
		Items: []Item{
			{ProductID: "prod_1", Name: "Widget", Quantity: 2, Price: 9.99},
		},
		Total:  19.98,
		Status: "pending",
	}

	jsonStr, err := FormatOrder(order)
	require.NoError(t, err)

	assert.True(t, strings.Contains(jsonStr, `"id": "ord_1"`))
	assert.True(t, strings.Contains(jsonStr, `"total": 19.98`))
	assert.True(t, strings.Contains(jsonStr, `"status": "pending"`))
	assert.True(t, strings.HasPrefix(jsonStr, "{"))
}

func TestParseOrder(t *testing.T) {
	jsonStr := `{"id":"ord_1","user_id":"usr_1","total":29.97,"status":"completed","items":[]}`
	order, err := ParseOrder(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, "ord_1", order.ID)
	assert.Equal(t, float64(29.97), order.Total)
	assert.Equal(t, "completed", order.Status)
}

func TestParseOrder_InvalidJSON(t *testing.T) {
	_, err := ParseOrder(`{bad json}`)
	assert.Error(t, err)
}

func TestBuildSuccessResponse(t *testing.T) {
	resp := BuildSuccessResponse("hello", nil)
	assert.True(t, resp.Success)
	assert.Equal(t, "hello", resp.Data)
	assert.Nil(t, resp.Meta)
}

func TestBuildErrorResponse(t *testing.T) {
	resp := BuildErrorResponse("something went wrong")
	assert.False(t, resp.Success)
	assert.Equal(t, "something went wrong", resp.Error)
	assert.Empty(t, resp.Data)
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

	t.Run("invalid JSON", func(t *testing.T) {
		err := ValidateOrderJSON(`{invalid}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}

func TestIsJSON(t *testing.T) {
	assert.True(t, IsJSON(`{"a":1}`))
	assert.False(t, IsJSON(`not json`))
	assert.False(t, IsJSON(``))
}

func TestNormalizeJSON(t *testing.T) {
	normalized, err := NormalizeJSON(`{  "b" : 2 , "a" : 1 }`)
	require.NoError(t, err)
	assert.Equal(t, `{"a":1,"b":2}`, normalized)
}

func TestCompactJSON(t *testing.T) {
	result := CompactJSON(`{
		"name": "test",
		"value": 42
	}`)
	assert.Equal(t, `{"name":"test","value":42}`, result)
}

func TestFormatJSON(t *testing.T) {
	formatted, err := FormatJSON(`{"nested":{"a":1,"b":2}}`)
	require.NoError(t, err)
	assert.Contains(t, formatted, "  ")
	assert.Contains(t, formatted, "\n")
}

func TestContainsJSONKey(t *testing.T) {
	assert.True(t, ContainsJSONKey(`{"found":true}`, "found"))
	assert.False(t, ContainsJSONKey(`{"found":true}`, "missing"))
	assert.False(t, ContainsJSONKey(`invalid`, "key"))
}

func TestGetJSONValue(t *testing.T) {
	v, err := GetJSONValue(`{"count":42}`, "count")
	require.NoError(t, err)
	assert.Equal(t, float64(42), v)
}

func TestGetJSONValue_Missing(t *testing.T) {
	_, err := GetJSONValue(`{"a":1}`, "b")
	assert.Error(t, err)
}

func TestSerializeInventory(t *testing.T) {
	products := []Product{
		{ID: "p1", Name: "Widget", Price: 9.99},
	}
	jsonStr, err := SerializeInventory(products)
	require.NoError(t, err)

	var resp APIResponse
	err = json.Unmarshal([]byte(jsonStr), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestJSONEqAssert(t *testing.T) {
	expected := `{"id":"1","user_id":"u1","items":[],"total":0,"status":"","created_at":"0001-01-01T00:00:00Z"}`

	order, err := ParseOrder(`{"id":"1","user_id":"u1","items":[]}`)
	require.NoError(t, err)

	actual, err := FormatOrder(*order)
	require.NoError(t, err)
	assert.JSONEq(t, expected, actual)
}

func TestFormatOrder_RoundTrip(t *testing.T) {
	original := Order{
		ID:  "ord_1",
		Items: []Item{
			{ProductID: "p1", Name: "Gadget", Quantity: 1, Price: 49.99},
		},
		Total:  49.99,
		Status: "shipped",
	}

	jsonStr, err := FormatOrder(original)
	require.NoError(t, err)

	parsed, err := ParseOrder(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, original.ID, parsed.ID)
	assert.Equal(t, original.Total, parsed.Total)
	assert.Len(t, parsed.Items, 1)
	assert.Equal(t, "Gadget", parsed.Items[0].Name)
}

func TestBuildSuccessResponse_Marshal(t *testing.T) {
	resp := BuildSuccessResponse(Product{ID: "p1", Name: "Test"}, &Meta{Total: 1})
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded APIResponse
	json.Unmarshal(data, &decoded)
	assert.True(t, decoded.Success)
}
