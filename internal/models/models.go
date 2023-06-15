package models

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type DBModel struct {
	DB *sql.DB
}

type Models struct {
	DB DBModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

func (model *DBModel) GetWidgetByID(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		SELECT id, name, description, inventory_level, price, coalesce(image, ''), is_recurring, plan_id,
			created_at, updated_at
		FROM widgets
		WHERE id = $1
	`

	var widget Widget

	row := model.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&widget.ID, &widget.Name, &widget.Description, &widget.InventoryLevel, &widget.Price, &widget.Image,
		&widget.IsRecurring, &widget.PlanID, &widget.CreatedAt, &widget.UpdatedAt,
	)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

func (model *DBModel) InsertTransactionGetTransactionID(transaction Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		INSERT into transactions
			(amount, currency, last_four, bank_return_code, transaction_status_id, created_at, updated_at, expiry_month,
			 expiry_year, payment_intent, payment_method)
		values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		returning id
	`

	row := model.DB.QueryRowContext(
		ctx,
		query,
		transaction.Amount,
		transaction.Currency,
		transaction.LastFour,
		transaction.BankReturnCode,
		transaction.TransactionStatusID,
		time.Now(),
		time.Now(),
		transaction.ExpiryMonth,
		transaction.ExpiryYear,
		transaction.PaymentIntent,
		transaction.PaymentMethod,
	)

	var transactionId int

	err := row.Scan(&transactionId)
	if err != nil {
		return 0, err
	}

	return transactionId, nil

}

func (model *DBModel) InsertOrderGetOrderID(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		INSERT into orders
			(widget_id, transaction_id, status_id, quantity, amount, created_at, updated_at, customer_id)
		values($1, $2, $3, $4, $5, $6, $7, $8)
		returning id
	`

	row := model.DB.QueryRowContext(
		ctx,
		query,
		order.WidgetId,
		order.TransactionId,
		order.StatusId,
		order.Quantity,
		order.Amount,
		time.Now(),
		time.Now(),
		order.CustomerID,
	)

	var orderId int

	err := row.Scan(&orderId)
	if err != nil {
		return 0, err
	}

	return orderId, nil

}

func (model *DBModel) InsertCustomerGetCustomerID(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		INSERT into customers
			(first_name, last_name, email, created_at, updated_at)
		values($1, $2, $3, $4, $5)
		returning id
	`
	var customerId int

	row := model.DB.QueryRowContext(ctx, query,
		customer.FirstName, customer.LastName, customer.Email, time.Now(), time.Now(),
	)

	err := row.Scan(&customerId)
	if err != nil {
		return 0, err
	}

	return customerId, nil
}

func (model *DBModel) GetUserByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		select
			id, first_name, last_name, email, password, created_at, updated_at
		from users
		where email = $1
	`
	var user User

	row := model.DB.QueryRowContext(ctx, query, strings.ToLower(email))
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return &user, err
	}

	return &user, nil
}

func (model *DBModel) GetUserByToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))

	const query = `
		select u.id, u.first_name, u.last_name, u.email
		from users u
		inner join tokens t on(u.id = t.user_id)
		where t.token_hash = $1
	`

	var user User

	err := model.DB.QueryRowContext(
		ctx,
		query,
		tokenHash[:],
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Models for the postgres db

type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventoryLevel"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"isRecurring"`
	PlanID         string    `json:"planId"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

type Order struct {
	ID            int       `json:"id"`
	WidgetId      int       `json:"widgetId"`
	TransactionId int       `json:"transactionId"`
	CustomerID    int       `json:"customerID"`
	StatusId      int       `json:"statusId"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"lastFour"`
	BankReturnCode      string    `json:"bankReturnCode"`
	PaymentIntent       string    `json:"paymentIntent"`
	PaymentMethod       string    `json:"paymentMethod"`
	TransactionStatusID int       `json:"transactionStatusId"`
	ExpiryMonth         int       `json:"expiryMonth"`
	ExpiryYear          int       `json:"expiryYear"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type TransactionData struct {
	Email           string
	FirstName       string
	LastName        string
	PaymentMethod   string
	PaymentIntent   string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     uint64
	ExpiryYear      uint64
	BankReturnCode  string
	WidgetId        int
}

func (model *DBModel) SetNewPassword(newPassword, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	const stmt = `
		update users
		set password = $1
		where email = $2
	`

	_, err = model.DB.ExecContext(
		ctx,
		stmt,
		passwordHash,
		email,
	)

	if err != nil {
		return err
	}

	return nil
}
