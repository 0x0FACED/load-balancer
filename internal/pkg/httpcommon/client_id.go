package httpcommon

import (
	"net/http"
	"strings"
)

func ClientIDFromRequest(r *http.Request) string {
	// TODO: Think about X-Client-Id and RemoteAddr.
	//
	// There is might be collision with header.
	// Maybe its important to correlate client id and remoteaddr.
	// Problem:
	// 	User1 uses X-Client-ID: "uuidtest1" and has remoteaddr 1.1.1.1
	// 	User2 uses X-Client-ID: "uuidtest1" and has remoteaddr 2.2.2.2
	// Different users, but rate limiter will think its 1 user.
	if r.Header.Get("X-Client-ID") != "" {
		return r.Header.Get("X-Client-ID")
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}
