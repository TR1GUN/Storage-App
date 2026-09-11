package storage

import (
	"sync"
)

// ─── StorageManager ─────────────────────────────────────────

type StorageManager struct {
	settings    SettingsApp
	hotStorage  *HotStorage
	coldStorage *ColdStorage
}

func NewStorageManager() *StorageManager {
	settings := SettingsApp{
		PathToHotStorage:           "data/hot_storage.jsonl",
		PathToColdStorageDir:       "data/cold/",
		PathToColdStorageIndexFile: "data/cold/index.txt",
	}
	return &StorageManager{
		settings:    settings,
		hotStorage:  NewHotStorage(settings),
		coldStorage: NewColdStorage(settings),
	}
}

// Startup — работа при старте.
// Инициализируем файлы, переносим записи из Hot в Cold.
func (m *StorageManager) Startup() error {
	if err := m.hotStorage.CheckUpFiles(); err != nil {
		return err
	}
	if err := m.coldStorage.CheckUpFiles(); err != nil {
		return err
	}
	return m.migrateFromTemp()
}

// migrateFromTemp — миграция из горячего хранилища в холодное + очистка temp.
func (m *StorageManager) migrateFromTemp() error {
	records, err := m.hotStorage.GetAllRecords()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	// Аналог asyncio.gather — запускаем сохранение всех записей конкурентно.
	// В Go для этого используем errgroup или WaitGroup + горутины.
	var wg sync.WaitGroup
	errChan := make(chan error, len(records))

	for _, record := range records {
		wg.Add(1)
		go func(r main.Record) {
			defer wg.Done()
			if err := m.coldStorage.SaveRecord(r); err != nil {
				errChan <- err
			}
		}(record)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err // возвращаем первую ошибку
		}
	}

	// Очистка горячего хранилища после успешной миграции
	return m.hotStorage.ClearUp()
}

// Shutdown — работа при завершении.
// Пока заглушка.
func (m *StorageManager) Shutdown() error {
	return nil
}

// GetAllRecords — все записи с сервера (Hot + Cold).
func (m *StorageManager) GetAllRecords() ([]main.Record, error) {
	var records []main.Record

	// Шаг 1: Hot Storage
	hotRecords, err := m.hotStorage.GetAllRecords()
	if err != nil {
		return nil, err
	}
	records = append(records, hotRecords...)

	// Шаг 2: Cold Storage
	coldRecords, err := m.coldStorage.ReadAll()
	if err != nil {
		return nil, err
	}
	records = append(records, coldRecords...)

	return records, nil
}

// GetRecordByID — запись по ID (сначала Hot, потом Cold).
func (m *StorageManager) GetRecordByID(idx int) (*main.Record, error) {
	// Шаг 1: Hot Storage
	record, err := m.hotStorage.GetRecord(idx)
	if err != nil {
		return nil, err
	}
	if record != nil {
		return record, nil
	}

	// Шаг 2: Cold Storage
	record, err = m.coldStorage.ReadRecordByID(idx)
	if err != nil {
		return nil, err
	}
	if record != nil {
		return record, nil
	}

	return nil, nil // не найдено
}

// CreateRecord — создаём запись в хранилище.
// Возвращает текст ошибки (пустая строка = успех) и системную ошибку.
func (m *StorageManager) CreateRecord(record main.Record) (string, error) {
	return m.hotStorage.SaveRecord(record)
}
