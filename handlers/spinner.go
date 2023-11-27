package handlers

import "net/http"

func ServeSpinner(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/spinner.svg")
}