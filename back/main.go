package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"main/config"
	"main/internal/api"
	"main/internal/llm"
	"main/internal/session"
	"main/internal/session/store"
	"main/logx"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/errgroup"
)

var (
	version = "dev"
	commit  = "unknown" //nolint:gochecknoglobals // this is built into the binary using ldflags
	built   = "unknown" //nolint:gochecknoglobals // this is built into the binary using ldflags
)

func run() int {
	configPath := flag.String("config", "", "Path to the YAML config file")
	flag.Parse()

	// load and validate config
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("invalid config: " + err.Error())
	}

	// init the logger
	logger := logx.InitLogger(cfg.Log)
	logger.Info(
		"application info",
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("built", built),
	)
	logger = logger.With(
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("built", built),
	)

	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// app-lifecycle ctx (cancels on SIGINT/SIGTERM)
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// short init timeout for dependencies
	initCtx, cancel := context.WithTimeout(appCtx, time.Second)
	defer cancel()

	// connect to the llm
	ollamaLLM, err := llm.NewOllamaLLM(initCtx, logger, cfg.Ollama.BaseURL)
	if err != nil {
		logger.Error("ollama init failed", slog.String("error", err.Error()))
		return 1
	}
	logger.Info("ollama server is accessible")

	// connect to the store
	vkStore, err := store.NewValkeyStore(initCtx, logger, valkey.ClientOption{InitAddress: []string{cfg.Valkey.Address}})
	if err != nil {
		logger.Error("unable to create valkey store", slog.String("error", err.Error()))
		return 1
	}
	logger.Info("connected to valkey")
	defer vkStore.Close()

	sessions := session.New(logger, vkStore)

	// router and app
	router := gin.New()
	app := api.New(logger, sessions, ollamaLLM)
	app.RegisterRoutes(router)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           router,
		ReadHeaderTimeout: cfg.App.ReadHeaderTimeout,
		ReadTimeout:       cfg.App.ReadTimeout,
		WriteTimeout:      cfg.App.WriteTimeout,
		IdleTimeout:       cfg.App.IdleTimeout,
	}

	// errgroup drives server run and graceful shutdown
	group, ctx := errgroup.WithContext(appCtx)

	// server goroutine: report only real errors (ignore ErrServerClosed)
	group.Go(func() error {
		logger.Info("server starting", slog.String("addr", srv.Addr))
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// shutdown goroutine: wait for signal/context cancel, then graceful shutdown
	group.Go(func() error {
		<-ctx.Done() // triggered by signal or parent cancel
		logger.Info("shutdown signal received")

		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shCtx); err != nil {
			return err
		}
		logger.Info("http server shut down cleanly")
		return nil
	})

	// Wait for either server error (non-zero exit) or graceful shutdown (zero).
	if err := group.Wait(); err != nil {
		logger.Error("unable to shutdown gracefully", slog.String("error", err.Error()))
		return 1
	}

	logger.Info("shutdown complete")
	return 0
}

// allow for cleanup in run, but return the correct exit code
func main() {
	os.Exit(run())
}
