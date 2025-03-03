package service

import (
	"context"
	"net/http"
	"time"
	"vgo-balancer/pkg/algo"
	"vgo-balancer/pkg/backend"
	"vgo-balancer/pkg/config"

	"go.uber.org/zap"
)

type Service struct {
	Name   string          // Unique name of the service.
	Host   string          // Host address where the service is accessible.
	Port   int             // Port number on which the service listens.
	BEPool *backend.BEPool // Backend Pool
	Algo   algo.Algorithm
	Hc     *HealthCheck // HealthCheck is the health check configuration.
	Ctx    context.Context
	Logger *zap.Logger // Logger is used to log information and errors.
}

type Header struct {
	RequestHeaders        map[string]string // RequestHeaders is a list of headers to be added to the request.
	ResponseHeaders       map[string]string // ResponseHeaders is a list of headers to be added to the response.
	RemoveRequestHeaders  []string          // RemoveRequestHeaders is a list of headers to be removed from the request.
	RemoveResponseHeaders []string          // RemoveResponseHeaders is a list of headers to be removed from the response.
}

func NewService(svc *config.Service, ctx context.Context, logger *zap.Logger) *Service {
	bePool := backend.NewBEPool(svc.Backends, svc.RequestTimeout, svc.Headers, logger)
	hc := NewHealthCheck(svc.HealthCheck, logger, ctx)
	return &Service{
		Name:   svc.Name,
		BEPool: bePool,
		Algo:   algo.CreateAlgorithm(svc.LBtype, bePool.Backends), // Initialize the Algo field
		Hc:     hc,
		Ctx:    ctx,
		Logger: logger,
	}
}

func (s *Service) StartService() {
	s.Hc.StartHealthCheck(s.BEPool.Backends)
}

func (s *Service) ServeRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	currentBE := s.Algo.NextBackend(s.BEPool.Backends, w, r)
	if currentBE == nil {
		s.Logger.Error("Failed to select backend, No available backend found.")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	s.Logger.Info("Selected backend", zap.String("backend", currentBE.URL.String()))
	currentBE.Proxy.ServeHTTP(w, r)
	duration := time.Since(start)
	currentBE.ResponseTime = duration
	s.Logger.Info("Request served", zap.String("backend", currentBE.URL.String()))
}
