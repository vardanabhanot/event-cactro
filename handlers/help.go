package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed help.html
var helpHTML []byte

// Help handles GET /help — renders the API reference UI
func Help(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(helpHTML)
}
