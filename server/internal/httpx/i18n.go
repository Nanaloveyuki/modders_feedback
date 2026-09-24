package httpx

import (
	"net/http"
	"strings"
)

func Locale(r *http.Request) string {
	header := strings.ToLower(r.Header.Get("Accept-Language"))
	if strings.HasPrefix(header, "en") {
		return "en"
	}
	return "zh"
}

func Text(r *http.Request, zh, en string) string {
	if Locale(r) == "en" {
		return en
	}
	return zh
}
