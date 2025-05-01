package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/load-balancer/internal/pkg/httpcommon"
)

type ProxyMiddleware struct {
	balancer balancer.Balancer
}

func NewProxyMiddleware(balancer balancer.Balancer) *ProxyMiddleware {
	return &ProxyMiddleware{
		balancer: balancer,
	}
}

func (m *ProxyMiddleware) Proxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendAddr, err := m.balancer.Next()
		if err != nil {
			httpcommon.JSONError(w, http.StatusServiceUnavailable, err)
			return
		}

		backendURL, err := url.Parse(backendAddr)
		if err != nil {
			httpcommon.JSONError(w, http.StatusInternalServerError, err)
			return
		}

		wrapped := &responseObserver{
			ResponseWriter: w,
			onFinish: func() {
				if lcb, ok := m.balancer.(interface {
					Release(string)
				}); ok {
					lcb.Release(backendAddr)
				}
			},
		}
		defer wrapped.finishOnce()

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = backendURL.Scheme
				req.URL.Host = backendURL.Host
				req.URL.Path = singleJoiningSlash(backendURL.Path, req.URL.Path)
				req.Header.Set("X-Forwarded-Host", req.Host)

				if _, ok := req.Header["User-Agent"]; !ok {
					req.Header.Set("User-Agent", "")
				}

				log.Println("Proxying to", backendURL.String())
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				httpcommon.JSONError(w, http.StatusServiceUnavailable, err)
			},
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
			},
		}

		proxy.ServeHTTP(wrapped, r)
	})
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type responseObserver struct {
	http.ResponseWriter
	onFinish func()
	once     sync.Once
}

func (o *responseObserver) Write(b []byte) (int, error) {
	o.finishOnce()
	return o.ResponseWriter.Write(b)
}

func (o *responseObserver) WriteHeader(statusCode int) {
	o.finishOnce()
	o.ResponseWriter.WriteHeader(statusCode)
}

func (o *responseObserver) finishOnce() {
	o.once.Do(func() {
		if o.onFinish != nil {
			o.onFinish()
		}
	})
}
