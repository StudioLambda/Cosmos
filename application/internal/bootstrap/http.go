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
	config := framework.ServerConfigFrom(configuration, "http.server")

	return framework.NewServer(config, router)
}

// NewHTTPRouter creates an HTTP router with configured middleware.
func NewHTTPRouter(configuration *contract.Configuration, logger *contract.Logger, cache *contract.Cache) *framework.Router {
	router := framework.New()
	cors := middleware.CORSConfigFrom(configuration, "http.cors")
	csrf := middleware.CSRFConfigFrom(configuration, "http.csrf")
	secureHeaders := middleware.SecureHeadersConfigFrom(configuration, "http.secure_headers")
	rateLimit := middleware.RateLimitConfigFrom(configuration, "http.rate_limit")
	correlation := middleware.CorrelationConfigFrom(configuration, "observability.correlation")

	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recover())
	router.Use(middleware.Correlation(correlation))
	router.Use(middleware.SecureHeaders(secureHeaders))
	router.Use(middleware.CORS(cors))
	router.Use(middleware.CSRF(csrf))
	router.Use(middleware.RateLimit(cache, rateLimit))

	handler.
		NewHelloWorld().
		Routes(router)

	return router
}
