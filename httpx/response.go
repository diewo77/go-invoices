package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			// best-effort error response; avoid writing partial JSON
			http.Error(w, `{"error":"encode_error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		body = []byte("null")
	}
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		// nothing we can do at this point
		_ = err
	}
}

func JSONError(w http.ResponseWriter, status int, msg string, details any) {
	JSON(w, status, ErrorResponse{Error: msg, Details: details})
}
