package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
)

func (application *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Get("/api/widget-by-id/{id}", application.handlerGetWidgetById)

	mux.Post("/api/payment-intent", application.handlerPostPaymentIntent)
	mux.Post("/api/create-customer-and-subscribe-the-plan", application.handlerPostCreateCustomerAndSubscribePlan)

	return mux
}
