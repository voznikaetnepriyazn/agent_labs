package tool

import (
	"context"
	"encoding/json"
	"errors"
	//"github.com/sashabaranov/go-openai"
)

type Tool interface {
	// Name возвращает имя инструмента (для LLM и реестра)
	Name() string

	// Description возвращает описание для LLM (из .md файла)
	Description() string

	// Parameters возвращает JSON Schema аргументов
	Parameters() json.RawMessage

	// Execute выполняет инструмент с заданными аргументами
	Execute(ctx context.Context, argsJSON string) (string, error)
}

var (
	// ErrInvalidArgs — аргументы не соответствуют JSON Schema
	ErrInvalidArgs = errors.New("invalid arguments")

	// ErrNotFound — запрошенный ресурс не найден
	ErrNotFound = errors.New("not found")

	// ErrPermission — нет прав на выполнение операции
	ErrPermission = errors.New("permission denied")
)

func ParseArgs[T any](argsJSON string) (T, error) {
	var args T
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return args, ErrInvalidArgs
	}
	return args, nil
}

func Errorf(format string, args ...any) error {
	return errors.New(format)
}
