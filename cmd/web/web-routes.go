package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (application *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Get("/virtual-terminal", application.handlerGetVirtualTerminal)

	mux.Post("/payment-succeed", application.handlerPostPaymentSucceeded)

	return mux
}
