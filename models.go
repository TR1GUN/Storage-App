package main

// import (
//
//	"encoding/json"
//	"fmt"
//
// )
type Record struct {
	ID   int    `json:"id"`
	Data string `json:"record"`
}

//Поскольку мы работаем сразу со структурой,
//логичнее бы инкапсулировать логику сериализации прямо в нее.

//
//func NewRecord[R Record]() *Record[R] {
//	return &Record[R]{Model: new(T)}
//}
//
//// Dump — сериализация модели в JSON-строку.
//// Аналог dump(record) -> str | None.
//func (s *Serialization[T]) Dump(record *T) (string, error) {
//	if record == nil {
//		return "", fmt.Errorf("record is nil")
//	}
//	data, err := json.Marshal(record)
//	if err != nil {
//		return "", err
//	}
//	return string(data), nil
//}
//
//// Load — десериализация JSON-строки в модель.
//// Аналог load(record: str) -> BaseModel | None.
//func (s *Serialization[T]) Load(raw string) (*T, error) {
//	if raw == "" {
//		return nil, fmt.Errorf("empty input")
//	}
//	var rec T
//	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
//		return nil, err // аналог except ValidationError
//	}
//	return &rec, nil
//}
