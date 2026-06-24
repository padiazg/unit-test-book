package fields_struct_inputs

import "fmt"

type Product struct {
	Code        string
	Name        string
	Category    string
	Description string
	UnitPrice   float64
	Currency    string
}

func (p Product) String() string {
	return fmt.Sprintf("%s | %s | %s | %.2f %s",
		p.Code, p.Name, p.Category, p.UnitPrice, p.Currency)
}

func (p Product) FormatShort() string {
	return fmt.Sprintf("[%s] %s", p.Code, p.Name)
}
