package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"text/template"
)

func Raw(w http.ResponseWriter, status int, data []byte) error {
	w.WriteHeader(status)

	_, err := w.Write(data)

	return err
}

func Bytes(w http.ResponseWriter, status int, data []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")

	return Raw(w, status, data)
}

func String(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	return Raw(w, status, []byte(data))
}

func StringTemplate(w http.ResponseWriter, status int, tmpl template.Template, data any) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	return tmpl.Execute(w, data)
}

func HTML(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/html")

	return Raw(w, status, []byte(data))
}

func HTMLTemplate(w http.ResponseWriter, status int, tmpl template.Template, data any) error {
	w.Header().Set("Content-Type", "text/html")

	return tmpl.Execute(w, data)
}

// JSON serializes and writes the given json value to the
// response writer.
//
// It automatically sets the content-type to `application/json`
func JSON[T any](w http.ResponseWriter, status int, data T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func XML(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	return xml.NewEncoder(w).Encode(data)
}
