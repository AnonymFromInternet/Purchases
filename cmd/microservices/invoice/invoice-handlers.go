package main

import (
	"fmt"
	"net/http"
	"time"
)

type Order struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (application *application) handlerGetCreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	// receive json
	var order Order
	application.readJSONInto(&order, w, r)

	// generate pdf with invoice data

	// create mail

	// send mail with attachment

	// send response
	var answerPayload AnswerPayload
	answerPayload.Error = false
	answerPayload.Message = fmt.Sprintf("Invoice %d.pdf created and sent to %s", order.ID, order.Email)
	application.convertToJsonAndSend(answerPayload, w)
}
