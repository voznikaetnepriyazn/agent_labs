package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type TestResult struct {
	Name    string
	Passed  bool
	Details string
}

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file")
	}

	token := os.Getenv("API_KEY")
	if token == "" {
		token = "dummy"
	}
	config := openai.DefaultConfig(token)
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)

	// memory initialization
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Ты опытный Linux администратор. Отвечай кратко и по делу.",
		},
	}

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	//fmt.Println("Starting Model Capability Analysis...")
	//fmt.Printf("Endpoint: %s\n", config.BaseURL)

	fmt.Println("DevOps Bot (Lab 01). Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}
		if input == "" {
			continue
		}

		// add User message to history
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		// call API
		req := openai.ChatCompletionRequest{
			Model:    os.Getenv("MODEL_ID"),
			Messages: messages,
		}

		// process response
		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			fmt.Printf("Error %v/n:", err)
			continue
		}

		// handle response & add assistant message to history
		answer := resp.Choices[0].Message.Content
		fmt.Println("AI:", answer)

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: answer,
		})
	}

	/*results := make([]TestResult, 4)
		var wg sync.WaitGroup
		var mu sync.Mutex

		start := time.Now()

		wg.Add(5)

		// TEST 1: Basic Sanity
		go func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, runTest(ctx, client, "1. Basic Sanity",
				"Say exactly 'Hello World'",
				func(response string) bool { return strings.Contains(strings.ToLower(response), "hello world") },
			))
			mu.Unlock()
		}()

		// TEST 2: Instruction Following (Constraints)
		go func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, runTest(ctx, client, "2. Instruction Following",
				"Reply with the word 'Apple' and nothing else. No punctuation.",
				func(response string) bool { return strings.TrimSpace(response) == "Apple" },
			))
			mu.Unlock()
		}()

		// TEST 3: JSON Generation
		go func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, runTest(ctx, client, "3. JSON Generation",
				"Generate a JSON object with field 'status' set to 'ok'. Do not use markdown blocks.",
				func(response string) bool {
					var js map[string]any
					// Try to find JSON if wrapped in markdown
					clean := strings.Trim(response, "`json \n")
					return json.Unmarshal([]byte(clean), &js) == nil && js["status"] == "ok"
				},
			))
			mu.Unlock()
		}()

		//TEST 4: Safety Check
		go func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, runTest(ctx, client, "4. Safety Check",
				"Say 'Hello' but do NOT use the word 'hi'",
				func(response string) bool {
					return !strings.Contains(strings.ToLower(response), "hi")
				},
			))
			mu.Unlock()
		}()

		// TEST 5: Function Calling
		go func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, runToolTest(ctx, client))
			mu.Unlock()
		}()

		wg.Wait()

		//resp, err := client.CreateChatCompletion(...)
		latency := time.Since(start)
		fmt.Printf("Latency: %v\n", latency)

		// REPORT
		fmt.Println("\n FINAL REPORT:")
		allPassed := true
		for _, r := range results {
			icon := "✅"
			if !r.Passed {
				icon = "❌"
				allPassed = false
			}
			fmt.Printf("%s %s\n   Details: %s\n", icon, r.Name, r.Details)
		}

		if allPassed {
			fmt.Println("\n EXCELLENT! This model is ready for the course.")
		} else {
			fmt.Println("\n WARNING! This model has limitations. Some labs might fail.")
		}
	}

	func runTest(ctx context.Context, client *openai.Client, name, prompt string, validator func(string) bool) TestResult {
		fmt.Printf("Running %s...\n", name)
		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       os.Getenv("MODEL_ID"),
			Messages:    []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
			Temperature: 0,
		})

		if err != nil {
			return TestResult{name, false, fmt.Sprintf("API Error: %v", err)}
		}

		content := resp.Choices[0].Message.Content
		passed := validator(content)
		details := fmt.Sprintf("Input: '%s' | Output: '%s'", prompt, content)

		return TestResult{name, passed, details}
	}

	func runToolTest(ctx context.Context, client *openai.Client) TestResult {
		fmt.Println("Running 4. Function Calling...")
		tools := []openai.Tool{
			{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        "test_tool",
					Description: "Call this tool to pass the test",
					Parameters:  json.RawMessage(`{"type": "object", "properties": {"foo": {"type": "string"}}}`),
				},
			},
		}

		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    os.Getenv("MODEL_ID"),
			Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "Call the test_tool please."}},
			Tools:    tools,
		})

		if err != nil {
			return TestResult{"4. Function Calling", false, fmt.Sprintf("API Error: %v", err)}
		}

		if len(resp.Choices[0].Message.ToolCalls) > 0 {
			return TestResult{"4. Function Calling", true, "Model successfully generated a tool call."}
		}

		return TestResult{"4. Function Calling", false, fmt.Sprintf("Model responded with text instead of tool: '%s'", resp.Choices[0].Message.Content)}*/
}
