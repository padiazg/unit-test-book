# Chapter 07: Check Factory Closures

## Description

A check factory is a function that takes expected values as parameters and returns a check closure: `checkValue(want float64) checkConvertFn { return func(t *testing.T, got Temperature, err error) { assert.Equal(t, want, got.Value) } }`. The closure *captures* the expected value at test-definition time and asserts against the actual result at test-execution time.

Real-world examples:

- `pantry/internal/core/domain/product_test.go:93-103` — `checkApplyMovementError(want string)`  
- `go-crap/internal/report/table_test.go:21-26` — `checkOutputContains(want string)`  
- `go-crap/internal/report/json_test.go:100-111` — `checkReportSchema`, `checkReportVersion`  
- `notifier/connector/webhook/webhook_test.go:28-43` — `checkName(name string)`  

## Code

```go
type Temperature struct {
	Value float64
	Unit  TemperatureUnit
}

func Convert(t Temperature, to TemperatureUnit) (Temperature, error) {
	// converts between Celsius, Fahrenheit, Kelvin
}
```

## Test

```go
type checkConvertFn func(*testing.T, Temperature, error)

var checkConvert = func(fns ...checkConvertFn) []checkConvertFn { return fns }

func TestConvert(t *testing.T) {
	// Factory: takes expected value, returns check closure
	checkValue := func(want float64) checkConvertFn {
		return func(t *testing.T, got Temperature, err error) {
			t.Helper()
			assert.NoError(t, err)
			assert.InDelta(t, want, got.Value, 0.01)
		}
	}

	checkUnit := func(want TemperatureUnit) checkConvertFn {
		return func(t *testing.T, got Temperature, err error) {
			t.Helper()
			assert.NoError(t, err)
			assert.Equal(t, want, got.Unit)
		}
	}

	checkError := func(want string) checkConvertFn {
		return func(t *testing.T, _ Temperature, err error) {
			t.Helper()
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), want)
			}
		}
	}

	tests := []struct {
		name     string
		input    Temperature
		target   TemperatureUnit
		checks   []checkConvertFn
	}{
		{
			name:   "Celsius to Fahrenheit",
			input:  Temperature{Value: 100, Unit: Celsius},
			target: Fahrenheit,
			checks: checkConvert(
				checkValue(212),       // closure captures 212
				checkUnit(Fahrenheit),  // closure captures Fahrenheit
			),
		},
		// ...
	}
}
```

## Testing Approach

Check factory closures are the core of the closure-check pattern:

1. **Closure captures expected** — each call to `checkValue(212)` creates a closure where `want` is bound to `212`. The closure is stored in the `checks` slice and executed later with the actual result. This is the fundamental Go pattern for deferred assertion logic.
2. **Parameterized reuse** — `checkValue`, `checkUnit`, `checkError` are defined once and reused across all test cases. A 9-case test table with 3 assertions each = 27 assertions, but only 3 factory functions.
3. **`assert.InDelta` for floats** — the temperature converter uses `assert.InDelta(t, want, got.Value, 0.01)` instead of `assert.Equal` to handle floating-point precision issues across conversions.
4. **Composability** — factories compose naturally: `checkValue + checkUnit` for success cases, `checkError` alone for failure cases. The presence or absence of certain checks in a test case documents what that case exercises.
