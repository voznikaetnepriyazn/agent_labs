package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/agent_labs/internal/agent/tool"
)

var CheckStatusTool = tool.Tool{
	Schema: openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "check_status",
			Description: "Check the status of a server by hostname.",
			Parameters: json.RawMessage(`{
                "type": "object",
                "properties": {
                    "hostname": {"type": "string"}
                },
                "required": ["hostname"]
            }`),
		},
	},
	Handler: checkStatusHandler,
}

func checkStatusHandler(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Hostname string `json:"hostname"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	return fmt.Sprintf("Server %s is running", args.Hostname), nil //logic
}

func Tools() []tool.Tool {
	return []tool.Tool{
		CheckStatusTool,
	}
}
