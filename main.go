package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Запись сохранена", "id", 123, "path", "/data/hot.jsonl")
	logger.Error("Не удалось записать", "error")
}
