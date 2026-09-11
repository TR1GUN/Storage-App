package storage

import (
	"fmt"
	"storage_app"
	"storage_app/utils"
	"strings"
)

// ─── SettingsApp (заглушка, подставь свою реализацию) ───────

type SettingsApp struct {
	PathToHotStorage  string
	PathToColdStorage string
}

// ─── HotStorage ─────────────────────────────────────────────

type HotStorage struct {
	indexes   map[int]PositionToFile
	settings  SettingsApp
	fileIO    *FileIO
	serialize *utils.Serialization[main.Record]
}

func NewHotStorage(settings SettingsApp) *HotStorage {
	return &HotStorage{
		indexes:   make(map[int]PositionToFile),
		settings:  settings,
		fileIO:    NewFileIO(),
		serialize: utils.NewSerialization[main.Record](),
	}
}

// ClearUp — очистка хранилища.
func (h *HotStorage) ClearUp() error {
	if err := h.fileIO.ClearFile(h.settings.PathToHotStorage); err != nil {
		return err
	}
	h.indexes = make(map[int]PositionToFile)
	return nil
}

// CheckUpFiles — проверка, что необходимые файлы существуют.
func (h *HotStorage) CheckUpFiles() error {
	exists, err := FileExists(h.settings.PathToHotStorage)
	if err != nil {
		return err
	}
	if !exists {
		if err := h.fileIO.CreateFile(h.settings.PathToHotStorage); err != nil {
			return err
		}
		h.indexes = make(map[int]PositionToFile)
	}
	return nil
}

// SaveRecord — сохранение записи в файл.
// Возвращает nil при успехе, строку с ошибкой — если запись уже существует
// или данные не сериализуются.
func (h *HotStorage) SaveRecord(record main.Record) (string, error) {
	if _, exists := h.indexes[record.ID]; exists {
		return fmt.Sprintf("Запись c id %d уже создана", record.ID), nil
	}

	// Сериализация
	dataToRecord, err := h.serialize.Dump(&record)
	if err != nil {
		return "Данные невозможно сериализовать в JSON", nil
	}

	// Запись в конец файла
	position, err := h.fileIO.AppendToFile(h.settings.PathToHotStorage, dataToRecord)
	if err != nil {
		return "", err
	}

	h.indexes[record.ID] = *position
	return "", nil // успех
}

// GetRecord — получение записи по ID.
func (h *HotStorage) GetRecord(idx int) (*main.Record, error) {
	pos, exists := h.indexes[idx]
	if !exists {
		return nil, nil // записи нет в кэше — считаем, что не существует
	}

	dataRecord, err := h.fileIO.ReadDataFromFileByLine(h.settings.PathToHotStorage, pos)
	if err != nil {
		return nil, err
	}

	if dataRecord == "" {
		return nil, nil // битые/пустые данные
	}

	rec, err := h.serialize.Load(dataRecord)
	if err != nil {
		// TODO: что-то предпринять, если данные битые
		return nil, err
	}
	return rec, nil
}

// GetAllRecords — читаем и отдаём все доступные записи.
func (h *HotStorage) GetAllRecords() ([]main.Record, error) {
	dataFromFile, err := h.fileIO.ReadDataFromFile(h.settings.PathToHotStorage)
	if err != nil {
		return nil, err
	}

	// Разбиваем на строки (аналог .splitlines())
	lines := strings.Split(dataFromFile, "\n")

	var allRecords []main.Record
	for _, line := range lines {
		if line == "" {
			continue
		}
		rec, err := h.serialize.Load(line)
		if err != nil {
			continue // пропускаем битые строки
		}
		allRecords = append(allRecords, *rec)
	}
	return allRecords, nil
}
