package promotion

import "github.com/mercadolibre/backend-challenge/internal/basket"

const (
	lanaTshirtQuantityStrategy = 3
	lanaTshirtDiscount         = 0.25
)

// BuyXOrMore is responsible for promotion methods.
type BuyXOrMore struct{}

// Compute calculate the amount for buyXOrMore promotion.
func (s *BuyXOrMore) Compute(product basket.Product, quantity int) float64 {
	if quantity >= lanaTshirtQuantityStrategy {
		return (product.Price * float64(quantity)) - ((product.Price * float64(quantity)) * lanaTshirtDiscount)
	}
	return product.Price * float64(quantity)
}
