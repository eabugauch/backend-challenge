package handler

import "github.com/go-chi/chi"

// BasketRoutes mapping cases endpoints.
func BasketRoutes(r *chi.Mux, bktService BktService) *chi.Mux {
	bktHandler := New(bktService)
	r.Get("/ping", bktHandler.Ping)
	r.Post("/basket", bktHandler.CreateBkt)
	r.Get("/basket/{basket_id}", bktHandler.GetBkt)
	r.Put("/basket/{basket_id}/product", bktHandler.AddProduct)
	r.Get("/basket/{basket_id}/amount", bktHandler.GetAmount)
	r.Delete("/basket/{basket_id}", bktHandler.RemoveBkt)
	return r
}
