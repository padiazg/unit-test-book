package inline_check_closures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkValidateFn func(*testing.T, error)

var checkValidate = func(fns ...checkValidateFn) []checkValidateFn { return fns }

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		checks []checkValidateFn
	}{
		{
			name: "valid config",
			cfg:  Config{Host: "example.com", Port: 443, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					assert.NoError(t, err)
				},
			),
		},
		{
			name: "missing host",
			cfg:  Config{Port: 443, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "host is required")
					}
				},
			),
		},
		{
			name: "zero port",
			cfg:  Config{Host: "example.com", Port: 0, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "port must be between 1 and 65535")
					}
				},
			),
		},
		{
			name: "port too high",
			cfg:  Config{Host: "example.com", Port: 70000, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "port must be between 1 and 65535")
					}
				},
			),
		},
		{
			name: "missing api key",
			cfg:  Config{Host: "example.com", Port: 443, Key: ""},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "api key is required")
					}
				},
			),
		},
		{
			name: "multiple errors returns first",
			cfg:  Config{},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "host is required")
						assert.NotContains(t, err.Error(), "port")
					}
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

// ────────────────────────────────────────────────────────────────
// Over-factored version — same tests with unnecessary factory
// extraction. Each factory is used once yet wrapped in a named
// closure. Compare with the inline version above: same logic,
// more ceremony.
// ────────────────────────────────────────────────────────────────

// var checkValidate = func(fns ...checkValidateFn) []checkValidateFn { return fns }
//
// func checkSuccess() checkValidateFn {
// 	return func(t *testing.T, err error) {
// 		t.Helper()
// 		assert.NoError(t, err)
// 	}
// }
//
// func checkError(want string) checkValidateFn {
// 	return func(t *testing.T, err error) {
// 		t.Helper()
// 		if assert.Error(t, err) {
// 			assert.Contains(t, err.Error(), want)
// 		}
// 	}
// }
//
// func checkErrorNotContains(want string) checkValidateFn {
// 	return func(t *testing.T, err error) {
// 		t.Helper()
// 		if assert.Error(t, err) {
// 			assert.NotContains(t, err.Error(), want)
// 		}
// 	}
// }
//
// func TestValidateConfig_overfactored(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		cfg    Config
// 		checks []checkValidateFn
// 	}{
// 		{name: "valid", cfg: Config{Host: "e.com", Port: 443, Key: "sk-abc"}, checks: checkValidate(checkSuccess())},
// 		{name: "missing host", cfg: Config{Port: 443, Key: "sk-abc"}, checks: checkValidate(checkError("host is required"))},
// 		{name: "zero port", cfg: Config{Host: "e.com", Port: 0, Key: "sk-abc"}, checks: checkValidate(checkError("port must be between 1 and 65535"))},
// 		{name: "port high", cfg: Config{Host: "e.com", Port: 70000, Key: "sk-abc"}, checks: checkValidate(checkError("port must be between 1 and 65535"))},
// 		{name: "missing key", cfg: Config{Host: "e.com", Port: 443, Key: ""}, checks: checkValidate(checkError("api key is required"))},
// 		{name: "first error only", cfg: Config{}, checks: checkValidate(checkError("host is required"), checkErrorNotContains("port"))},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := ValidateConfig(tt.cfg)
// 			for _, c := range tt.checks {
// 				c(t, err)
// 			}
// 		})
// 	}
// }
