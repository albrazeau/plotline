package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"
)

var (
	version = "dev"
	commit  = "unknown"
	built   = "unknown" // ISO-8601, UTC
)

func main() {

	log.Printf("version=%s commit=%s built=%s\n", version, commit, built)

	u, err := url.Parse("http://ollama:11434")
	if err != nil {
		log.Fatal(err)
	}

	client := api.NewClient(u, http.DefaultClient)

	accessible := false
	for range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := client.Heartbeat(ctx)
		if err == nil {
			accessible = true
			cancel()
			break
		}
		log.Println("unable to ping ollama client, spinning...")
		cancel()
		time.Sleep(time.Second)
	}

	if !accessible {
		log.Fatal("unable to access ollama server, exiting")
	}

	log.Println("ollama server is accessible")

	for {
		log.Println("on")
		time.Sleep((time.Second * 5))
	}
}
