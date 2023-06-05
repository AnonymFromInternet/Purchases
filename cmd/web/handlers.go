package main

import (
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func (application *application) handlerGetVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "virtual-terminal", nil, "stripe-js")
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
	idAsString := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(idAsString)
	if err != nil {
		application.errorLog.Println("cannot convert widget id from url param into int", err)

		return
	}

	widget, err := application.DB.GetWidgetBy(idAsInt)

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

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
	}

	pi, err := card.RetrievePaymentIntentBy(paymentIntent)
	if err != nil {
		application.errorLog.Println(err)

		return nil
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		application.errorLog.Println(err)

		return nil
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	data := make(map[string]interface{})
	data["email"] = email
	data["cardholderName"] = cardHolderName
	data["paymentMethod"] = paymentMethod
	data["paymentIntent"] = paymentIntent
	data["paymentAmount"] = paymentAmountAsInt / 100
	data["paymentCurrency"] = paymentCurrency
	data["lastFour"] = lastFour
	data["expiryMonth"] = expiryMonth
	data["expiryYear"] = expiryYear
	data["bankReturnCode"] = pi.Charges.Data[0].ID

	return data
}
