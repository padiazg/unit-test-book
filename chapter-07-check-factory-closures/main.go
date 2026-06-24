package check_factory_closures

import "fmt"

type TemperatureUnit string

const (
	Celsius    TemperatureUnit = "C"
	Fahrenheit TemperatureUnit = "F"
	Kelvin     TemperatureUnit = "K"
)

type Temperature struct {
	Unit  TemperatureUnit
	Value float64
}

func Convert(t Temperature, to TemperatureUnit) (Temperature, error) {
	if t.Unit == "" {
		return Temperature{}, fmt.Errorf("source unit is empty")
	}

	var celsius float64
	switch t.Unit {
	case Celsius:
		celsius = t.Value
	case Fahrenheit:
		celsius = (t.Value - 32) * 5 / 9
	case Kelvin:
		celsius = t.Value - 273.15
	default:
		return Temperature{}, fmt.Errorf("unknown source unit: %s", t.Unit)
	}

	var result float64
	switch to {
	case Celsius:
		result = celsius
	case Fahrenheit:
		result = celsius*9/5 + 32
	case Kelvin:
		result = celsius + 273.15
	default:
		return Temperature{}, fmt.Errorf("unknown target unit: %s", to)
	}

	return Temperature{Value: result, Unit: to}, nil
}
