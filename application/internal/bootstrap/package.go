package bootstrap

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewConfig),
	do.Lazy(NewLogger),
	do.Lazy(NewCache),
	do.Lazy(NewCacheDriver),
	do.Lazy(NewCrypto),
	do.Lazy(NewDatabase),
	do.Lazy(NewDatabaseDriver),
	do.Lazy(NewEvents),
	do.Lazy(NewEventDriver),
	do.Lazy(NewHasher),
	do.Lazy(NewHTTPRouter),
	do.Lazy(NewHTTPServer),
)
