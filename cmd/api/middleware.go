package main

import "net/http"

func (application *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := application.checkTokenValidityGetUser(r)
		if err != nil {
			application.errorLog.Println("error in Auth middleware :", err)
			application.sendInvalidCredentials(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
