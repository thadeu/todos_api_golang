package todo

import (
	fab "github.com/Goldziher/fabricator"
)

func NewTodo[T any](customData ...map[string]any) T {
	instance := fab.New(*new(T))

	if len(customData) > 0 {
		return instance.Build(customData...)
	}

	return instance.Build()
}
