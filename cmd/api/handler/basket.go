package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/mercadolibre/backend-challenge/internal/basket"
	localMap "github.com/mercadolibre/backend-challenge/internal/basket/local-map"
	localLib "github.com/mercadolibre/backend-challenge/local-library"
)

const (
	bktIDParam              = "basket_id"
	bktNotFoundMsg          = "basket not found"
	bktIDRequiredMsg        = "basket_id is required"
	bktInternalServerErrMsg = "internal server error"
)

// A BktService interface is used to manage the Basket methods.
type BktService interface {
	Create() basket.Basket
	Get(bktID string) (basket.Basket, error)
	GetAmount(bktID string) (float64, error)
	Delete(bktID string) error
	AddProduct(bktID string, prdID string, quantity int) (basket.Basket, error)
}

// BktHandler is responsible for handle methods related to basket service.
type BktHandler struct {
	bktService BktService
}

// New return an instance of BktHandler.
func New(bktService BktService) BktHandler {
	return BktHandler{
		bktService: bktService,
	}
}

// CreateBkt creates an empty basket.
func (rh *BktHandler) CreateBkt(w http.ResponseWriter, r *http.Request) {
	if !isValidCaller(w, r) {
		return
	}
	localLib.RespondJSON(w, rh.bktService.Create(), http.StatusCreated)
}

// GetBkt returns the basket corresponding to the id sent by parameter.
func (rh *BktHandler) GetBkt(w http.ResponseWriter, r *http.Request) {
	if !isValidCaller(w, r) {
		return
	}

	bktID := chi.URLParam(r, bktIDParam)
	if bktID == "" {
		localLib.RespondJSON(w, localLib.Error{Message: bktIDRequiredMsg, StatusCode: http.StatusBadRequest}, http.StatusBadRequest)
		return
	}

	bkt, err := rh.bktService.Get(bktID)
	if err != nil {
		if err == localMap.ErrBktNotFound {
			localLib.RespondJSON(w, localLib.Error{Message: bktNotFoundMsg, StatusCode: http.StatusNotFound}, http.StatusNotFound)
			return
		}
		// TODO add metrics
		log.Printf("error in get basket: %s", err.Error())
		localLib.RespondJSON(w, localLib.Error{Message: bktInternalServerErrMsg, StatusCode: http.StatusInternalServerError}, http.StatusInternalServerError)
		return
	}
	localLib.RespondJSON(w, bkt, http.StatusOK)
}

// AddProduct adds a product to the basket passed by parameters.
func (rh *BktHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	if !isValidCaller(w, r) {
		return
	}

	bktID := chi.URLParam(r, bktIDParam)
	if bktID == "" {
		localLib.RespondJSON(w, localLib.Error{Message: bktIDRequiredMsg, StatusCode: http.StatusBadRequest}, http.StatusBadRequest)
		return
	}

	var body basket.AddProduct
	if err := localLib.Bind(r, &body); err != nil {
		localLib.RespondJSON(w, localLib.Error{Message: "invalid body", StatusCode: http.StatusBadRequest}, http.StatusBadRequest)
		return
	}

	bkt, err := rh.bktService.AddProduct(bktID, body.Code, body.Quantity)
	// Not found is not implemented as giving an error here means that we have not previously loaded the product.
	if err != nil {
		if err == localMap.ErrBktNotFound {
			localLib.RespondJSON(w, localLib.Error{Message: bktNotFoundMsg, StatusCode: http.StatusNotFound}, http.StatusNotFound)
			return
		}
		if err == localMap.ErrInvalidProductCode {
			localLib.RespondJSON(w, localLib.Error{Message: "invalid product code", StatusCode: http.StatusBadRequest}, http.StatusBadRequest)
			return
		}
		// TODO add metrics
		log.Printf("error in add product: %s", err.Error())
		localLib.RespondJSON(w, localLib.Error{Message: bktInternalServerErrMsg, StatusCode: http.StatusInternalServerError}, http.StatusInternalServerError)
		return
	}

	localLib.RespondJSON(w, bkt, http.StatusOK)
}

func isValidCaller(w http.ResponseWriter, r *http.Request) bool {
	callerScope := r.Header.Get("x-client-key")
	secretCaller := os.Getenv("X_CLIENT_KEY")
	if secretCaller == "" || callerScope != secretCaller {
		log.Printf("unauthorized caller")
		localLib.RespondJSON(w, localLib.Error{Message: "Forbidden", StatusCode: http.StatusForbidden}, http.StatusForbidden)
		return false
	}
	return true
}

// GetAmount returns the amount in the basket.
func (rh *BktHandler) GetAmount(w http.ResponseWriter, r *http.Request) {
	if !isValidCaller(w, r) {
		return
	}
	bktID := chi.URLParam(r, bktIDParam)
	if bktID == "" {
		localLib.RespondJSON(w, localLib.Error{Message: bktIDRequiredMsg, StatusCode: http.StatusBadRequest}, http.StatusBadRequest)
		return
	}

	amount, err := rh.bktService.GetAmount(bktID)
	if err != nil {
		if err == localMap.ErrBktNotFound {
			localLib.RespondJSON(w, localLib.Error{Message: bktNotFoundMsg, StatusCode: http.StatusNotFound}, http.StatusNotFound)
			return
		}
		// TODO add metrics
		log.Printf("error in get amount: %s", err.Error())
		localLib.RespondJSON(w, localLib.Error{Message: bktInternalServerErrMsg, StatusCode: http.StatusInternalServerError}, http.StatusInternalServerError)
		return
	}

	localLib.RespondJSON(w, basket.GetAmount{BktID: bktID, Amount: amount}, http.StatusOK)

}

// RemoveBkt deletes the basket sent by parameter.
func (rh *BktHandler) RemoveBkt(w http.ResponseWriter, r *http.Request) {
	if !isValidCaller(w, r) {
		return
	}
	bktID := chi.URLParam(r, bktIDParam)
	if bktID == "" {
		localLib.RespondJSON(w, localLib.Error{Message: bktIDRequiredMsg, StatusCode: http.StatusBadRequest}, http.StatusBadRequest)
		return
	}

	err := rh.bktService.Delete(bktID)
	if err != nil {
		if err == localMap.ErrBktNotFound {
			localLib.RespondJSON(w, localLib.Error{Message: bktNotFoundMsg, StatusCode: http.StatusNotFound}, http.StatusNotFound)
			return
		}
		// TODO add metrics
		log.Printf("error in delete basket: %s", err.Error())
		localLib.RespondJSON(w, localLib.Error{Message: bktInternalServerErrMsg, StatusCode: http.StatusInternalServerError}, http.StatusInternalServerError)
		return
	}

	localLib.RespondJSON(w, nil, http.StatusNoContent)
}

// Ping is the endpoint to validate that the application was up correctly.
func (rh *BktHandler) Ping(w http.ResponseWriter, _ *http.Request) {
	localLib.RespondJSON(w, "pong", http.StatusOK)
}
