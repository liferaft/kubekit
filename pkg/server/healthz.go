package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc/health"
)

// WithHealthCheck enable the server to have health check
func (s *Server) WithHealthCheck(port string) *Server {
	s.withHealthCheck = true
	s.healthPort = port
	return s
}

// RegisterHealthCheck enable the Health Check for the gRPC server. Health is
// checked with gRPC using the same gRRPC server using the proto from
// 'google.golang.org/grpc/health/grpc_health_v1'
func (s *Server) RegisterHealthCheck() *Server {
	if len(s.err) != 0 {
		return s
	}

	if s.grpcServer == nil {
		return s.addErr(fmt.Errorf("cannot start HealthCheck before gRPC service"))
	}
	if s.healthServer != nil {
		return s
	}

	s.healthServer = health.NewServer()
	// Do not use ResumeServices because its does not insert the services.
	for name := range s.services {
		s.ui.Log.Debugf("registering service %q for Health Check", name)
		s.healthServer.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)
	}
	grpc_health_v1.RegisterHealthServer(s.grpcServer, s.healthServer)

	return s
}

func (s *Server) setHealthCheck(mux *http.ServeMux) {
	if len(s.healthPort) != 0 && s.healthPort != s.httpPort && s.healthPort != s.muxPort && s.healthPort != s.grpcPort {
		s.StartHealthz()
		return
	}
	s.setHealthzToMux(mux)
}

func (s *Server) setHealthzToMux(mux *http.ServeMux) {
	for name, serv := range s.services {
		path := fmt.Sprintf("/healthz/%s/%s", serv.Version, serv.Name)
		s.ui.Log.Debugf("registering healthz for service %s on HTTP/REST (%s)", name, path)
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.EscapedPath()
			name := strings.TrimPrefix(path, "/healthz/")
			if len(name) == 0 {
				Errorf(http.StatusBadRequest, "healthz", "cannot identify the service name from the path %q", path).Write(w)
				return
			}
			name = strings.Replace(name, "/", ".", -1)

			code := s.checkService(name)
			HealthzResponse(int(code), name).Write(w)
			// if s.IsServiceNotServing(name) {
			// 	HealthzResponse()
			// 	w.WriteHeader(http.StatusServiceUnavailable)
			// 	return
			// }
			// if s.IsServiceServing(name) {
			// 	w.WriteHeader(http.StatusOK)
			// 	return
			// }
			// Errorf(http.StatusInternalServerError, "healthz", "unknown service %q", path).Write(w)
		})
		mux.HandleFunc("healthz", func(w http.ResponseWriter, r *http.Request) {
			if s.IsHealthy() {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		})
	}
}

// StartHealthz starts the Health Check Server on HTTP/REST on the given port
func (s *Server) StartHealthz() *Server {
	if len(s.err) != 0 {
		return s
	}
	if s.healthServer == nil {
		return s.addErr(fmt.Errorf("not found HealthCheck server for the gRPC service"))
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.ctx = ctx

	// httpHealthServer := http.NewServeMux()
	mux := http.NewServeMux()
	s.setHealthzToMux(mux)
	s.healthHTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.healthPort),
		Handler: mux,
	}

	s.ui.Log.Infof("starting HTTP/REST HealthCheck Server on %s...", s.healthHTTPServer.Addr)

	go func() {
		defer cancel()
		s.ctx = context.WithValue(s.ctx, contextKey("healthz"), true)
		err := s.healthHTTPServer.ListenAndServe()
		s.ui.Log.Warnf("the Health Check HTTP/REST server stoped serving. Error: %s", err)
		s.errCh <- err
	}()

	return s
}
