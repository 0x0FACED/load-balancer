package httpcommon

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func EmptyResponse(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

func DefaultResponse(w http.ResponseWriter, statusCode int, data []byte) {
	w.WriteHeader(statusCode)
	_, _ = w.Write(data)
}
