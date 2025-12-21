package hook

import (
	"net/http"
	"slices"
	"sync"
)

// Manager is a structure that can be used to
// attach specific hooks to the request / response
// lifecycle. It requires the use of the hooks middleware.
//
// It is safe for concurrent use.
type Manager struct {
	mutex                  sync.Mutex
	afterResponseFuncs     []AfterResponseFunc
	beforeWriteHeaderFuncs []BeforeWriteHeaderFunc
	beforeWriteFuncs       []BeforeWriteFunc
}

type AfterResponseFunc func(err error)
type BeforeWriteHeaderFunc func(w http.ResponseWriter, status int)
type BeforeWriteFunc func(w http.ResponseWriter, content []byte)
type key struct{}

var Key = key{}

func NewManager() *Manager {
	return &Manager{
		mutex:                  sync.Mutex{},
		beforeWriteHeaderFuncs: []BeforeWriteHeaderFunc{},
		beforeWriteFuncs:       []BeforeWriteFunc{},
	}
}

func (h *Manager) BeforeWriteHeader(callbacks ...BeforeWriteHeaderFunc) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.beforeWriteHeaderFuncs = append(h.beforeWriteHeaderFuncs, callbacks...)
}

func (h *Manager) BeforeWriteHeaderFuncs() []BeforeWriteHeaderFunc {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.beforeWriteHeaderFuncs)
	slices.Reverse(clone)

	return clone
}

func (h *Manager) BeforeWrite(callbacks ...BeforeWriteFunc) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.beforeWriteFuncs = append(h.beforeWriteFuncs, callbacks...)
}

func (h *Manager) BeforeWriteFuncs() []BeforeWriteFunc {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.beforeWriteFuncs)
	slices.Reverse(clone)

	return clone
}

func (h *Manager) AfterResponse(callbacks ...AfterResponseFunc) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.afterResponseFuncs = append(h.afterResponseFuncs, callbacks...)
}

func (h *Manager) AfterResponseFuncs() []AfterResponseFunc {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.afterResponseFuncs)
	slices.Reverse(clone)

	return clone
}
