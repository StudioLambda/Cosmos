package request

import (
	"io"
	"net/http"
)

func Bytes(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}
