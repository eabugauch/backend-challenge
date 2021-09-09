package basket

// Basket represents the Basket response.
type Basket struct {
	ID              string         `json:"id"`
	Products        map[string]int `json:"products"`
	Amount          float64        `json:"total_amount"`
	DateCreated     string         `json:"date_created"`
	DateLastUpdated string         `json:"date_last_updated"`
	Status          string         `json:"-"` // Active or Inactive
}

// AddProduct represents the AddProduct request.
type AddProduct struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

// Product is used to store the information of each product.
type Product struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// GetAmount represents the GetAmount response.
type GetAmount struct {
	BktID  string  `json:"basket_id"`
	Amount float64 `json:"amount"`
}
