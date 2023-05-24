package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

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
		var convertedErrorMessage string

		if stripeError, ok := err.(*stripe.Error); ok {
			convertedErrorMessage = getStringFromStripeMessage(stripeError.Code)
		}

		return nil, convertedErrorMessage, err
	}

	return pi, "", nil
}

func getStringFromStripeMessage(code stripe.ErrorCode) string {
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
		message = "Unknown Error and operation was declined"
	}

	return message
}