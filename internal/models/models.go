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

func (model *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `
		SELECT id, name from widgets
		WHERE id = $1
	`

	var widget Widget

	row := model.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&widget.Id, &widget.Name)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

// Models for the postgres db

type Widget struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventoryLevel"`
	Price          int       `json:"price"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

type Order struct {
	Id            int
	WidgetIs      int
	TransactionId string
	StatusId      int
	Quantity      int
	Amount        int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Status struct {
	Id        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransactionStatus struct {
	Id        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	Id                  int
	Amount              int
	Currency            string
	LastFour            string
	BackReturnCode      string
	TransactionStatusId int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type User struct {
	Id        int
	FirstName string
	LastName  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
