package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// PositionToFile — позиция в файле (смещение начала и конца).
type PositionToFile struct {
	OffsetStart int64
	OffsetEnd   int64
}

// FileIO — работа с файлами с кэшированием и потокобезопасностью.
type FileIO struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
	cache map[string][]byte
}

// NewFileIO создаёт новый экземпляр FileIO.
func NewFileIO() *FileIO {
	return &FileIO{
		locks: make(map[string]*sync.Mutex),
		cache: make(map[string][]byte),
	}
}

// getLock возвращает (или создаёт) мьютекс для конкретного файла.
func (f *FileIO) getLock(key string) *sync.Mutex {
	f.mu.Lock()
	defer f.mu.Unlock()

	if lock, ok := f.locks[key]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	f.locks[key] = lock
	return lock
}

// normalizeKey приводит путь к единому виду.
func normalizeKey(file string) string {
	if abs, err := filepath.Abs(file); err == nil {
		return abs
	}
	return file
}

// ReadDataFromFile — прочитать файл целиком с кэшированием.
func (f *FileIO) ReadDataFromFile(file string) (string, error) {
	key := normalizeKey(file)

	// Проверяем кэш
	f.mu.Lock()
	if cached, ok := f.cache[key]; ok {
		f.mu.Unlock()
		return string(cached), nil
	}
	f.mu.Unlock()

	lock := f.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	// Повторная проверка под локом
	f.mu.Lock()
	if cached, ok := f.cache[key]; ok {
		f.mu.Unlock()
		return string(cached), nil
	}
	f.mu.Unlock()

	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	f.mu.Lock()
	f.cache[key] = data
	f.mu.Unlock()

	return string(data), nil
}

// WriteDataInFile — записать данные в файл (перезапись).
func (f *FileIO) WriteDataInFile(file string, data string) error {
	key := normalizeKey(file)

	lock := f.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		return err
	}

	// Инвалидируем кэш
	f.mu.Lock()
	delete(f.cache, key)
	f.mu.Unlock()

	return nil
}

// FileExists — проверка существования файла.
func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// CreateFile — создать пустой файл.
func (f *FileIO) CreateFile(filePath string) error {
	return f.WriteDataInFile(filePath, "")
}

// ClearFile — очистить файл.
func (f *FileIO) ClearFile(filePath string) error {
	return f.WriteDataInFile(filePath, "")
}

// AppendToFile — добавить данные в конец файла.
// Возвращает позицию до и после записи.
func (f *FileIO) AppendToFile(filePath string, data string) (*PositionToFile, error) {
	key := normalizeKey(filePath)

	lock := f.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	// Размер файла до записи
	info, err := os.Stat(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	var offsetBefore int64
	if err == nil {
		offsetBefore = info.Size()
	}

	// Открываем файл в режиме append
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := []byte(data + "\n")
	if _, err := file.Write(buf); err != nil {
		return nil, err
	}

	// Размер файла после записи
	infoAfter, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	offsetAfter := infoAfter.Size()

	// Инвалидируем кэш
	f.mu.Lock()
	delete(f.cache, key)
	f.mu.Unlock()

	return &PositionToFile{
		OffsetStart: offsetBefore,
		OffsetEnd:   offsetAfter,
	}, nil
}

// ReadDataFromFileByLine — прочитать фрагмент файла по offset.
func (f *FileIO) ReadDataFromFileByLine(filePath string, line PositionToFile) (string, error) {
	key := normalizeKey(filePath)

	lock := f.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Seek на offset_start
	if _, err := file.Seek(line.OffsetStart, 0); err != nil {
		return "", err
	}

	// Читаем ровно (offset_end - offset_start) байт
	size := line.OffsetEnd - line.OffsetStart
	buf := make([]byte, size)
	if _, err := file.Read(buf); err != nil {
		return "", err
	}

	// Убираем trailing "\n"
	text := string(bytes.TrimRight(buf, "\n"))
	if text == "" {
		return "", nil
	}
	return text, nil
}
