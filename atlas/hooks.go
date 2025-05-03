package atlas

type BeforeStart interface {
	// BeforeStart is a hook that allows calling
	// code before the application start.
	//
	// An [App] may implement this method.
	BeforeStart()
}

type AfterStart interface {
	// AfterStart is a hook that allows calling
	// code after the application start.
	//
	// An [App] may implement this method.
	AfterStart()
}

type BeforeShutdown interface {
	// BeforeShutdown is a hook that allows calling
	// code before the application shutdown.
	//
	// An [App] may implement this method.
	BeforeShutdown()
}

type AfterShutdown interface {
	// AfterShutdown is a hook that allows calling
	// code after the application shutdown.
	//
	// An [App] may implement this method.
	AfterShutdown()
}
