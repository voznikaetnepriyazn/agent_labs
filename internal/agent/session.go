package agent

import (
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

type Session struct {
	mu       sync.Mutex
	messages []openai.ChatCompletionMessage
}

func NewSession(systemPrompt string) *Session {
	return &Session{
		messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: systemPrompt},
		},
	}
}

func (s *Session) AddMessage(msg openai.ChatCompletionMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *Session) Messages() []openai.ChatCompletionMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages
}
