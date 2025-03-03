package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"vgo-balancer/pkg/config"
	"vgo-balancer/pkg/service"

	"go.uber.org/zap"
)

// default configurations
const (
	DefaultHTTPPort     = 80
	DefaultHTTPSPort    = 443
	ReadTimeout         = 15 * time.Second
	WriteTimeout        = 15 * time.Second
	IdleTimeout         = 60 * time.Second
	ShutdownGracePeriod = 30 * time.Second
)

type Server struct {
	logger *zap.Logger
	config *config.VgoBalancer
	ctx    context.Context
}

// Service Map stores the services registered with the load balancer.
var serviceMap map[string]*service.Service

func NewServer(ctx context.Context, logger *zap.Logger, config *config.VgoBalancer) *Server {
	server := &Server{
		logger: logger,
		config: config,
		ctx:    ctx,
	}
	return server
}

func (s *Server) Start() {
	if s.config.Port == 0 {
		s.config.Port = DefaultHTTPPort
	}
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.logger.Info("Parsing configuration and registering services")
	serviceMap = make(map[string]*service.Service)	
	for _, svc := range s.config.Services {
		if _, ok := serviceMap[svc.Name]; !ok {
			svcLogger := s.logger.With(zap.String("service", svc.Name))
			service := service.NewService(&svc, s.ctx, svcLogger)
			service.StartService()
			serviceMap[svc.Name] = service
			s.logger.Info(fmt.Sprintf("Service: %s, registered successfully.", svc.Name))
		} else {
			s.logger.Warn(fmt.Sprintf("Service: %s already exists. Please change the service name to avoid conflicts.", svc.Name))
		}
	}

	s.logger.Info("Starting Load Balancer", zap.String("address", addr))
	http.ListenAndServe(addr, http.HandlerFunc(s.handleRequest))
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received request", zap.String("method", r.Method), zap.String("url", r.URL.String()))
	svcName, err := s.GetServiceName(r)
	if err != nil {
		s.logger.Error("Failed to parse the request URL", zap.Error(err))
		s.logger.Error("failed to parse the request URL", zap.Error(err))
	}

	s.logger.Info("Service name extracted from URL", zap.String("service", svcName))
	if svc, ok := serviceMap[svcName]; ok {
		svc.ServeRequest(w, r)
	} else {
		s.logger.Error("service not found", zap.String("service", svcName))
		http.Error(w, "Service not found", http.StatusNotFound)
	}

}

func (s *Server) GetServiceName(r *http.Request) (string, error) {
	svcUrl, err := url.Parse(r.URL.Path)
	if err != nil {
		return "", err
	}
	svcPath := svcUrl.Path[1:] // remove the leading slash
	svcName := strings.Split(svcPath, "/")[0]
	return svcName, nil // assuming the service name is the first part of the path
}


