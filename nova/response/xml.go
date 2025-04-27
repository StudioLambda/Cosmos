package response

import (
	"encoding/xml"
	"net/http"
)

func XML(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	return xml.NewEncoder(w).Encode(data)
}
