package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// StartGRPC starts the gRPC server
func (s *Server) StartGRPC(port string) *Server {
	if len(s.err) != 0 {
		return s
	}

	if len(port) == 0 || port == "0" {
		port = s.grpcPort
	} else {
		s.grpcPort = port
	}

	s.ui.Log.Debugf("using the following ports. Mux: %s, gRPC: %s, HTTP: %s", s.muxPort, s.grpcPort, s.httpPort)

	// grpcAddress is the address the service is bind to, it's better `:port` than
	// `host:port`/`localhost:port`. The former allows to accept connections over
	// the network
	grpcAddress := fmt.Sprintf(":%s", s.grpcPort)

	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return s.addErr(err)
	}

	opts := []grpc.ServerOption{}
	secureMsg := "insecure "
	if !s.insecure {
		withClientCredentials := len(s.muxPort) != 0 || len(s.httpPort) != 0

		var creds credentials.TransportCredentials
		if withClientCredentials {
			s.ui.Log.Debugf("setting up the gRPC secure with client credentials to be used by a gRPC gateway")
			creds = credentials.NewClientTLSFromCert(s.certPool, s.host)
		} else {
			// Only when running only secure gRPC, no HTTP
			s.ui.Log.Debugf("setting up the gRPC secure with server credentials")
			creds = credentials.NewTLS(&tls.Config{
				ClientAuth:               tls.RequireAndVerifyClientCert,
				Certificates:             []tls.Certificate{s.certificate},
				ClientCAs:                s.certPool,
				MinVersion:               tls.VersionTLS12,
				NextProtos:               []string{"h2"},
				PreferServerCipherSuites: true,
				CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				},
			})
		}
		opts = append(opts, grpc.Creds(creds))
		secureMsg = "secure "
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

	s.makeSignalCh()

	s.ui.Log.Infof("starting %sgRPC server on %s...", secureMsg, lis.Addr())

	go func() {
		s.ctx = context.WithValue(s.ctx, contextKey("grpc"), true)
		err := s.grpcServer.Serve(lis)
		s.ui.Log.Warnf("the gRPC server stoped serving. Error: %s", err)
		s.errCh <- err
	}()
	s.running = true

	return s
}

// StopGRPC stops the GRPC server if it's running
func (s *Server) StopGRPC() *Server {
	if s.grpcServer == nil {
		return s
	}
	s.ui.Log.Warnf("shutting down gRPC server...")
	s.grpcServer.GracefulStop()

	// if HTTP is running, wait for it to complete shutdown ...
	if s.ctx.Value(contextKey("http")) == true {
		s.ui.Log.Debugf("gRPC is waiting for HTTP/REST to complete shutdown")
		<-s.ctx.Done()
		return s
	}
	// ... otherwise, cancel the context so the Wait() can finish
	ctx, cancel := context.WithCancel(s.ctx)
	s.ctx = ctx
	cancel()
	return s
}
