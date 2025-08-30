package llm

import "context"

type LLM interface {
	Models(context.Context) ([]string, error)
}
