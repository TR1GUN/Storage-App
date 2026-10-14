package handlers

import (
	"net/http"
	"storage_app/storage"
)

// Routes — настраивает маршруты HTTP сервера.
func Routes(mux *http.ServeMux, sm *storage.StorageManager) {

	// POST /records — создать запись
	mux.HandleFunc("POST /records", CreateRecordHandler(sm))

	// GET /records/{id} — запись по ID
	mux.HandleFunc("GET /records/", GetRecordByIDHandler(sm))

	// GET /records — все записи (должен быть после /records/, чтобы ServeMux корректно роутит)
	mux.HandleFunc("GET /records", GetAllRecordsHandler(sm))
}
