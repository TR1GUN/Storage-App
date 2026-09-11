package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// Serialization — обёртка для работы непосредственно с Record:
// Dump — сериализует Record в JSON-строку (чтобы писать в файл),
// Load — десериализует JSON-строку из файла обратно в Record.
// Аналог model_dump_json() / model_validate_json().
type Serialization[T Record] struct {
	Model *T
}

func NewSerialization[T Record]() *Serialization[T] {
	return &Serialization[T]{Model: new(T)}
}

// Dump — Record -> JSON-строка.
// Аналог record.model_dump_json() -> str.
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

// Load — JSON-строка -> Record.
// Аналог Record.model_validate_json(raw) -> Record.
func (s *Serialization[T]) Load(raw string) (*T, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty input")
	}
	var rec T
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return nil, err // аналог except ValidationError
	}
	return &rec, nil
}

// ----------------

// Serialization — обёртка над моделью для dump/load.
// Аналог Python-класса Serialization.
type Serialization[T any] struct {
	Model *T
	logger *L

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func NewSerialization[T any]() *Serialization[T] {
	return &Serialization[T]{Model: new(T)}
}

// Dump — сериализация модели в JSON-строку.
// Аналог dump(record) -> str | None.
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

// Load — десериализация JSON-строки в модель.
// Аналог load(record: str) -> BaseModel | None.
func (s *Serialization[T]) Load(raw string) (*T, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty input")
	}
	var rec T
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return nil, err // аналог except ValidationError
	}
	return &rec, nil
}
