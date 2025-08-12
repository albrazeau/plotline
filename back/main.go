package main

import (
	"context"
	"errors"
	"log/slog"
	"main/internal/api"
	"main/internal/llm"
	"main/internal/logx"
	"main/internal/session"
	"main/internal/session/store"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/errgroup"
)

var (
	version = "dev"
	commit  = "unknown"
	built   = "unknown"
)

func run() int {

	// init the logger and determine if it is in production
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	isProd := env == "production" || env == "prod"
	logger := logx.InitLogger(isProd)

	logger.Info(
		"application info",
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("built", built),
		slog.Bool("is_production", isProd),
	)

	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}

	// app-lifecycle ctx (cancels on SIGINT/SIGTERM)
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// short init timeout for dependencies
	initCtx, cancel := context.WithTimeout(appCtx, time.Second)
	defer cancel()

	ollamaLLM, err := llm.NewOllamaLLM(initCtx, "http://ollama:11434")
	if err != nil {
		logger.Error("ollama init failed", slog.String("error", err.Error()))
		return 1
	}
	logger.Info("ollama server is accessible")

	vkStore, err := store.NewValkeyStore(initCtx, valkey.ClientOption{InitAddress: []string{"valkey:6379"}})
	if err != nil {
		logger.Error("unable to create valkey store", slog.String("error", err.Error()))
		return 1
	}
	logger.Info("connected to valkey")
	defer vkStore.Close()

	sessions := session.New(vkStore)

	// router and app
	router := gin.New()
	app := api.New(logger, sessions, ollamaLLM)
	app.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
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
