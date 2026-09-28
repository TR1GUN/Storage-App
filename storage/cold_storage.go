package storage

import (
	"fmt"
	"storage_app/schemas"
	"storage_app/utils"
	"strings"
)

// ColdStorage — долговременное (холодное) хранилище.
// Каждая запись хранится в отдельном JSON файле.
type ColdStorage struct {
	settings  SettingsApp
	fileIO    *FileIO
	serialize *utils.Serialization[schemas.Record]
}

// NewColdStorage создаёт новый ColdStorage.
func NewColdStorage(settings SettingsApp) *ColdStorage {
	return &ColdStorage{
		settings:  settings,
		fileIO:    NewFileIO(),
		serialize: utils.NewSerialization[schemas.Record](),
	}
}

// CheckUpFiles — проверяет наличие индексного файла.
func (c *ColdStorage) CheckUpFiles() error {
	exists, err := FileExists(c.settings.PathToColdStorageIndexFile)
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

// SaveRecord — сохраняет запись в cold storage.
func (c *ColdStorage) SaveRecord(record schemas.Record) error {
	filepath := c.settings.GetPathToColdStorage(record.ID)

	// Сериализуем и записываем в отдельный файл
	data, err := c.serialize.Dump(&record)
	if err != nil {
		return fmt.Errorf("serialize record %d: %w", record.ID, err)
	}

	if err := c.fileIO.WriteDataInFile(filepath, data); err != nil {
		return fmt.Errorf("write record file %d: %w", record.ID, err)
	}

	// Добавляем путь в индекс-файл
	if _, err := c.fileIO.AppendToFile(c.settings.PathToColdStorageIndexFile, filepath); err != nil {
		return fmt.Errorf("append to index: %w", err)
	}

	return nil
}

// ReadRecordByID — читает запись по ID.
func (c *ColdStorage) ReadRecordByID(id int) (*schemas.Record, error) {
	filepath := c.settings.GetPathToColdStorage(id)

	exists, err := FileExists(filepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return c.getRecordByPath(filepath)
}

// getRecordByPath — читает запись по пути к файлу.
func (c *ColdStorage) getRecordByPath(file string) (*schemas.Record, error) {
	fileData, err := c.fileIO.ReadDataFromFile(file)
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

// ReadAll — читает все записи cold storage.
func (c *ColdStorage) ReadAll() ([]schemas.Record, error) {
	fileData, err := c.fileIO.ReadDataFromFile(c.settings.PathToColdStorageIndexFile)
	if err != nil {
		return nil, err
	}

	filepaths := strings.Split(fileData, "\n")

	var records []schemas.Record
	for _, fp := range filepaths {
		if fp == "" {
			continue
		}
		rec, err := c.getRecordByPath(fp)
		if err != nil {
			continue
		}
		if rec != nil {
			records = append(records, *rec)
		}
	}
	return records, nil
}
