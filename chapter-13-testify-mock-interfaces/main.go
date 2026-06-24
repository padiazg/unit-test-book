package testify_mock_interfaces

type Product struct {
	ID    string
	Name  string
	Price float64
}

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

func (s *ProductService) CreateProduct(name string, price float64) (*Product, error) {
	product := &Product{
		ID:    generateID(name),
		Name:  name,
		Price: price,
	}
	if err := s.repo.Save(product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) RemoveProduct(id string) error {
	return s.repo.Delete(id)
}

func generateID(name string) string {
	if name == "" {
		return "unknown"
	}
	return "prod-" + name
}
