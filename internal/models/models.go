package models

import (
	"context"
	"database/sql"
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

func (model *DBModel) GetWidgetBy(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		SELECT id, name, description, inventory_level, price, coalesce(image, ''),
			created_at, updated_at
		FROM widgets
		WHERE id = $1
	`

	var widget Widget

	row := model.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&widget.ID, &widget.Name, &widget.Description, &widget.InventoryLevel, &widget.Price, &widget.Image,
		&widget.CreatedAt, &widget.UpdatedAt,
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
			(amount, currency, last_four, backend_return_code, transaction_status_id, created_at, updated_at)
		values($1, $2, $3, $4, $5, $6, $7)
	`

	result, err := model.DB.ExecContext(
		ctx,
		query,
		transaction.Amount,
		transaction.Currency,
		transaction.LastFour,
		transaction.BankReturnCode,
		transaction.TransactionStatusID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

func (model *DBModel) InsertOrderGetOrderID(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		INSERT into orders
			(widget_id, transaction_id, status_id, quantity, amount, created_at, updated_at)
		values($1, $2, $3, $4, $5, $6, $7)
	`

	result, err := model.DB.ExecContext(
		ctx,
		query,
		order.WidgetId,
		order.TransactionId,
		order.StatusId,
		order.Quantity,
		order.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

func (model *DBModel) InsertCustomerGetCustomerID(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		INSERT into customers
			(first_name, last_name, email, created_at, updated_at)
		values($1, $2, $3, $4, $5)
	`
	result, err := model.DB.ExecContext(ctx, query,
		customer.FirstName, customer.LastName, customer.Email, time.Now(), time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Models for the postgres db

type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventoryLevel"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
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

type TemplateData struct {
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
