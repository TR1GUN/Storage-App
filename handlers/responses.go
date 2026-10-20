package handlers

import (
	"encoding/json"
	"net/http"
)

// SendErrorResponse — возвращает JSON ошибку.
func SendErrorResponse(w http.ResponseWriter, code int, message, detail string) {
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
