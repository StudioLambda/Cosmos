package config

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewMemoryCache),
	do.Lazy(NewRedisCache),
	do.Lazy(NewAES),
	do.Lazy(NewChaCha20),
	do.Lazy(NewSQLDatabase),
	do.Lazy(NewEventMemory),
	do.Lazy(NewEventAMQP),
	do.Lazy(NewEventMQTT),
	do.Lazy(NewEventNATS),
	do.Lazy(NewEventRedis),
	do.Lazy(NewBcryptHash),
	do.Lazy(NewArgon2Hash),
	do.Lazy(NewHTTPCORS),
	do.Lazy(NewHTTPCSRF),
	do.Lazy(NewHTTPSecureHeaders),
	do.Lazy(NewHTTPRateLimit),
	do.Lazy(NewHTTPServer),
	do.Lazy(NewObservabilityCorrelation),
)
