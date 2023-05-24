package main

import (
	"encoding/json"
	"net/http"
)

type stripePayload struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
	Content string `json:"content"`
	Id      int    `json:"id"`
}

func (application *application) handlerGetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var err error
	response := jsonResponse{
		Ok: true,
	}

	output, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		application.errorLog.Println("[Backend]:[main]:[handlerGetPaymentIntent] - cannot send response")

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(output)
	if err != nil {
		application.errorLog.Println("[Backend]:[main]:[handlerGetPaymentIntent] - cannot write response to writer")

		return
	}
}
