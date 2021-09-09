package promotion

import "github.com/mercadolibre/backend-challenge/internal/basket"

// WithoutPromo is responsible for promotion methods.
type WithoutPromo struct{}

// Compute calculate the amount for withoutPromo promotion.
func (s *WithoutPromo) Compute(product basket.Product, quantity int) float64 {
	return product.Price * float64(quantity)
}
