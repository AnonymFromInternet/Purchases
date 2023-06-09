package main

import (
	"encoding/json"
	"fmt"
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"github.com/AnonymFromInternet/Purchases/internal/status"
	"github.com/AnonymFromInternet/Purchases/internal/transactionStatus"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"strconv"
	"time"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	Plan          string `json:"plan"`
	PaymentMethod string `json:"paymentMethod"`
	Email         string `json:"email"`
	LastFour      string `json:"lastFour"`
	CardBrand     string `json:"cardBrand"`
	ExpiryMonth   int    `json:"expiryMonth"`
	ExpiryYear    int    `json:"expiryYear"`
	ProductID     string `json:"productID"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
}

type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	Id      int    `json:"id,omitempty"`
}

func (application *application) handlerGetWidgetById(w http.ResponseWriter, r *http.Request) {
	idAsString := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(idAsString)
	if err != nil {
		application.errorLog.Println("cannot convert widget id from url param into int", err)

		return
	}

	widget, err := application.DB.GetWidgetBy(idAsInt)
	widgetAsResponse, err := json.MarshalIndent(widget, "", " ")

	w.Header().Set(contentTypes.ContentTypeKey, contentTypes.ApplicationJSON)
	_, err = w.Write(widgetAsResponse)
	if err != nil {
		application.errorLog.Println("cannot sent widget as a response", err)

		return
	}
}

func (application *application) handlerPostCreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var loginPagePayload LoginPagePayload

	application.readJSON(w, r, &loginPagePayload)

	var payload AnswerPayload
	payload.Error = false
	payload.Message = "Authentication was successful"

	output, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		application.errorLog.Println("cannot convert payload for user into slice of bytes", err)

		return
	}

	w.Header().Set(contentTypes.ContentTypeKey, contentTypes.ApplicationJSON)
	_, _ = w.Write(output)
}

func (application *application) handlerPostCreateCustomerAndSubscribePlan(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload
	var err error

	// TODO: should make something with this variable
	var subscription *stripe.Subscription

	fmt.Println(subscription)

	badResponse := jsonResponse{
		Ok: false,
	}

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		application.errorLog.Println("cannot decode payload", err)

		return
	}

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
		Currency:  payload.Currency,
	}

	newCustomer, errorMessage, err := card.CreateCustomer(payload.PaymentMethod, payload.Email)
	if err != nil {
		application.errorLog.Println(errorMessage)
		badResponse.Message = errorMessage
		application.convertToJsonAndSend(badResponse, w)

		return
	}

	subscription, errorMessage, err = card.CreateSubscription(newCustomer, payload.Plan, payload.Email, payload.LastFour, "")
	if err != nil {
		application.errorLog.Println("cannot subscribe to plan", err)
		badResponse.Message = errorMessage
		application.convertToJsonAndSend(badResponse, w)

		return
	}

	// if all is ok -> send ok response / else -> send error response
	productId, err := strconv.Atoi(payload.ProductID)
	if err != nil {
		application.errorLog.Println("cannot convert product id into int", err)
		return
	}

	customerID := application.saveCustomerGetCustomerID(payload.FirstName, payload.LastName, payload.Email)
	if err != nil {
		application.errorLog.Println("cannot save customer into database", err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		application.errorLog.Println("cannot convert amount into int", err)
		return
	}

	transaction := models.Transaction{
		Amount:              amount,
		Currency:            payload.Currency,
		LastFour:            payload.LastFour,
		TransactionStatusID: transactionStatus.Cleared,
		ExpiryMonth:         payload.ExpiryMonth,
		ExpiryYear:          payload.ExpiryYear,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	transactionId := application.saveTransactionGetTransactionID(transaction)
	if err != nil {
		application.errorLog.Println("cannot save transaction", err)
		return
	}

	order := models.Order{
		WidgetId:      productId,
		TransactionId: transactionId,
		CustomerID:    customerID,
		StatusId:      status.Cleared,
		Quantity:      1,
		Amount:        amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	application.saveOrderGetOrderID(order)

	response := jsonResponse{
		Ok:      true,
		Message: "Subscription was successful",
	}

	application.convertToJsonAndSend(response, w)
}

func (application *application) saveOrderGetOrderID(order models.Order) {
	_, err := application.DB.InsertOrderGetOrderID(order)
	if err != nil {
		application.errorLog.Println("cannot get transaction id", err)
	}
}

func (application *application) saveTransactionGetTransactionID(transaction models.Transaction) int {
	txnID, err := application.DB.InsertTransactionGetTransactionID(transaction)
	if err != nil {
		application.errorLog.Println("cannot insert transaction into database", err)
	}

	return txnID
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

func (application *application) handlerPostPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var err error
	var payload stripePayload

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		application.errorLog.Println("cannot get payload from the request body", err)

		return
	}

	amount, err := strconv.ParseFloat(payload.Amount, 64)
	if err != nil {
		application.errorLog.Println("cannot convert amount into int", err)

		return
	}

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
		Currency:  payload.Currency,
	}

	paymentIntent, errorMessage, err := card.ChargeCard(payload.Currency, int(amount*100))
	if err != nil {
		application.errorLog.Println("cannot get paymentIntent from charge card function", err)

		errorResponse := jsonResponse{
			Ok:      false,
			Message: errorMessage,
			Content: "",
			Id:      0,
		}

		application.convertToJsonAndSend(errorResponse, w)

		return
	}

	application.convertToJsonAndSend(paymentIntent, w)
}

func (application *application) convertToJsonAndSend(data interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		application.errorLog.Println("cannot convert data into json", err)

		return
	}

	w.Header().Set(contentTypes.ContentTypeKey, contentTypes.ApplicationJSON)
	_, _ = w.Write(output)
}
