package main

import (
	"net/http"
	"strconv"
)

func (application *application) handlerGetVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := application.renderTemplate(w, r, "virtual-terminal", nil); err != nil {
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
