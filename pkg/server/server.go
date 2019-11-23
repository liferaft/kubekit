package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"strings"

	"fmt"

	"os/signal"

	"github.com/kraken/ui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	servertls "github.com/kubekit/kubekit/pkg/server/tls"
)

// Server encapsulate the server configuration
type Server struct {
	Name     string
	host     string
	ui       *ui.UI
	ctx      context.Context
	running  bool
	services Services

	// Channels
	err    []error
	errCh  chan error
	sigCh  chan os.Signal
	doneCh chan struct{}

	// TLS
	insecure    bool
	serverCert  *servertls.Certificate
	caCert      *servertls.Certificate
	certificate tls.Certificate
	certPool    *x509.CertPool

	// Servers
	grpcServer       *grpc.Server
	grpcPort         string
	httpServer       *http.Server
	httpPort         string
	muxServer        *http.Server
	muxPort          string
	healthServer     *health.Server
	healthHTTPServer *http.Server
	healthPort       string
	withSwagger      bool
	withHealthCheck  bool
	allowCORS        bool
}

type contextKey string

// New creates a new Server to expose the HTTP/REST and gRPC API as
// well as the Swagger configuration
func New(ctx context.Context, name, host string, parentUI *ui.UI) *Server {
	s := &Server{
		Name:     name,
		host:     host,
		ui:       parentUI,
		insecure: true,
	}

	s.host = "localhost"

	s.services = make(map[string]*Service, 0)

	s.ctx = ctx
	s.sigCh = make(chan os.Signal, 1)
	s.errCh = make(chan error, 1)

	return s
}

// NewTLS creates a new server using TLS
func NewTLS(ctx context.Context, name, host string, certDir, pubKeyFile, privKeyFile, clientCAFile string, parentUI *ui.UI) *Server {
	s := New(ctx, name, host, parentUI)
	return s.WithTLS(certDir, pubKeyFile, privKeyFile, clientCAFile)
}

// StartAndWait starts the HTTP/REST and gRPC server on the same port. It
// returns when the server is manually stopped or due to an error, returning
// such error
func StartAndWait(address string, parentUI *ui.UI, services ...*Service) error {
	return startAndWait(address, "", "", "", "", parentUI, services...)
}

// StartSecureAndWait starts the HTTP/REST and gRPC server on the same port
// using TLS. It returns when the server is manually stopped or due to an error,
// returning such error
func StartSecureAndWait(address string, certDir, pubKeyFile, privKeyFile, clientCAFile string, parentUI *ui.UI, services ...*Service) error {
	return startAndWait(address, certDir, pubKeyFile, privKeyFile, clientCAFile, parentUI, services...)
}

func startAndWait(address string, certDir, pubKeyFile, privKeyFile, clientCAFile string, parentUI *ui.UI, services ...*Service) error {
	ctx := context.Background()

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("found no service to serve")
	}

	s := New(ctx, "server", host, parentUI)

	if len(certDir)+len(pubKeyFile)+len(privKeyFile)+len(clientCAFile) != 0 {
		s.WithTLS(certDir, pubKeyFile, privKeyFile, clientCAFile)
	}

	for i, service := range services {
		s.AddService(fmt.Sprintf("service-%d", i), service)
	}
	return s.StartMux(port).
		Wait().
		Err()
}

// Err returns the errors ocurred
func (s *Server) Err() error {
	if len(s.err) == 0 {
		return nil
	}
	if len(s.err) == 1 {
		return s.err[0]
	}

	var errStr []string
	for _, e := range s.err {
		errStr = append(errStr, e.Error())
	}
	return fmt.Errorf("Errors: %q", strings.Join(errStr, ", "))
}

func (s *Server) addErr(e error) *Server {
	if e != nil {
		s.err = append(s.err, e)
	}
	return s
}

func (s *Server) makeSignalCh() {
	if s.running {
		return
	}
	if s.sigCh == nil {
		s.sigCh = make(chan os.Signal, 1)
	}
	signal.Notify(s.sigCh, os.Interrupt)
}

// Wait waits for any server to stop either by the user or an error
func (s *Server) Wait() *Server {
	if !s.running {
		return s
	}
	if len(s.err) != 0 {
		return s
	}

	for {
		select {
		case <-s.ctx.Done():
			return s.addErr(s.ctx.Err())
		case e := <-s.errCh:
			s.addErr(e)
			ctx, cancel := context.WithCancel(s.ctx)
			s.ctx = ctx
			cancel()
		case <-s.sigCh:
			s.ui.Log.Warn("received a ^C signal, shutting down the servers")
			s.Stop()
		}
	}
}

// Start starts a server to expose the provisioner API using HTTP/REST and gRPC
func (s *Server) Start(port string, ports ...string) *Server {
	var err error
	if s.muxPort, s.httpPort, s.grpcPort, err = getPorts(port, ports...); err != nil {
		return s.addErr(err)
	}

	s.ui.Log.Debugf("using the following ports. Mux: %s, gRPC: %s, HTTP: %s", s.muxPort, s.grpcPort, s.httpPort)

	if s.muxPort != "" {
		return s.StartMux(s.muxPort)
	}
	// return s.StartHTTP(s.httpPort)
	return s.StartGRPC(s.grpcPort).StartHTTP(s.httpPort)
}

// IsServing returns true if given service ("grpc", "http", "mux") is running
func (s *Server) IsServing(service string) bool {
	if s.ctx == nil {
		return false
	}
	if s.ctx.Value(contextKey(service)) == nil {
		return false
	}
	_, ok := s.ctx.Value(contextKey(service)).(bool)
	return ok
}

// Stop stops both servers GRPC and HTTP/REST
func (s *Server) Stop() *Server {
	return s.StopHTTP().StopGRPC().StopMux()
}
