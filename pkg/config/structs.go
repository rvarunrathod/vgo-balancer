package config

import "time"

type VgoBalancer struct {
	Host     string    `yaml:"host,omitempty"` // Host is the host address where the balancer is accessible.
	Port     int       `yaml:"port"`           // Port is the port number on which the balancer listens.
	Services []Service `yaml:"services"`       // Services is a list of services
}

type Backend struct {
	URL            string `yaml:"url"`            // URL is the URL of the backend
	Weight         int    `yaml:"weight"`         // Weight is the weight of the backend
	ConnectionPool *Pool  `yaml:"pool,omitempty"` // The Connection pool configuration.
	MaxConnection  int    `yaml:"max_connection"` // MaxConnection is the maximum number of connections allowed.
}

type Service struct {
	Name           string        `yaml:"name"`                   // Unique name of the service.
	Headers        Header        `yaml:"headers,omitempty"`      // Headers is a list of headers to be added to the request.
	Backends       []Backend     `yaml:"backends"`               // Backends is a list of backends.
	RequestTimeout time.Duration `yaml:"request_timeout"`        // RequestTimeout is the timeout for the request. e.g. 60s
	LBtype         string        `yaml:"lb_type"`                // Load balancing policy.
	HealthCheck    *HealthCheck  `yaml:"health_check,omitempty"` // HealthCheck is the health check configuration.
}

type Pool struct {
	MaxIdle       int `yaml:"max_idle"`     // The maximum number of idle connections in the pool.
	MaxConnection int `yaml:"max_conn"`     // The maximum number of open connections in the pool.
	IdleTimeout   int `yaml:"idle_timeout"` // The idle timeout for the connection in seconds.
}

type Header struct {
	RequestHeaders        map[string]string `yaml:"request_headers,omitempty"`         // RequestHeaders is a list of headers to be added to the request.
	ResponseHeaders       map[string]string `yaml:"response_headers,omitempty"`        // ResponseHeaders is a list of headers to be added to the response.
	RemoveRequestHeaders  []string          `yaml:"remove_request_headers,omitempty"`  // RemoveRequestHeaders is a list of headers to be removed from the request.
	RemoveResponseHeaders []string          `yaml:"remove_response_headers,omitempty"` // RemoveResponseHeaders is a list of headers to be removed from the response.
}

type HealthCheck struct {
	Endpoint        string        `yaml:"endpoint"`          // The endpoint to check the health of the service.
	Interval        time.Duration `yaml:"interval"`          // The interval to check the health of the service.
	Timeout         time.Duration `yaml:"timeout"`           // The timeout for the health check.
	Retries         int           `yaml:"retries"`           // The number of retries, before marking the service as unhealthy.
	HealthCheckType string        `yaml:"health_check_type"` // The type of health check. e.g. http, tcp
}
