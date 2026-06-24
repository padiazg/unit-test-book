# Chapter 03: Fields Struct for Inputs

## Description

When a constructor or method takes many parameters, inline all of them in the test case table makes rows wide and hard to read. The solution: define a `fields` struct type in the test that groups related inputs. Each test case has a `fields` field, and the test body unpacks it into the SUT constructor.

Real-world example:
- `hexago/pkg/version/version_test.go:141` — `TestVersionString` uses `type fields struct { Version, Commit, BuildDate string }`

## Code

```go
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
```

## Test

```go
func TestProduct_String(t *testing.T) {
	type fields struct {
		Code        string
		Name        string
		Category    string
		Description string
		UnitPrice   float64
		Currency    string
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
```

## Testing Approach

The `fields` struct pattern solves a readability problem:

1. **Named grouping** — the `fields` type gives a name to the input parameter group. When the same struct appears in multiple test functions, it communicates "these inputs belong together."
2. **Compact tables** — without the `fields` struct, the test case would need 6 inline fields for `Code`, `Name`, `Category`, `Description`, `UnitPrice`, `Currency`. The `fields` struct wraps them into one column, keeping the table at 3-4 columns.
3. **Partial initialization** — Go's zero-value initialization means you only set the fields you care about. The "minimal product" case only sets 4 fields; the rest get zero values.
4. **No external dependency** — the `fields` type is declared inside the test file (often inside the test function for single-use cases). It's not exported and doesn't pollute the production API.
