package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/mercadolibre/backend-challenge/internal/basket"
	localMap "github.com/mercadolibre/backend-challenge/internal/basket/local-map"
	localLib "github.com/mercadolibre/backend-challenge/local-library"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	XClientKey       = "x-client-key"
	XClientKeyValue  = "admin1"
	XClientKeyEnvVar = "X_CLIENT_KEY"
)

var (
	errForbidden = localLib.Error{Message: "Forbidden", StatusCode: http.StatusForbidden}
	bktCreated   = basket.Basket{
		ID: "RANDOM123",
		Products: map[string]int{
			"PEN": 1,
		},
		Amount:          5.0,
		DateCreated:     time.Now().String(),
		DateLastUpdated: time.Now().String(),
	}
)

type ServiceBktMock struct {
	mock.Mock
}

func (s *ServiceBktMock) Create() basket.Basket {
	args := s.Called()
	return args.Get(0).(basket.Basket)
}

func (s *ServiceBktMock) Get(_ string) (basket.Basket, error) {
	args := s.Called()
	return args.Get(0).(basket.Basket), args.Error(1)
}

func (s *ServiceBktMock) GetAmount(_ string) (float64, error) {
	args := s.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (s *ServiceBktMock) Delete(_ string) error {
	args := s.Called()
	return args.Error(0)
}

func (s *ServiceBktMock) AddProduct(_ string, _ string, _ int) (basket.Basket, error) {
	args := s.Called()
	return args.Get(0).(basket.Basket), args.Error(1)
}

func Test_CreateBkt(t *testing.T) {
	var tests = []struct {
		name            string
		wantStatus      int
		mockBktServFunc func() BktService
		expectedErr     localLib.Error
	}{
		{
			name:       "Create basket - Created",
			wantStatus: http.StatusCreated,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("Create", mock.Anything).Return(bktCreated)
				return &mockTableUpdate
			},
		},
		{
			name:       "Create basket - Forbidden",
			wantStatus: http.StatusForbidden,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				return &mockTableUpdate
			},
			expectedErr: errForbidden,
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(XClientKeyEnvVar, XClientKeyValue)
			require.NoError(t, err)

			bktHandler := New(test.mockBktServFunc())
			r := chi.NewRouter()
			r.Post("/basket", bktHandler.CreateBkt)

			rq := httptest.NewRequest(http.MethodPost, "/basket", nil)
			if tt.wantStatus != http.StatusForbidden {
				rq.Header.Set(XClientKey, XClientKeyValue)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, rq)

			resp := rr.Result()
			require.Equal(t, test.wantStatus, resp.StatusCode)
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			if test.expectedErr.StatusCode == 0 {
				var response basket.Basket
				err = json.Unmarshal(body, &response)
				require.NoError(t, err)
				require.Equal(t, response.ID, bktCreated.ID)
			} else {
				var response localLib.Error
				err = json.Unmarshal(body, &response)
				require.NoError(t, err)
				require.Equal(t, response.Message, tt.expectedErr.Message)
			}
		})
	}
}

// TODO Validate response
func Test_GetAmount(t *testing.T) {
	var tests = []struct {
		name            string
		wantStatus      int
		mockBktServFunc func() BktService
		msgErrExpected  string
	}{
		{
			name:       "Get Amount - Ok",
			wantStatus: http.StatusOK,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("GetAmount", mock.Anything).Return(10.0, nil)
				return &mockTableUpdate
			},
		},
		{
			name:       "Get Amount - Forbidden",
			wantStatus: http.StatusForbidden,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				return &mockTableUpdate
			},
		},
		{
			name:       "Get Amount - BadRequest - basket_id is required ",
			wantStatus: http.StatusBadRequest,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				return &mockTableUpdate
			},
			msgErrExpected: bktIDRequiredMsg,
		},
		{
			name:       "Get Amount - Bkt not found",
			wantStatus: http.StatusNotFound,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("GetAmount", mock.Anything).Return(0.0, localMap.ErrBktNotFound)
				return &mockTableUpdate
			},
		},
		{
			name:       "Get Amount - Internal server error",
			wantStatus: http.StatusInternalServerError,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("GetAmount", mock.Anything).Return(0.0, errors.New("random error"))
				return &mockTableUpdate
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			var idAndAmount string
			if tt.msgErrExpected == bktIDRequiredMsg {
				idAndAmount = "/amount"
			} else {
				idAndAmount = bktCreated.ID + "/amount"
			}
			err := os.Setenv(XClientKeyEnvVar, XClientKeyValue)
			require.NoError(t, err)

			bktHandler := New(test.mockBktServFunc())
			r := chi.NewRouter()
			r.Get("/basket/{basket_id}/amount", bktHandler.GetAmount)

			rq := httptest.NewRequest(http.MethodGet, "/basket/"+idAndAmount, nil)
			if tt.wantStatus != http.StatusForbidden {
				rq.Header.Set(XClientKey, XClientKeyValue)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, rq)

			resp := rr.Result()
			require.Equal(t, test.wantStatus, resp.StatusCode)
		})
	}
}

// TODO Validate response
func Test_AddProduct(t *testing.T) {
	productOk, err := json.Marshal(basket.AddProduct{
		Code:     "PEN",
		Quantity: 1,
	})
	require.NoError(t, err)

	invalidProduct, err := json.Marshal(basket.AddProduct{
		Code:     "RANDOMCODE",
		Quantity: 1,
	})
	require.NoError(t, err)

	var tests = []struct {
		name            string
		wantStatus      int
		mockBktServFunc func() BktService
		msgErrExpected  string
		giveRequest     *bytes.Reader
	}{
		{
			name:        "Add Product - Ok",
			wantStatus:  http.StatusOK,
			giveRequest: bytes.NewReader(productOk),
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("AddProduct", mock.Anything).Return(bktCreated, nil)
				return &mockTableUpdate
			},
		},
		{
			name:        "Add Product - Forbidden",
			wantStatus:  http.StatusForbidden,
			giveRequest: bytes.NewReader(productOk),
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				return &mockTableUpdate
			},
		},
		{
			name:        "Add Product  - BadRequest - basket_id is required ",
			wantStatus:  http.StatusBadRequest,
			giveRequest: bytes.NewReader(productOk),
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				return &mockTableUpdate
			},
			msgErrExpected: bktIDRequiredMsg,
		},
		{
			name:        "Add Product - Bkt not found",
			wantStatus:  http.StatusNotFound,
			giveRequest: bytes.NewReader(productOk),
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("AddProduct", mock.Anything).Return(basket.Basket{}, localMap.ErrBktNotFound)
				return &mockTableUpdate
			},
		},
		{
			name:        "Add Product - BadRequest - Invalid product codee",
			wantStatus:  http.StatusBadRequest,
			giveRequest: bytes.NewReader(invalidProduct),
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("AddProduct", mock.Anything).Return(basket.Basket{}, localMap.ErrInvalidProductCode)
				return &mockTableUpdate
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			var idAndProd string
			if tt.msgErrExpected == bktIDRequiredMsg {
				idAndProd = "/product"
			} else {
				idAndProd = bktCreated.ID + "/product"
			}
			err := os.Setenv(XClientKeyEnvVar, XClientKeyValue)
			require.NoError(t, err)

			bktHandler := New(test.mockBktServFunc())
			r := chi.NewRouter()
			r.Put("/basket/{basket_id}/product", bktHandler.AddProduct)

			rq := httptest.NewRequest(http.MethodPut, "/basket/"+idAndProd, tt.giveRequest)
			if tt.wantStatus != http.StatusForbidden {
				rq.Header.Set(XClientKey, XClientKeyValue)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, rq)

			resp := rr.Result()
			require.Equal(t, test.wantStatus, resp.StatusCode)
		})
	}
}

// TODO Validate response
func Test_RemoveBkt(t *testing.T) {
	var tests = []struct {
		name            string
		wantStatus      int
		mockBktServFunc func() BktService
		msgErrExpected  string
	}{
		{
			name:       "Remove basket - No Content",
			wantStatus: http.StatusNoContent,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("Delete", mock.Anything).Return(nil)
				return &mockTableUpdate
			},
		},
		{
			name:       "Remove basket - Forbidden",
			wantStatus: http.StatusForbidden,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				return &mockTableUpdate
			},
		},
		{
			name:       "Remove basket - Bkt not found",
			wantStatus: http.StatusNotFound,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("Delete", mock.Anything).Return(localMap.ErrBktNotFound)
				return &mockTableUpdate
			},
		},
		{
			name:       "Remove basket - internal server error",
			wantStatus: http.StatusInternalServerError,
			mockBktServFunc: func() BktService {
				mockTableUpdate := ServiceBktMock{}
				mockTableUpdate.On("Delete", mock.Anything).Return(errors.New("random error"))
				return &mockTableUpdate
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(XClientKeyEnvVar, XClientKeyValue)
			require.NoError(t, err)

			bktHandler := New(test.mockBktServFunc())
			r := chi.NewRouter()
			r.Delete("/basket/{basket_id}", bktHandler.RemoveBkt)

			rq := httptest.NewRequest(http.MethodDelete, "/basket/"+bktCreated.ID, nil)
			if tt.wantStatus != http.StatusForbidden {
				rq.Header.Set(XClientKey, XClientKeyValue)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, rq)

			resp := rr.Result()
			require.Equal(t, test.wantStatus, resp.StatusCode)
		})
	}
}
