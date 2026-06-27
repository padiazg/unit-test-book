# Chapter 13: testify/mock Interfaces

## Description

Use `github.com/stretchr/testify/mock` to generate interface mocks at test time. Embed `mock.Mock` into a struct implementing your interface, then use `On("Method", args...).Return(values...)` to set expectations and `AssertExpectations(t)` to verify every expected call was made. No code generation, no mockgen, just testify.

Real-world example: `pantry/internal/adapters/primary/http/product_handler_test.go:199` — `mockProductService` with testify/mock.

## Code

```go
type ProductRepository interface {
	FindByID(id string) (*Product, error)
	Save(product *Product) error
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetProduct(id string) (*Product, error) {
	// validates input, applies business logic, then delegates to repo
	if id == "" {
		return nil, fmt.Errorf("product ID required")
	}
	return s.repo.FindByID(id)
}
```

## Test

```go
type mockProductRepository struct {
	mock.Mock
}

func (m *mockProductRepository) FindByID(id string) (*Product, error) {
	args := m.Called(id)
	return args.Get(0).(*Product), args.Error(1)
}

func (m *mockProductRepository) Save(product *Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func TestProductService_GetProduct(t *testing.T) {
	tests := []struct {
		name   string
		prodID string
		mockFn func(*mockProductRepository)
		checks []checkProductServiceFn
	}{
		{
			name:   "product found",
			prodID: "prod-1",
			mockFn: func(m *mockProductRepository) {
				m.On("FindByID", "prod-1").Return(&Product{ID: "prod-1", Name: "Widget", Price: 9.99}, nil)
			},
			checks: checkProductService(checkProduct("prod-1")),
		},
		{
			name:   "product not found",
			prodID: "prod-42",
			mockFn: func(m *mockProductRepository) {
				m.On("FindByID", "prod-42").Return(nil, errors.New("not found"))
			},
			checks: checkProductService(checkError("not found")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProductRepository{}
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}
			svc := NewProductService(mockRepo)
			p, err := svc.GetProduct(tt.prodID)
			for _, fn := range tt.checks {
				fn(t, p, err)
			}
			mockRepo.AssertExpectations(t) // verifies all expected calls happened
		})
	}
}
```

## Testing Approach

testify/mock:

1. **Explicit expectations** — `m.On("FindByID", "prod-1").Return(...)` documents exactly what call is expected with what argument. The mock panics on unexpected calls — catching bugs fast.
2. **`AssertExpectations`** — the final line in each test case verifies every `On(...)` was actually called. Missed assertions (e.g. a cached result skips the repo call) are caught.
3. **Typed return helpers** — `args.Get(0).(*Product)` extracts the first return value with a type assertion. testify doesn't know your return types; this cast is the standard pattern.
4. **Per-table mock setup** — `mockFn` closures configure expectations inline in each table row. The mock is fresh for every case via `&mockProductRepository{}` in the loop.
