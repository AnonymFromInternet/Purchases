package main

import (
	"encoding/json"
	"errors"
	"github.com/AnonymFromInternet/Purchases/internal/contentTypes"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"io"
	"net/http"
	"os"
)

type AnswerPayload struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Token   *models.Token `json:"token"` // TODO: Зачем указатель?
}

// readJSONInto check if request size is no greater than maxSizeInBytes and decode payload from request into arg
func (application *application) readJSONInto(ptrToData interface{}, w http.ResponseWriter, r *http.Request) {
	maxSizeInBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxSizeInBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(ptrToData)
	if err != nil {
		application.errorLog.Println("cannot decode data from payload", err)

		return
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		application.errorLog.Println(errors.New("body must have only a single JSON value"))
	}
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

func (application *application) createDirIfNotExist(path string) {
	const mode = 0755

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			application.errorLog.Println("cannot create dir ", err)
		}
	}
}
