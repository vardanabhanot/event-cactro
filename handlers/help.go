package handlers

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed help.html
var helpHTMLTmpl string

var helpTmpl = template.Must(template.New("help").Parse(helpHTMLTmpl))

// Help handles GET /help — renders the API reference UI with a dynamic base URL
func Help(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	data := map[string]string{
		"BaseURL": scheme + "://" + r.Host + "/api",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	helpTmpl.Execute(w, data)
}
