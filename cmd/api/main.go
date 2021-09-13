package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mercadolibre/backend-challenge/cmd/api/handler"
	localMap "github.com/mercadolibre/backend-challenge/internal/basket/local-map"
)

const (
	ExitCodeOK = iota
	ExitCodeFailToCreateWebApplication
	defaultWebApplicationPort = "8080"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	bktService := localMap.New()
	r = handler.BasketRoutes(r, bktService)

	log.Print("listen in port: " + defaultWebApplicationPort)
	err := http.ListenAndServe(":"+defaultWebApplicationPort, r)
	if err != nil {
		log.Print(err.Error())
		os.Exit(ExitCodeFailToCreateWebApplication)
	}
	log.Print("server exit")
	os.Exit(ExitCodeOK)
}
