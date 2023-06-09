package main

import (
	"encoding/json"
	"errors"
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"io"
	"net/http"
)

type LoginPagePayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AnswerPayload struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// readJSON check if request size is no greater than maxSizeInBytes and decode payload from request into arg
func (application *application) readJSON(w http.ResponseWriter, r *http.Request, ptrToData interface{}) {
	maxSizeInBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxSizeInBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(ptrToData)
	if err != nil {
		application.errorLog.Println("cannot decode data from payload", err)
		application.sendBadRequest(w, r, err)

		return
	}

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
