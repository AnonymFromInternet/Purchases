package main

import (
	"encoding/json"
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"net/http"
	"strconv"
)

type stripePayload struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	Id      int    `json:"id,omitempty"`
}

func (application *application) handlerGetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var err error
	var stripePayload stripePayload

	err = json.NewDecoder(r.Body).Decode(&stripePayload)
	if err != nil {
		application.errorLog.Println("cannot get payload from the request body", err)

		return
	}

	amount, err := strconv.Atoi(stripePayload.Amount)
	if err != nil {
		application.errorLog.Println("cannot convert amount into int", err)

		return
	}

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
		Currency:  stripePayload.Currency,
	}

	okay := true
	pi, errorMessage, err := card.ChargeCard(stripePayload.Currency, amount)
	if err != nil {
		okay = false
	}

	if okay {
		output, err := json.MarshalIndent(pi, "", " ")
		if err != nil {
			application.errorLog.Println("cannot convert pi into json", err)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(output)
	} else {
		response := jsonResponse{
			Ok:      false,
			Message: errorMessage,
			Content: "",
			Id:      0,
		}

		output, err := json.MarshalIndent(response, "", " ")
		if err != nil {
			application.errorLog.Println("cannot convert response into json", err)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(output)
	}
}
