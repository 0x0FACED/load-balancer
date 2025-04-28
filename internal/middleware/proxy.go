package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/0x0FACED/load-balancer/internal/balancer"
)

type ProxyMiddleware struct {
	proxy    *httputil.ReverseProxy
	balancer balancer.Balancer
}

func NewProxyMiddleware(balancer balancer.Balancer) *ProxyMiddleware {
	pm := &ProxyMiddleware{
		balancer: balancer,
	}

	pm.proxy = &httputil.ReverseProxy{
		Director:     pm.director,
		ErrorHandler: pm.errorHandler,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  true,
		},
	}

	return pm
}

func (m *ProxyMiddleware) Proxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.proxy.ServeHTTP(w, r)
	})
}

func (m *ProxyMiddleware) director(req *http.Request) {
	backend, err := m.balancer.Next()
	if err != nil {
		// todo: handle error
		log.Println("Error getting backend:", err)
		return
	}

	backend = "http://" + backend

	backendURL, err := url.Parse(backend)
	if err != nil {
		// todo: handle error
		return
	}

	req.URL.Scheme = backendURL.Scheme
	req.URL.Host = backendURL.Host
	req.URL.Path = singleJoiningSlash(backendURL.Path, req.URL.Path)
	req.Header.Set("X-Forwarded-Host", req.Host)

	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", "")
	}

	log.Println("Proxying to ", backendURL.String())
}

func (m *ProxyMiddleware) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
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
