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
			SendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}

		records, err := sm.GetAllRecords()
		if err != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to get records", err.Error())
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
			SendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}

		// Извлекаем ID из пути /records/{id}
		idStr := r.URL.Path[len("/records/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			SendErrorResponse(w, http.StatusBadRequest, "Invalid ID", err.Error())
			return
		}

		record, err := sm.GetRecordByID(id)
		if err != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to get record", err.Error())
			return
		}
		if record == nil {
			SendErrorResponse(w, http.StatusNotFound, "Record not found", "")
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
			SendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}

		var record schemas.Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			SendErrorResponse(w, http.StatusBadRequest, "Invalid JSON body", err.Error())
			return
		}

		if err := sm.CreateRecord(record); err != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to create record", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)
	}
}
