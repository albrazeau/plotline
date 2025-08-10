package main

import (
	"context"
	"log"
	"main/internal/llm"
	"main/internal/session/store"
	"os/signal"
	"syscall"
	"time"

	"github.com/valkey-io/valkey-go"
)

var (
	version = "dev"
	commit  = "unknown"
	built   = "unknown" // ISO-8601, UTC
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("BACK: version=%s commit=%s built=%s\n", version, commit, built)
}
func main() {

	appLifecycleCtx, appLifecycleCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appLifecycleCancel()

	ctx, cancel := context.WithTimeout(appLifecycleCtx, time.Second)
	defer cancel()

	_, err := llm.NewOllamaLLM(ctx, "http://ollama:11434")
	log.Println("ollama server is accessible")

	_, err = store.NewValkeyStore(ctx, valkey.ClientOption{InitAddress: []string{"valkey:6379"}})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to valkey")

	// here, do graceful shutdown after the app is given a shutdown signal
	<-appLifecycleCtx.Done()

	log.Println("shutting down")
}
