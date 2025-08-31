package llm

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"

	"github.com/ollama/ollama/api"
)

var _ LLM = &OllamaLLM{}

type OllamaLLM struct {
	client *api.Client
	logger *slog.Logger
}

func NewOllamaLLM(ctx context.Context, logger *slog.Logger, uri string) (*OllamaLLM, error) {
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
		logger: logger.With(slog.String("component", "ollama_llm")),
	}, nil
}

func (o *OllamaLLM) Models(ctx context.Context) ([]string, error) {
	resp, err := o.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list models: %w", err)
	}

	models := make([]string, 0)
	for _, model := range resp.Models {
		models = append(models, model.Name)
	}
	slices.Sort(models)
	return models, nil
}

// func (o *OllamaLLM) Chat(ctx context.Context, model string) {
// 	o.client.Chat(ctx, &api.ChatRequest{
// 		Model: model,
// 	})
// }
