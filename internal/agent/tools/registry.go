package tool

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) RegisterAll(ts []Tool) {
	for _, t := range ts {
		r.Register(t)
	}
}

// Has returns true if the tool is registered
func (r *Registry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
}

// Schemas get the schemas for all registered tools
func (r *Registry) Schemas() []openai.Tool {
	result := make([]openai.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name(),
				Description: t.Description(), //from .md
				Parameters:  t.Parameters(),
			},
		})
	}
	return result
}

func (r *Registry) Execute(ctx context.Context, name, argsJSON string) string {
	// check if tool is registered
	currentTool, ok := r.tools[name]
	if !ok {
		return fmt.Sprintf("Error: unknown tool %q", name)
	}

	result, err := currentTool.Execute(ctx, argsJSON)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return result
}
