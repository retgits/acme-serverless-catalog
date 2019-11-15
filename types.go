package catalog

import "encoding/json"

type Product struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ShortDescription string   `json:"shortDescription"`
	Description      string   `json:"description"`
	ImageURL1        string   `json:"imageUrl1"`
	ImageURL2        string   `json:"imageUrl2"`
	ImageURL3        string   `json:"imageUrl3"`
	Price            float32  `json:"price"`
	Tags             []string `json:"tags"`
}

func UnmarshalProduct(data string) (Product, error) {
	var r Product
	err := json.Unmarshal([]byte(data), &r)
	return r, err
}

func (r *Product) Marshal() (string, error) {
	s, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func (r *ProductCreateResponse) Marshal() (string, error) {
	s, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

type ProductCreateResponse struct {
	Message    string  `json:"message"`
	ResourceID Product `json:"resourceId"`
	Status     int     `json:"status"`
}

func (r *AllProductsResponse) Marshal() (string, error) {
	s, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

type AllProductsResponse struct {
	Data []Product `json:"data"`
}
