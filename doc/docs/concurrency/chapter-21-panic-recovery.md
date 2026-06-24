# Chapter 21: Panic Recovery

## Description

Test panic recovery with `recover()` inside `defer` to convert panics into errors. Production code uses `defer/recover` to prevent crashes from division by zero, nil pointer dereference, or invalid input. Tests exercise both the panic path (via `assert.Panics`) and the recovered-error path (via `SafeDivide`, `SafeExecute`).

## Code

```go
var ErrPanic = errors.New("panic recovered")

func SafeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanic, r)
		}
	}()
	return a / b, nil
}

func SafeExecute(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanic, r)
		}
	}()
	fn()
	return nil
}
```

## Test

```go
func TestSafeDivide(t *testing.T) {
	t.Run("normal division", func(t *testing.T) {
		r, err := SafeDivide(10, 3)
		require.NoError(t, err)
		assert.Equal(t, 3, r)
	})
	t.Run("division by zero", func(t *testing.T) {
		_, err := SafeDivide(1, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPanic)
		assert.Contains(t, err.Error(), "integer divide by zero")
	})
}

func TestMustParse(t *testing.T) {
	t.Run("valid number", func(t *testing.T) {
		n, err := MustParse("42")
		require.NoError(t, err)
		assert.Equal(t, 42, n)
	})
	t.Run("empty string", func(t *testing.T) {
		_, err := MustParse("")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPanic)
	})
	t.Run("invalid character", func(t *testing.T) {
		_, err := MustParse("12a34")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPanic)
	})
}

func TestPanicIfNegative(t *testing.T) {
	assert.Equal(t, 10, PanicIfNegative(5))
	assert.Panics(t, func() { PanicIfNegative(-1) })
	assert.PanicsWithValue(t, "negative value not allowed", func() {
		PanicIfNegative(-5)
	})
}

func TestSafeExecute(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		called := false
		err := SafeExecute(func() { called = true })
		require.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("panic recovered", func(t *testing.T) {
		err := SafeExecute(func() { panic("boom") })
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPanic)
	})
}
```

## Testing Approach

Panic recovery testing:

1. **Named returns for defer** — `(result int, err error)` gives the deferred closure access to the return values. The deferred `recover` sets `err` before the function returns.
2. **`assert.Panics` for expected panics** — `PanicIfNegative(-1)` should panic. `assert.Panics` catches it and verifies the panic happened. `PanicsWithValue` checks the exact panic value.
3. **`assert.ErrorIs` for sentinel errors** — `ErrPanic` wraps the original panic value. Tests assert `ErrorIs(ErrPanic)` to confirm recovery happened, plus `Contains` for the underlying reason.
4. **Safe wrapper pattern** — `SafeExecute(fn)` demonstrates recovering a panic from any function. Useful for wrapping third-party code or callbacks that might panic.
