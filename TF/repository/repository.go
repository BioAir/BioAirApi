package repository

type Product struct {
	ID        int     `json:"id,omitempty"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"price"`
}

var products []Product
var nextID = 1

func GetProducts() []Product {
	return products
}

func AddProduct(product Product) int {
	product.ID = nextID
	nextID++
	products = append(products, product)
	return product.ID

}
