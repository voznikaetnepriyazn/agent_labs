package agent

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/sashabaranov/go-openai"
	tool "github.com/voznikaetnepriyazn/agent_labs/internal/agent/tools"
)

const (
	maxSteps   = 10
	llmTimeout = 180 * time.Second
)

//go:embed promt.md
var systemPromt string

type Agent struct {
	client   *openai.Client
	registry *tool.Registry
	*Session
}

func NewAgent(client *openai.Client, registry *tool.Registry) *Agent {
	return &Agent{
		client:   client,
		registry: registry,
		Session:  NewSession(systemPromt),
	}
}

// RunAgent prosesses user input and returns agent's response
func (a *Agent) RunAgent(ctx context.Context, userInput string, model string) (string, error) {
	a.AddUserInput(userInput)

	for i := 0; i < maxSteps; i++ {
		slog.Debug("agent step", "step", i)

		llmCtx, cancel := context.WithTimeout(ctx, llmTimeout)
		defer cancel()

		start := time.Now()

		//  sending request to LLM
		response, err := a.client.CreateChatCompletion(llmCtx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: a.Messages(),
			Tools:    a.registry.Schemas(),
			/*ToolChoice: openai.ToolChoiceOneOfTool{
				OneOfTool: &openai.ToolChoiceOneOfToolOneOfTool{
					ToolId: a.registry.Schemas()[0].ID,
				},
			},*/
			Temperature: 0,
		})

		elapsed := time.Since(start)
		slog.Debug("LLM request took", "duration", elapsed)

		if err != nil {
			// context timeout
			if errors.Is(err, context.DeadlineExceeded) {
				return "", fmt.Errorf("LLM timed out after %s on step %d", llmTimeout, i+1)
			}
			//context canceled
			if errors.Is(err, context.Canceled) {
				return "", fmt.Errorf("LLM canceled: %w on step %d", err, i+1)
			}

			return "", fmt.Errorf("LLM error: %w on step %d", err, i+1)
		}

		// check if LLM returned any choices
		if len(response.Choices) == 0 {
			return "", fmt.Errorf("no choices returned from LLM on step %d", i+1)
		}

		// add LLM response to session
		msg := response.Choices[0].Message
		a.AddMessage(msg)

		// check if LLM wants to call a tool
		if len(msg.ToolCalls) == 0 {
			slog.Info("final answer", "steps", i+1, "total_time", elapsed)
			return msg.Content, nil
		}

		// loop over tool calls
		for _, tc := range msg.ToolCalls {
			slog.Debug("tool calling",
				"tool_name", tc.Function.Name,
				"tool_args", tc.Function.Arguments,
			)

			if err := validateToolCall(tc, a.registry); err != nil {
				slog.Warn("invalid tool_call", "error", err)
				a.AddToolResult(tc.ID, fmt.Sprintf("Error: %v", err))
				continue
			}

			// execute appropriate tool and add result to session
			res := a.registry.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			slog.Info("tool result", "tool_name", tc.Function.Name, "result", res)

			a.AddToolResult(tc.ID, res)
		}
	}

	return "", fmt.Errorf("превышено максимальное количество шагов (%d)", maxSteps)
}
