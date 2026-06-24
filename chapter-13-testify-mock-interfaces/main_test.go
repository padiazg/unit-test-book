package testify_mock_interfaces

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func (m *mockProductRepository) Save(product *Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type checkProductServiceFn func(*testing.T, *Product, error)

var checkProductService = func(fns ...checkProductServiceFn) []checkProductServiceFn { return fns }

func TestProductService_GetProduct(t *testing.T) {
	checkError := func(want string) checkProductServiceFn {
		return func(t *testing.T, _ *Product, err error) {
			t.Helper()
			require.Error(t, err)
			assert.Contains(t, err.Error(), want)
		}
	}

	checkProduct := func(want string) checkProductServiceFn {
		return func(t *testing.T, p *Product, err error) {
			t.Helper()
			require.NoError(t, err)
			assert.NotNil(t, p)
			assert.Equal(t, want, p.ID)
		}
	}

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
			checks: checkProductService(
				checkProduct("prod-1"),
			),
		},
		{
			name:   "product not found",
			prodID: "prod-42",
			mockFn: func(m *mockProductRepository) {
				m.On("FindByID", "prod-42").Return(nil, errors.New("not found"))
			},
			checks: checkProductService(
				checkError("not found"),
			),
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

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name       string
		productName string
		price      float64
		mockFn     func(*mockProductRepository)
		wantErr    bool
	}{
		{
			name:        "success",
			productName: "Laptop",
			price:       1299.99,
			mockFn: func(m *mockProductRepository) {
				m.On("Save", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "save fails",
			productName: "Tablet",
			price:       499.99,
			mockFn: func(m *mockProductRepository) {
				m.On("Save", mock.Anything).Return(errors.New("connection error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProductRepository{}
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}

			svc := NewProductService(mockRepo)
			p, err := svc.CreateProduct(tt.productName, tt.price)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, p)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, tt.productName, p.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
