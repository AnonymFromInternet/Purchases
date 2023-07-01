package main

import (
	"fmt"
	"net/http"
)

func (application *application) handlerGetCreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "microservice works")
	if err != nil {
		return
	}
}
