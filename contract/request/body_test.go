package request_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/request"
)

func TestBytesReadsEntireBody(t *testing.T) {
	t.Parallel()

	body := "hello world"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.Bytes(r)

	require.NoError(t, err)
	require.Equal(t, []byte(body), result)
}

func TestBytesEmptyBody(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	result, err := request.Bytes(r)

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestBytesErrorOnFailedRead(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/", errReader{})

	_, err := request.Bytes(r)

	require.Error(t, err)
}

func TestLimitedBytesReturnsErrorWhenBodyExceedsLimit(t *testing.T) {
	t.Parallel()

	body := "abcdefghij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	_, err := request.LimitedBytes(r, 5)

	require.ErrorIs(t, err, request.ErrBodyTooLarge)
}

func TestLimitedBytesReadsFullBodyUnderLimit(t *testing.T) {
	t.Parallel()

	body := "abc"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedBytes(r, 100)

	require.NoError(t, err)
	require.Equal(t, []byte("abc"), result)
}

func TestLimitedBytesUsesDefaultOnNegativeMaxSize(t *testing.T) {
	t.Parallel()

	body := "short"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedBytes(r, -1)

	require.NoError(t, err)
	require.Equal(t, []byte("short"), result)
}

func TestStringReadsBodyAsString(t *testing.T) {
	t.Parallel()

	body := "hello string"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.String(r)

	require.NoError(t, err)
	require.Equal(t, body, result)
}

func TestStringEmptyBody(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	result, err := request.String(r)

	require.NoError(t, err)
	require.Equal(t, "", result)
}

func TestStringErrorOnFailedRead(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/", errReader{})

	_, err := request.String(r)

	require.Error(t, err)
}

func TestLimitedStringReturnsErrorWhenBodyExceedsLimit(t *testing.T) {
	t.Parallel()

	body := "abcdefghij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	_, err := request.LimitedString(r, 5)

	require.ErrorIs(t, err, request.ErrBodyTooLarge)
}

func TestLimitedStringReadsFullBodyUnderLimit(t *testing.T) {
	t.Parallel()

	body := "abc"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedString(r, 100)

	require.NoError(t, err)
	require.Equal(t, "abc", result)
}

func TestLimitedStringUsesDefaultOnNegativeMaxSize(t *testing.T) {
	t.Parallel()

	body := "short"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedString(r, -1)

	require.NoError(t, err)
	require.Equal(t, "short", result)
}

func TestLimitedStringErrorOnFailedRead(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/", errReader{})

	_, err := request.LimitedString(r, 10)

	require.Error(t, err)
}

func TestJSONDecodesValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := `{"name":"alice","age":30}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.JSON[payload](r)

	require.NoError(t, err)
	require.Equal(t, "alice", result.Name)
	require.Equal(t, 30, result.Age)
}

func TestJSONReturnsErrorOnInvalidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{invalid json}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.JSON[payload](r)

	require.Error(t, err)
}

func TestJSONIgnoresUnknownFields(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"alice","extra":"ignored"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.JSON[payload](r)

	require.NoError(t, err)
	require.Equal(t, "alice", result.Name)
}

func TestStrictJSONDecodesValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"bob"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.StrictJSON[payload](r)

	require.NoError(t, err)
	require.Equal(t, "bob", result.Name)
}

func TestStrictJSONRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"bob","extra":"bad"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.StrictJSON[payload](r)

	require.Error(t, err)
}

func TestStrictJSONReturnsErrorOnInvalidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `not json`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.StrictJSON[payload](r)

	require.Error(t, err)
}

func TestLimitedJSONDecodesValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"carol"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.LimitedJSON[payload](r, 1024)

	require.NoError(t, err)
	require.Equal(t, "carol", result.Name)
}

func TestLimitedJSONUsesDefaultOnNegativeMaxSize(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"carol"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.LimitedJSON[payload](r, -1)

	require.NoError(t, err)
	require.Equal(t, "carol", result.Name)
}

func TestLimitedJSONReturnsErrorOnInvalidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `bad`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.LimitedJSON[payload](r, 1024)

	require.Error(t, err)
}

func TestStrictLimitedJSONDecodesValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"dan"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.StrictLimitedJSON[payload](r, 1024)

	require.NoError(t, err)
	require.Equal(t, "dan", result.Name)
}

func TestStrictLimitedJSONRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"dan","extra":"bad"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.StrictLimitedJSON[payload](r, 1024)

	require.Error(t, err)
}

func TestStrictLimitedJSONUsesDefaultOnNegativeMaxSize(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"dan"}`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.StrictLimitedJSON[payload](r, -1)

	require.NoError(t, err)
	require.Equal(t, "dan", result.Name)
}

func TestStrictLimitedJSONReturnsErrorOnInvalidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `not json`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.StrictLimitedJSON[payload](r, 1024)

	require.Error(t, err)
}

func TestXMLDecodesValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `xml:"name"`
	}

	body := `<payload><name>eve</name></payload>`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.XML[payload](r)

	require.NoError(t, err)
	require.Equal(t, "eve", result.Name)
}

func TestXMLReturnsErrorOnInvalidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `xml:"name"`
	}

	body := `not xml at all <<<`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.XML[payload](r)

	require.Error(t, err)
}

func TestLimitedXMLDecodesValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `xml:"name"`
	}

	body := `<payload><name>frank</name></payload>`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.LimitedXML[payload](r, 1024)

	require.NoError(t, err)
	require.Equal(t, "frank", result.Name)
}

func TestLimitedXMLUsesDefaultOnNegativeMaxSize(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `xml:"name"`
	}

	body := `<payload><name>frank</name></payload>`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	result, err := request.LimitedXML[payload](r, -1)

	require.NoError(t, err)
	require.Equal(t, "frank", result.Name)
}

func TestLimitedXMLReturnsErrorOnInvalidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `xml:"name"`
	}

	body := `not xml <<<`
	r := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(body),
	)

	_, err := request.LimitedXML[payload](r, 1024)

	require.Error(t, err)
}

func TestLimitedBytesReturnsDataWhenBodyWithinLimit(t *testing.T) {
	t.Parallel()

	body := "hello"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedBytes(r, 10)

	require.NoError(t, err)
	require.Equal(t, []byte("hello"), result)
}

func TestLimitedBytesReturnsDataWhenBodyExactlyAtLimit(t *testing.T) {
	t.Parallel()

	body := "12345"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedBytes(r, 5)

	require.NoError(t, err)
	require.Equal(t, []byte("12345"), result)
}

func TestLimitedStringReturnsDataWhenBodyWithinLimit(t *testing.T) {
	t.Parallel()

	body := "hello"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := request.LimitedString(r, 10)

	require.NoError(t, err)
	require.Equal(t, "hello", result)
}

func TestLimitedJSONReturnsErrorWhenBodyExceedsLimit(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"this is a very long name that exceeds the limit"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	_, err := request.LimitedJSON[payload](r, 10)

	require.ErrorIs(t, err, request.ErrBodyTooLarge)
}

func TestStrictLimitedJSONReturnsErrorWhenBodyExceedsLimit(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"this is a very long name that exceeds the limit"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	_, err := request.StrictLimitedJSON[payload](r, 10)

	require.ErrorIs(t, err, request.ErrBodyTooLarge)
}

func TestLimitedXMLReturnsErrorWhenBodyExceedsLimit(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `xml:"name"`
	}

	body := `<payload><name>this is a very long name that exceeds the limit</name></payload>`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	_, err := request.LimitedXML[payload](r, 10)

	require.ErrorIs(t, err, request.ErrBodyTooLarge)
}

func TestErrBodyTooLargeIsCheckableWithErrorsIs(t *testing.T) {
	t.Parallel()

	body := "abcdefghij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	_, err := request.LimitedBytes(r, 5)

	require.True(t, errors.Is(err, request.ErrBodyTooLarge))
}

// errReader is an io.Reader that always returns an error.
type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
