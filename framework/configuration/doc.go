// Package configuration provides framework-owned configuration driver
// adapters.
//
// It bridges external configuration libraries to the contract-level
// [contract.ConfigurationDriver] abstraction so applications can depend on a
// stable, backend-agnostic configuration API.
//
// # Current adapters
//
// The package currently includes a Koanf-backed driver via [NewKoanf],
// [NewKoanfFrom], and [NewKoanfWith].
//
// Example
//
//	config := configuration.NewKoanf()
//	err := config.Instance().Load(rawbytes.Provider([]byte("http:\n  port: 8080\n")), yaml.Parser())
//	if err != nil {
//		return err
//	}
//
//	value := contract.NewConfiguration(config)
//	port := value.GetOr("http.port", 3000)
//	_ = port
package configuration
