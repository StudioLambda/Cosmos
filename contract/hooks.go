package contract

import "net/http"

type hooksKey struct{}

var HooksKey = hooksKey{}

type AfterResponseHook func(err error)
type BeforeWriteHeaderHook func(w http.ResponseWriter, status int)
type BeforeWriteHook func(w http.ResponseWriter, content []byte)

type Hooks interface {
	AfterResponse(callbacks ...AfterResponseHook)
	AfterResponseFuncs() []AfterResponseHook
	BeforeWrite(callbacks ...BeforeWriteHook)
	BeforeWriteFuncs() []BeforeWriteHook
	BeforeWriteHeader(callbacks ...BeforeWriteHeaderHook)
	BeforeWriteHeaderFuncs() []BeforeWriteHeaderHook
}
