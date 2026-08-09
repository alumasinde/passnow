package httpx

import (
	"encoding/json"
	"net/http"
)

// maxBodyBytes caps request bodies platform-wide. Change here, not per handler.
const maxBodyBytes = 1 << 20 // 1 MiB

// DecodeJSON reads and decodes a JSON request body with a size limit, and
// writes a standard error response on failure. Returns false if the
// caller should stop (response already written).
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields() // reject unexpected fields outright — cheap defense against mass-assignment typos
	if err := dec.Decode(dst); err != nil {
		WriteError(w, ErrBadRequestBody)
		return false
	}
	return true
}
