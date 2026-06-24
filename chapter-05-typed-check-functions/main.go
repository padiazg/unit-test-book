package typed_check_functions

import "errors"

type AccountType string

const (
	Savings  AccountType = "savings"
	Checking AccountType = "checking"
)

type Account struct {
	Type    AccountType
	ID      string
	Owner   string
	Balance float64
}

type TransferResult struct {
	From *Account
	To   *Account
	Fee  float64
}

func Transfer(from, to *Account, amount float64) (*TransferResult, error) {
	if from == nil || to == nil {
		return nil, errors.New("both accounts are required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if from.Balance < amount {
		return nil, errors.New("insufficient funds")
	}

	fee := 0.0
	if from.Type == Checking {
		fee = amount * 0.01
	}

	from.Balance -= amount + fee
	to.Balance += amount

	return &TransferResult{From: from, To: to, Fee: fee}, nil
}
