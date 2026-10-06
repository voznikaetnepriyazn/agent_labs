package tool

import (
	"context"
	"encoding/json"
	"errors"
)

type Tool interface {
	// Name returns name of tool for LLM and registry
	Name() string

	// Description returns description for LLM from .md file
	Description() string

	// Parameters returns JSON Schema
	Parameters() json.RawMessage

	// Execute executes tool с заданными аргументами
	Execute(ctx context.Context, argsJSON string) (string, error)
}

var (
	// ErrInvalidArgs — arguments do not match JSON Schema
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
