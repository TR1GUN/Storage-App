package handlers

import (
	"net/http"
	"storage_app/storage"
)

// Routes — настраивает маршруты HTTP сервера.
func Routes(mux *http.ServeMux, sm *storage.StorageManager) {
	// GET /records — все записи
	mux.HandleFunc("GET /records", GetAllRecordsHandler(sm))

	// POST /records — создать запись
	mux.HandleFunc("POST /records", CreateRecordHandler(sm))

	// GET /records/{id} — запись по ID
	mux.HandleFunc("GET /records/", GetRecordByIDHandler(sm))
}
