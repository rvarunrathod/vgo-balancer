package backend

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
	"vgo-balancer/pkg/config"

	"go.uber.org/zap"
)

type BEPool struct {
	Backends []*Backend // list of backends
	Headers  *Header    // Headers is a list of headers to be added to the request.
}

type Backend struct {
	URL            *url.URL               // URL is the URL of the backend
	Weight         int                    // Weight is the weight of the backend
	IsAlive        atomic.Bool            // IsAlive is the status of the backend.
	ResponseTime   time.Duration          // ResponseTime is the response time of the backend.
	RequestTimeout time.Duration          // RequestTimeout is the timeout for the request. e.g. 60s
	Proxy          *httputil.ReverseProxy // proxy is the reverse proxy for the backend.
	Logger         *zap.Logger
}

type Header struct {
	RequestHeaders        map[string]string // RequestHeaders is a list of headers to be added to the request.
	ResponseHeaders       map[string]string // ResponseHeaders is a list of headers to be added to the response.
	RemoveRequestHeaders  []string          // RemoveRequestHeaders is a list of headers to be removed from the request.
	RemoveResponseHeaders []string          // RemoveResponseHeaders is a list of headers to be removed from the response.
}

func NewHeaders(h config.Header) *Header {
	return &Header{
		RequestHeaders:        h.RequestHeaders,
		ResponseHeaders:       h.ResponseHeaders,
		RemoveRequestHeaders:  h.RemoveRequestHeaders,
		RemoveResponseHeaders: h.RemoveResponseHeaders,
	}
}

func NewBEPool(backends []config.Backend, requestTimeout time.Duration, headers config.Header, logger *zap.Logger) *BEPool {
	fHeader := NewHeaders(headers)
	var b []*Backend
	for _, backend := range backends {
		cb := &Backend{
			Weight:  backend.Weight,
			IsAlive: atomic.Bool{},
			Logger:  logger,
		}

		backendURL, err := url.Parse(backend.URL)
		if err != nil {
			logger.Warn("failed to parse the backend URL", zap.String("URL", backend.URL), zap.Error(err))
			continue
		}

		// Handle if requestTimeout is empty, set it to 60s
		if requestTimeout == 0 {
			requestTimeout = 60 * time.Second
		}

		cb.URL = backendURL
		cb.IsAlive.Store(true)
		cb.Proxy = httputil.NewSingleHostReverseProxy(backendURL)
		cb.Proxy.Transport = http.DefaultTransport
		cb.Proxy.Transport = &http.Transport{
			MaxIdleConns:    getOrDefault(backend.ConnectionPool.MaxIdle, 10),
			MaxConnsPerHost: getOrDefault(backend.ConnectionPool.MaxConnection, 10),
			IdleConnTimeout: time.Duration(getOrDefault(backend.ConnectionPool.IdleTimeout, 90)) * time.Second,
			DialContext: defaultTransportDialContext(&net.Dialer{
				Timeout: requestTimeout,
			}),
		}
		cb.Proxy.ModifyResponse = func(response *http.Response) error {
			RemoveResponseHeaders(fHeader, response)
			AddResponseHeaders(fHeader, response)
			return nil
		}

		// Modify requests
		originalDirector := cb.Proxy.Director
		cb.Proxy.Director = func(req *http.Request) {
			originalDirector(req)
			RemoveRequestHeaders(fHeader, req)
			AddRequestHeaders(fHeader, req)
		}

		b = append(b, cb)
	}

	return &BEPool{
		Backends: b,
		Headers:  fHeader,
	}
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func getOrDefault(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func RemoveRequestHeaders(h *Header, r *http.Request) {
	for _, header := range h.RemoveRequestHeaders {
		r.Header.Del(header)
	}
}

func AddRequestHeaders(h *Header, r *http.Request) {
	for key, value := range h.RequestHeaders {
		r.Header.Set(key, value)
	}
}

func RemoveResponseHeaders(h *Header, w *http.Response) {
	for _, header := range h.RemoveResponseHeaders {
		w.Header.Del(header)
	}
}

func AddResponseHeaders(h *Header, w *http.Response) {
	for key, value := range h.ResponseHeaders {
		w.Header.Set(key, value)
	}
}
