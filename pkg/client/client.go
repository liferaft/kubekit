package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/johandry/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

// Config is the kubekit client configuration
type Config struct {
	APIVersion string
	Host       string
	Logger     *log.Logger
	err        error

	GrpcPort string
	GrpcConn *grpc.ClientConn

	HTTPClient         *http.Client
	HTTPBaseURL        *url.URL
	HTTPHealthzClient  *http.Client
	HTTPHealthzBaseURL *url.URL

	Insecure    bool
	Certificate tls.Certificate
	CertPool    *x509.CertPool
}

// grpcGatewayErrorBody is the same struct `errorBody` defined
// at https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
type grpcGatewayErrorBody struct {
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Code    int32         `json:"code"`
	Details []interface{} `json:"details,omitempty"`
}

// New creates a new KubeKit Client
func New(apiVersion string, logger *log.Logger) *Config {
	return &Config{
		APIVersion: apiVersion,
		Insecure:   true,
		Logger:     logger,
	}
}

func (c *Config) Error() (*Config, error) {
	return c, c.err
}

// WithHTTP makes the KubeKit client to have a REST/HTTP client
func (c *Config) WithHTTP(host, httpPort, healthzPort string) {
	httpBaseURL, httpClient := c.getHTTPClient(host, httpPort)
	httpHealthzBaseURL, httpHealthzClient := c.getHTTPHealthzClient(host, healthzPort, httpPort)

	c.HTTPClient = httpClient
	c.HTTPBaseURL = httpBaseURL

	c.HTTPHealthzClient = httpHealthzClient
	c.HTTPHealthzBaseURL = httpHealthzBaseURL

	if c.Host == "" {
		c.Host = host
	}
}

// WithTLS makes the client secure by loading the given certificates
func (c *Config) WithTLS(certDir, certFile, keyFile, caFile string) error {
	var err error

	c.Insecure = false

	c.Certificate, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("could not load key pair. %s", err)
	}

	c.CertPool = x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("could not read ca certificate: %s", err)
	}

	if ok := c.CertPool.AppendCertsFromPEM(ca); !ok {
		return fmt.Errorf("failed to append CA certs")
	}

	return nil
}

// GetGRPCConn create a gRPC connection to the given host:port
func (c *Config) GetGRPCConn(host, port string) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%s", host, port)
	if c.Insecure {
		return grpc.Dial(address, grpc.WithInsecure())
	}

	var opts []grpc.DialOption

	// creds := credentials.NewClientTLSFromCert(certPool, address)
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   host,
		Certificates: []tls.Certificate{c.Certificate},
		RootCAs:      c.CertPool,
	})

	opts = append(opts, grpc.WithTransportCredentials(creds))

	return grpc.Dial(address, opts...)
}

func (c *Config) getHTTPClient(host, port string) (*url.URL, *http.Client) {
	client := &http.Client{}
	schema := "http"

	if !c.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				// InsecureSkipVerify: true,
				ServerName:               host,
				ClientAuth:               tls.RequireAndVerifyClientCert,
				Certificates:             []tls.Certificate{c.Certificate},
				RootCAs:                  c.CertPool,
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
			},
		}
		schema = "https"
	}

	baseURL, _ := url.Parse(fmt.Sprintf("%s://%s:%s", schema, host, port))

	return baseURL, client
}

func (c *Config) getHTTPHealthzClient(host, port, httpPort string) (*url.URL, *http.Client) {
	client := &http.Client{}
	schema := "http"

	if !c.Insecure && (httpPort == port) {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				// There is no need to enforce security with health check, this make is slower
				InsecureSkipVerify: true,
			},
		}
		schema = "https"
	}
	healthzBaseURL, _ := url.Parse(fmt.Sprintf("%s://%s:%s", schema, host, port))

	return healthzBaseURL, client
}

// ProtoFunc is a function to call inside a function to access the GRPC or REST
// API. It always return the API result and an error
type ProtoFunc func() (string, error)

// TODO: Add modes to replace `firstTry`. Mode could be: `first-succeed`, `no-http`, `no-grpc`, `only-grpc`, `only-http`

// RunGRPCnRESTFunc execute the given functions to access the GRPC and REST API
func (c *Config) RunGRPCnRESTFunc(name string, firstTry bool, fnGRPC ProtoFunc, fnREST ProtoFunc) (string, error) {
	var (
		outputGRPC string
		outputHTTP string
		errGRPC    error
		errHTTP    error
	)

	if c.GrpcConn != nil {
		outputGRPC, errGRPC = fnGRPC()
		if errGRPC != nil {
			c.Logger.Warnf("gRPC request for %q failed with error: %s", name, errGRPC)
		} else if firstTry {
			c.Logger.Debug("gRPC succeed and HTTP/REST won't be executed, as requested")
			return outputGRPC, nil
		}
	}
	if c.HTTPClient != nil {
		if outputHTTP, errHTTP = fnREST(); errHTTP != nil {
			c.Logger.Warnf("HTTP/REST request for %q failed with error: %s", name, errHTTP)
		}
		if isErr, err := IsError(outputHTTP); isErr {
			errHTTP = err
			c.Logger.Warnf("HTTP/REST request for %q returned error: %s", name, errHTTP)
		}
	}

	return c.validOutput(outputGRPC, outputHTTP, errGRPC, errHTTP)
}

// ValidOutput returns a correct output and error (if any) when both protocols where call
func (c *Config) validOutput(grpcOutput, httpOutput string, grpcErr, httpErr error) (string, error) {
	if grpcErr != nil && httpErr != nil {
		return "", fmt.Errorf("failed the gRPC & REST request.\ngRPC error: %s\nHTTP/REST error: %s", grpcErr, httpErr)
	}
	// Only one error is not nil
	if grpcErr != nil {
		if httpOutput != "" {
			c.Logger.Warnf("the REST request succeeded but the gRPC request failed with error %s", grpcErr)
			return httpOutput, nil
		}
		return "", grpcErr
	}
	if httpErr != nil {
		if grpcOutput != "" {
			c.Logger.Warnf("the gRPC request succeeded but the REST request failed with error %s", httpErr)
			return grpcOutput, nil
		}
		return "", httpErr
	}
	// No error happened for both protocols
	if grpcOutput != "" && httpOutput != "" && grpcOutput != httpOutput {
		return "", fmt.Errorf("the gRPC & REST succeeded but with different results")
	}
	// One of the outputs is not empty or both are the same, so return one of them that is not empty
	if grpcOutput != "" {
		return grpcOutput, nil
	}
	// Either REST has something or both are empty, with no error
	return httpOutput, nil
}

// IsError returns true and the error message if the REST output is an error
// response. The gRPC-Gateway returns an error struct when an error is found but
// it's not an error type. The structure is defined
func IsError(output string) (isErr bool, err error) {
	if !strings.Contains(output, "\"error\"") {
		return isErr, err
	}
	isErr = true
	errOutput := &grpcGatewayErrorBody{}
	if err = json.Unmarshal([]byte(output), errOutput); err != nil {
		return isErr, err
	}
	var details string
	if len(errOutput.Details) != 0 {
		details = fmt.Sprintf(". Details: %s", errOutput.Details)
	}
	codeGRPC := codes.Code(errOutput.Code)
	codeHTTP := http.StatusText(runtime.HTTPStatusFromCode(codeGRPC))

	err = fmt.Errorf("%s. gRPC Code: %s. HTTP Code: %s%s", errOutput.Error, codeGRPC, codeHTTP, details)

	return isErr, err
}
