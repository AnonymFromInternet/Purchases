package main

import (
	"encoding/json"
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type AnswerPayload struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Token   *models.Token `json:"token"` // TODO: Зачем указатель?
}

func (application *application) isPasswordValid(hash, passwordGivenByUser string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordGivenByUser))
	if err != nil {
		return false
	}

	return true
}

func (application *application) sendInvalidCredentials(w http.ResponseWriter) {
	var answerPayload AnswerPayload

	answerPayload.Error = true
	answerPayload.Message = "invalid authentication credentials"

	output, err := json.MarshalIndent(answerPayload, "", " ")
	if err != nil {
		application.errorLog.Println("cannot convert data into json", err)

		return
	}

	w.Header().Set(contentTypes.ContentTypeKey, contentTypes.ApplicationJSON)
	_, _ = w.Write(output)
}
