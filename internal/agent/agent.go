package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/sashabaranov/go-openai"
	tool "github.com/voznikaetnepriyazn/agent_labs/internal/agent/tools"
)

const (
	maxSteps     = 10
	llmTimeout   = 180 * time.Second
	systemPrompt = `Ты DevOps инженер. Используй инструменты для проверки сервисов. У тебя есть инструмент check_status для проверки статуса сервера по hostname.

	ПРАВИЛА:
	1. Когда пользователь просит проверить статус сервера — ВСЕГДА вызывай check_status.
	2. Никогда не выдумывай результаты проверки.
	3. Никогда не предлагай пользователю команды — просто вызывай инструмент.
	4. Если hostname не указан — спроси его у пользователя.
	5. После получения результата — кратко перескажи его.`
	//userInput    = "Проверь статус сервера web-01"
)

type Agent struct {
	client   *openai.Client
	registry *tool.Registry
	*Session
}

func NewAgent(client *openai.Client, registry *tool.Registry) *Agent {
	return &Agent{
		client:   client,
		registry: registry,
		Session:  NewSession(systemPrompt),
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
			if errors.Is(err, context.DeadlineExceeded) {
				return "", fmt.Errorf("LLM timed out after %s on step %d", llmTimeout, i+1)
			}
			if errors.Is(err, context.Canceled) {
				return "", fmt.Errorf("LLM canceled: %w on step %d", err, i+1)
			}

			return "", fmt.Errorf("LLM error: %w on step %d", err, i+1)
		}

		if len(response.Choices) == 0 {
			return "", fmt.Errorf("no choices returned from LLM on step %d", i+1)
		}

		msg := response.Choices[0].Message
		a.AddMessage(msg)

		if len(msg.ToolCalls) == 0 {
			slog.Info("final answer", "steps", i+1, "total_time", elapsed)
			return msg.Content, nil
		}

		for _, tc := range msg.ToolCalls {
			slog.Debug("tool calling",
				"tool_name", tc.Function.Name,
				"tool_args", tc.Function.Arguments,
			)

			res := a.registry.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			slog.Info("tool result", "tool_name", tc.Function.Name, "result", res)

			a.AddToolResult(tc.ID, res)
		}
	}

	return "", fmt.Errorf("превышено максимальное количество шагов (%d)", maxSteps)
}
