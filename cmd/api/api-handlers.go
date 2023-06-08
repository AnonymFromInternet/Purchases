package main

import (
	"encoding/json"
	"fmt"
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	Plan          string `json:"plan"`
	PaymentMethod string `json:"paymentMethod"`
	Email         string `json:"email"`
	LastFour      string `json:"lastFour"`
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

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(widgetAsResponse)
	if err != nil {
		application.errorLog.Println("cannot sent widget as a response", err)

		return
	}
}

func (application *application) handlerPostCreateCustomerAndSubscribePlan(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload
	var err error

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		application.errorLog.Println("cannot decode payload", err)

		return
	}

	fmt.Println("payload :", payload)

	response := jsonResponse{
		Ok:      true,
		Message: "Test message from new handler from api",
		Content: "Test",
		Id:      8,
	}

	application.convertToJsonAndSend(response, w)
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

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(output)
}
