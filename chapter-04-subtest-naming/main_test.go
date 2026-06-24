package subtest_naming

import (
	"testing"
)

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
	o.Confirm() // move to confirmed

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
			name:    orderID_001,
			setup:   nil, // fresh order, status = pending
			wantErr: false,
		},
		{
			name: "already delivered",
			setup: func(o *Order) {
				o.Status = StatusDelivered
			},
			wantErr: true,
		},
		{
			name: "already cancelled",
			setup: func(o *Order) {
				o.Status = StatusCancelled
			},
			wantErr: true,
		},
		{
			name: "shipped order",
			setup: func(o *Order) {
				o.Status = StatusShipped
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrder(orderID_001, orderAmount)
			if tt.setup != nil {
				tt.setup(o)
			}

			err := o.Cancel()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Order.Cancel() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("Order.Cancel() unexpected error = %v", err)
			}
			if o.Status != StatusCancelled {
				t.Errorf("Order.Cancel() status = %s, want %s", o.Status, StatusCancelled)
			}
		})
	}
}
