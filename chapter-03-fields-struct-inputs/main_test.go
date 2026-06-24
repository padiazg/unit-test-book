package fields_struct_inputs

import (
	"testing"
)

func TestProduct_String(t *testing.T) {
	type fields struct {
		Category    string
		Code        string
		Currency    string
		Description string
		Name        string
		UnitPrice   float64
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "full product",
			fields: fields{
				Code: "P001", Name: "Wireless Mouse", Category: "Electronics",
				Description: "Ergonomic wireless mouse", UnitPrice: 29.99, Currency: "USD",
			},
			want: "P001 | Wireless Mouse | Electronics | 29.99 USD",
		},
		{
			name: "minimal product",
			fields: fields{
				Code: "P002", Name: "Notebook", UnitPrice: 3.50, Currency: "EUR",
			},
			want: "P002 | Notebook |  | 3.50 EUR",
		},
		{
			name:   "zero value",
			fields: fields{},
			want:   " |  |  | 0.00 ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Product{
				Code:        tt.fields.Code,
				Name:        tt.fields.Name,
				Category:    tt.fields.Category,
				Description: tt.fields.Description,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("Product.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProduct_FormatShort(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		pname string
		want  string
	}{
		{name: "standard", code: "P001", pname: "Mouse", want: "[P001] Mouse"},
		{name: "empty code", code: "", pname: "Item", want: "[] Item"},
		{name: "empty name", code: "X99", pname: "", want: "[X99] "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Product{Code: tt.code, Name: tt.pname}
			if got := p.FormatShort(); got != tt.want {
				t.Errorf("Product.FormatShort() = %v, want %v", got, tt.want)
			}
		})
	}
}
