package tool

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/agent_labs/internal/agent/tool/status"
)

type Tool struct {
	Schema  openai.Tool
	Handler func(ctx context.Context, argsJSON string) (string, error)
}

type Registry struct {
	tools []openai.Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: []openai.Tool{
			"check_status": status.Tools(),
		},
	}
}
