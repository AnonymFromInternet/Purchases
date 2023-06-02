package main

import (
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"net/http"
	"strconv"
	"time"
)

func (application *application) handlerGetVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["publicKey"] = application.config.stripe.publicKey

	err := application.renderTemplate(w, r, "virtual-terminal", &templateData{StringMap: stringMap}, "stripe-js")
	if err != nil {
		application.errorLog.Println(err)

		return
	}
}

func (application *application) handlerPostPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var err error

	err = r.ParseForm()
	if err != nil {
		application.errorLog.Println(err)

		return
	}

	tmplData := application.getTemplateData(r)

	err = application.renderTemplate(w, r, "payment-succeeded", &templateData{
		Data: tmplData,
	})
	if err != nil {
		application.errorLog.Println(err)

		return
	}

}

func (application *application) handlerGetBuyOnce(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["publicKey"] = application.config.stripe.publicKey

	widget := models.Widget{
		Id:             1,
		Name:           "Test Widget",
		Description:    "Very nice Widget",
		InventoryLevel: 10,
		Price:          1000,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	err := application.renderTemplate(w, r, "buy-once", &templateData{StringMap: stringMap, Data: data}, "stripe-js")
	if err != nil {
		application.errorLog.Println(err)
	}
}

func (application *application) getTemplateData(r *http.Request) map[string]interface{} {
	email := r.Form.Get("email")
	cardHolderName := r.Form.Get("cardholder-name")
	paymentMethod := r.Form.Get("payment-method")
	paymentIntent := r.Form.Get("payment-intent")
	paymentAmount := r.Form.Get("payment-amount")
	paymentCurrency := r.Form.Get("payment-currency")

	paymentAmountAsInt, err := strconv.Atoi(paymentAmount)
	if err != nil {
		application.errorLog.Println(err)

		return nil
	}

	data := make(map[string]interface{})
	data["email"] = email
	data["cardholderName"] = cardHolderName
	data["paymentMethod"] = paymentMethod
	data["paymentIntent"] = paymentIntent
	data["paymentAmount"] = paymentAmountAsInt / 100
	data["paymentCurrency"] = paymentCurrency

	return data
}
