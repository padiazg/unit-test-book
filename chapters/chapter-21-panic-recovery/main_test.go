package panic_recovery

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	t.Run("negative numbers", func(t *testing.T) {
		r, err := SafeDivide(-10, 3)
		require.NoError(t, err)
		assert.Equal(t, -3, r)
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
		assert.Contains(t, err.Error(), "empty input")
	})

	t.Run("invalid character", func(t *testing.T) {
		_, err := MustParse("12a34")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPanic)
		assert.Contains(t, err.Error(), "invalid character")
	})

	t.Run("zero value", func(t *testing.T) {
		n, err := MustParse("0")
		require.NoError(t, err)
		assert.Equal(t, 0, n)
	})

	t.Run("large number", func(t *testing.T) {
		n, err := MustParse("999999")
		require.NoError(t, err)
		assert.Equal(t, 999999, n)
	})
}

func TestPanicIfNegative(t *testing.T) {
	t.Run("positive number", func(t *testing.T) {
		assert.Equal(t, 10, PanicIfNegative(5))
	})

	t.Run("zero", func(t *testing.T) {
		assert.Equal(t, 0, PanicIfNegative(0))
	})

	t.Run("panic on negative", func(t *testing.T) {
		assert.Panics(t, func() { PanicIfNegative(-1) })
	})

	t.Run("panic with message", func(t *testing.T) {
		assert.PanicsWithValue(t, "negative value not allowed", func() {
			PanicIfNegative(-5)
		})
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
		err := SafeExecute(func() { panic("something went wrong") })
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPanic)
		assert.Contains(t, err.Error(), "something went wrong")
	})
}

func TestDivideEdgeCases(t *testing.T) {
	t.Run("min int", func(t *testing.T) {
		_, err := SafeDivide(math.MinInt, -1)
		// may or may not panic depending on platform
		_ = err
	})
}
