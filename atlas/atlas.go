package atlas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/studiolambda/cosmos/nova"
	"github.com/studiolambda/cosmos/nova/middleware"
)

type Atlas struct {
	app    App
	ops    Options
	mux    sync.Mutex
	wg     sync.WaitGroup
	router *nova.Router
	server *http.Server
}

var (
	ErrAlreadyStarted = errors.New("app already started")
	ErrNotStarted     = errors.New("app not started")
)

func New(app App) *Atlas {
	return &Atlas{
		app: app,
		ops: DefaultOptions,
	}
}

func (a *Atlas) isStarted() bool {
	return a.server != nil
}

func (a *Atlas) IsStarted() bool {
	a.mux.Lock()
	defer a.mux.Unlock()

	return a.isStarted()
}

func (a *Atlas) Start(ops Options) (e error) {
	a.mux.Lock()

	if a.isStarted() {
		a.mux.Unlock()

		return ErrAlreadyStarted
	}

	a.ops = ops

	// Ensure that we always have a logger
	// to avoid nil pointers. The default logger
	// does discard all the content it logs to.
	if a.ops.Logger == nil {
		a.ops.Logger = DefaultOptions.Logger
	}

	a.ops.Logger.Info(
		"application starting",
		"addr", a.ops.Addr,
	)

	if h, ok := a.app.(BeforeStart); ok {
		h.BeforeStart()
	}

	// Initialize the nova router and register all the
	// application routes using the specific app
	// [App.Register] method.
	a.router = nova.New()

	if a.ops.RouterMiddlewares {
		// Handles errors returned by handlers, including the ones
		// recovered by [middleware.Recover].
		a.router.Use(middleware.ErrorHandler(middleware.ErrorHandlerOptions{
			Logger: a.ops.Logger,
			IsDev:  a.ops.IsDev,
		}))

		// Recovers panics from HTTP handlers.
		a.router.Use(middleware.Recover())
	}

	a.app.Register(a.router)

	// Create the HTTP server using the sensible settings
	// provided in the [Atlas.Start] method.
	a.server = &http.Server{
		Addr:              a.ops.Addr,
		Handler:           a.router,
		ReadTimeout:       a.ops.ReadTimeout,
		ReadHeaderTimeout: a.ops.ReadHeaderTimeout,
		WriteTimeout:      a.ops.WriteTimeout,
		IdleTimeout:       a.ops.IdleTimeout,
		MaxHeaderBytes:    a.ops.MaxHeaderBytes,
	}

	// Graceful shutdown context
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	exit := make(chan error)
	ready := make(chan struct{})

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		ln, err := net.Listen("tcp", a.ops.Addr)

		if err != nil {
			slog.Info("A")
			ready <- struct{}{}
			slog.Info("B")
			exit <- err
			slog.Info("C")
			return
		}

		ready <- struct{}{}

		if err := a.server.Serve(ln); err != http.ErrServerClosed {
			exit <- err
			return
		}

		exit <- nil
	}()

	_ = <-ready

	a.ops.Logger.Info("application started")

	if h, ok := a.app.(AfterStart); ok {
		h.AfterStart()
	}

	a.mux.Unlock()

	select {
	case <-stop:
		fmt.Fprint(os.Stdout, "\r") // Clear the `^C` on stdout
		return a.Stop()
	case err := <-exit:
		return errors.Join(err, a.Stop())
	}
}

func (a *Atlas) reset() {
	a.router = nil
	a.server = nil
	a.ops = DefaultOptions
}

func (a *Atlas) Stop() error {
	return a.StopContext(context.Background())
}

func (a *Atlas) StopContext(ctx context.Context) error {
	a.mux.Lock()

	if !a.isStarted() {
		a.mux.Unlock()

		return ErrNotStarted
	}

	a.ops.Logger.Info("application shutting down")

	defer a.reset()

	ctx, cancel := context.WithTimeout(ctx, a.ops.ShutdownTimeout)
	defer cancel()

	if h, ok := a.app.(BeforeShutdown); ok {
		h.BeforeShutdown()
	}

	if err := a.server.Shutdown(ctx); err != nil {
		a.mux.Unlock()
		return err
	}

	a.ops.Logger.Info("application shut down")

	if h, ok := a.app.(AfterShutdown); ok {
		h.AfterShutdown()
	}

	a.mux.Unlock()

	return nil
}
