package framework

import (
	"net/http"
	"slices"
	"sync"
)

// Hooks is a structure that can be used to
// attach specific hooks to the request / response
// lifecycle. It requires the use of the hooks middleware.
//
// It is safe for concurrent use.
type Hooks struct {
	mutex                  sync.Mutex
	afterResponseHooks     []AfterResponseHook
	beforeWriteHeaderHooks []BeforeWriteHeaderHook
	beforeWriteHooks       []BeforeWriteHook
}

type AfterResponseHook func(err error)
type BeforeWriteHeaderHook func(w http.ResponseWriter, status int)
type BeforeWriteHook func(w http.ResponseWriter, content []byte)

func NewHooks() *Hooks {
	return &Hooks{
		mutex:                  sync.Mutex{},
		beforeWriteHeaderHooks: []BeforeWriteHeaderHook{},
		beforeWriteHooks:       []BeforeWriteHook{},
	}
}

func (h *Hooks) BeforeWriteHeader(callbacks ...BeforeWriteHeaderHook) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.beforeWriteHeaderHooks = append(h.beforeWriteHeaderHooks, callbacks...)
}

func (h *Hooks) BeforeWriteHeaderFuncs() []BeforeWriteHeaderHook {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.beforeWriteHeaderHooks)
	slices.Reverse(clone)

	return clone
}

func (h *Hooks) BeforeWrite(callbacks ...BeforeWriteHook) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.beforeWriteHooks = append(h.beforeWriteHooks, callbacks...)
}

func (h *Hooks) BeforeWriteFuncs() []BeforeWriteHook {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.beforeWriteHooks)
	slices.Reverse(clone)

	return clone
}

func (h *Hooks) AfterResponse(callbacks ...AfterResponseHook) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.afterResponseHooks = append(h.afterResponseHooks, callbacks...)
}

func (h *Hooks) AfterResponseFuncs() []AfterResponseHook {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.afterResponseHooks)
	slices.Reverse(clone)

	return clone
}
