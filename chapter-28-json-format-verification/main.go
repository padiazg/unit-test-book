package json_format_verification

import (
	"encoding/json"
	"fmt"
	"time"
)

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Items     []Item    `json:"items"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Total   int `json:"total"`
	Offset  int `json:"offset"`
	Limit   int `json:"limit"`
}

func FormatOrder(order Order) (string, error) {
	data, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling order: %w", err)
	}
	return string(data), nil
}

func ParseOrder(jsonStr string) (*Order, error) {
	var order Order
	if err := json.Unmarshal([]byte(jsonStr), &order); err != nil {
		return nil, fmt.Errorf("parsing order: %w", err)
	}
	return &order, nil
}

func BuildSuccessResponse(data interface{}, meta *Meta) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

func BuildErrorResponse(err string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   err,
	}
}

func MustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
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

	if _, ok := raw["total"].(float64); !ok {
		return fmt.Errorf("total must be numeric")
	}

	return nil
}

func IsJSON(s string) bool {
	return json.Valid([]byte(s))
}

func NormalizeJSON(s string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func CompactJSON(s string) string {
	var v interface{}
	json.Unmarshal([]byte(s), &v)
	data, _ := json.Marshal(v)
	return string(data)
}

func FormatJSON(s string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("formatting: %w", err)
	}
	return string(data), nil
}

func ContainsJSONKey(jsonStr, key string) bool {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func GetJSONValue(jsonStr, key string) (interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	v, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found", key)
	}
	return v, nil
}

type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
}

func SerializeInventory(products []Product) (string, error) {
	resp := BuildSuccessResponse(products, &Meta{Total: len(products)})
	return FormatJSON(Serialize(resp))
}

func Serialize(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func MustFormat(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
