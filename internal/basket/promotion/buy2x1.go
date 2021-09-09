package promotion

import "github.com/mercadolibre/backend-challenge/internal/basket"

// Buy2Get1Free is responsible for promotion methods.
type Buy2Get1Free struct{}

// Compute calculate the amount for buy2get1free promotion.
func (s *Buy2Get1Free) Compute(product basket.Product, quantity int) float64 {
	if quantity%2 == 0 {
		return product.Price * float64(quantity/2)
	}
	return product.Price * float64((quantity/2)+1)
}
