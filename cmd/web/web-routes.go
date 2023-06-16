package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (application *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(SessionLoadMiddleWare)

	mux.Route("/admin", func(chiRouter chi.Router) {
		chiRouter.Use(application.AuthMiddleware)

		chiRouter.Get("/virtual-terminal", application.handlerGetVirtualTerminal)
		chiRouter.Get("/all-sales", application.handlerGetAllSales)
		chiRouter.Get("/all-subscriptions", application.handlerGetAllSubscriptions)
		chiRouter.Get("/sales/{id}", application.handlerGetSaleDescription)
		chiRouter.Get("/subscriptions/{id}", application.handlerGetSubscriptionsDescription)
	})

	mux.Get("/main", application.handlerGetMainPage)

	mux.Get("/widget/{id}", application.handlerGetBuyOnce)
	mux.Get("/receipt-buy-once", application.handlerGetReceiptAfterBuyOnce)
	// mux.Get("/receipt-virtual-terminal", application.handlerGetReceiptAfterVirtualTerminal)
	mux.Get("/receipt-gold-plan", application.handlerGetReceiptGoldPlan)
	mux.Get("/gold-plan", application.handlerGetGoldPlan)
	mux.Get("/login", application.handlerGetLoginPage)
	mux.Get("/logout", application.handlerPostLogoutPage)
	mux.Get("/forget-password", application.handlerGetForgetPassword)
	mux.Get("/reset-password", application.handlerGetResetPassword)

	mux.Post("/payment-succeeded-buy-once", application.handlerPostPaymentSucceededByOnce)
	mux.Post("/login", application.handlerPostLoginPage)
	// mux.Post("/payment-succeeded-virtual-terminal", application.handlerPostPaymentSucceededVirtualTerminal)

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static*", http.StripPrefix("/static", fileServer))

	return mux
}
