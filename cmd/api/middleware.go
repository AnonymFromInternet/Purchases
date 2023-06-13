package main

import (
	"fmt"
	"net/http"
)

func (application *application) AuthMiddleware(next http.Handler) http.Handler {
	fmt.Print("AuthMiddleware()")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := application.checkTokenValidityGetUser(r)
		if err != nil {
			application.errorLog.Println("error in auth middleware :", err)
			application.sendInvalidCredentials(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
