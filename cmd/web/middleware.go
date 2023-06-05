package main

import "net/http"

func SessionLoadMiddleWare(next http.Handler) http.Handler {
	return SessionManager.LoadAndSave(next)
}
