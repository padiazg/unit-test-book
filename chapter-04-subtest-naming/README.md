# Chapter 04: Subtest Naming Strategies

## Description

Using package-level constants for test case names ensures consistency across multiple test functions that exercise the same domain concept. When the same constant (`orderID_001`) appears in `TestOrder_Confirm`, `TestOrder_Cancel`, and `TestOrder_Ship`, it signals these tests are testing the same entity. Constants also serve as documentation — they name the test case and the test value simultaneously.

Real-world example:

- `hexago/pkg/version/version_test.go:8-13` — constants like `version_0_0_1`, `version_0_0_1_rc_1` reuse across `TestVersionParseVersion`, `TestVersionParseDate`, `TestVersionString`  

## Code

```go
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	Status OrderStatus
	ID     string
	Amount float64
}

func NewOrder(id string, amount float64) *Order {
	return &Order{ID: id, Amount: amount, Status: StatusPending}
}

func (o *Order) Confirm() error { ... }
func (o *Order) Cancel() error { ... }
func (o *Order) Ship() error { ... }
func (o *Order) Deliver() error { ... }

// Full implementation in main.go
```

## Test

```go
const (
	orderID_001  = "ORD-001"
	orderAmount  = 99.50
)

func TestOrder_Confirm(t *testing.T) {
	o := NewOrder(orderID_001, orderAmount)

	if err := o.Confirm(); err != nil {
		t.Errorf("Order.Confirm() unexpected error = %v", err)
	}
	if o.Status != StatusConfirmed {
		t.Errorf("Order.Confirm() status = %s, want %s", o.Status, StatusConfirmed)
	}
}

func TestOrder_Confirm_AlreadyConfirmed(t *testing.T) {
	o := NewOrder(orderID_001, orderAmount)
	o.Confirm()

	err := o.Confirm()
	if err == nil {
		t.Error("Order.Confirm() expected error for already confirmed order")
	}
}

func TestOrder_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Order)
		wantErr bool
	}{
		{
			name:    orderID_001,  // uses constant as subtest name
			setup:   nil,
			wantErr: false,
		},
		{
			name: "already delivered",
			setup: func(o *Order) { o.Status = StatusDelivered },
			wantErr: true,
		},
		// ... more cases
	}
	// ...
}
```

## Testing Approach

Constant-based subtest naming:

1. **Cross-test consistency** — the constant `orderID_001` is used in `TestOrder_Confirm`, `TestOrder_Cancel`, and could appear in `TestOrder_Ship` and `TestOrder_Deliver`. If the ID format ever changes, update one constant.
2. **Dual purpose** — a constant serves as both a test value (the order ID) and a subtest name (in `TestOrder_Cancel`). This links the value to its test case at a glance.
3. **Domain vocabulary** — constants like `StatusPending`, `StatusConfirmed`, `StatusShipped`, `StatusDelivered`, `StatusCancelled` document the order lifecycle state machine. Tests read as: "in status _Shipped_, Cancel should succeed".
4. **Subtest naming as documentation** — when `t.Run(orderID_001, ...)` runs, it prints `=== RUN   TestOrder_Cancel/ORD-001`. The output immediately tells you which order was being tested.
5. **Combined styles** — the example mixes constant-based names (`orderID_001`) with descriptive strings (`"already delivered"`). Use constants for values that appear across tests, descriptive strings for one-off test scenarios.
