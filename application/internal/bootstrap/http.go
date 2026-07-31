package bootstrap

import (
	"net/http"

	"github.com/studiolambda/cosmos/application/internal/http/handler"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

// NewHTTPServer creates an HTTP server from http.server configuration.
func NewHTTPServer(configuration *contract.Configuration, router *framework.Router) *http.Server {
	config := configuration.From[framework.ServerConfig]("http.server")

	return framework.NewServer(config, router)
}

// NewHTTPRouter creates an HTTP router with configured middleware.
func NewHTTPRouter(configuration *contract.Configuration, logger *contract.Logger) *framework.Router {
	router := framework.New()
	cors := configuration.From[middleware.CORSConfig]("http.cors")
	csrf := configuration.From[middleware.CSRFConfig]("http.csrf")
	secureHeaders := configuration.From[middleware.SecureHeadersConfig]("http.secure_headers")
	correlation := configuration.From[middleware.CorrelationConfig]("observability.correlation")

	router.Use(
		middleware.Logger(logger),
		middleware.Recover(),
		middleware.Correlation(correlation),
		middleware.SecureHeaders(secureHeaders),
		middleware.CORS(cors),
		middleware.CSRF(csrf),
	)

	handler.
		NewHelloWorld().
		Routes(router)

	return router
}
