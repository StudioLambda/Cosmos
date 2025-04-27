package response

import (
	"html/template"
	"net/http"
)

func HTML(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/html")

	return Raw(w, status, []byte(data))
}

func HTMLTemplate(w http.ResponseWriter, status int, tmpl template.Template, data any) error {
	w.Header().Set("Content-Type", "text/html")

	return tmpl.Execute(w, data)
}
