package handlers

import (
	"encoding/json"
	"net/http"
)

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
