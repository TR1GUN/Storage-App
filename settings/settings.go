package settings

import (
	"fmt"
	"path/filepath"
)

// ─── Расширение SettingsApp ─────────────────────────────────

// SettingsApp Настройки нашего приложения
type SettingsApp struct {
	PathToHotStorage           string
	PathToColdStorageDir       string
	PathToColdStorageIndexFile string
}

// GetPathToColdStorage — путь к файлу конкретной записи в холодном хранилище.
func (s *SettingsApp) GetPathToColdStorage(idx int) string {
	return filepath.Join(s.PathToColdStorageDir, fmt.Sprintf("record_%d.json", idx))
}

func Get_settings() SettingsApp {
	return SettingsApp{}
}
