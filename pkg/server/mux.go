package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// StartMux starts a Multiplex server to redirect the traffic to the HTTP/REST
// and GRPC services, allowing to have both services in one port
func (s *Server) StartMux(port string) *Server {
	if len(s.err) != 0 {
		return s
	}

	if len(port) == 0 || port == "0" {
		port = s.muxPort
	} else {
		s.muxPort = port
	}
	serveAddress := fmt.Sprintf("%s:%s", s.host, port)

	// if s.insecure {
	// 	return s.addErr(fmt.Errorf("HTTP/REST and gRPC cannot run insecure on the same port. Set keys to enable TLS or use flag --grpc-port (and optionally --port) to run insecure using different ports"))
	// }

	// if s.grpcServer == nil {
	// 	return s.addErr(fmt.Errorf("cannot start Multiplex server, gRPC server is not running"))
	// }
	// if s.httpServer == nil {
	// 	return s.addErr(fmt.Errorf("cannot start Multiplex server, HTTP/REST server is not running"))
	// }

	// s.grpcPort, s.httpPort = port, port

	opts := []grpc.ServerOption{}
	if !s.insecure {
		opts = append(opts, grpc.Creds(credentials.NewClientTLSFromCert(s.certPool, s.host)))
	}
	s.grpcServer = grpc.NewServer(opts...)

	for name, serv := range s.services {
		s.ui.Log.Debugf("registering service %q for gRPC", name)
		serv.ServiceRegister.Register(s.grpcServer)
	}
	if s.withHealthCheck {
		s.RegisterHealthCheck()
	}
	reflection.Register(s.grpcServer)

	// Because we run our REST endpoint on the same port as the GRPC the address is the same.
	upstreamGRPCServerAddress := serveAddress

	ctx, cancel := context.WithCancel(s.ctx)
	s.ctx = ctx

	var gwopts []grpc.DialOption
	if s.insecure {
		gwopts = []grpc.DialOption{grpc.WithInsecure()}
	} else {
		creds := credentials.NewTLS(&tls.Config{
			ServerName:               s.host,
			RootCAs:                  s.certPool,
			MinVersion:               tls.VersionTLS12,
			NextProtos:               []string{"h2"},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		})
		gwopts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	}

	gwmux := runtime.NewServeMux(runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true},
	))
	// Or:?
	// gwmux := runtime.NewServeMux()

	for name, serv := range s.services {
		s.ui.Log.Debugf("registering service %q for HTTP/REST", name)
		if err := serv.RegisterHandlerFromEndpoint(s.ctx, gwmux, upstreamGRPCServerAddress, gwopts); err != nil {
			defer cancel()
			return s.
				addErr(fmt.Errorf("failed to register service %s. %s", name, err)).
				Stop()
		}
	}

	mux := http.NewServeMux()
	if s.withSwagger {
		s.setSwagger(mux)
	}
	if s.healthServer != nil {
		s.setHealthCheck(mux)
	}
	mux.Handle("/", gwmux)

	var httpMux http.Handler

	if s.allowCORS {
		httpMux = allowCORS(mux)
	} else {
		httpMux = mux
	}

	// Accepting connections over the network with `:port` (better) instead of `host:port`/`localhost:port`.
	httpAddress := fmt.Sprintf(":%s", port) // Or?: serveAddress

	conn, err := net.Listen("tcp", httpAddress)
	if err != nil {
		defer cancel()
		return s.
			addErr(fmt.Errorf("failed to create connection on %q. %s", ":%s"+port, err)).
			Stop()
	}

	httpServer := &http.Server{
		Addr:    serveAddress,                                  // fmt.Sprintf("%s:%s", "", port), // "" -> accept connections over the network
		Handler: muxHandlerFuncInsecure(s.grpcServer, httpMux), // s.httpServer.Handler), // HTTP is required to be running first
	}

	secureMsg := "insecure "
	if !s.insecure {
		httpServer.Handler = muxHandlerFuncSecure(s.grpcServer, httpMux)
		httpServer.TLSConfig = &tls.Config{
			ServerName:               s.host,
			Certificates:             []tls.Certificate{s.certificate},
			NextProtos:               []string{"h2"},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		}
		conn = tls.NewListener(conn, httpServer.TLSConfig)
		secureMsg = "secure "
	}

	s.makeSignalCh()

	s.ui.Log.Infof("starting %sHTTP/REST gateway and gRPC server on %s...", secureMsg, httpServer.Addr)

	go func() {
		defer cancel()
		s.ctx = context.WithValue(s.ctx, contextKey("mux"), true)
		s.httpServer = httpServer
		err := s.httpServer.Serve(conn)
		s.ui.Log.Warnf("the HTTP/REST gateway and gRPC server stoped serving. Error: %s", err)
		s.errCh <- err
	}()
	s.running = true

	return s
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

func allowCORS(httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
			httpHandler.ServeHTTP(w, r)
		}
	})

}

// muxHandlerFuncSecure is used by the Multiplex HTTP Server to route the requests to
// the gRPC server or the HTTP muxer (HTTP/REST) at runtime. As it uses HTTP/2, TLS is required.
func muxHandlerFuncSecure(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	})
}

// muxHandlerFuncInsecure is used by the Multiplex HTTP Server to route the requests to
// the gRPC server or the HTTP muxer (HTTP/REST) at runtime. This function works without TLS
func muxHandlerFuncInsecure(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				httpHandler.ServeHTTP(w, r)
			}
		}), &http2.Server{})
}

// StopMux stops the Multiplex server if it's running
func (s *Server) StopMux() *Server {
	if s.muxServer == nil {
		return s
	}
	timeout := 5 * time.Second
	s.ui.Log.Warnf("shutting down Multiplex server in %s seconds...", timeout)
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()
	s.ctx = ctx

	err := s.httpServer.Shutdown(s.ctx)
	return s.addErr(err)
}

// AllowCORS makes the server to allow CORS (Cross-Origin Resource Sharing)
func (s *Server) AllowCORS() *Server {
	s.allowCORS = true
	return s
}

// SetCORS enable CORS if and only if the first parmeter (enabled) is `true`
func (s *Server) SetCORS(enabled bool) *Server {
	if enabled {
		return s.AllowCORS()
	}
	return s
}
