package framework

import (
	"slices"
	"sync"

	"github.com/studiolambda/cosmos/contract"
)

// Hooks is a structure that can be used to
// attach specific hooks to the request / response
// lifecycle. It requires the use of the hooks middleware.
//
// It is safe for concurrent use.
type Hooks struct {
	mutex                  sync.Mutex
	afterResponseHooks     []contract.AfterResponseHook
	beforeWriteHeaderHooks []contract.BeforeWriteHeaderHook
	beforeWriteHooks       []contract.BeforeWriteHook
}

func NewHooks() *Hooks {
	return &Hooks{
		mutex:                  sync.Mutex{},
		beforeWriteHeaderHooks: []contract.BeforeWriteHeaderHook{},
		beforeWriteHooks:       []contract.BeforeWriteHook{},
	}
}

func (h *Hooks) BeforeWriteHeader(callbacks ...contract.BeforeWriteHeaderHook) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.beforeWriteHeaderHooks = append(h.beforeWriteHeaderHooks, callbacks...)
}

func (h *Hooks) BeforeWriteHeaderFuncs() []contract.BeforeWriteHeaderHook {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.beforeWriteHeaderHooks)
	slices.Reverse(clone)

	return clone
}

func (h *Hooks) BeforeWrite(callbacks ...contract.BeforeWriteHook) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.beforeWriteHooks = append(h.beforeWriteHooks, callbacks...)
}

func (h *Hooks) BeforeWriteFuncs() []contract.BeforeWriteHook {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.beforeWriteHooks)
	slices.Reverse(clone)

	return clone
}

func (h *Hooks) AfterResponse(callbacks ...contract.AfterResponseHook) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.afterResponseHooks = append(h.afterResponseHooks, callbacks...)
}

func (h *Hooks) AfterResponseFuncs() []contract.AfterResponseHook {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clone := slices.Clone(h.afterResponseHooks)
	slices.Reverse(clone)

	return clone
}
