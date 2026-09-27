package status

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

var tools = []openai.Tool{
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "check_status",
			Description: "Check the status of a server by hostname. Use this when user asks about server status.",
			Parameters: json.RawMessage(`{
                    "type": "object",
                    "properties": {
                        "hostname": {
                            "type": "string",
                            "description": "Server hostname"
                        }
                    },
                    "required": ["hostname"]
                }`),
		},
	},
}
