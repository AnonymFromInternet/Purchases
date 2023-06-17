package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"strings"
)

type LoginPagePayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AnswerPayload struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Token   *models.Token `json:"token"` // TODO: Зачем указатель?
}

// readJSONInto check if request size is no greater than maxSizeInBytes and decode payload from request into arg
func (application *application) readJSONInto(ptrToData interface{}, w http.ResponseWriter, r *http.Request) {
	fmt.Println("readJSONInto()")
	maxSizeInBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxSizeInBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(ptrToData)
	if err != nil {
		application.errorLog.Println("cannot decode data from payload", err)
		application.sendBadRequest(w, r, err)

		return
	}

	fmt.Println("")

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		application.errorLog.Println(errors.New("body must have only a single JSON value"))
		application.sendBadRequest(w, r, err)
	}
}

func (application *application) sendBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	var payload AnswerPayload

	payload.Error = true
	payload.Message = err.Error()

	output, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		application.errorLog.Println("cannot marshal payload into slice of bytes", err)

		return
	}

	w.Header().Set(contentTypes.ContentTypeKey, contentTypes.ApplicationJSON)
	_, _ = w.Write(output)
}

func (application *application) convertToJsonAndSend(data interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		application.errorLog.Println("cannot convert data into json", err)

		return
	}

	w.Header().Set(contentTypes.ContentTypeKey, contentTypes.ApplicationJSON)
	_, _ = w.Write(output)
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

func (application *application) isPasswordValid(hash, passwordGivenByUser string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordGivenByUser))
	if err != nil {
		return false
	}

	return true
}

func (application *application) checkTokenValidityGetUser(r *http.Request) (*models.User, error) {
	authorizationData := r.Header.Get("Authorization")
	if authorizationData == "" {
		application.errorLog.Println("cannot get auth data from headers")

		return nil, errors.New("authentication was not successful")
	}

	sliceFromAuthorizationData := strings.Split(authorizationData, " ")
	if len(sliceFromAuthorizationData) < 2 || sliceFromAuthorizationData[0] != "Bearer" {
		application.errorLog.Println("cannot slice auth data")

		return nil, errors.New("authentication was not successful")
	}

	token := sliceFromAuthorizationData[1]
	if len(token) < 26 {
		application.errorLog.Println("token length is less than 26")

		return nil, errors.New("invalid token length")
	}

	user, err := application.DB.GetUserByToken(token)
	if err != nil {
		application.errorLog.Println("cannot get user by token :", err)

		return nil, err
	}

	return user, nil
}
