package utils

import (
	"encoding/json"
	"fmt"
)

// Serialization — обёртка для сериализации/десериализации произвольного типа.
type Serialization[T any] struct{}

// NewSerialization создаёт новый экземпляр сериализатора.
func NewSerialization[T any]() *Serialization[T] {
	return &Serialization[T]{}
}

// Dump — сериализует значение в JSON-строку.
func (s *Serialization[T]) Dump(record *T) (string, error) {
	if record == nil {
		return "", fmt.Errorf("record is nil")
	}
	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Load — десериализует JSON-строку в значение типа T.
func (s *Serialization[T]) Load(raw string) (*T, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty input")
	}
	var rec T
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}
