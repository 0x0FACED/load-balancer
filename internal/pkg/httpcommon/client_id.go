package httpcommon

import (
	"net/http"
	"strings"
)

func ClientIDFromRequest(r *http.Request) string {
	if r.Header.Get("X-Client-ID") != "" {
		return r.Header.Get("X-Client-ID")
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
