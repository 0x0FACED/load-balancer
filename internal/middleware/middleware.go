package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/zlog"
)

type Middleware struct {
	log      *zlog.ZerologLogger
	balancer balancer.Balancer
}

func (m *Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		queryParams := r.URL.Query()

		var formData map[string]any
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" || r.Header.Get("Content-Type") == "multipart/form-data" {
			if err := r.ParseForm(); err == nil {
				formData = make(map[string]any)
				for key, values := range r.Form {
					if len(values) == 1 {
						formData[key] = values[0]
					} else {
						formData[key] = values
					}
				}
			}
		}

		var jsonBody map[string]any
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if err := json.Unmarshal(bodyBytes, &jsonBody); err != nil {
				jsonBody = nil // just ingore it of err != nil
			}

		}

		next.ServeHTTP(w, r)

		logEvent := m.log.Info().
			Str("method", r.Method).
			Str("addr", r.RemoteAddr).
			Str("host", r.Host).
			Str("request_uri", r.RequestURI).
			TimeDiff("duration(ms)", time.Now(), start).
			Str("content_type", r.Header.Get("Content-Type"))

		if len(queryParams) > 0 {
			logEvent = logEvent.Interface("query_params", queryParams)
		}

		if len(formData) > 0 {
			logEvent = logEvent.Interface("form_data", formData)
		}

		if len(jsonBody) > 0 {
			logEvent = logEvent.Interface("json_body", jsonBody)
		}

		logEvent.Msg("Request")
	})
}
