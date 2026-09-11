package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

//
//// --- StorageManager (твой класс) ---
//
//type StorageManager struct {
//	mu sync.Mutex
//	// здесь твои поля Hot/Cold, пути, кэш и т.д.
//}
//
//func NewStorageManager() *StorageManager {
//	return &StorageManager{}
//}
//
//func (s *StorageManager) startup() error {
//	// Миграция Hot -> Cold при старте
//	log.Println("StorageManager: startup (Hot -> Cold migration)")
//	// TODO: реализация миграции
//	return nil
//}
//
//func (s *StorageManager) shutdown() error {
//	// Очистка, финализация
//	log.Println("StorageManager: shutdown")
//	// TODO: финализация
//	return nil
//}
//
//func (s *StorageManager) getRecords() ([]Record, error) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	// TODO: объединить Hot и Cold, вернуть список
//	// Для примера возвращаем пустой список
//	return []Record{}, nil
//}
//
//func (s *StorageManager) getRecordByID(id int) (*Record, error) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	// TODO: поиск по ID в Hot/Cold
//	return nil, nil // заглушка
//}
//
//func (s *StorageManager) createRecord(r Record) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	// TODO: запись в Hot, затем миграция в Cold и т.д.
//	return nil // заглушка
//}

// --- HTTP handlers ---

//func makeGetAllRecordsHandler(storage *StorageManager) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		records, err := storage.getRecords()
//		if err != nil {
//			writeError(w, http.StatusInternalServerError, "Failed to get records", err.Error())
//			return
//		}
//
//		w.Header().Set("Content-Type", "application/json")
//		if err := json.NewEncoder(w).Encode(records); err != nil {
//			log.Printf("Error encoding response: %v", err)
//		}
//	}
//}
//
//func makeGetRecordByIDHandler(storage *StorageManager) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		idStr := r.URL.Path[len("/records/"):] // упрощённо; лучше использовать mux
//		id, err := strconv.Atoi(idStr)
//		if err != nil {
//			writeError(w, http.StatusBadRequest, "Invalid ID", err.Error())
//			return
//		}
//
//		record, err := storage.getRecordByID(id)
//		if err != nil {
//			writeError(w, http.StatusInternalServerError, "Failed to get record", err.Error())
//			return
//		}
//		if record == nil {
//			writeError(w, http.StatusNotFound, "Record not found", "")
//			return
//		}
//
//		w.Header().Set("Content-Type", "application/json")
//		if err := json.NewEncoder(w).Encode(record); err != nil {
//			log.Printf("Error encoding response: %v", err)
//		}
//	}
//}
//
//func makeCreateRecordHandler(storage *StorageManager) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		var record Record
//		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
//			writeError(w, http.StatusUnprocessableEntity, "Invalid JSON body", err.Error())
//			return
//		}
//
//		err := storage.createRecord(record)
//		if err != nil {
//			writeError(w, http.StatusInternalServerError, "Failed to create record", err.Error())
//			return
//		}
//
//		w.WriteHeader(http.StatusCreated)
//	}
//}

func writeError(w http.ResponseWriter, code int, message, detail string) {
	type errorResp struct {
		Error  string `json:"error"`
		Detail string `json:"detail,omitempty"`
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResp{
		Error:  message,
		Detail: detail,
	})
}

// --- Маршрутизация (упрощённая без mux) ---

func setupRoutes(mux *http.ServeMux, storage *StorageManager) {
	mux.HandleFunc("/records", makeGetAllRecordsHandler(storage))
	mux.HandleFunc("/records/", func(w http.ResponseWriter, r *http.Request) {
		// обрабатываем /records/{id} только для GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		makeGetRecordByIDHandler(storage)(w, r)
	})
	mux.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
		// POST /records
		if r.Method == http.MethodPost {
			makeCreateRecordHandler(storage)(w, r)
			return
		}
		// остальные методы — ошибка
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

// Для корректной обработки /records/{id} лучше использовать полноценный роутер (mux).
// Ниже — более чистая версия с отдельным паттерном для ID.

func setupRoutesClean(mux *http.ServeMux, storage *StorageManager) {
	mux.HandleFunc("/records", makeGetAllRecordsHandler(storage))

	mux.HandleFunc("/records/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		prefix := "/records/"
		if len(path) <= len(prefix) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		idPart := path[len(prefix):]

		// Если это /records/{id}, то дальше проверяем метод
		switch r.Method {
		case http.MethodGet:
			// извлекаем ID из path
			id, err := strconv.Atoi(idPart)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Invalid ID", err.Error())
				return
			}
			// создаём временный запрос с правильным path для handler'а
			req := *r
			req.URL.Path = fmt.Sprintf("/records/%d", id)
			makeGetRecordByIDHandler(storage)(w, &req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			makeCreateRecordHandler(storage)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

// --- Lifespan / запуск и остановка ---

func main() {
	storage := NewStorageManager()

	if err := storage.startup(); err != nil {
		log.Fatalf("Startup failed: %v", err)
	}

	mux := http.NewServeMux()
	setupRoutesClean(mux, storage)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Server starting on :8000")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	// signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM) // раскомментировать при необходимости

	select {
	case <-shutdownChan:
		log.Println("Shutdown signal received")
	}

	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	if err := storage.shutdown(); err != nil {
		log.Printf("Storage shutdown error: %v", err)
	}
}
