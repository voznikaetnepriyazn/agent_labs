package status

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	tool "github.com/voznikaetnepriyazn/agent_labs/internal/agent/tools"
)

//go:embed status.md
var description string

// ReadTool reads file
type ReadTool struct{}

func New() *ReadTool {
	return &ReadTool{}
}

func (t *ReadTool) Name() string {
	return "read"
}

func (t *ReadTool) Description() string {
	return description //содержимое status.md
}

func (t *ReadTool) Parameters() json.RawMessage {
	return json.RawMessage(`{
        "type": "object",
        "properties": {
            "path": {
                "type": "string",
                "description": "path to file"
            }
        },
        "required": ["path"]
    }`)
}

func (t *ReadTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	args, err := tool.ParseArgs[struct {
		Path string `json:"path"`
	}](argsJSON)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(args.Path)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", args.Path, err)
	}

	return string(content), nil
}
