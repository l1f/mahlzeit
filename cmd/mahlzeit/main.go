package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/l1f/mahlzeit/db/queries"
	"github.com/l1f/mahlzeit/internal/app"
	"github.com/l1f/mahlzeit/internal/http/routes"
	"go.uber.org/zap"
)

const ExitCodeOnError = 1

func main() {
	defer recoverPanic()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unexpected error during execution: %s", err)

		// Since os.Exit would skip the deferred statements, the context cancellation is invoked
		// manually at this point.
		cancel()

		os.Exit(ExitCodeOnError) // nolint:gocritic
	}
}

// We deliberately use the main function as the entrypoint to wrap basic command
// execution into run. That itself makes it testable and the provided [context.Context] can be
// used for downstream goroutines to cancel their operations.
func run(ctx context.Context, args []string) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("logger setup failed: %w", err)
	}
	defer logger.Sync()
	zap.RedirectStdLog(logger)

	var cfg app.Configuration
	_, err = toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		return fmt.Errorf("parsing config.toml failed: %w", err)
	}

	// The split up between config parsing and connecting is done to enable lazy connects.
	// With pgx/v5 this is no longer required, so the following code could be simplified again.
	dbConf, err := pgxpool.ParseConfig(cfg.Database.ConnectionString)
	if err != nil {
		return fmt.Errorf("parsing database configuration failed: %w", err)
	}

	dbConf.LazyConnect = true

	pool, err := pgxpool.ConnectConfig(ctx, dbConf)
	if err != nil {
		return err
	}

	app := &app.Application{
		Queries: queries.New(pool),
		Logger:  logger,
	}

	logger.Info("starting server", zap.String("endpoint", cfg.Web.Endpoint))
	h := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:              cfg.Web.Endpoint,
		Handler:           routes.All(app),
		ReadHeaderTimeout: time.Second, // protect against SLOWLORIS attack
	}
	return h.ListenAndServe()
}

func recoverPanic() {
	if rec := recover(); rec != nil {
		err := rec.(error)
		log.Printf("unhandled error: %v", err)
		fmt.Fprintf(os.Stderr, "Program quit unexpectedly; please check your logs\n")
		os.Exit(ExitCodeOnError)
	}
}
