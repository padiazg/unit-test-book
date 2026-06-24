# Chapter 05: Typed Check Functions

## Description

A typed check function is a function type alias for test assertions: `type checkTransferFn func(*testing.T, *TransferResult, error)`. Instead of inline `if` blocks inside each subtest, check functions are defined once and reused across test cases. They become composable, named assertion blocks that separate *what to check* from *how to check it*.

Real-world examples:
- `pantry/internal/core/domain/product_test.go:89` — `checkProductApplyMovementFn`
- `jokes/internal/adapters/secondary/external/joke_client_test.go:11` — `NewJokeClientFn`
- `notifier/model/test_utils.go:10-11` — `TestCheckNotifierFn`, `TestCheckResultFn`

## Code

```go
type TransferResult struct {
	From *Account
	To   *Account
	Fee  float64
}

func Transfer(from, to *Account, amount float64) (*TransferResult, error) {
	// validates accounts, applies fee for checking accounts,
	// updates balances
}
```

## Test

```go
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

	// Test cases use []checkTransferFn to compose assertions
	tests := []struct {
		name   string
		from   *Account
		to     *Account
		amount float64
		checks []checkTransferFn  // <--- slice of check functions
	}{
		{
			name:   "successful transfer",
			from:   &Account{Balance: 1000},
			to:     &Account{Balance: 500},
			amount: 200,
			checks: []checkTransferFn{
				checkSuccess,
				checkFromBalance(800),
				checkToBalance(700),
				checkFee(0),
			},
		},
		// ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Transfer(tt.from, tt.to, tt.amount)
			for _, c := range tt.checks {
				c(t, r, err)   // each check gets (t, result, error)
			}
		})
	}
}
```

## Testing Approach

The typed check function pattern:

1. **Explicit signature** — `type checkTransferFn func(*testing.T, *TransferResult, error)` communicates exactly what the check function receives. Readers know immediately that the production function returns `(*TransferResult, error)`.
2. **Composable assertions** — test cases use `[]checkTransferFn` to mix and match assertions. One case might check success+balances+fee; another might only check error. The test loop runs all checks regardless.
3. **`t.Helper()`** — every check function calls `t.Helper()`, which removes the check function itself from the stack trace on failure. The reported line points directly to the failing test case in the table.
4. **Reusable building blocks** — `checkError(want string)` and `checkSuccess()` can be shared across test functions testing different methods of the same type. No duplication of assertion logic.
