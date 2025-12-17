package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
)

func Session(r *http.Request, key any) (contract.Session, bool) {
	s, ok := r.Context().Value(key).(contract.Session)

	return s, ok
}
