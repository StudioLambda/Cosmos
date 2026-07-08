// Package response contains response-writing helpers for HTTP handlers.
//
// It provides focused helpers for common response types while keeping standard
// net/http response semantics.
//
// # Initialization behavior
//
// This package has no global setup and can be used directly from handlers.
//
// Example
//
//	if err := response.JSON(w, http.StatusCreated, payload); err != nil {
//		return err
//	}
//
//	return nil
package response
