package package_level_var_swap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id, err := GenerateID()
		require.NoError(t, err)
		assert.Len(t, id, 32) // 16 bytes = 32 hex chars
	})

	t.Run("rand read error", func(t *testing.T) {
		orig := randRead
		randRead = func(b []byte) (n int, err error) {
			return 0, errors.New("crypto failure")
		}
		defer func() { randRead = orig }()

		id, err := GenerateID()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "crypto failure")
		assert.Empty(t, id)
	})

	t.Run("partial read", func(t *testing.T) {
		orig := randRead
		randRead = func(b []byte) (n int, err error) {
			return 2, nil // only wrote 2 of 16 bytes
		}
		defer func() { randRead = orig }()

		id, err := GenerateID()
		require.NoError(t, err)
		assert.Len(t, id, 32)
	})
}

func TestGenerateShortCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		code := GenerateShortCode()
		assert.Len(t, code, 8) // 4 bytes = 8 hex chars
	})

	t.Run("rand read error returns fallback", func(t *testing.T) {
		orig := randRead
		randRead = func(b []byte) (n int, err error) {
			return 0, errors.New("random error")
		}
		defer func() { randRead = orig }()

		code := GenerateShortCode()
		assert.Equal(t, "fallback", code)
	})

	t.Run("restores original after test", func(t *testing.T) {
		orig := randRead
		randRead = func(b []byte) (n int, err error) {
			return 0, errors.New("mock")
		}
		randRead = orig // restore immediately

		code := GenerateShortCode()
		assert.Len(t, code, 8) // original works
	})
}
