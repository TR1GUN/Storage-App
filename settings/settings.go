package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

// SettingsApp — конфигурация приложения.
type SettingsApp struct {
	PathToHotStorage           string
	PathToColdStorageDir       string
	PathToColdStorageIndexFile string
}

// LoadSettings загружает настройки из переменных окружения с дефолтными значениями.
func LoadSettings() SettingsApp {
	return SettingsApp{
		PathToHotStorage:           getEnv("HOT_STORAGE_PATH", "data/hot_storage.jsonl"),
		PathToColdStorageDir:       getEnv("COLD_STORAGE_DIR", "data/cold/"),
		PathToColdStorageIndexFile: getEnv("COLD_STORAGE_INDEX", "data/cold/index.txt"),
	}
}

// GetPathToColdStorage возвращает путь к файлу конкретной записи в холодном хранилище.
func (s *SettingsApp) GetPathToColdStorage(idx int) string {
	return filepath.Join(s.PathToColdStorageDir, fmt.Sprintf("record_%d.json", idx))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
