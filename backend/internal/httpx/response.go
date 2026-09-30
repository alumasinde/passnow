// Package httpx is the ONE place that knows how JSON responses and errors
// are shaped and worded. Change the envelope format or any message text
// here — nowhere else should ever construct a response body by hand.
package httpx

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

// WriteJSON writes any success payload. All handlers should use this
// instead of calling json.NewEncoder directly, so the envelope stays
// consistent if it ever needs a wrapper (e.g. {"data": ...}).
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a known AppError. This is what every handler and
// middleware in the codebase should call — never construct an error
// response inline.
func WriteError(w http.ResponseWriter, err AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorBody{Code: err.Code, Message: err.Message},
	})
}
