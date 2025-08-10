package llm

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

var _ LLM = &OllamaLLM{}

type OllamaLLM struct {
	client *api.Client
}

func NewOllamaLLM(ctx context.Context, uri string) (*OllamaLLM, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	client := api.NewClient(u, http.DefaultClient)

	err = client.Heartbeat(ctx)
	if err != nil {
		return nil, err
	}

	return &OllamaLLM{
		client: client,
	}, nil
}
