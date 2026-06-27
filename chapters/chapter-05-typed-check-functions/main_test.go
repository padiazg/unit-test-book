package typed_check_functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkTransferFn func(*testing.T, *TransferResult, error)

func TestTransfer(t *testing.T) {
	checkSuccess := func(t *testing.T, r *TransferResult, err error) {
		t.Helper()
		assert.NoError(t, err)
		assert.NotNil(t, r)
	}

	checkError := func(want string) checkTransferFn {
		return func(t *testing.T, r *TransferResult, err error) {
			t.Helper()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), want)
			assert.Nil(t, r)
		}
	}

	checkFromBalance := func(want float64) checkTransferFn {
		return func(t *testing.T, r *TransferResult, err error) {
			t.Helper()
			assert.NoError(t, err)
			assert.Equal(t, want, r.From.Balance)
		}
	}

	checkToBalance := func(want float64) checkTransferFn {
		return func(t *testing.T, r *TransferResult, err error) {
			t.Helper()
			assert.NoError(t, err)
			assert.Equal(t, want, r.To.Balance)
		}
	}

	checkFee := func(want float64) checkTransferFn {
		return func(t *testing.T, r *TransferResult, err error) {
			t.Helper()
			assert.NoError(t, err)
			assert.Equal(t, want, r.Fee)
		}
	}

	tests := []struct {
		name   string
		from   *Account
		to     *Account
		amount float64
		checks []checkTransferFn
	}{
		{
			name:   "successful transfer",
			from:   &Account{ID: "1", Owner: "Alice", Type: Savings, Balance: 1000},
			to:     &Account{ID: "2", Owner: "Bob", Type: Savings, Balance: 500},
			amount: 200,
			checks: []checkTransferFn{
				checkSuccess,
				checkFromBalance(800),
				checkToBalance(700),
				checkFee(0),
			},
		},
		{
			name:   "checking account fee applied",
			from:   &Account{ID: "3", Owner: "Charlie", Type: Checking, Balance: 1000},
			to:     &Account{ID: "4", Owner: "Diana", Type: Savings, Balance: 500},
			amount: 200,
			checks: []checkTransferFn{
				checkSuccess,
				checkFromBalance(798),
				checkToBalance(700),
				checkFee(2),
			},
		},
		{
			name:   "insufficient funds",
			from:   &Account{ID: "5", Owner: "Eve", Type: Savings, Balance: 50},
			to:     &Account{ID: "6", Owner: "Frank", Type: Savings, Balance: 100},
			amount: 100,
			checks: []checkTransferFn{
				checkError("insufficient funds"),
			},
		},
		{
			name:   "zero amount",
			from:   &Account{ID: "7", Owner: "Grace", Type: Savings, Balance: 1000},
			to:     &Account{ID: "8", Owner: "Hank", Type: Savings, Balance: 500},
			amount: 0,
			checks: []checkTransferFn{
				checkError("amount must be positive"),
			},
		},
		{
			name:   "nil from account",
			from:   nil,
			to:     &Account{ID: "9", Owner: "Ivy", Type: Savings, Balance: 500},
			amount: 100,
			checks: []checkTransferFn{
				checkError("both accounts are required"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Transfer(tt.from, tt.to, tt.amount)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}
