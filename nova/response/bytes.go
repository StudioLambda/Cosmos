package response

import "net/http"

func Bytes(w http.ResponseWriter, status int, data []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")

	return Raw(w, status, data)
}

func Raw(w http.ResponseWriter, status int, data []byte) error {
	w.WriteHeader(status)

	_, err := w.Write(data)

	return err
}
