package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// PositionToFile — Структура которая хранит позицию в файле.
type PositionToFile struct {
	OffsetStart int64
	OffsetEnd   int64
}

// FileIO — класс для работы с файлами. Кэш + per-file мьютексы.
type FileIO struct {
	mu    sync.Mutex // Делаем мьютекс для блокировки кэшей\файлов
	locks map[string]*sync.Mutex // Делаем лок для блокировки кэшей\файлов
	cache map[string][]byte // Наш кэш для быстрого доступа к закешированным данным
	encode_file string
}

// Это наш констуктор
func NewFileIO() *FileIO {
	return &FileIO{
		locks: make(map[string]*sync.Mutex),
		cache: make(map[string][]byte),
		encode_file: "utf-8",
	}
}

// getLock возвращает (или создаёт) мьютекс для конкретного файла.
func (f *FileIO) getLock(key string) *sync.Mutex {
    // Поскольку у нас мапа не потокобезопасна, необходимо сделать лок на нее.
    // Обьявляем лок
	f.mu.Lock()
    // Говорим об освобождении лока, после завершения функции.
	defer f.mu.Unlock()

    // 	получаем лок для данного файла
	if lock, ok := f.locks[key]; ok {
		return lock
	}
    //Если лока нет - ставим его
	lock := &sync.Mutex{}
	f.locks[key] = lock
	return lock
}

// normalizeKey приводит путь к единому виду, чтобы
// "dir/file" и "dir/../dir/file" указывали на один мьютекс.
func normalizeKey(file string) string {
	if abs, err := filepath.Abs(file); err == nil {
		return abs
	}
	return file
}

// get_encoding Определяем кодировку в которой нужно работать.
func (f *FileIO) get_encoding(encoding string) string {
    enc := f.encode_file
    if len(encoding) > 0 {
        return encoding
    }
    return enc
}

// ReadDataFromFile — прочитать файл целиком.
// Так же используется кэширование записей с одинаковым path.
func (f *FileIO) ReadDataFromFile(file string, encoding string) (string, error) {
	enc := f.get_encoding(encoding)
	// в Go строки всегда []byte; кодировка учитывается только при decode/encode
	_ = enc
	key := normalizeKey(file)

	// Проверяем кэш без блокировки файла
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
// Аналог write_data_in_file.
func (f *FileIO) WriteDataInFile(file string, data string, encoding string) error {
	enc := f.get_encoding(encoding)
	_ = enc

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
// Аналог file_exists (static method).
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
// Аналог create_file.
func (f *FileIO) CreateFile(filePath string, encoding string) error {
	return f.WriteDataInFile(filePath, "", encoding)
}

// ClearFile — очистить файл (записать пустую строку).
// Аналог clear_file.
func (f *FileIO) ClearFile(filePath string) error {
	return f.WriteDataInFile(filePath, "", f.encode_file)
}

// AppendToFile — добавить данные в конец файла.
// Возвращает PositionToFile с offset до и после записи.
// Аналог append_to_file.
func (f *FileIO) AppendToFile(filePath string, data string, encoding string) (*PositionToFile, error) {
	enc := f.get_encoding(encoding)
	_ = enc

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

	// Открываем файл в режиме append, пишем data + "\n"
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
// Аналог read_data_from_file_by_line.
func (f *FileIO) ReadDataFromFileByLine(filePath string, line PositionToFile, encoding string) (string, error) {
	enc := f.get_encoding(encoding)
	_ = enc

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

	// Убираем trailing "\n" — аналог .rstrip("\n")
	text := string(bytes.TrimRight(buf, "\n"))
	if text == "" {
		return "", nil // возвращаем "пусто" как None в Python
	}
	return text, nil
}
