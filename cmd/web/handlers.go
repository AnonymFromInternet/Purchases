package main

import (
	"net/http"
)

func (application *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := application.renderTemplate(w, r, "virtual-terminal", nil); err != nil {
		application.errorLog.Println(err)

		return
	}
}
