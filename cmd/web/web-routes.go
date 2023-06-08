package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (application *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(SessionLoadMiddleWare)

	mux.Get("/main", application.handlerGetMainPage)
	mux.Get("/virtual-terminal", application.handlerGetVirtualTerminal)
	mux.Get("/widget/{id}", application.handlerGetBuyOnce)
	mux.Get("/receipt-buy-once", application.handlerGetReceiptAfterBuyOnce)
	mux.Get("/receipt-virtual-terminal", application.handlerGetReceiptAfterVirtualTerminal)
	mux.Get("/gold-plan", application.handlerGetGoldPlan)

	mux.Post("/payment-succeed-buy-once", application.handlerPostPaymentSucceededByOnce)
	mux.Post("/payment-succeed-virtual-terminal", application.handlerPostPaymentSucceededVirtualTerminal)

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static*", http.StripPrefix("/static", fileServer))

	return mux
}
