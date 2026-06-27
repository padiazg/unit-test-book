package check_factory_closures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkConvertFn func(*testing.T, Temperature, error)

var checkConvert = func(fns ...checkConvertFn) []checkConvertFn { return fns }

func TestConvert(t *testing.T) {
	checkNoError := func() checkConvertFn {
		return func(t *testing.T, _ Temperature, err error) {
			t.Helper()
			assert.NoError(t, err)
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

	tests := []struct {
		name   string
		input  Temperature
		target TemperatureUnit
		checks []checkConvertFn
	}{
		{
			name:   "Celsius to Fahrenheit",
			input:  Temperature{Value: 100, Unit: Celsius},
			target: Fahrenheit,
			checks: checkConvert(
				checkNoError(),
				checkValue(212),
				checkUnit(Fahrenheit),
			),
		},
		{
			name:   "Fahrenheit to Celsius",
			input:  Temperature{Value: 32, Unit: Fahrenheit},
			target: Celsius,
			checks: checkConvert(
				checkNoError(),
				checkValue(0),
				checkUnit(Celsius),
			),
		},
		{
			name:   "Celsius to Kelvin",
			input:  Temperature{Value: 0, Unit: Celsius},
			target: Kelvin,
			checks: checkConvert(
				checkNoError(),
				checkValue(273.15),
				checkUnit(Kelvin),
			),
		},
		{
			name:   "Kelvin to Celsius",
			input:  Temperature{Value: 373.15, Unit: Kelvin},
			target: Celsius,
			checks: checkConvert(
				checkNoError(),
				checkValue(100),
				checkUnit(Celsius),
			),
		},
		{
			name:   "absolute zero in Fahrenheit",
			input:  Temperature{Value: 0, Unit: Kelvin},
			target: Fahrenheit,
			checks: checkConvert(
				checkNoError(),
				checkValue(-459.67),
				checkUnit(Fahrenheit),
			),
		},
		{
			name:   "unknown source unit",
			input:  Temperature{Value: 100, Unit: "R"},
			target: Celsius,
			checks: checkConvert(
				checkError("unknown source unit"),
			),
		},
		{
			name:   "unknown target unit",
			input:  Temperature{Value: 100, Unit: Celsius},
			target: "R",
			checks: checkConvert(
				checkError("unknown target unit"),
			),
		},
		{
			name:   "empty source unit",
			input:  Temperature{Value: 100, Unit: ""},
			target: Celsius,
			checks: checkConvert(
				checkError("source unit is empty"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.input, tt.target)
			for _, c := range tt.checks {
				c(t, got, err)
			}
		})
	}
}
