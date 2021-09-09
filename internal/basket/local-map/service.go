package local_map

import (
	"errors"
	"sync"
	"time"

	"github.com/mercadolibre/backend-challenge/internal/basket"
	"github.com/mercadolibre/backend-challenge/internal/basket/promotion"
	"github.com/rs/xid"
)

const (
	lanaPenCode    = "PEN"
	lanaTshirtCode = "TSHIRT"
	lanaMugCode    = "MUG"
	statusActive   = "active"
	statusInactive = "inactive"
)

var (
	productMap = map[string]basket.Product{
		lanaPenCode:    {Code: lanaPenCode, Name: "Lana Pen", Price: 5.00},
		lanaTshirtCode: {Code: lanaTshirtCode, Name: "Lana T-Shirt", Price: 20.00},
		lanaMugCode:    {Code: lanaMugCode, Name: "Lana Coffee Mug ", Price: 7.50},
	}
	promotionMap = map[string]Promotion{
		lanaPenCode:    &promotion.Buy2Get1Free{},
		lanaTshirtCode: &promotion.BuyXOrMore{},
	}
)

var (
	// ErrBktNotFound is used when the basket_id is not found on the map.
	ErrBktNotFound = errors.New("basket not found")
	// ErrInvalidProductCode is used when the product code does not belong to one supported
	ErrInvalidProductCode = errors.New("invalid product code")
)

// Service is responsible for service methods.
type Service struct {
	bktMutex   sync.Mutex
	bktStorage map[string]basket.Basket
	prdStorage map[string]basket.Product
	promotions map[string]Promotion
}

// New returns a Service implementation.
func New() *Service {
	return &Service{
		bktStorage: make(map[string]basket.Basket),
		prdStorage: uploadProducts(),
		promotions: buildPromotion(),
	}
}

func buildPromotion() map[string]Promotion {
	return promotionMap
}

func uploadProducts() map[string]basket.Product {
	return productMap
}

// Create creates a basket with empty values.
func (s *Service) Create() basket.Basket {
	s.bktMutex.Lock()
	defer s.bktMutex.Unlock()

	bkt := buildBkt()

	s.bktStorage[bkt.ID] = bkt
	return bkt
}

func buildBkt() basket.Basket {
	return basket.Basket{
		ID:          xid.New().String(),
		DateCreated: time.Now().UTC().Format("01-02-2006 15:04:05"),
		Products:    make(map[string]int),
		Status:      statusActive,
	}
}

func (s *Service) Get(bktID string) (basket.Basket, error) {
	s.bktMutex.Lock()
	defer s.bktMutex.Unlock()
	bkt, exist := s.bktStorage[bktID]
	if !exist || bkt.Status == statusInactive {
		return basket.Basket{}, ErrBktNotFound
	}
	return bkt, nil
}

// Promotion interface is used to manage the Promotion methods.
type Promotion interface {
	Compute(basket basket.Product, quantity int) float64
}

// GetAmount returns the amount of the basket and an error if any.
func (s *Service) GetAmount(bktID string) (float64, error) {
	s.bktMutex.Lock()
	defer s.bktMutex.Unlock()
	bkt, exist := s.bktStorage[bktID]
	if !exist || bkt.Status == statusInactive {
		return 0, ErrBktNotFound
	}
	return bkt.Amount, nil
}

func (s *Service) calculateAmount(bktID string) (float64, error) {
	withoutPromo := promotion.WithoutPromo{}

	bkt, exist := s.bktStorage[bktID]
	if !exist {
		return 0, ErrBktNotFound
	}
	var amount float64
	for productCode, quantity := range bkt.Products {
		promo, exists := s.promotions[productCode]
		if !exists {
			amount += withoutPromo.Compute(s.prdStorage[productCode], quantity)
		} else {
			amount += promo.Compute(s.prdStorage[productCode], quantity)
		}
	}
	return amount, nil
}

// Delete deletes the basket sent by parameter.
func (s *Service) Delete(bktID string) error {
	s.bktMutex.Lock()
	defer s.bktMutex.Unlock()
	bkt, exist := s.bktStorage[bktID]
	if !exist || bkt.Status == statusInactive {
		return ErrBktNotFound
	}

	// TODO add validation of status transitions
	bkt.Status = statusInactive
	s.bktStorage[bktID] = bkt
	return nil
}

// AddProduct add a product to the basket.
func (s *Service) AddProduct(bktID string, prdID string, quantity int) (basket.Basket, error) {
	s.bktMutex.Lock()
	defer s.bktMutex.Unlock()
	bkt, exist := s.bktStorage[bktID]
	if !exist || bkt.Status == statusInactive {
		return basket.Basket{}, ErrBktNotFound
	}

	_, exist = s.prdStorage[prdID]
	if !exist {
		return basket.Basket{}, ErrInvalidProductCode
	}

	bkt.Products[prdID] += quantity
	bkt.DateLastUpdated = time.Now().UTC().Format("01-02-2006 15:04:05")
	amount, err := s.calculateAmount(bktID)
	if err != nil {
		return basket.Basket{}, err
	}
	bkt.Amount = amount
	s.bktStorage[bktID] = bkt
	return bkt, nil
}
