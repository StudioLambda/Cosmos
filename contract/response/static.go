package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"text/template"
)

// Raw writes raw byte data to the response writer with the specified status code.
// It does not set any Content-Type header, allowing the caller full control over
// the response format. This is the most basic response function that other
// response functions build upon.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
//   - data: The raw byte data to write to the response
func Raw(w http.ResponseWriter, status int, data []byte) error {
	w.WriteHeader(status)

	_, err := w.Write(data)

	return err
}

// Status sets the HTTP status code for the response without writing any body content.
// This is useful for responses that only need to communicate a status (like 204 No Content,
// 201 Created, or various error codes) without additional data. The response body will
// remain empty after calling this function.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
func Status(w http.ResponseWriter, status int) error {
	w.WriteHeader(status)

	return nil
}

// Bytes writes binary data to the response writer with the appropriate
// Content-Type header for binary content (application/octet-stream).
// This is useful for serving files, images, or any binary data that
// should be treated as a download or binary stream.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
//   - data: The binary data to write to the response
func Bytes(w http.ResponseWriter, status int, data []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")

	return Raw(w, status, data)
}

// String writes plain text data to the response writer with the appropriate
// Content-Type header for text content (text/plain; charset=utf-8).
// This is ideal for serving plain text responses, logs, or simple text data.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
//   - data: The string data to write to the response
func String(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	return Raw(w, status, []byte(data))
}

// StringTemplate executes a text template with the provided data and writes
// the result as plain text to the response writer. The Content-Type is set
// to text/plain; charset=utf-8. This is useful for generating plain text
// content from templates, such as emails or configuration files.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set (note: this is set after template execution)
//   - tmpl: The text template to execute
//   - data: The data to pass to the template for execution
func StringTemplate(w http.ResponseWriter, status int, tmpl template.Template, data any) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	return tmpl.Execute(w, data)
}

// HTML writes HTML content to the response writer with the appropriate
// Content-Type header for HTML (text/html). This is used for serving
// static HTML content or HTML strings that have been pre-generated.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
//   - data: The HTML string to write to the response
func HTML(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/html")

	return Raw(w, status, []byte(data))
}

// HTMLTemplate executes an HTML template with the provided data and writes
// the result as HTML to the response writer. The Content-Type is set to
// text/html. This is the standard way to serve dynamic HTML pages in web
// applications using Go's html/template package.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set (note: this is set after template execution)
//   - tmpl: The HTML template to execute
//   - data: The data to pass to the template for execution
func HTMLTemplate(w http.ResponseWriter, status int, tmpl template.Template, data any) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)

	return tmpl.Execute(w, data)
}

// JSON serializes the given data to JSON format and writes it to the
// response writer. It automatically sets the Content-Type header to
// application/json and uses Go's json.Encoder for efficient streaming
// serialization. This is the standard way to serve JSON API responses.
//
// The function uses generics to provide type safety for the input data
// while still accepting any serializable type.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
//   - data: The data to serialize as JSON (must be JSON-serializable)
func JSON[T any](w http.ResponseWriter, status int, data T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// XML serializes the given data to XML format and writes it to the
// response writer. It automatically sets the Content-Type header to
// application/xml and uses Go's xml.Encoder for streaming serialization.
// This is useful for serving XML API responses or RSS feeds.
//
// The data should be a struct with appropriate xml tags for proper
// marshaling, or implement xml.Marshaler interface for custom serialization.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to set
//   - data: The data to serialize as XML (must be XML-serializable)
func XML(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	return xml.NewEncoder(w).Encode(data)
}

// Redirect sends an HTTP redirect response to the specified URL with the given status code.
// This is a generic redirect function that allows you to specify any redirect status code.
// The Location header is set to the provided URL and the appropriate status code is returned.
//
// Common redirect status codes:
//   - 301: Moved Permanently
//   - 302: Found (temporary redirect)
//   - 303: See Other
//   - 307: Temporary Redirect
//   - 308: Permanent Redirect
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP redirect status code to set
//   - url: The URL to redirect the user to
func Redirect(w http.ResponseWriter, status int, url string) error {
	w.Header().Set("Location", url)

	return Status(w, status)
}
