package main

import (
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
)

func (application *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", contentTypes.ContentTypeKey, "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Get("/api/widget-by-id/{id}", application.handlerGetWidgetById)

	mux.Post("/api/payment-intent", application.handlerPostPaymentIntent)
	mux.Post("/api/create-customer-and-subscribe-the-plan", application.handlerPostCreateCustomerAndSubscribePlan)
	mux.Post("/api/authenticate", application.handlerPostCreateAuthToken)
	mux.Post("/api/is-authenticated", application.handlerPostIsAuthenticated)
	mux.Post("/api/forget-password", application.handlerPostForgetPassword)
	mux.Post("/api/set-new-password", application.handlerPostSetNewPassword)

	mux.Route("/api/admin", func(chiRouter chi.Router) {
		chiRouter.Use(application.AuthMiddleware)

		chiRouter.Post("/payment-succeeded-virtual-terminal", application.handlerPostPaymentSucceededVirtualTerminal)
		chiRouter.Post("/all-sales", application.handlerPostAllSales)
		chiRouter.Post("/all-subscriptions", application.handlerPostAllSubscriptions)
		chiRouter.Post("/subscription-or-sale-description/{id}", application.handlerPostSubscriptionOrSaleDescription)
		chiRouter.Post("/refund", application.handlerPostRefund)
		chiRouter.Post("/cancel-subscription", application.handlerPostCancelSubscription)
		chiRouter.Post("/all-admin-users", application.handlerPostAllAdminUsers)
		chiRouter.Post("/all-admin-users/{id}", application.handlerPostOneUser)
	})

	return mux
}
