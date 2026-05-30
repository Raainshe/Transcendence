package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteJSON is the exported form for use by other packages (server, middleware).
func WriteJSON(w http.ResponseWriter, code int, v any) {
	writeJSON(w, code, v)
}

// WriteError writes a uniform JSON error envelope.
func WriteError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// parseQueryInt returns the integer value of a query param.
// Missing/empty params return def; non-empty but malformed values return ok=false.
func parseQueryInt(r *http.Request, key string, def int) (int, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}
