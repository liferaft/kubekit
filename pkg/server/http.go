package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// StartHTTP starts the HTTP/REST server
func (s *Server) StartHTTP(port string) *Server {
	if len(s.err) != 0 {
		return s
	}

	if len(port) == 0 || port == "0" {
		port = s.httpPort
	} else {
		s.httpPort = port
	}

	// A solo HTTP/REST API service is not part of the scope of this package. If
	// you like a solo HTTP, use a HTTP framework suitable to implement API's.
	// This package relies on gRPC to run
	if len(s.grpcPort) == 0 { // || !s.IsServing("grpc") {
		return s.addErr(fmt.Errorf("the HTTP/REST service should not start without the gRPC service. The gRPC port is 0"))
	}

	// serveAddress is used for the secure TLS connection
	// serveAddress := fmt.Sprintf("%s:%s", s.host, s.httpPort)

	// httpAddress is the address the service is bind to, it's better `:port` than
	// `host:port`/`localhost:port`. The former allows to accept connections over
	// the network
	httpAddress := fmt.Sprintf(":%s", s.httpPort)

	upstreamGRPCServerAddress := fmt.Sprintf("%s:%s", s.host, s.grpcPort)

	ctx, cancel := context.WithCancel(s.ctx)
	s.ctx = ctx

	var opts []grpc.DialOption
	var secureMsg string
	if s.insecure {
		opts = []grpc.DialOption{grpc.WithInsecure()}
		secureMsg = "insecure "
	} else {
		tlsConf := &tls.Config{
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
		}
		creds := credentials.NewTLS(tlsConf)

		opts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
		secureMsg = "secure "
	}

	gwmux := runtime.NewServeMux(runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true},
	))
	// Or:?
	// gwmux := runtime.NewServeMux()

	for name, serv := range s.services {
		s.ui.Log.Debugf("registering service %q for HTTP/REST", name)
		if err := serv.RegisterHandlerFromEndpoint(s.ctx, gwmux, upstreamGRPCServerAddress, opts); err != nil {
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

	s.httpServer = &http.Server{
		Addr:    httpAddress,
		Handler: mux,
	}

	if !s.insecure {
		s.httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{s.certificate},
		}
	}

	s.makeSignalCh()

	s.ui.Log.Infof("starting %sHTTP/REST gateway on %s...", secureMsg, s.httpServer.Addr)

	go func() {
		defer cancel()
		s.ctx = context.WithValue(s.ctx, contextKey("http"), true)
		err := s.httpServer.ListenAndServe()
		s.ui.Log.Warnf("the HTTP/REST server stoped serving. Error: %s", err)
		s.errCh <- err
	}()
	s.running = true

	return s
}

// StopHTTP stops the HTTP/REST server if it's running
func (s *Server) StopHTTP() *Server {
	if s.httpServer == nil {
		return s
	}

	timeout := 5 * time.Second
	s.ui.Log.Warnf("shutting down HTTP/REST server in %s seconds...", timeout)
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()
	s.ctx = ctx

	err := s.httpServer.Shutdown(s.ctx)
	return s.addErr(err)
}

// WithSwagger enable the HTTP server to present the swagger JSON
func (s *Server) WithSwagger() *Server {
	s.withSwagger = true
	return s
}

func (s *Server) setSwagger(mux *http.ServeMux) {
	for name, serv := range s.services {
		if serv.SwaggerBytes == nil {
			continue
		}
		path := fmt.Sprintf("/swagger/%s.json", strings.ToLower(serv.Name))
		s.ui.Log.Debugf("registering swagger for service %s on HTTP/REST (%s)", name, path)
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			io.Copy(w, bytes.NewReader(serv.SwaggerBytes))
		})
	}
}
