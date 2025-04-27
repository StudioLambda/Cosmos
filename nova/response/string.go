package response

import (
	"net/http"
	"text/template"
)

func String(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	return Raw(w, status, []byte(data))
}

func StringTemplate(w http.ResponseWriter, status int, tmpl template.Template, data any) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	return tmpl.Execute(w, data)
}
