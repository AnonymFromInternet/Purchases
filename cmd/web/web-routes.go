package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (application *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Get("/virtual-terminal", application.handlerGetVirtualTerminal)
	mux.Get("/widget/{id}", application.handlerGetBuyOnce)

	mux.Post("/payment-succeed", application.handlerPostPaymentSucceeded)

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static*", http.StripPrefix("/static", fileServer))

	return mux
}
