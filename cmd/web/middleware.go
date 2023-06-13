package main

import "net/http"

func SessionLoadMiddleWare(next http.Handler) http.Handler {
	return SessionManager.LoadAndSave(next)
}

func (application *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !application.SessionManager.Exists(r.Context(), "userID") {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)

			return
		}

		next.ServeHTTP(w, r)
	})
}
