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
    return &Order{
        ID:     id,
        Amount: amount,
        Status: StatusPending,
    }
}

func (o *Order) Confirm() error {
    if o.Status != StatusPending {
        return fmt.Errorf("cannot confirm order in status: %s", o.Status)
    }
    o.Status = StatusConfirmed
    return nil
}

func (o *Order) Cancel() error {
    if o.Status == StatusDelivered || o.Status == StatusCancelled {
        return fmt.Errorf("cannot cancel order in status: %s", o.Status)
    }
    o.Status = StatusCancelled
    return nil
}

func (o *Order) Ship() error {
    if o.Status != StatusConfirmed {
        return fmt.Errorf("cannot ship order in status: %s", o.Status)
    }
    o.Status = StatusShipped
    return nil
}

func (o *Order) Deliver() error {
    if o.Status != StatusShipped {
        return fmt.Errorf("cannot deliver order in status: %s", o.Status)
    }
    o.Status = StatusDelivered
    return nil
}
```

## Test

```go
const (
    orderID_001 = "ORD-001"
    orderAmount = 99.50
)

type checkOrderCancelFn func(*testing.T, *Order, error)

var checkOrderCancel = func(fns ...checkOrderCancelFn) []checkOrderCancelFn { return fns }

func checkCancelError(want string) checkOrderCancelFn {
    return func(t *testing.T, s *Order, err error) {
        t.Helper()
        if want == "" {
            assert.NoErrorf(t, err, "checkCancelError: expected no error, got %v", err)
            return
        }
        if assert.Errorf(t, err, "checkCancelError: expected error %q", want) {
            assert.Containsf(t, err.Error(), want, "checkCancelError mismatch")
        }
    }
}

func TestOrder_Cancel(t *testing.T) {
    checkCancelled := func(want OrderStatus) checkOrderCancelFn {
        return func(t *testing.T, o *Order, err error) {
            t.Helper()
            assert.Equal(t, want, o.Status)
        }
    }

    tests := []struct {
        name   string
        checks []checkOrderCancelFn
        id     string
        amount float64
        before func(*Order)
    }{
        {
            name:   orderID_001,
            checks: checkOrderCancel(checkCancelError(""), checkCancelled(StatusCancelled)),
        },
        {
            name:   "confirm then cancel",
            id:     "ORD-002",
            amount: 50.0,
            before: func(o *Order) { o.Confirm() },
            checks: checkOrderCancel(checkCancelError(""), checkCancelled(StatusCancelled)),
        },
        {
            name:   "shipped then cancel",
            id:     "ORD-003",
            amount: 200.0,
            before: func(o *Order) { o.Confirm(); o.Ship() },
            checks: checkOrderCancel(checkCancelError(""), checkCancelled(StatusCancelled)),
        },
        {
            name:   "delivered",
            id:     "ORD-004",
            amount: 30.0,
            before: func(o *Order) { o.Confirm(); o.Ship(); o.Deliver() },
            checks: checkOrderCancel(checkCancelError("cannot cancel order in status: delivered")),
        },
        {
            name:   "already cancelled",
            id:     "ORD-005",
            amount: 75.0,
            before: func(o *Order) { o.Cancel() },
            checks: checkOrderCancel(checkCancelError("cannot cancel order in status: cancelled")),
        },
    }
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            s := NewOrder(tt.id, tt.amount)
            if tt.before != nil {
                tt.before(s)
            }
            err := s.Cancel()
            for _, c := range tt.checks {
                c(t, s, err)
            }
        })
    }
}

type checkOrderConfirmFn func(*testing.T, *Order, error)

var checkOrderConfirm = func(fns ...checkOrderConfirmFn) []checkOrderConfirmFn { return fns }

func checkConfirmError(want string) checkOrderConfirmFn {
    return func(t *testing.T, s *Order, err error) {
        t.Helper()
        if want == "" {
            assert.NoErrorf(t, err, "checkConfirmError: expected no error, got %v", err)
            return
        }
        if assert.Errorf(t, err, "checkConfirmError: expected error %q", want) {
            assert.Containsf(t, err.Error(), want, "checkConfirmError mismatch")
        }
    }
}

func TestOrder_Confirm(t *testing.T) {
    checkConfirmed := func(want OrderStatus) checkOrderConfirmFn {
        return func(t *testing.T, o *Order, err error) {
            t.Helper()
            assert.Equal(t, want, o.Status)
        }
    }

    tests := []struct {
        name   string
        checks []checkOrderConfirmFn
        id     string
        amount float64
        before func(*Order)
    }{
        {
            name:   orderID_001,
            checks: checkOrderConfirm(checkConfirmError(""), checkConfirmed(StatusConfirmed)),
        },
        {
            name:   "shipped",
            id:     "ORD-002",
            amount: 200.0,
            before: func(o *Order) { o.Confirm(); o.Ship() },
            checks: checkOrderConfirm(checkConfirmError("cannot confirm order in status: shipped")),
        },
        {
            name:   "delivered",
            id:     "ORD-003",
            amount: 30.0,
            before: func(o *Order) { o.Confirm(); o.Ship(); o.Deliver() },
            checks: checkOrderConfirm(checkConfirmError("cannot confirm order in status: delivered")),
        },
        {
            name:   "already cancelled",
            id:     "ORD-004",
            amount: 75.0,
            before: func(o *Order) { o.Cancel() },
            checks: checkOrderConfirm(checkConfirmError("cannot confirm order in status: cancelled")),
        },
    }
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            s := NewOrder(tt.id, tt.amount)
            if tt.before != nil {
                tt.before(s)
            }
            err := s.Confirm()
            for _, c := range tt.checks {
                c(t, s, err)
            }
        })
    }
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Order.Cancel
go-testgen gen . Order.Confirm
```

## Testing Approach

Constant-based subtest naming:

1. **Cross-test consistency** — the constant `orderID_001` is used in `TestOrder_Confirm`, `TestOrder_Cancel`, and could appear in `TestOrder_Ship` and `TestOrder_Deliver`. If the ID format ever changes, update one constant.
2. **Dual purpose** — a constant serves as both a test value (the order ID) and a subtest name (in `TestOrder_Cancel`). This links the value to its test case at a glance.
3. **Domain vocabulary** — constants like `StatusPending`, `StatusConfirmed`, `StatusShipped`, `StatusDelivered`, `StatusCancelled` document the order lifecycle state machine. Tests read as: "in status _Shipped_, Cancel should succeed".
4. **Subtest naming as documentation** — when `t.Run(orderID_001, ...)` runs, it prints `=== RUN   TestOrder_Cancel/ORD-001`. The output immediately tells you which order was being tested.
5. **Combined styles** — the example mixes constant-based names (`orderID_001`) with descriptive strings (`"already delivered"`). Use constants for values that appear across tests, descriptive strings for one-off test scenarios.
