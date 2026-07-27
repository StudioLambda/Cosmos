package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/framework/middleware"
)

func NewHTTPServer(i do.Injector) (framework.ServerConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := framework.ServerConfig{
		Host:              k.String("http.server.host"),
		Port:              k.Int("http.server.port"),
		ReadHeaderTimeout: k.Duration("http.server.read_header_timeout"),
		ReadTimeout:       k.Duration("http.server.read_timeout"),
		WriteTimeout:      k.Duration("http.server.write_timeout"),
		IdleTimeout:       k.Duration("http.server.idle_timeout"),
		MaxHeaderBytes:    k.Int("http.server.max_header_bytes"),
	}

	return config, nil
}

func NewHTTPCORS(i do.Injector) (middleware.CORSConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := middleware.CORSConfig{
		AllowedOrigins:   k.Strings("http.cors.allowed_origins"),
		AllowedMethods:   k.Strings("http.cors.allowed_methods"),
		AllowedHeaders:   k.Strings("http.cors.allowed_headers"),
		ExposedHeaders:   k.Strings("http.cors.exposed_headers"),
		AllowCredentials: k.Bool("http.cors.allow_credentials"),
		MaxAge:           k.Int("http.cors.max_age"),
	}

	return config, nil
}

func NewHTTPCSRF(i do.Injector) (middleware.CSRFConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := middleware.CSRFConfig{
		TrustedOrigins: k.Strings("http.csrf.trusted_origins"),
	}

	return config, nil
}

func NewHTTPSecureHeaders(i do.Injector) (middleware.SecureHeadersConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := middleware.SecureHeadersConfig{
		ContentTypeOptions:      k.String("http.secure_headers.content_type_options"),
		FrameOptions:            k.String("http.secure_headers.frame_options"),
		ReferrerPolicy:          k.String("http.secure_headers.referrer_policy"),
		XSSProtection:           k.String("http.secure_headers.xss_protection"),
		StrictTransportSecurity: k.String("http.secure_headers.strict_transport_security"),
		ContentSecurityPolicy:   k.String("http.secure_headers.content_security_policy"),
		PermissionsPolicy:       k.String("http.secure_headers.permissions_policy"),
	}

	return config, nil
}

func NewHTTPRateLimit(i do.Injector) (middleware.RateLimitConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := middleware.RateLimitConfig{
		Name:   k.String("http.rate_limit.name"),
		Limit:  k.Int("http.rate_limit.limit"),
		Window: k.Duration("http.rate_limit.window"),
	}

	return config, nil
}
