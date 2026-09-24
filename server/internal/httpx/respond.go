package httpx

import (
	"encoding/json"
	"io"
	"net/http"
)

const maxBody = 1 << 20

func Decode(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		Error(w, http.StatusBadRequest, Text(r, "请求内容无效", "Invalid request body"))
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		Error(w, http.StatusBadRequest, Text(r, "请求内容无效", "Invalid request body"))
		return false
	}
	return true
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
