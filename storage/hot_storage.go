package storage

import (
	"storage_app/schemas"
	"storage_app/utils"
	"strings"
)

// HotStorage — временное (горячее) хранилище.
// Данные хранятся в одном файле (JSONL), индексы — в памяти.
type HotStorage struct {
	indexes   map[int]PositionToFile
	settings  SettingsApp
	fileIO    *FileIO
	serialize *utils.Serialization[schemas.Record]
}

// NewHotStorage создаёт новый HotStorage.
func NewHotStorage(settings SettingsApp) *HotStorage {
	return &HotStorage{
		indexes:   make(map[int]PositionToFile),
		settings:  settings,
		fileIO:    NewFileIO(),
		serialize: utils.NewSerialization[schemas.Record](),
	}
}

// CheckUpFiles — проверяет, существует ли файл хранилища, создаёт при необходимости.
func (h *HotStorage) CheckUpFiles() error {
	exists, err := FileExists(h.settings.PathToHotStorage)
	if err != nil {
		return err
	}
	if !exists {
		if err := h.fileIO.CreateFile(h.settings.PathToHotStorage); err != nil {
			return err
		}
	}
	return nil
}

// SaveRecord — сохраняет запись в hot storage.
// Если запись с таким ID уже существует — перезаписывает её.
func (h *HotStorage) SaveRecord(record schemas.Record) error {
	data, err := h.serialize.Dump(&record)
	if err != nil {
		return err
	}

	// Если запись уже существует — удаляем старый индекс и позицию
	if _, exists := h.indexes[record.ID]; exists {
		// Удаляем старую позицию из файла (перезапись)
		if err := h.overwriteRecord(record.ID, data); err != nil {
			return err
		}
		return nil
	}

	// Новая запись — добавляем в конец файла
	position, err := h.fileIO.AppendToFile(h.settings.PathToHotStorage, data)
	if err != nil {
		return err
	}

	h.indexes[record.ID] = *position
	return nil
}

// overwriteRecord — перезаписывает запись в файле по ID.
func (h *HotStorage) overwriteRecord(id int, newData string) error {
	// Читаем весь файл
	allData, err := h.fileIO.ReadDataFromFile(h.settings.PathToHotStorage)
	if err != nil {
		return err
	}

	lines := strings.Split(allData, "\n")
	var updatedLines []string
	found := false

	for _, line := range lines {
		if line == "" {
			updatedLines = append(updatedLines, "")
			continue
		}
		rec, err := h.serialize.Load(line)
		if err != nil {
			updatedLines = append(updatedLines, line)
			continue
		}
		if rec.ID == id {
			updatedLines = append(updatedLines, newData)
			found = true
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !found {
		// Запись не найдена — добавляем в конец
		updatedLines = append(updatedLines, newData)
	}

	// Записываем обновлённый файл
	return h.fileIO.WriteDataInFile(h.settings.PathToHotStorage, strings.Join(updatedLines, "\n"))
}

// GetRecord — получает запись по ID из hot storage.
func (h *HotStorage) GetRecord(id int) (*schemas.Record, error) {
	pos, exists := h.indexes[id]
	if !exists {
		return nil, nil
	}

	dataRecord, err := h.fileIO.ReadDataFromFileByLine(h.settings.PathToHotStorage, pos)
	if err != nil {
		return nil, err
	}

	if dataRecord == "" {
		return nil, nil
	}

	rec, err := h.serialize.Load(dataRecord)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetAllRecords — возвращает все записи из hot storage.
func (h *HotStorage) GetAllRecords() ([]schemas.Record, error) {
	dataFromFile, err := h.fileIO.ReadDataFromFile(h.settings.PathToHotStorage)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(dataFromFile, "\n")

	var allRecords []schemas.Record
	for _, line := range lines {
		if line == "" {
			continue
		}
		rec, err := h.serialize.Load(line)
		if err != nil {
			continue
		}
		allRecords = append(allRecords, *rec)
	}
	return allRecords, nil
}

// ClearUp — очищает hot storage.
func (h *HotStorage) ClearUp() error {
	if err := h.fileIO.ClearFile(h.settings.PathToHotStorage); err != nil {
		return err
	}
	h.indexes = make(map[int]PositionToFile)
	return nil
}
