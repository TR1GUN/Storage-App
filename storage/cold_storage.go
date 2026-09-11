package storage

import (
	"fmt"
	"storage_app/schemas"
	"storage_app/utils"
	"strings"
)

// ─── ColdStorage ───────────────────────────────────────────
// ColdStorage - Наше долговременное (Холодное) хранилище
type ColdStorage struct {
	settings  SettingsApp
	fileIO    *FileIO
	serialize *utils.Serialization[schemas.Record]
}

func NewColdStorage(settings SettingsApp) *ColdStorage {
	return &ColdStorage{
		settings:  settings,
		fileIO:    main.NewFileIO(),
		serialize: main.NewSerialization[main.Record](),
	}
}

// CheckUpFiles — проверка необходимых файлов.
// Поскольку файлы существуют отдельно для каждой записи —
// проверяем только наличие файла индексации.
func (c *ColdStorage) CheckUpFiles() error {
	exists, err := main.FileExists(c.settings.PathToColdStorageIndexFile)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.fileIO.CreateFile(c.settings.PathToColdStorageIndexFile); err != nil {
			return err
		}
	}
	return nil
}

// SaveRecord — сохраняем запись в долговременное хранилище.
func (c *ColdStorage) SaveRecord(record main.Record) error {
	filepathStr := c.settings.GetPathToColdStorage(record.ID)

	// Сериализуем и записываем в отдельный файл
	data, err := c.serialize.Dump(&record)
	if err != nil {
		return fmt.Errorf("serialize record %d: %w", record.ID, err)
	}

	if err := c.fileIO.WriteDataInFile(filepathStr, data); err != nil {
		return fmt.Errorf("write record file %d: %w", record.ID, err)
	}

	// Добавляем путь в индекс-файл
	if _, err := c.fileIO.AppendToFile(c.settings.PathToColdStorageIndexFile, filepathStr); err != nil {
		return fmt.Errorf("append to index: %w", err)
	}

	return nil
}

// ReadRecordByID — читаем запись по ID.
func (c *ColdStorage) ReadRecordByID(idx int) (*main.Record, error) {
	filepathStr := c.settings.GetPathToColdStorage(idx)

	exists, err := main.FileExists(filepathStr)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil // аналог return None
	}

	return c.getRecordByPath(filepathStr)
}

// getRecordByPath — получение данных по пути.
func (c *ColdStorage) getRecordByPath(filepathStr string) (*main.Record, error) {
	fileData, err := c.fileIO.ReadDataFromFile(filepathStr)
	if err != nil {
		return nil, err
	}

	if fileData == "" {
		return nil, nil
	}

	rec, err := c.serialize.Load(fileData)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// ReadAll — читаем все записи долговременного хранилища.
func (c *ColdStorage) ReadAll() ([]main.Record, error) {
	fileData, err := c.fileIO.ReadDataFromFile(c.settings.PathToColdStorageIndexFile)
	if err != nil {
		return nil, err
	}

	filepaths := strings.Split(fileData, "\n")

	var records []main.Record
	for _, fp := range filepaths {
		if fp == "" {
			continue
		}
		rec, err := c.getRecordByPath(fp)
		if err != nil {
			continue // пропускаем битые/недоступные файлы
		}
		if rec != nil {
			records = append(records, *rec)
		}
	}
	return records, nil
}
