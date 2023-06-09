package main

import (
	"encoding/json"
	"fmt"
	"github.com/AnonymFromInternet/Purchases/internal/cards"
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"github.com/AnonymFromInternet/Purchases/internal/status"
	"github.com/AnonymFromInternet/Purchases/internal/transactionStatus"
	"github.com/AnonymFromInternet/Purchases/internal/urlsigner"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
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

	widget, err := application.DB.GetWidgetByID(idAsInt)
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

	application.readJSONInto(&loginPagePayload, w, r)

	user, err := application.DB.GetUserByEmail(loginPagePayload.Email)
	if err != nil {
		application.errorLog.Println("cannot get user from database", err)
		application.sendInvalidCredentials(w)
		return
	}

	isPasswordValid := application.isPasswordValid(user.Password, loginPagePayload.Password)
	if !isPasswordValid {
		application.errorLog.Println("invalid password or an error by the comparing process")
		application.sendInvalidCredentials(w)
		return
	}

	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		application.errorLog.Println("invalid password or an error by the comparing process", err)
		application.sendInvalidCredentials(w)
		return
	}

	err = application.DB.InsertToken(token.Hash, user)
	if err != nil {
		application.errorLog.Println("cannot save token into database", err)
		application.sendBadRequest(w, r, err)
		return
	}

	var payload AnswerPayload
	payload.Error = false
	payload.Message = fmt.Sprintf("your token for %s", user.Email)
	payload.Token = token

	application.convertToJsonAndSend(payload, w)
}

func (application *application) handlerPostForgetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	application.readJSONInto(&payload, w, r)

	// verify if that email exists in the database
	_, err := application.DB.GetUserByEmail(payload.Email)
	if err != nil {
		var response struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}

		response.Error = true
		response.Message = "You are not registered user"
		application.convertToJsonAndSend(response, w)

		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", application.config.frontendURLForPasswordReset, payload.Email)
	sign := urlsigner.Signer{
		Secret: []byte(application.config.secretKeyForPasswordReset),
	}

	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}

	data.Link = signedLink

	err = application.SendEmail("widgets@widgets.com", payload.Email, "Password reset", "password-reset", data)
	if err != nil {
		application.errorLog.Println("cannot send email for password reset :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	response.Error = false
	response.Message = "Email was sent"

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostCreateCustomerAndSubscribePlan(w http.ResponseWriter, r *http.Request) {
	var stripePayload stripePayload
	var err error

	badResponse := jsonResponse{
		Ok: false,
	}

	err = json.NewDecoder(r.Body).Decode(&stripePayload)
	if err != nil {
		application.errorLog.Println("cannot decode stripePayload", err)

		return
	}

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
		Currency:  stripePayload.Currency,
	}

	newCustomer, errorMessage, err := card.CreateCustomer(stripePayload.PaymentMethod, stripePayload.Email)
	if err != nil {
		application.errorLog.Println(errorMessage)
		badResponse.Message = errorMessage
		application.convertToJsonAndSend(badResponse, w)

		return
	}

	subscription, errorMessage, err := card.CreateSubscription(newCustomer, stripePayload.Plan, stripePayload.Email, stripePayload.LastFour, "")
	if err != nil {
		application.errorLog.Println("cannot subscribe to plan", err)
		badResponse.Message = errorMessage
		application.convertToJsonAndSend(badResponse, w)

		return
	}

	// if all is ok -> send ok response / else -> send error response
	productId, err := strconv.Atoi(stripePayload.ProductID)
	if err != nil {
		application.errorLog.Println("cannot convert product id into int", err)
		return
	}

	customerID := application.saveCustomerGetCustomerID(stripePayload.FirstName, stripePayload.LastName, stripePayload.Email)
	if err != nil {
		application.errorLog.Println("cannot save customer into database", err)
		return
	}

	amount, err := strconv.Atoi(stripePayload.Amount)
	if err != nil {
		application.errorLog.Println("cannot convert amount into int", err)
		return
	}

	transaction := models.Transaction{
		Amount:              amount,
		Currency:            stripePayload.Currency,
		LastFour:            stripePayload.LastFour,
		TransactionStatusID: transactionStatus.Cleared,
		ExpiryMonth:         stripePayload.ExpiryMonth,
		ExpiryYear:          stripePayload.ExpiryYear,
		PaymentIntent:       subscription.ID,
		PaymentMethod:       stripePayload.PaymentMethod,
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

	paymentIntent, errorMessage, err := card.ChargeCard(payload.Currency, int(amount))
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

func (application *application) handlerPostIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	user, err := application.checkTokenValidityGetUser(r)
	if err != nil {
		application.errorLog.Println("problem by token validation :", err)
		application.sendInvalidCredentials(w)
		return
	}

	var answerPayload AnswerPayload
	answerPayload.Error = false
	answerPayload.Message = fmt.Sprintf("authentication for %s was successful", user.Email)

	application.convertToJsonAndSend(answerPayload, w)
}

func (application *application) handlerPostPaymentSucceededVirtualTerminal(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handlerPostPaymentSucceededVirtualTerminal()")
	var transactionData models.TransactionData

	application.readJSONInto(&transactionData, w, r)

	fmt.Println("transactionData :", transactionData)

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
	}

	paymentIntent, err := card.RetrievePaymentIntent(transactionData.PaymentIntent)
	if err != nil {
		application.errorLog.Println("cannot retrieve payment intent :", err)
		return
	}

	paymentMethod, err := card.CreatePaymentMethod(transactionData.PaymentMethod)
	if err != nil {
		application.errorLog.Println("cannot retrieve payment method :", err)
		return
	}

	transactionData.LastFour = paymentMethod.Card.Last4
	transactionData.ExpiryMonth = paymentMethod.Card.ExpMonth
	transactionData.ExpiryYear = paymentMethod.Card.ExpYear

	transaction := models.Transaction{
		Amount:              transactionData.Amount,
		Currency:            transactionData.Currency,
		LastFour:            transactionData.LastFour,
		BankReturnCode:      paymentIntent.Charges.Data[0].ID,
		TransactionStatusID: transactionStatus.Cleared,
		ExpiryMonth:         int(transactionData.ExpiryMonth),
		ExpiryYear:          int(transactionData.ExpiryYear),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		PaymentMethod:       transactionData.PaymentMethod,
		PaymentIntent:       transactionData.PaymentIntent,
	}

	_ = application.saveTransactionGetTransactionID(transaction)

	application.convertToJsonAndSend(transaction, w)

}

func (application *application) handlerPostSetNewPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}

	application.readJSONInto(&payload, w, r)

	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	err := application.DB.SetNewPassword(payload.NewPassword, payload.Email)
	if err != nil {
		application.errorLog.Println("cannot set new password :", err)
		response.Error = true
		response.Message = "an error by the password reset"
		application.convertToJsonAndSend(response, w)

		return
	}

	response.Error = false
	response.Message = "password reset was successful"

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostAllSales(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ItemsAmount int `json:"itemsAmount"`
		CurrentPage int `json:"currentPage"`
	}

	application.readJSONInto(&payload, w, r)

	allOrders, lastPage, totalRecords, err := application.DB.GetAllSalesPaginated(payload.ItemsAmount, payload.CurrentPage)
	if err != nil {
		application.errorLog.Println("cannot get all allOrders from database :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	if len(allOrders) < 1 {
		allOrders = make([]*models.Order, 0)
	}

	var response struct {
		ItemsAmount  int             `json:"itemsAmount"`
		CurrentPage  int             `json:"currentPage"`
		LastPage     int             `json:"lastPage"`
		TotalRecords int             `json:"totalRecords"`
		AllOrders    []*models.Order `json:"allOrders"`
	}

	response.ItemsAmount = payload.ItemsAmount
	response.CurrentPage = 1
	response.LastPage = lastPage
	response.TotalRecords = totalRecords
	response.AllOrders = allOrders

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ItemsAmount int `json:"itemsAmount"`
		CurrentPage int `json:"currentPage"`
	}

	application.readJSONInto(&payload, w, r)
	allSubscriptions, lastPage, totalRecords, err := application.DB.GetAllSubscriptionsPaginated(payload.ItemsAmount, payload.CurrentPage)
	if err != nil {
		application.errorLog.Println("cannot get all subscriptions from database :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	if len(allSubscriptions) < 1 {
		allSubscriptions = make([]*models.Order, 0)
	}

	var response struct {
		ItemsAmount  int             `json:"itemsAmount"`
		CurrentPage  int             `json:"currentPage"`
		LastPage     int             `json:"lastPage"`
		TotalRecords int             `json:"totalRecords"`
		AllOrders    []*models.Order `json:"allOrders"`
	}

	response.ItemsAmount = payload.ItemsAmount
	response.CurrentPage = 1
	response.LastPage = lastPage
	response.TotalRecords = totalRecords
	response.AllOrders = allSubscriptions

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostSubscriptionOrSaleDescription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		application.errorLog.Println("cannot convert id into int :", err)
		return
	}

	order, err := application.DB.GetSaleByID(idAsInt)
	if err != nil {
		application.errorLog.Println("cannot get order from database with given id :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	var response struct {
		Error bool         `json:"error"`
		Order models.Order `json:"order"`
	}

	response.Error = false
	response.Order = order

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostRefund(w http.ResponseWriter, r *http.Request) {
	var chargeToRefund struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"paymentIntent"`
		Amount        int    `json:"amount"`
		Currency      string `json:"currency"`
	}

	application.readJSONInto(&chargeToRefund, w, r)

	card := cards.Card{
		PublicKey: application.config.stripe.publicKey,
		SecretKey: application.config.stripe.secretKey,
		Currency:  chargeToRefund.Currency,
	}

	err := card.Refund(chargeToRefund.PaymentIntent, chargeToRefund.Amount)
	if err != nil {
		application.errorLog.Println("cannot refund a charge :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	// Update status of sale in database
	err = application.DB.UpdateOrderStatus(2, chargeToRefund.ID)
	if err != nil {
		application.errorLog.Println("cannot update order status :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	var response AnswerPayload
	response.Error = false
	response.Message = "Refund was successful"

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostCancelSubscription(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"paymentIntent"`
		Currency      string `json:"currency"`
	}

	application.readJSONInto(&payload, w, r)

	card := cards.Card{
		SecretKey: application.config.stripe.secretKey,
		PublicKey: application.config.stripe.publicKey,
		Currency:  payload.Currency,
	}

	err := card.CancelSubscription(payload.PaymentIntent)
	if err != nil {
		application.sendBadRequest(w, r, err)
		application.errorLog.Println("cannot cancel the subscription :", err)
		return
	}

	// Update status of subscription in database
	err = application.DB.UpdateOrderStatus(status.Cancelled, payload.ID)
	if err != nil {
		application.errorLog.Println("cannot update order status :", err)
		application.sendBadRequest(w, r, err)
		return
	}

	var response AnswerPayload
	response.Error = false
	response.Message = "Cancelling the subscription was successful"

	application.convertToJsonAndSend(response, w)
}

func (application *application) handlerPostAllAdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := application.DB.GetAllUsers()
	if err != nil {
		application.sendBadRequest(w, r, err)
		return
	}

	application.convertToJsonAndSend(users, w)
}

func (application *application) handlerPostOneUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)

	user, err := application.DB.GetUserById(idAsInt)
	if err != nil {
		application.sendBadRequest(w, r, err)
		return
	}

	application.convertToJsonAndSend(user, w)
}

func (application *application) handlerPostEditUserOrAddNew(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		application.errorLog.Println("cannot convert id into int", err)
		return
	}

	var user models.User

	application.readJSONInto(&user, w, r)
	if err != nil {
		application.errorLog.Println("cannot read payload into user", err)
		application.sendBadRequest(w, r, err)
		return
	}

	adminWantsToEditExistingUser := idAsInt > 0
	if adminWantsToEditExistingUser {

		userWantsToChangePassword := user.Password != ""
		if userWantsToChangePassword {
			newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
			if err != nil {
				application.errorLog.Println("cannot generate hash from password ", err)
				application.sendBadRequest(w, r, err)
				return
			}

			user.Password = string(newPasswordHash)

		}

		err := application.DB.EditUser(user)
		if err != nil {
			application.errorLog.Println("cannot edit user", err)
			application.sendBadRequest(w, r, err)
			return
		}
	} else {
		newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
		if err != nil {
			application.errorLog.Println("cannot generate hash from password ", err)
			application.sendBadRequest(w, r, err)
			return
		}
		user.Password = string(newPasswordHash)

		err = application.DB.InsertUser(user)
		if err != nil {
			application.errorLog.Println("cannot insert new user ", err)
			application.sendBadRequest(w, r, err)
			return
		}
	}

	var answerPayload AnswerPayload
	answerPayload.Error = false
	answerPayload.Message = ""

	application.convertToJsonAndSend(answerPayload, w)
}

func (application *application) handlerPostDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		application.errorLog.Println("cannot convert id into int", err)
		return
	}

	err = application.DB.DeleteUser(idAsInt)
	if err != nil {
		application.errorLog.Println("cannot delete user ", err)
		application.sendBadRequest(w, r, err)
		return
	}

	var answerPayload AnswerPayload
	answerPayload.Error = false
	answerPayload.Message = "User was successfully deleted"

	application.convertToJsonAndSend(answerPayload, w)
}
