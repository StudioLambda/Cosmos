package response_test

import (
	"encoding/json"
	"encoding/xml"
	htmltemplate "html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/response"
)

func TestRawWritesBytesWithStatus(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Raw(w, http.StatusOK, []byte("hello"))

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello", w.Body.String())
}

func TestRawSetsDefaultContentType(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Raw(w, http.StatusOK, []byte("data"))

	require.NoError(t, err)
	require.Equal(
		t,
		"application/octet-stream",
		w.Header().Get("Content-Type"),
	)
}

func TestRawPreservesExistingContentType(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "text/plain")

	err := response.Raw(w, http.StatusOK, []byte("data"))

	require.NoError(t, err)
	require.Equal(t, "text/plain", w.Header().Get("Content-Type"))
}

func TestRawWithEmptyBody(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Raw(w, http.StatusOK, []byte{})

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Body.String())
}

func TestStatusSetsStatusCode(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Status(w, http.StatusNoContent)

	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestStatusWritesNoBody(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Status(w, http.StatusCreated)

	require.NoError(t, err)
	require.Empty(t, w.Body.String())
}

func TestBytesWritesWithOctetStreamContentType(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Bytes(
		w, http.StatusOK, []byte{0x01, 0x02, 0x03},
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"application/octet-stream",
		w.Header().Get("Content-Type"),
	)
	require.Equal(t, []byte{0x01, 0x02, 0x03}, w.Body.Bytes())
}

func TestStringWritesPlainText(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.String(w, http.StatusOK, "hello world")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"text/plain; charset=utf-8",
		w.Header().Get("Content-Type"),
	)
	require.Equal(t, "hello world", w.Body.String())
}

func TestStringEmptyBody(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.String(w, http.StatusOK, "")

	require.NoError(t, err)
	require.Empty(t, w.Body.String())
}

func TestStringTemplateExecutes(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	tmpl := template.Must(
		template.New("test").Parse("Hello, {{.Name}}!"),
	)

	err := response.StringTemplate(
		w, http.StatusOK, *tmpl, map[string]string{"Name": "World"},
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"text/plain; charset=utf-8",
		w.Header().Get("Content-Type"),
	)
	require.Equal(t, "Hello, World!", w.Body.String())
}

func TestStringTemplateReturnsErrorOnBadTemplate(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	tmpl := template.Must(
		template.New("test").Parse("{{.Name}}"),
	)

	err := response.StringTemplate(w, http.StatusOK, *tmpl, 42)

	require.Error(t, err)
}

func TestHTMLWritesHTMLContent(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.HTML(
		w, http.StatusOK, "<h1>Hello</h1>",
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"text/html; charset=utf-8",
		w.Header().Get("Content-Type"),
	)
	require.Equal(t, "<h1>Hello</h1>", w.Body.String())
}

func TestHTMLTemplateExecutes(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	tmpl := htmltemplate.Must(
		htmltemplate.New("test").Parse("<p>{{.Name}}</p>"),
	)

	err := response.HTMLTemplate(
		w, http.StatusOK, *tmpl, map[string]string{"Name": "World"},
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"text/html; charset=utf-8",
		w.Header().Get("Content-Type"),
	)
	require.Equal(t, "<p>World</p>", w.Body.String())
}

func TestHTMLTemplateReturnsErrorOnBadTemplate(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	tmpl := htmltemplate.Must(
		htmltemplate.New("test").Parse("{{.Name}}"),
	)

	err := response.HTMLTemplate(w, http.StatusOK, *tmpl, 42)

	require.Error(t, err)
}

func TestJSONWritesJSONContent(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	w := httptest.NewRecorder()

	err := response.JSON(
		w, http.StatusOK, payload{Name: "alice", Age: 30},
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"application/json",
		w.Header().Get("Content-Type"),
	)

	var result payload
	decodeErr := json.Unmarshal(w.Body.Bytes(), &result)

	require.NoError(t, decodeErr)
	require.Equal(t, "alice", result.Name)
	require.Equal(t, 30, result.Age)
}

func TestJSONWritesNullForNil(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.JSON[any](w, http.StatusOK, nil)

	require.NoError(t, err)
	require.Equal(t, "null\n", w.Body.String())
}

func TestJSONSetsStatusCode(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.JSON(w, http.StatusCreated, map[string]string{
		"id": "1",
	})

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestXMLWritesXMLContent(t *testing.T) {
	t.Parallel()

	type payload struct {
		XMLName xml.Name `xml:"item"`
		Name    string   `xml:"name"`
	}

	w := httptest.NewRecorder()

	err := response.XML(
		w, http.StatusOK, payload{Name: "bob"},
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		"application/xml",
		w.Header().Get("Content-Type"),
	)
	require.Contains(t, w.Body.String(), "<name>bob</name>")
}

func TestXMLSetsStatusCode(t *testing.T) {
	t.Parallel()

	type payload struct {
		XMLName xml.Name `xml:"item"`
		ID      int      `xml:"id"`
	}

	w := httptest.NewRecorder()

	err := response.XML(
		w, http.StatusCreated, payload{ID: 1},
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestRedirectSetsLocationHeader(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Redirect(
		w, http.StatusFound, "https://example.com",
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusFound, w.Code)
	require.Equal(
		t,
		"https://example.com",
		w.Header().Get("Location"),
	)
}

func TestRedirectWritesNoBody(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.Redirect(
		w, http.StatusMovedPermanently, "/new-path",
	)

	require.NoError(t, err)
	require.Empty(t, w.Body.String())
}

func TestSafeRedirectAllowsRelativePath(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(
		w, http.StatusFound, "/dashboard",
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusFound, w.Code)
	require.Equal(t, "/dashboard", w.Header().Get("Location"))
}

func TestSafeRedirectAllowsRelativePathWithQuery(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(
		w, http.StatusFound, "/search?q=hello",
	)

	require.NoError(t, err)
	require.Equal(
		t,
		"/search?q=hello",
		w.Header().Get("Location"),
	)
}

func TestSafeRedirectRejectsAbsoluteURL(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(
		w, http.StatusFound, "https://evil.com",
	)

	require.ErrorIs(t, err, response.ErrUnsafeRedirect)
}

func TestSafeRedirectRejectsProtocolRelativeURL(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(
		w, http.StatusFound, "//evil.com",
	)

	require.ErrorIs(t, err, response.ErrUnsafeRedirect)
}

func TestSafeRedirectRejectsNonSlashPrefix(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(
		w, http.StatusFound, "evil.com/path",
	)

	require.ErrorIs(t, err, response.ErrUnsafeRedirect)
}

func TestSafeRedirectRejectsEmptyURL(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(w, http.StatusFound, "")

	require.ErrorIs(t, err, response.ErrUnsafeRedirect)
}

func TestSafeRedirectRejectsUnparseableURL(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	err := response.SafeRedirect(
		w, http.StatusFound, "/path\x7f",
	)

	require.ErrorIs(t, err, response.ErrUnsafeRedirect)
}

func TestErrUnsafeRedirectMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"unsafe redirect URL: must be a relative path",
		response.ErrUnsafeRedirect.Error(),
	)
}
