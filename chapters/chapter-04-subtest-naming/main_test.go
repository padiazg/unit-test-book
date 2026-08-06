package subtest_naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
