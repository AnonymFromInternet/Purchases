package main

import (
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"github.com/AnonymFromInternet/Purchases/internal/status"
	"github.com/AnonymFromInternet/Purchases/internal/transactionStatus"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"
)

func (application *application) handlerGetVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "virtual-terminal", nil, "stripe-js")
	if err != nil {
		application.errorLog.Println(err)

		return
	}
}

func (application *application) handlerGetMainPage(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "main", nil)
	if err != nil {
		application.errorLog.Println(err)

		return
	}
}

func (application *application) handlerPostPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var err error

	tmplData := application.getTemplateData(r)
	customerID := application.saveCustomerGetCustomerID(tmplData.FirstName, tmplData.LastName, tmplData.Email)

	transaction := models.Transaction{
		Amount:              tmplData.PaymentAmount,
		Currency:            tmplData.PaymentCurrency,
		LastFour:            tmplData.LastFour,
		BankReturnCode:      tmplData.BankReturnCode,
		TransactionStatusID: transactionStatus.Cleared,
		ExpiryMonth:         int(tmplData.ExpiryMonth),
		ExpiryYear:          int(tmplData.ExpiryYear),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	transactionID := application.saveTransactionGetTransactionID(transaction)

	order := models.Order{
		WidgetId:      tmplData.WidgetId,
		TransactionId: transactionID,
		CustomerID:    customerID,
		StatusId:      status.Cleared,
		Quantity:      1,
		Amount:        tmplData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	application.saveOrderGetOrderID(order)

	data := make(map[string]interface{})
	data["tmplData"] = tmplData

	err = application.renderTemplate(w, r, "payment-succeeded", &templateData{
		Data: data,
	})
	if err != nil {
		application.errorLog.Println(err)

		return
	}

	// TODO: here should be redirect after charging?
}

func (application *application) saveCustomerGetCustomerID(firstName, lastName, email string) int {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	customerID, err := application.DB.InsertCustomerGetCustomerID(customer)
	if err != nil {
		application.errorLog.Println("cannot get customer id", err)

		return 0
	}

	return customerID
}

func (application *application) saveTransactionGetTransactionID(transaction models.Transaction) int {
	transactionID, err := application.DB.InsertTransactionGetTransactionID(transaction)
	if err != nil {
		application.errorLog.Println("cannot get transaction id", err)
	}

	return transactionID
}

func (application *application) saveOrderGetOrderID(order models.Order) {
	_, err := application.DB.InsertOrderGetOrderID(order)
	if err != nil {
		application.errorLog.Println("cannot get transaction id", err)
	}
}

func (application *application) handlerGetBuyOnce(w http.ResponseWriter, r *http.Request) {
	idAsString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idAsString)
	if err != nil {
		application.errorLog.Println("cannot convert widget id from url param into int", err)

		return
	}

	widget, err := application.DB.GetWidgetBy(id)

	if err != nil {
		application.errorLog.Println("cannot get widget from db", err)

		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	err = application.renderTemplate(w, r, "buy-once", &templateData{Data: data}, "stripe-js")
	if err != nil {
		application.errorLog.Println(err)
	}
}

func (application *application) getTemplateData(r *http.Request) models.TemplateData {
	var tmplData models.TemplateData
	err := r.ParseForm()
	if err != nil {
		application.errorLog.Println("cannot parse a form", err)

		return tmplData
	}

	email := r.Form.Get("email")
	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	paymentMethod := r.Form.Get("payment-method")
	paymentIntent := r.Form.Get("payment-intent")
	paymentAmount := r.Form.Get("payment-amount")
	paymentCurrency := r.Form.Get("payment-currency")

	widgetId, err := strconv.Atoi(r.Form.Get("widgetId"))
	if err != nil {
		application.errorLog.Println("cannot convert widget id into int", err)

		return tmplData
	}

	paymentAmountAsInt, err := strconv.Atoi(paymentAmount)
	if err != nil {
		application.errorLog.Println(err)

		return tmplData
	}

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		application.errorLog.Println(err)

		return tmplData
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		application.errorLog.Println(err)

		return tmplData
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	tmplData = models.TemplateData{
		Email:           email,
		FirstName:       firstName,
		LastName:        lastName,
		PaymentMethod:   paymentMethod,
		PaymentIntent:   paymentCurrency,
		PaymentAmount:   paymentAmountAsInt,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     expiryMonth,
		ExpiryYear:      expiryYear,
		BankReturnCode:  pi.Charges.Data[0].ID,
		WidgetId:        widgetId,
	}

	return tmplData
}
