package atlas

import "github.com/studiolambda/cosmos/nova"

type App interface {
	// Register is called when the application
	// should register the http routes.
	Register(router *nova.Router) error
}
