package service

import (
	"context"
	"net"
	"net/http"
	"time"
	"vgo-balancer/pkg/config"
	bc "vgo-balancer/pkg/backend"

	"go.uber.org/zap"
)

const (
	DefaultHealthCheckInterval = 30 * time.Second
	DefaultHealthCheckTimeout  = 5 * time.Second
	DefaultHealthCheckRetries  = 3
)

// Supported health check types
const (
	HealthCheckTypeHTTP = "http"
	HealthCheckTypeTCP  = "tcp"
)

type HealthCheck struct {
	endpoint        string        // The endpoint to check the health of the service.
	interval        time.Duration // The interval to check the health of the service.
	timeout         time.Duration // The timeout for the health check.
	retries         int           // The number of retries, before marking the service as unhealthy.
	healthCheckType string        // The type of health check. default is http.
	client          *http.Client
	logger          *zap.Logger
	ctx             context.Context
}

func NewHealthCheck(hc *config.HealthCheck, logger *zap.Logger, ctx context.Context) *HealthCheck {
	hcObj := &HealthCheck{
		endpoint:        hc.Endpoint,
		interval:        hc.Interval,
		timeout:         hc.Timeout,
		retries:         hc.Retries,
		healthCheckType: hc.HealthCheckType,
		logger:          logger,
		ctx:             ctx,
	}

	if hcObj.endpoint == "" {
		logger.Warn("Health check endpoint is not provided. TCP based health check will be done.")
		hcObj.healthCheckType = HealthCheckTypeTCP
	}

	if hcObj.interval == 0 {
		hcObj.interval = DefaultHealthCheckInterval
	}

	if hcObj.timeout == 0 {
		hcObj.timeout = DefaultHealthCheckTimeout
	}

	if hcObj.retries == 0 {
		hcObj.retries = DefaultHealthCheckRetries
	}

	if hcObj.healthCheckType == HealthCheckTypeHTTP {
		hcObj.client = &http.Client{
			Timeout: hcObj.timeout,
		}
	}

	return hcObj
}

func (hc *HealthCheck) StartHealthCheck(backends []*bc.Backend) {
	for i, backend := range backends {
		go func(b *bc.Backend, delay time.Duration) {
			time.Sleep(delay)
			for {
				select {
				case <-hc.ctx.Done():
					hc.logger.Info("Health check stopped for the backend", zap.String("backend", b.URL.String()))
					return
				case <-time.After(hc.interval):
					// Perform health check
					if !hc.performHealthCheck(b) {
						return
					}
				}
			}
		}(backend, time.Duration(i)*time.Second) // Stagger by 1 second for each backend
	}
}

func (hc *HealthCheck) performHealthCheck(b *bc.Backend) bool {
	for i := 0; i < hc.retries; i++ {
		switch hc.healthCheckType {
		case HealthCheckTypeHTTP:
			if hc.performHTTPHealthCheck(b) {
				return true
			}
		case HealthCheckTypeTCP:
			if hc.performTCPHealthCheck(b) {
				return true
			}
		default:
			hc.logger.Warn("Unsupported health check type", zap.String("type", hc.healthCheckType))
			return false
		}
		time.Sleep(hc.interval)
	}
	hc.logger.Warn("Health check failed after retries", zap.String("backend", b.URL.String()))
	return false
}

// http-based health check
func (hc *HealthCheck) performHTTPHealthCheck(b *bc.Backend) bool {
	healthURL := *b.URL
	healthURL.Path = hc.endpoint
	req, err := http.NewRequest("GET", healthURL.String(), nil)
	if err != nil {
		hc.logger.Warn("Failed to create HTTP health check request", zap.String("backend", b.URL.String()), zap.Error(err))
		b.IsAlive.Store(false)
		return false
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		hc.logger.Warn("HTTP health check failed for", zap.String("backend", b.URL.String()), zap.Error(err))
		b.IsAlive.Store(false)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		b.IsAlive.Store(true)
		return true
	} else {
		hc.logger.Warn("HTTP health check returned non-2xx", zap.String("backend", b.URL.String()), zap.Int("status", resp.StatusCode))
		b.IsAlive.Store(false)
		return false
	}
}

// TCP-based health check.
func (hc *HealthCheck) performTCPHealthCheck(b *bc.Backend) bool {
	healthAddress := b.URL.Host
	host, port, err := net.SplitHostPort(healthAddress)
	if err != nil {
		// If port is missing, infer from scheme
		host = healthAddress
		if b.URL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	tcpAddress := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", tcpAddress, hc.timeout)
	if err != nil {
		hc.logger.Warn("TCP health check failed for", zap.String("backend", b.URL.String()), zap.Error(err))
		b.IsAlive.Store(false)
		return false
	}

	conn.Close()
	b.IsAlive.Store(true)
	return true
}
