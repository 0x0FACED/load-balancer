package httpcommon

import (
	"net/http"
	"strings"
)

func ClientIDFromRequest(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}
