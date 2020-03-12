// Package catalog contains all events that the Catalog service
// in the ACME Serverless Fitness Shop can send and receive.
package catalog

import "encoding/json"

const (
	// Domain is the domain where the services reside
	Domain = "Catalog"
)

// Product represents the products as they are stored in the data store
type Product struct {
	// ID is the unique identifier of the product
	ID string `json:"id"`

	// Name is the name of the product
	Name string `json:"name"`

	// ShortDescription is a short description of the product
	// suited for Point of Sales or mobile apps
	ShortDescription string `json:"shortDescription"`

	// Description is a longer description of the product
	// suited for websites
	Description string `json:"description"`

	// ImageURL1 is the location of the first image
	ImageURL1 string `json:"imageUrl1"`

	// ImageURL2 is the location of the second image
	ImageURL2 string `json:"imageUrl2"`

	// ImageURL3 is the location of the third image
	ImageURL3 string `json:"imageUrl3"`

	// Price is the monetary value of the product
	Price float32 `json:"price"`

	// Tags are keys that represent additional sorting
	// information for front-end displays
	Tags []string `json:"tags"`
}

// UnmarshalProduct parses the JSON-encoded data and stores the result
// in a Product
func UnmarshalProduct(data string) (Product, error) {
	var r Product
	err := json.Unmarshal([]byte(data), &r)
	return r, err
}

// Marshal returns the JSON encoding of Product
func (r *Product) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// ProductCreateResponse is the respons that is sent back to the API after a product
// has been created.
type ProductCreateResponse struct {
	// Message is a status message indicating success or failure
	Message string `json:"message"`

	// ResourceID represents the product that has been created
	// together with the new product ID
	ResourceID Product `json:"resourceId"`

	// Status is the HTTP status code indicating success or failure
	Status int `json:"status"`
}

// Marshal returns the JSON encoding of ProductCreateResponse
func (r *ProductCreateResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// AllProductsResponse is the response struct for the reply to
// the API call to get all products.
type AllProductsResponse struct {
	Data []Product `json:"data"`
}

// Marshal returns the JSON encoding of AllProductsResponse
func (r *AllProductsResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
