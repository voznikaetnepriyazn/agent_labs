package helper

import (
	"unicode"

	"github.com/sashabaranov/go-openai"
)

// CountTokens counts the number of tokens in a string
func CountTokens(text string) int {
	ru, en := 0, 0
	for _, r := range text {
		if unicode.Is(unicode.Cyrillic, r) {
			ru++
		} else if unicode.Is(unicode.Latin, r) {
			en++
		}
	}
	return ru/3 + en/4
}

// TrimHistory trims the history but saves System Promt
func TrimHistory(messages []openai.ChatCompletionMessage, maxTokens int) []openai.ChatCompletionMessage {
	if len(messages) == 0 {
		return messages
	}

	var system *openai.ChatCompletionMessage
	start := 0
	if messages[0].Role == openai.ChatMessageRoleSystem {
		system = &messages[0]
		start = 1
	}

	var kept []openai.ChatCompletionMessage
	usedTokens := 0
	if system != nil {
		usedTokens += CountTokens(system.Content)
	}

	for i := start; i <= len(messages)-1; i++ {
		t := CountTokens(messages[i].Content)
		if usedTokens+t > maxTokens {
			break
		}
		usedTokens += t
		kept = append(kept, messages[i])
	}

	if system != nil {
		return append([]openai.ChatCompletionMessage{*system}, kept...)
	}
	return kept
}
