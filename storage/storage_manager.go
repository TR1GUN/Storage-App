package storage

import (
	"storage_app/schemas"
	"storage_app/settings"
	"sync"
)

// SettingsApp — алиас для настроек.
type SettingsApp = settings.SettingsApp

// StorageManager — управляет hot и cold хранилищами.
type StorageManager struct {
	mu          sync.Mutex
	settings    SettingsApp
	hotStorage  *HotStorage
	coldStorage *ColdStorage
}

// NewStorageManager создаёт новый StorageManager.
func NewStorageManager() *StorageManager {
	s := settings.LoadSettings()
	return &StorageManager{
		settings:    s,
		hotStorage:  NewHotStorage(s),
		coldStorage: NewColdStorage(s),
	}
}

// Startup — инициализация при старте.
// Проверяет файлы и мигрирует данные из Hot в Cold.
func (m *StorageManager) Startup() error {
	if err := m.hotStorage.CheckUpFiles(); err != nil {
		return err
	}
	if err := m.coldStorage.CheckUpFiles(); err != nil {
		return err
	}
	return m.migrateFromHotToCold()
}

// migrateFromHotToCold — миграция из hot в cold хранилище.
func (m *StorageManager) migrateFromHotToCold() error {
	records, err := m.hotStorage.GetAllRecords()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	// Сохраняем все записи последовательно (без гонки за индексный файл)
	for _, record := range records {
		if err := m.coldStorage.SaveRecord(record); err != nil {
			return err
		}
	}

	// Очищаем hot storage после успешной миграции
	return m.hotStorage.ClearUp()
}

// Shutdown — завершение работы.
func (m *StorageManager) Shutdown() error {
	return nil
}

// GetAllRecords — возвращает все записи (hot + cold).
func (m *StorageManager) GetAllRecords() ([]schemas.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var records []schemas.Record

	hotRecords, err := m.hotStorage.GetAllRecords()
	if err != nil {
		return nil, err
	}
	records = append(records, hotRecords...)

	coldRecords, err := m.coldStorage.ReadAll()
	if err != nil {
		return nil, err
	}
	records = append(records, coldRecords...)

	return records, nil
}

// GetRecordByID — ищет запись по ID (сначала hot, потом cold).
func (m *StorageManager) GetRecordByID(id int) (*schemas.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Сначала ищем в hot storage
	record, err := m.hotStorage.GetRecord(id)
	if err != nil {
		return nil, err
	}
	if record != nil {
		return record, nil
	}

	// Потом в cold storage
	record, err = m.coldStorage.ReadRecordByID(id)
	if err != nil {
		return nil, err
	}
	if record != nil {
		return record, nil
	}

	return nil, nil
}

// CreateRecord — создаёт запись в hot storage.
func (m *StorageManager) CreateRecord(record schemas.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.hotStorage.SaveRecord(record)
}
