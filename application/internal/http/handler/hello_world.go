package handler

import (
	"net/http"

	"github.com/studiolambda/cosmos/application/internal/http/api"
	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"
)

type HelloWorld struct {
	//
}

func NewHelloWorld() *HelloWorld {
	return &HelloWorld{
		//
	}
}

func (h *HelloWorld) Routes(router *framework.Router) {
	router.Get("/", h.HelloWorld)
}

func (h *HelloWorld) HelloWorld(w http.ResponseWriter, r *http.Request) error {
	res := api.HelloWorldResponse{
		Hello: "World!",
	}

	return response.JSON(w, http.StatusOK, res)
}
