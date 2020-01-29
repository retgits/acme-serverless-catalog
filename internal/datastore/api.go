package datastore

import catalog "github.com/retgits/acme-serverless-catalog"

// Manager ...
type Manager interface {
	AddProduct(p catalog.Product) error
	GetProduct(productID string) (catalog.Product, error)
	GetProducts() ([]catalog.Product, error)
}
