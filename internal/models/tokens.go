package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken generates a token that lasts up to ttl and returns it
func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

// func (model *DBModel) InsertToken(tokenHash []byte, user *User) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()
//
// 	var tokenLastUpdatedTime time.Time
//
// 	query := `
// 		select updated_at from tokens
// 		where user_id = $1
// 	`
//
// 	err := model.DB.QueryRowContext(ctx, query, user.ID).Scan(&tokenLastUpdatedTime)
// 	if err != nil && err != sql.ErrNoRows {
// 		fmt.Println("IF 1")
// 		return err
// 	}
//
// 	if err == sql.ErrNoRows {
// 		fmt.Println("IF 2")
// 		statement := `
// 		insert into tokens (user_id, name, email, token_hash, created_at, updated_at)
// 		values ($1, $2, $3, $4, $5, $6)
// 	`
// 		_, err = model.DB.ExecContext(
// 			ctx,
// 			statement,
// 			user.ID,
// 			user.LastName,
// 			user.Email,
// 			tokenHash,
// 			time.Now(),
// 			time.Now(),
// 		)
// 		if err != nil {
// 			fmt.Println("IF 3")
// 			return err
// 		}
//
// 		return nil
// 	}
//
// 	tokenExpiryUpTo := tokenLastUpdatedTime.Add(24 * time.Hour)
//
// 	if tokenLastUpdatedTime.After(tokenExpiryUpTo) {
// 		fmt.Println("IF 4")
// 		statement := `
// 		delete from tokens
// 		where user_id = $1
// 	`
//
// 		_, err = model.DB.ExecContext(
// 			ctx,
// 			statement,
// 			user.ID,
// 		)
// 		if err != nil {
// 			return err
// 		}
//
// 		statement = `
// 		insert into tokens (user_id, name, email, token_hash, created_at, updated_at)
// 		values ($1, $2, $3, $4, $5, $6)
// 	`
// 		_, err = model.DB.ExecContext(
// 			ctx,
// 			query,
// 			user.ID,
// 			user.LastName,
// 			user.Email,
// 			tokenHash,
// 			time.Now(),
// 			time.Now(),
// 		)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	fmt.Println("NOTHING CASE")
//
// 	return nil
// }

func (model *DBModel) InsertToken(tokenHash []byte, user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	statement := `
		delete from tokens
		where user_id = $1
	`

	_, err := model.DB.ExecContext(
		ctx,
		statement,
		user.ID,
	)
	if err != nil {
		return err
	}

	statement = `
		insert into tokens (user_id, name, email, token_hash, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6)
	`
	_, err = model.DB.ExecContext(
		ctx,
		statement,
		user.ID,
		user.LastName,
		user.Email,
		tokenHash,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}
