package subtest_naming

import "fmt"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID     string
	Amount float64
	Status OrderStatus
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
