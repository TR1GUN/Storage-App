package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"storage_app/schemas"
	"storage_app/storage"
)

// GetAllRecordsHandler — GET /records — возвращает все записи.
func GetAllRecordsHandler(sm *storage.StorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}

		records, err := sm.GetAllRecords()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get records", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// GetRecordByIDHandler — GET /records/{id} — возвращает запись по ID.
func GetRecordByIDHandler(sm *storage.StorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}

		// Извлекаем ID из пути /records/{id}
		idStr := r.URL.Path[len("/records/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid ID", err.Error())
			return
		}

		record, err := sm.GetRecordByID(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get record", err.Error())
			return
		}
		if record == nil {
			writeError(w, http.StatusNotFound, "Record not found", "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(record); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// CreateRecordHandler — POST /records — создаёт новую запись.
func CreateRecordHandler(sm *storage.StorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}

		var record schemas.Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", err.Error())
			return
		}

		if err := sm.CreateRecord(record); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create record", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)
	}
}

// Routes — регистрирует все маршруты.
func Routes(mux *http.ServeMux, sm *storage.StorageManager) {
	mux.HandleFunc("GET /records", GetAllRecordsHandler(sm))
	mux.HandleFunc("POST /records", CreateRecordHandler(sm))
	mux.HandleFunc("GET /records/", GetRecordByIDHandler(sm))
}

// writeError — возвращает JSON ошибку.
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
