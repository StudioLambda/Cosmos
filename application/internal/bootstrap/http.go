package bootstrap

import (
	"log/slog"
	"net/http"

	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/correlation"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func NewHTTPServer(i do.Injector) (*http.Server, error) {
	config := do.MustInvoke[framework.ServerConfig](i)
	router := do.MustInvoke[*framework.Router](i)

	return framework.NewServer(config, router), nil
}

func NewHTTPRouter(i do.Injector) (*framework.Router, error) {
	router := framework.New()
	logger := do.MustInvoke[*slog.Logger](i)
	cors := do.MustInvoke[middleware.CORSConfig](i)
	csrf := do.MustInvoke[middleware.CSRFConfig](i)
	secureHeaders := do.MustInvoke[middleware.SecureHeadersConfig](i)
	rateLimit := do.MustInvoke[middleware.RateLimitConfig](i)
	cache := do.MustInvoke[*contract.Cache](i)
	corr := do.MustInvoke[correlation.MiddlewareConfig](i)

	router.Use(middleware.Recover())
	router.Use(middleware.CORS(cors))
	router.Use(middleware.CSRF(csrf))
	router.Use(middleware.SecureHeaders(secureHeaders))
	router.Use(middleware.RateLimit(cache, rateLimit))
	router.Use(correlation.Middleware(corr))
	router.Use(middleware.Logger(logger))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) error {
		return response.String(w, http.StatusOK, "Hello, world!")
	})

	return router, nil
}
