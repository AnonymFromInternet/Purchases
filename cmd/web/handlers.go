package main

import (
	"fmt"
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"github.com/AnonymFromInternet/Purchases/internal/status"
	"github.com/AnonymFromInternet/Purchases/internal/transactionStatus"
	"github.com/AnonymFromInternet/Purchases/internal/urlsigner"
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

func (application *application) handlerGetReceiptGoldPlan(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "payment-succeeded-gold-plan", nil)
	if err != nil {
		application.errorLog.Println("cannot render template", err)

		return
	}
}

func (application *application) handlerGetLoginPage(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "login", nil)
	if err != nil {
		application.errorLog.Println("cannot render login page")

		return
	}
}

func (application *application) handlerPostLoginPage(w http.ResponseWriter, r *http.Request) {
	err := application.SessionManager.RenewToken(r.Context())
	if err != nil {
		application.errorLog.Println("cannot renew token :", err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		application.errorLog.Println("cannot parse form :", err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := application.DB.GetUserByEmail(email)
	if err != nil {
		application.errorLog.Println(fmt.Sprintf("cannot get user from database with email %s", email), err)
		return
	}

	isPasswordValid := application.isPasswordValid(user.Password, password)
	if !isPasswordValid {
		application.errorLog.Println("invalid password")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	application.SessionManager.Put(r.Context(), "userID", user.ID)

	http.Redirect(w, r, "/main", http.StatusSeeOther)
}

func (application *application) handlerGetReceiptAfterBuyOnce(w http.ResponseWriter, r *http.Request) {
	sessionData := application.SessionManager.Get(r.Context(), "receipt").(models.TransactionData)
	data := make(map[string]interface{})
	data["tmplData"] = sessionData

	application.SessionManager.Put(r.Context(), "receipt", nil)

	err := application.renderTemplate(w, r, "payment-succeeded-buy-once", &templateData{
		Data: data,
	})
	if err != nil {
		application.errorLog.Println(err)

		return
	}
}

func (application *application) handlerGetReceiptAfterVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "payment-succeeded-virtual-terminal", nil)
	if err != nil {
		application.errorLog.Println("cannot render template with the name payment-succeeded-virtual-terminal", err)
	}
}

func (application *application) handlerGetBuyOnce(w http.ResponseWriter, r *http.Request) {
	idAsString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idAsString)
	if err != nil {
		application.errorLog.Println("cannot convert widget id from url param into int", err)

		return
	}

	widget, err := application.DB.GetWidgetByID(id)

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

func (application *application) handlerGetGoldPlan(w http.ResponseWriter, r *http.Request) {
	widget, err := application.DB.GetWidgetByID(2)
	if err != nil {
		application.errorLog.Println("cannot get widget from database", err)

		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	err = application.renderTemplate(w, r, "gold-plan", &templateData{
		Data: data,
	}, "stripe-js")

	if err != nil {
		application.errorLog.Println("cannot render template", err)

		return
	}
}

// handlerPostPaymentSucceededByOnce is called from html, and only after the answer comes from stripe
func (application *application) handlerPostPaymentSucceededByOnce(w http.ResponseWriter, r *http.Request) {
	tmplData := application.getTransactionData(r)
	customerID := application.saveCustomerGetCustomerID(tmplData.FirstName, tmplData.LastName, tmplData.Email)

	transaction := models.Transaction{
		Amount:              tmplData.PaymentAmount,
		Currency:            tmplData.PaymentCurrency,
		LastFour:            tmplData.LastFour,
		BankReturnCode:      tmplData.BankReturnCode,
		TransactionStatusID: transactionStatus.Cleared,
		ExpiryMonth:         int(tmplData.ExpiryMonth),
		ExpiryYear:          int(tmplData.ExpiryYear),
		PaymentIntent:       tmplData.PaymentIntent,
		PaymentMethod:       tmplData.PaymentMethod,
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

	application.SessionManager.Put(r.Context(), "receipt", tmplData)
	http.Redirect(w, r, "/receipt-buy-once", http.StatusSeeOther)
}

func (application *application) handlerPostPaymentSucceededVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	formData := application.getFormData(r)

	fmt.Println("formData :", formData)

	http.Redirect(w, r, "/receipt-virtual-terminal", http.StatusSeeOther)
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

func (application *application) getTransactionData(r *http.Request) models.TransactionData {
	var transactionData models.TransactionData

	err := r.ParseForm()
	if err != nil {
		application.errorLog.Println("cannot parse a form", err)

		return transactionData
	}

	formData := application.getFormData(r)

	widgetId, err := strconv.Atoi(r.Form.Get("widgetId"))
	if err != nil {
		application.errorLog.Println("cannot convert widget id into int", err)

		return transactionData
	}

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
	}

	pi, err := card.RetrievePaymentIntent(formData.PaymentIntent)
	if err != nil {
		application.errorLog.Println(err)

		return transactionData
	}

	pm, err := card.CreatePaymentMethod(formData.PaymentMethod)
	if err != nil {
		application.errorLog.Println(err)

		return transactionData
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	transactionData = models.TransactionData{
		Email:           formData.Email,
		FirstName:       formData.FirstName,
		LastName:        formData.LastName,
		PaymentMethod:   formData.PaymentMethod,
		PaymentIntent:   formData.PaymentIntent,
		PaymentAmount:   formData.PaymentAmount,
		PaymentCurrency: formData.PaymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     expiryMonth,
		ExpiryYear:      expiryYear,
		BankReturnCode:  pi.Charges.Data[0].ID,
		WidgetId:        widgetId,
	}

	return transactionData
}

func (application *application) getFormData(r *http.Request) models.TransactionData {
	var formData models.TransactionData

	err := r.ParseForm()
	if err != nil {
		application.errorLog.Println("cannot parse a form", err)

		return formData
	}

	email := r.Form.Get("email")
	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	paymentMethod := r.Form.Get("payment-method")
	paymentIntent := r.Form.Get("payment-intent")
	paymentAmount := r.Form.Get("payment-amount")
	paymentCurrency := r.Form.Get("payment-currency")

	paymentAmountAsInt, err := strconv.Atoi(paymentAmount)
	if err != nil {
		application.errorLog.Println(err)

		return formData
	}

	formData = models.TransactionData{
		Email:           email,
		FirstName:       firstName,
		LastName:        lastName,
		PaymentMethod:   paymentMethod,
		PaymentIntent:   paymentIntent,
		PaymentAmount:   paymentAmountAsInt,
		PaymentCurrency: paymentCurrency,
	}

	return formData
}

func (application *application) handlerPostLogoutPage(w http.ResponseWriter, r *http.Request) {
	_ = application.SessionManager.Destroy(r.Context())
	_ = application.SessionManager.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (application *application) handlerGetForgetPassword(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "forget-password", &templateData{}, "stripe-js")
	if err != nil {
		application.errorLog.Println("cannot render template forget-password :", err)
		return
	}
}

func (application *application) handlerGetResetPassword(w http.ResponseWriter, r *http.Request) {
	requestURL := r.RequestURI
	fullURL := fmt.Sprintf("%s%s", application.config.frontendURLForPasswordReset, requestURL)

	signer := urlsigner.Signer{
		Secret: []byte(application.config.secretKeyForPasswordReset),
	}

	isTokenValid := signer.IsTokenValid(fullURL)
	if !isTokenValid {
		_, _ = w.Write([]byte("bad request or something went wrong"))
		return
	}

	isTokenActual := signer.IsTokenActual(fullURL, 15)
	if !isTokenActual {
		_, _ = w.Write([]byte("This link is already too old. Please go to the login page and reset the password again"))

		return
	}

	data := make(map[string]interface{})
	data["email"] = r.URL.Query().Get("email")

	err := application.renderTemplate(w, r, "set-new-password", &templateData{Data: data})
	if err != nil {
		application.errorLog.Println("cannot render the set-new-password template :", err)
		return
	}
}

func (application *application) handlerGetAllSales(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "all-sales", &templateData{}, "stripe-js")
	if err != nil {
		application.errorLog.Println("cannot render the all-sales template :", err)

		return
	}
}

func (application *application) handlerGetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "all-subscriptions", &templateData{}, "stripe-js")
	if err != nil {
		application.errorLog.Println("cannot render the all-sales template :", err)

		return
	}
}

func (application *application) handlerGetSaleDescription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)

	widget, err := application.DB.GetWidgetByID(idAsInt)
	data := make(map[string]interface{})

	data["widget"] = widget

	err = application.renderTemplate(w, r, "sale-description", &templateData{Data: data}, "stripe-js")
	if err != nil {
		application.errorLog.Println("cannot render the all-sales template :", err)

		return
	}
}

func (application *application) handlerGetSubscriptionsDescription(w http.ResponseWriter, r *http.Request) {
	err := application.renderTemplate(w, r, "subscription-description", &templateData{}, "stripe-js")
	if err != nil {
		application.errorLog.Println("cannot render the all-sales template :", err)

		return
	}
}
