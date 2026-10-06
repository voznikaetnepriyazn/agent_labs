package status

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	tool "github.com/voznikaetnepriyazn/agent_labs/internal/agent/tools"
)

//go:embed status.md
var description string

type StatusTool struct{}

func New() *StatusTool {
	return &StatusTool{}
}

func (t *StatusTool) Name() string {
	return "check_status"
}

func (t *StatusTool) Description() string {
	return description //содержимое status.md
}

func (t *StatusTool) Parameters() json.RawMessage {
	return json.RawMessage(`{
        "type": "object",
        "properties": {
            "hostname": {
                "type": "string",
                "description": "Server name for checking, web-01 or db-prod for example"
            }
        },
        "required": ["hostname"]
    }`)
}

func (t *StatusTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	args, err := tool.ParseArgs[struct {
		Hostname string `json:"hastname"`
	}](argsJSON)
	if err != nil {
		return "", err
	}

	//logic
	url := "https://" + args.Hostname + "/health"
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot reach %s: %w", args.Hostname, err)
	}
	defer resp.Body.Close()

	return fmt.Sprintf("Server %s: status=%d", args.Hostname, resp.StatusCode), nil
}
