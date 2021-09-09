package local_map

import (
	"testing"

	"github.com/mercadolibre/backend-challenge/internal/basket"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	service := New()
	require.Equal(t, service.bktStorage, make(map[string]basket.Basket))
	require.Equal(t, service.prdStorage, productMap)
	require.Equal(t, service.promotions, promotionMap)
}

func TestCreate(t *testing.T) {
	service := New()
	bkt := service.Create()
	require.Equal(t, bkt.Products, make(map[string]int))
	require.Equal(t, bkt.Status, statusActive)
}

func TestGet(t *testing.T) {
	service := New()
	bktAdded := service.Create()
	bktAddedInactive := service.Create()
	err := service.Delete(bktAddedInactive.ID)
	require.NoError(t, err)
	tests := []struct {
		name             string
		bktID            string
		expectedResponse basket.Basket
		expectedErr      error
	}{
		{
			name:             "Get basket - Ok",
			bktID:            bktAdded.ID,
			expectedResponse: bktAdded,
		},
		{
			name:             "Get basket - Not Found",
			bktID:            "randomID",
			expectedResponse: basket.Basket{},
			expectedErr:      ErrBktNotFound,
		},
		{
			name:             "Get basket - Not Found - Status inactive",
			bktID:            bktAddedInactive.ID,
			expectedResponse: basket.Basket{},
			expectedErr:      ErrBktNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.Get(tt.bktID)
			require.Equal(t, response, tt.expectedResponse)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}

func TestGetAmount(t *testing.T) {
	service := New()
	bktAdded := service.Create()
	bktAddedInactive := service.Create()
	err := service.Delete(bktAddedInactive.ID)
	require.NoError(t, err)
	tests := []struct {
		name             string
		bktID            string
		expectedResponse float64
		expectedErr      error
	}{
		{
			name:             "Get amount - Ok",
			bktID:            bktAdded.ID,
			expectedResponse: 0.00,
		},
		{
			name:             "Get amount - Not Found",
			bktID:            "randomID",
			expectedResponse: 0.00,
			expectedErr:      ErrBktNotFound,
		},
		{
			name:             "Get amount - Not Found - Status inactive",
			bktID:            bktAddedInactive.ID,
			expectedResponse: 0.00,
			expectedErr:      ErrBktNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.GetAmount(tt.bktID)
			require.Equal(t, response, tt.expectedResponse)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}

func TestDelete(t *testing.T) {
	service := New()
	bktAdded := service.Create()
	bktAddedInactive := service.Create()
	err := service.Delete(bktAddedInactive.ID)
	require.NoError(t, err)
	tests := []struct {
		name             string
		bktID            string
		expectedResponse error
	}{
		{
			name:  "Delete - Ok",
			bktID: bktAdded.ID,
		},
		{
			name:             "Get amount - Not Found",
			bktID:            "randomID",
			expectedResponse: ErrBktNotFound,
		},
		{
			name:             "Get amount - Not Found - Status inactive",
			bktID:            bktAddedInactive.ID,
			expectedResponse: ErrBktNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.bktID)
			require.Equal(t, err, tt.expectedResponse)
		})
	}
}

func TestAddProduct(t *testing.T) {
	service := New()
	bktAdded1 := service.Create()
	bktAdded2 := service.Create()
	bktAdded3 := service.Create()
	bktAdded4 := service.Create()
	bktAddedDeleted := service.Create()
	err := service.Delete(bktAddedDeleted.ID)
	require.NoError(t, err)
	tests := []struct {
		name           string
		bktID          string
		products       map[string]int
		expectedAmount float64
		expectedError  error
	}{
		{
			name:  "AddProduct - Ok - Case 1",
			bktID: bktAdded1.ID,
			products: map[string]int{
				lanaPenCode:    1,
				lanaTshirtCode: 1,
				lanaMugCode:    1,
			},
			expectedAmount: 32.50,
		},
		{
			name:  "AddProduct - Ok - Case 2",
			bktID: bktAdded2.ID,
			products: map[string]int{
				lanaPenCode:    2,
				lanaTshirtCode: 1,
			},
			expectedAmount: 25,
		},
		{
			name:  "AddProduct - Ok - Case 3",
			bktID: bktAdded3.ID,
			products: map[string]int{
				lanaPenCode:    1,
				lanaTshirtCode: 4,
			},
			expectedAmount: 65,
		},
		{
			name:  "AddProduct - Ok - Case 4",
			bktID: bktAdded4.ID,
			products: map[string]int{
				lanaPenCode:    3,
				lanaTshirtCode: 3,
				lanaMugCode:    1,
			},
			expectedAmount: 62.5,
		},
		{
			name:  "Get amount - Not Found - Status inactive",
			bktID: bktAddedDeleted.ID,
			products: map[string]int{
				lanaPenCode:    3,
				lanaTshirtCode: 3,
				lanaMugCode:    1,
			},
			expectedAmount: 0.00,
			expectedError:  ErrBktNotFound,
		},
		{
			name:  "Get amount - Not Found - Invalid product id",
			bktID: bktAdded1.ID,
			products: map[string]int{
				"randomProductID": 3,
			},
			expectedAmount: 0.00,
			expectedError:  ErrInvalidProductCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bkt basket.Basket
			for productCode, quantity := range tt.products {
				bkt, err = service.AddProduct(tt.bktID, productCode, quantity)
				if tt.expectedError != nil {
					require.Equal(t, err, tt.expectedError)
					require.Equal(t, bkt, basket.Basket{})
					return
				} else {
					require.NotNil(t, bkt)
					require.NoError(t, err)
				}
			}
			require.Equal(t, bkt.Amount, tt.expectedAmount)
		})
	}
}
