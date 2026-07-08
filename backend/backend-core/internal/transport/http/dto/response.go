package dto

import (
	"encoding/json"
	"net/http"
)

// --- API Error format ---

type ApiError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

func (e *ApiError) Write(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(e)
}

// --- API Success format ---

type ApiSuccess struct {
	Data interface{} `json:"data"`
}

func (s *ApiSuccess) Write(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ApiSuccess{Data: marshalPayload(payload)})
}

// --- Helpers ---

func marshalPayload(v interface{}) []byte {
	if v == nil {
		return []byte(`null`)
	}
	data, err := json.Marshal(v)
	if err != nil {
		panic(err) // для MVP это ок; в продакшене добавляем graceful handling
	}
	return data
}
