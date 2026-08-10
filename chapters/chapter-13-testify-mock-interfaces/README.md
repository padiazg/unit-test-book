---
section: mocking
---

# Chapter 13: testify/mock Interfaces

## Description

Use `github.com/stretchr/testify/mock` to generate interface mocks at test time. Embed `mock.Mock` into a struct implementing your interface, then use `On("Method", args...).Return(values...)` to set expectations and `AssertExpectations(t)` to verify every expected call was made. No code generation, no mockgen, just testify.

Real-world example: `pantry/internal/adapters/primary/http/product_handler_test.go:199` — `mockProductService` with testify/mock.

## Code

```go
type ProductRepository interface {
    FindByID(id string) (*Product, error)
    Save(product *Product) error
    Delete(id string) error
}

type ProductService struct {
    repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
    return &ProductService{repo: repo}
}

func (s *ProductService) GetProduct(id string) (*Product, error) {
    return s.repo.FindByID(id)
}
```

The interface is the seam. Production retrieves a product by delegating to the repository; tests swap the real repository for a testify mock.

## Test

```go
type mockProductRepository struct {
    mock.Mock
}

func (m *mockProductRepository) FindByID(id string) (*Product, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Product), args.Error(1)
}
```

Every method on the mock forwards to `m.Called(...)` so testify records the call and matches it against expectations. The `nil` guard matters: testify returns `nil` when no expectation is set, and the type assertion `args.Get(0).(*Product)` would panic on it.

```go
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

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Service.GetProduct
go-testgen gen . Service.CreateProduct
```

## Testing Approach

testify/mock:

1. **Explicit expectations** — `m.On("FindByID", "prod-1").Return(...)` documents exactly what call is expected with what argument. The mock panics on unexpected calls — catching bugs fast.
2. **`AssertExpectations`** — the final line in each test case verifies every `On(...)` was actually called. Missed assertions (e.g. a cached result skips the repo call) are caught.
3. **Typed return helpers** — `args.Get(0).(*Product)` extracts the first return value with a type assertion. testify doesn't know your return types; this cast is the standard pattern.
4. **Per-table mock setup** — `mockFn` closures configure expectations inline in each table row. The mock is fresh for every case via `&mockProductRepository{}` in the loop.

## Pros and Cons

### Pros

1. **Works out of the box** — no mock templates, no code generation. Embed `mock.Mock`, forward each method to `m.Called(...)`, and the framework handles recording, matching, and verification.
2. **One `On(...).Return(...)` line per scenario** — declaring the expected call with its arguments and canned result is concise and reads as a spec of the interaction.
3. **Strict verification built in** — `AssertExpectations(t)` fails the test if a declared expectation never fired, catching dead code paths and silent behavior changes.

### Cons

1. **Steep learning curve** — `On`, `Return`, `Called`, `args.Get(...)`, `mock.Anything`, `AssertExpectations` all need to be understood before the mock behaves as intended. Not something you pick up casually.
2. **Strictness fights legitimate flow** — the mock panics on any call you didn't declare with the exact arguments expected. A call that is perfectly valid in the production logic flow (a retry, a subtler argument, a second repository touch) throws an error unless you happened to anticipate it in the `On(...)` setup.
3. **Setup complexity grows with branching logic** — every code path the method under test can take needs its own expectation. Tests of less predictable logic become a long list of `On(...)` declarations just to keep the mock from exploding.
4. **Type assertions are manual and panic-prone** — `args.Get(0).(*Product)` requires the nil guard shown above; forget it and the "nothing returned" case crashes your test instead of failing cleanly.
5. **Overkill for simple stubs** — for a fixed response with no call verification, a hand-written stub (chapters 11–12) is six lines and zero framework knowledge.
