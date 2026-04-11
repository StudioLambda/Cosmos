package middleware_test

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func TestRecoverFromErrorPanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(errors.New("database connection lost"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "database connection lost")
}

func TestRecoverFromStringPanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic("something broke")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "something broke")
}

type testStringer struct {
	message string
}

func (stringer testStringer) String() string {
	return stringer.message
}

func TestRecoverFromStringerPanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(testStringer{message: "stringer panic"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "stringer panic")
}

func TestRecoverFromReaderPanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(strings.NewReader("reader body content"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "reader body content")
}

// testTextMarshaler implements encoding.TextMarshaler for testing.
type testTextMarshaler struct {
	text string
	err  error
}

// Ensure testTextMarshaler does NOT implement fmt.Stringer
// so it falls through to the TextMarshaler case.
var _ encoding.TextMarshaler = testTextMarshaler{}

func (marshaler testTextMarshaler) MarshalText() ([]byte, error) {
	if marshaler.err != nil {
		return nil, marshaler.err
	}

	return []byte(marshaler.text), nil
}

func TestRecoverFromTextMarshalerPanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(testTextMarshaler{text: "marshaled text"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "marshaled text")
}

func TestRecoverFromTextMarshalerError(t *testing.T) {
	t.Parallel()

	marshalErr := errors.New("marshal failed")
	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(testTextMarshaler{err: marshalErr})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrFailedRecovering)
	require.ErrorContains(t, err, "marshal failed")
}

func TestRecoverFromUnknownTypePanic(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(42)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "42")
}

type failingReader struct {
	err error
}

func (reader failingReader) Read(p []byte) (int, error) {
	return 0, reader.err
}

func TestRecoverFromReaderError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("read failed")
	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(failingReader{err: readErr})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrFailedRecovering)
	require.ErrorContains(t, err, "read failed")
}

func TestRecoverWithCustomHandler(t *testing.T) {
	t.Parallel()

	customErr := errors.New("custom recovery")
	handler := middleware.RecoverWith(func(value any) error {
		return fmt.Errorf("%w: %v", customErr, value)
	})(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, customErr)
	require.ErrorContains(t, err, "boom")
}

func TestRecoverNoPanicPassesThrough(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.Recover()(framework.Handler(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		called = true
		w.WriteHeader(http.StatusOK)

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := handler.Record(req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRecoverPassesThroughHandlerError(t *testing.T) {
	t.Parallel()

	expected := errors.New("normal error")
	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		return expected
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.ErrorIs(t, err, expected)
}

// readerStringer implements both io.Reader and fmt.Stringer.
// fmt.Stringer is checked before io.Reader in the type switch,
// so this should be handled as a Stringer.
type readerStringer struct {
	message string
}

func (rs readerStringer) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (rs readerStringer) String() string {
	return rs.message
}

func TestRecoverStringerTakesPrecedenceOverReader(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover()(func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		panic(readerStringer{message: "stringer wins"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	err := handler(rec, req)

	require.Error(t, err)
	require.ErrorIs(t, err, middleware.ErrRecoverUnexpected)
	require.ErrorContains(t, err, "stringer wins")
}
