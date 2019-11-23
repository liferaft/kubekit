package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// ServiceRegisterable is an interface for a service to implement a function to register
// itself to a gRPC server
type ServiceRegisterable interface {
	Register(*grpc.Server)
}

// RegisterHandlerFromEndpoint is a function from the gRPC GateWay API to dialing
// to "endpoint" and closing the connection when "ctx" gets done
type RegisterHandlerFromEndpoint func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

// Service contain the information a server needs from a service to operate
type Service struct {
	Version                     string
	Name                        string
	ServiceRegister             ServiceRegisterable
	RegisterHandlerFromEndpoint RegisterHandlerFromEndpoint
	SwaggerBytes                []byte
}

// Services is a collection of services
type Services map[string]*Service

// AddService appends a new service to serve
func (s *Server) AddService(name string, service *Service) *Server {
	s.services[name] = service
	return s
}

// AddServices appends all the given services to serve
func (s *Server) AddServices(services Services) *Server {
	for name, serv := range services {
		s.services[name] = serv
	}
	return s
}

// SetServiceStatus change the status of the given service
func (s *Server) SetServiceStatus(name string, status grpc_health_v1.HealthCheckResponse_ServingStatus) *Server {
	if _, ok := s.services[name]; !ok {
		return s.addErr(fmt.Errorf("service %s not found", name))
	}
	s.healthServer.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)
	return s
}

// ShutdownService sets all serving status to NOT_SERVING, and configures the
// server to ignore all future status changes.
func (s *Server) ShutdownService(name string) *Server {
	return s.SetServiceStatus(name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
}

// ResumeService sets all serving status to SERVING, and configures the server
// to accept all future status changes.
func (s *Server) ResumeService(name string) *Server {
	return s.SetServiceStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)
}

func (s *Server) checkService(name string) grpc_health_v1.HealthCheckResponse_ServingStatus {
	res, err := s.healthServer.Check(s.ctx, &grpc_health_v1.HealthCheckRequest{
		Service: name,
	})
	if err != nil {
		return grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}
	return res.GetStatus()
}

// IsServiceServing return true if the given service name is serving
func (s *Server) IsServiceServing(name string) bool {
	status := s.checkService(name)
	return status == grpc_health_v1.HealthCheckResponse_SERVING
}

// IsServiceNotServing return true if the given service name is serving
func (s *Server) IsServiceNotServing(name string) bool {
	status := s.checkService(name)
	return status == grpc_health_v1.HealthCheckResponse_NOT_SERVING
}

// IsHealthy return true if all the services are serving. If one service is not
// serving, the server is not healthy
func (s *Server) IsHealthy() bool {
	for name := range s.services {
		if s.IsServiceNotServing(name) {
			return false
		}
	}
	return true
}

// ShutdownServices sets all serving status to NOT_SERVING, and configures the
// server to ignore all future status changes.
func (s *Server) ShutdownServices() *Server {
	s.healthServer.Shutdown()
	return s
}

// ResumeServices sets all serving status to SERVING, and configures the server
// to accept all future status changes.
func (s *Server) ResumeServices() *Server {
	s.healthServer.Resume()
	return s
}
