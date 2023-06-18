package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/stripe/stripe-go/v72/sub"
)

// This package is used for the methods for interaction with stripe via stripe-go library

type Card struct {
	PublicKey string
	SecretKey string
	Currency  string
}

type Transaction struct {
	TransactionStatusId int
	Amount              int
	Currency            string
	LastFourCardNumbers string
	BankReturnedCode    string
}

// ChargeCard is an alias for CreatePaymentIntent method for the case if a cards analog will be used
func (card *Card) ChargeCard(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return card.CreatePaymentIntent(currency, amount)
}

// CreatePaymentIntent намерение
func (card *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = card.SecretKey

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		var stripeErrorMessage string

		if stripeError, ok := err.(*stripe.Error); ok {
			stripeErrorMessage = getStringFromStripeErrorMessage(stripeError.Code)
		}

		return nil, stripeErrorMessage, err
	}

	return pi, "", nil
}

// CreatePaymentMethod returns *stripe.PaymentMethod, which is the detailed information about payment method
func (card *Card) CreatePaymentMethod(paymentMethodId string) (*stripe.PaymentMethod, error) {
	stripe.Key = card.SecretKey

	return paymentmethod.Get(paymentMethodId, nil)
}

// RetrievePaymentIntent gets an existing pi by id
func (card *Card) RetrievePaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = card.SecretKey

	return paymentintent.Get(id, nil)
}

func (card *Card) CreateSubscription(customer *stripe.Customer, planID, email, last4, cardType string) (*stripe.Subscription, string, error) {
	stripeCustomerId := customer.ID
	items := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(planID)},
	}

	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerId),
		Items:    items,
	}

	subscriptionParams.AddMetadata("last_four", last4)
	subscriptionParams.AddMetadata("card_type", cardType)

	// TODO: For some reasons it doesn't work
	// subscriptionParams.AddExpand("last_invoice")

	subscription, err := sub.New(subscriptionParams)
	if err != nil {
		if stripeError, ok := err.(*stripe.Error); ok {
			stripeErrorMessage := getStringFromStripeErrorMessage(stripeError.Code)

			return nil, stripeErrorMessage, err
		}
	}

	return subscription, "", nil
}

func (card *Card) CreateCustomer(paymentMethod, email string) (*stripe.Customer, string, error) {
	stripe.Key = card.SecretKey

	newCustomerParams := stripe.CustomerParams{
		Email:         stripe.String(email),
		PaymentMethod: stripe.String(paymentMethod),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethod),
		},
	}

	newCustomer, err := customer.New(&newCustomerParams)
	if stripeError, ok := err.(*stripe.Error); ok {
		stripeErrorMessage := getStringFromStripeErrorMessage(stripeError.Code)

		return nil, stripeErrorMessage, err
	}

	return newCustomer, "", nil
}

func (card *Card) Refund(paymentIntent string, amount int) error {
	stripe.Key = card.SecretKey
	amountToRefund := int64(amount)

	refundParams := &stripe.RefundParams{
		Amount:        &amountToRefund,
		PaymentIntent: &paymentIntent,
	}

	_, err := refund.New(refundParams)
	if err != nil {
		return err
	}

	return nil
}

func (card *Card) CancelSubscription(subscriptionID string) error {
	stripe.Key = card.SecretKey

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	_, err := sub.Update(subscriptionID, params)
	if err != nil {
		return err
	}

	return nil
}

func getStringFromStripeErrorMessage(code stripe.ErrorCode) string {
	var message string

	switch code {
	case stripe.ErrorCodeCardDeclined:
		message = "Your Card was declined"
	case stripe.ErrorCodeExpiredCard:
		message = "Your Card is expired"
	case stripe.ErrorCodeIncorrectCVC:
		message = "Incorrect CVC Code"
	case stripe.ErrorCodeIncorrectZip:
		message = "Incorrect ZIP or Postal Code"
	case stripe.ErrorCodeAmountTooLarge:
		message = "The Amount is too large to charge to Your Card"
	case stripe.ErrorCodeAmountTooSmall:
		message = "The Amount is too small to charge to Your Card"
	case stripe.ErrorCodeBalanceInsufficient:
		message = "Insufficient Balance"
	default:
		message = "Unknown Error and the Operation was declined"
	}

	return message
}
