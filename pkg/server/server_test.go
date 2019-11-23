package server_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johandry/log"
	"github.com/kraken/ui"
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/manifest"
	"github.com/liferaft/kubekit/pkg/server"
	servertls "github.com/liferaft/kubekit/pkg/server/tls"
	"github.com/liferaft/kubekit/pkg/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const (
	waitForServerTimeout = 60 // timeout after 60 seconds = 1 minute
	waitForServerTicker  = 2  // check every 2 seconds
	defPassword          = "KubeKit"
	defBitSize           = "4096"
)

const do4 = false

var (
	parentUI *ui.UI
)

func init() {
	l := log.NewDefault()
	l.SetLevel(logrus.DebugLevel)
	parentUI = ui.New(false, l)
}

func TestServer_Start(t *testing.T) {
	clustersPath, err := ioutil.TempDir("", "clusters")
	if err != nil {
		t.Errorf("failed to create temporal dir to store the clusters. %s", err)
	}
	defer os.RemoveAll(clustersPath)

	v1Services, err := service.ServicesForVersion("v1", false, clustersPath, parentUI)
	if err != nil {
		t.Errorf("Server.Start() error: %v", err)
	}
	os.Setenv(server.EnvTLSServerPassword, defPassword)

	vAPI := "v1"
	vKubekit := manifest.Version
	vKubernetes, vDocker, vEtcd := "unknown", "unknown", "unknown"
	if release, ok := manifest.KubeManifest.Releases[vKubekit]; ok {
		vKubernetes = release.KubernetesVersion
		vDocker = release.DockerVersion
		vEtcd = release.EtcdVersion
	}
	versionResult := fmt.Sprintf(`{"api":"%s","kubekit":"%s","kubernetes":"%s","docker":"%s","etcd":"%s"}`, vAPI, vKubekit, vKubernetes, vDocker, vEtcd)

	type fields struct {
		name              string
		host              string
		ui                *ui.UI
		ctx               context.Context
		services          server.Services
		insecure          bool
		noHTTP            bool
		tlsCertFile       string // if given, the certificate will be generated manually, with `openssl` in the temporal directory. i.e. "kubekit.crt"
		tlsPrivateKeyFile string // if given, the certificate will be generated manually, with `openssl` in the temporal directory. i.e. "kubekit.key"
		caFile            string // if given, the certificate will be generated manually, with `openssl` in the temporal directory. i.e. "kubekit-ca.key"
		serverCert        *servertls.Certificate
		caCert            *servertls.Certificate
		certificate       tls.Certificate
		certPool          *x509.CertPool
		grpcPort          string
		httpPort          string
		healthPort        string
		withSwagger       bool
		withHealthCheck   bool
		apiService        []string
	}
	type results struct {
		versionResult       string
		healthzResult       string
		healthzStatusResult string
	}
	tests := []struct {
		name    string
		fields  fields
		results results
		wantErr bool
	}{
		{
			name: "Scenario #1: Test GRPC Insecure",
			fields: fields{
				ctx:             context.Background(),
				name:            "kubekit02",
				host:            "localhost",
				ui:              parentUI,
				services:        v1Services,
				insecure:        true,
				noHTTP:          true,
				grpcPort:        "new",
				httpPort:        "",
				healthPort:      "grpcPort",
				withSwagger:     true,
				withHealthCheck: true,
				apiService:      []string{"grpc"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		{
			name: "Scenario #2: Test GRPC Secure",
			fields: fields{
				ctx:             context.Background(),
				name:            "kubekit04",
				host:            "localhost",
				ui:              parentUI,
				services:        v1Services,
				insecure:        false,
				noHTTP:          true,
				grpcPort:        "new",
				httpPort:        "",
				healthPort:      "grpcPort",
				withSwagger:     true,
				withHealthCheck: true,
				apiService:      []string{"grpc"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		{
			name: "Scenario #2.a: Test GRPC Secure, with manually generated CA certificate",
			fields: fields{
				ctx:               context.Background(),
				name:              "kubekit04",
				host:              "localhost",
				ui:                parentUI,
				services:          v1Services,
				insecure:          false,
				noHTTP:            true,
				tlsCertFile:       "",
				tlsPrivateKeyFile: "",
				caFile:            "kubekit04-ca.key",
				grpcPort:          "new",
				httpPort:          "",
				healthPort:        "grpcPort",
				withSwagger:       true,
				withHealthCheck:   true,
				apiService:        []string{"grpc"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		{
			name: "Scenario #2.b: Test GRPC Secure, with manually generated certificates",
			fields: fields{
				ctx:               context.Background(),
				name:              "kubekit04",
				host:              "localhost",
				ui:                parentUI,
				services:          v1Services,
				insecure:          false,
				noHTTP:            true,
				tlsCertFile:       "kubekit04.crt",
				tlsPrivateKeyFile: "kubekit04.key",
				caFile:            "kubekit04-ca.key",
				grpcPort:          "new",
				httpPort:          "",
				healthPort:        "grpcPort",
				withSwagger:       true,
				withHealthCheck:   true,
				apiService:        []string{"grpc"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		{
			name: "Scenario #2.c: Test GRPC Secure, with incorrect CA certificate",
			fields: fields{
				ctx:               context.Background(),
				name:              "kubekit04",
				host:              "localhost",
				ui:                parentUI,
				services:          v1Services,
				insecure:          false,
				noHTTP:            true,
				tlsCertFile:       "kubekit04.crt",
				tlsPrivateKeyFile: "kubekit04.key",
				caFile:            "",
				grpcPort:          "new",
				httpPort:          "",
				healthPort:        "grpcPort",
				withSwagger:       true,
				withHealthCheck:   true,
				apiService:        []string{"grpc"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: true,
		},
		{
			name: "Scenario #2.c: Test GRPC Secure, with incorrect certificates",
			fields: fields{
				ctx:               context.Background(),
				name:              "kubekit04",
				host:              "localhost",
				ui:                parentUI,
				services:          v1Services,
				insecure:          false,
				noHTTP:            true,
				tlsCertFile:       "",
				tlsPrivateKeyFile: "kubekit04.key",
				caFile:            "",
				grpcPort:          "new",
				httpPort:          "",
				healthPort:        "grpcPort",
				withSwagger:       true,
				withHealthCheck:   true,
				apiService:        []string{"grpc"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: true,
		},
		{
			name: "Scenario #3: HTTP != GRPC  Insecure",
			fields: fields{
				ctx:             context.Background(),
				name:            "kubekit05",
				host:            "localhost",
				ui:              parentUI,
				services:        v1Services,
				insecure:        true,
				noHTTP:          false,
				grpcPort:        "new",
				httpPort:        "new",
				healthPort:      "httpPort",
				withSwagger:     true,
				withHealthCheck: true,
				apiService:      []string{"grpc", "http"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		{
			name: "Scenario #5: HTTP == GRPC Insecure",
			fields: fields{
				ctx:             context.Background(),
				name:            "kubekit07",
				host:            "localhost",
				ui:              parentUI,
				services:        v1Services,
				insecure:        true,
				noHTTP:          false,
				grpcPort:        "new",
				httpPort:        "grpcPort",
				healthPort:      "grpcPort",
				withSwagger:     true,
				withHealthCheck: true,
				apiService:      []string{"mux"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		{
			name: "Scenario #6: HTTP == GRPC Secure",
			fields: fields{
				ctx:             context.Background(),
				name:            "kubekit08",
				host:            "localhost",
				ui:              parentUI,
				services:        v1Services,
				insecure:        false,
				noHTTP:          false,
				grpcPort:        "new",
				httpPort:        "grpcPort",
				healthPort:      "grpcPort",
				withSwagger:     true,
				withHealthCheck: true,
				apiService:      []string{"mux"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
	}
	if do4 {
		tests = append(tests, struct {
			name    string
			fields  fields
			results results
			wantErr bool
		}{
			name: "Scenario #4: HTTP != GRPC  Secure",
			fields: fields{
				ctx:             context.Background(),
				name:            "kubekit",
				host:            "localhost",
				ui:              parentUI,
				services:        v1Services,
				insecure:        false,
				noHTTP:          false,
				grpcPort:        "new",
				httpPort:        "new",
				healthPort:      "httpPort",
				withSwagger:     true,
				withHealthCheck: true,
				apiService:      []string{"grpc", "http"},
			},
			results: results{
				versionResult:       versionResult,
				healthzStatusResult: "SERVING",
				healthzResult:       fmt.Sprintf(`{"code":1,"status":"SERVING","message":"service \"%s.Kubekit\" is serving","service":"%s.Kubekit"}`, vAPI, vAPI),
			},
			wantErr: false,
		},
		)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//can't run parallel as port #'s are not atomic and will overlap at the tests start
			//running parallel seems to cause servers to start real s
			tmpCertDir, err := ioutil.TempDir("", "test")
			if err != nil {
				t.Errorf("failed to create temporal dir to store the certificates. %s", err)
			}
			defer os.RemoveAll(tmpCertDir)

			ctx, cancel := context.WithTimeout(tt.fields.ctx, 5*time.Minute)
			defer cancel()

			getPort := func(value string) (string, error) {
				switch value {
				case "new":
					return server.GetFreePort()
				case "grpcPort":
					return tt.fields.grpcPort, nil
				case "httpPort":
					return tt.fields.httpPort, nil
				case "healthPort":
					return tt.fields.healthPort, nil
				case "":
					return "", nil
				}
				return "", errors.New("unexpected value")
			}

			s := server.New(ctx, tt.fields.name, tt.fields.host, tt.fields.ui).AddServices(tt.fields.services)
			if tt.fields.withSwagger {
				s.WithSwagger()
			}

			if !tt.fields.insecure {
				var tlsCertFile, tlsPrivateKeyFile, caFile string
				if len(tt.fields.tlsCertFile+tt.fields.tlsPrivateKeyFile+tt.fields.caFile) != 0 {
					tlsCertFile, tlsPrivateKeyFile, caFile, err = generateManualCertificates(tt.fields.host, tmpCertDir, tt.fields.tlsCertFile, tt.fields.tlsPrivateKeyFile, tt.fields.caFile)
					if err != nil {
						if !tt.wantErr { // (err != nil) != tt.wantErr
							t.Errorf(err.Error())
						}
						return
					}
				}
				s.WithTLS(tmpCertDir, tlsCertFile, tlsPrivateKeyFile, caFile)
			}

			loop_count := 0
			for {
				loop_count++
				//Get a port for the service
				tt.fields.grpcPort, err = getPort(tt.fields.grpcPort)
				if err != nil {
					t.Fatalf("Error getting port for grpcPort %s", err)
				}
				tt.fields.httpPort, err = getPort(tt.fields.httpPort)
				if err != nil {
					t.Fatalf("Error getting port for httpPort %s", err)
				}
				tt.fields.healthPort, err = getPort(tt.fields.healthPort)
				if err != nil {
					t.Fatalf("Error getting port for healthPort %s", err)
				}
				if tt.fields.withHealthCheck {
					s.WithHealthCheck(tt.fields.healthPort)
				}

				if tt.fields.noHTTP {

					s.StartGRPC(tt.fields.grpcPort)
					if err := s.Err(); err == nil {
						t.Logf("Starting gRPC server on port %s", tt.fields.grpcPort)
						break
					}
					tt.fields.healthPort = "grpcPort"

				} else {

					s.Start(tt.fields.httpPort, tt.fields.grpcPort)
					if err := s.Err(); err == nil {
						t.Logf("Starting HTTP & gRPC server on ports %s & %s", tt.fields.httpPort, tt.fields.grpcPort)
						break
					}

					if tt.fields.healthPort == tt.fields.grpcPort {
						tt.fields.healthPort = "grpcPort"
					} else {
						tt.fields.healthPort = "httpPort"
					}
				}
				if loop_count > 20 {
					//we've tried this 20x.  Something is up.
					break
				}
			}

			defer func() {
				t.Logf("Stopping the server")
				s.Stop()
			}()

			t.Logf("tt.fields.grpcPort:%s,  tt.fields.httpPort:%s,  tt.fields.healthPort:%s", tt.fields.grpcPort, tt.fields.httpPort, tt.fields.healthPort)

			if err := waitForServerRunning(ctx, s, tt.wantErr, tt.fields.apiService...); err != nil {
				if !tt.wantErr { // (err != nil) != tt.wantErr
					t.Errorf(err.Error())
				}
				return
			}

			if err := s.Err(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Start() error = %v, wantErr %v", err, tt.wantErr)
			}

			optClient := &clientOpt{
				name:                tt.fields.name,
				insecure:            tt.fields.insecure,
				certDir:             tmpCertDir,
				host:                tt.fields.host,
				noHTTP:              tt.fields.noHTTP,
				grpcPort:            tt.fields.grpcPort,
				httpPort:            tt.fields.httpPort,
				healthPort:          tt.fields.healthPort,
				healthzStatusResult: tt.results.healthzStatusResult,
				healthzResult:       tt.results.healthzResult,
				versionResult:       tt.results.versionResult,
			}

			t.Log("Running client tests")
			testClient(t, optClient)
		})
	}
}

func BenchmarkGenerateManualCerts(b *testing.B) {
	benchmarks := []struct {
		name              string
		host              string
		tlsCertFile       string // if given, the certificate will be generated manually, with `openssl` in the temporal directory. i.e. "kubekit.crt"
		tlsPrivateKeyFile string // if given, the certificate will be generated manually, with `openssl` in the temporal directory. i.e. "kubekit.key"
		caFile            string // if given, the certificate will be generated manually, with `openssl` in the temporal directory. i.e. "kubekit-ca.key"
	}{
		{"no certs", "localhost", "", "", ""},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			tmpCertDir, err := ioutil.TempDir("", "test")
			if err != nil {
				b.Errorf("failed to create temporal dir to store the certificates. %s", err)
			}
			defer os.RemoveAll(tmpCertDir)
			b.ResetTimer()
			_, _, _, err = generateManualCertificates(bm.host, tmpCertDir, bm.tlsCertFile, bm.tlsPrivateKeyFile, bm.caFile)
			return
		})
	}
}

func generateManualCertificates(host, certDir, tlsCertFile, tlsPrivateKeyFile, caFile string) (serverKeyFilename string, serverCrtFilename string, caKeyFilename string, err error) {
	var caCrtFilename string
	if caFile != "" {
		if caKeyFilename, caCrtFilename, err = generateManualCA(certDir, caFile); err != nil {
			return "", "", "", err
		}
	} else {
		if tlsCertFile != "" || tlsPrivateKeyFile != "" {
			return "", "", "", fmt.Errorf("cannot provide TLS certificates without a CA key. The CA key is required for mutua TLS")
		}
		// If CA is empty as well as the TLS certs return with all empty, so all of them will be auto-generated
		// impossible scenario, because the first reason to be in this function is that at least one is not empty
		return "", "", "", nil
	}

	if tlsCertFile != "" && tlsPrivateKeyFile != "" {
		if serverKeyFilename, serverCrtFilename, err = generateManualServerCertificates(host, certDir, tlsCertFile, tlsPrivateKeyFile, caKeyFilename, caCrtFilename); err != nil {
			return "", "", "", err
		}
	} else if tlsCertFile == "" && tlsPrivateKeyFile == "" {
		// If CA is not empty but the TLS certs are, then auto-generate the TLS certs with the CA
		return "", "", caKeyFilename, nil
	} else {
		// Means, either the key or cert is empty ... both are required or both empty
		return "", "", "", fmt.Errorf("cannot provide TLS certificates without a CA key. The CA key is required for mutua TLS")
	}

	return serverKeyFilename, serverCrtFilename, caKeyFilename, nil
}

func generateManualCA(certDir, caFile string) (string, string, error) {
	// # Input:
	// # Generate self-created Certificate Authority (CA) key and certificate:
	// openssl genrsa -des3 -passout pass:$(TLS_PASSWD) -out kubekit-ca.key 4096
	// openssl req -new -x509 -days 365 -key kubekit-ca.key -out kubekit-ca.crt -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=www.kubekit.io" -passin pass:$(TLS_PASSWD)
	// # Output: kubekit-ca.key & kubekit-ca.crt

	caKeyFilename := filepath.Join(certDir, caFile)
	cmd := exec.Command("openssl", "genrsa", "-des3", "-passout", "pass:"+defPassword, "-out", caKeyFilename, defBitSize)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate CA key. Error: %s\nOutput: %s", err, output)
	}

	caCrtFile := strings.Replace(caFile, ".key", ".crt", 1)
	caCrtFilename := filepath.Join(certDir, caCrtFile)
	cmd = exec.Command("openssl", "req", "-new", "-x509", "-days", "365", "-key", caKeyFilename, "-out", caCrtFilename, "-subj", "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=www.kubekit.io", "-passin", "pass:"+defPassword)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate CA certificate. Error: %s\nOutput: %s", err, output)
	}

	return caKeyFilename, caCrtFilename, nil
}

func generateManualServerCertificates(host, certDir, tlsCertFile, tlsPrivateKeyFile, caKeyFilename, caCrtFilename string) (serverKeyFilename string, serverCrtFilename string, err error) {
	// # Input: kubekit-ca.key & kubekit-ca.crt
	// # Generate KubeKit server key and request for signing (csr):
	// openssl genrsa -des3 -passout pass:$(TLS_PASSWD) -out kubekit.key 4096
	// openssl req -new -key kubekit.key -out kubekit.csr -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=$(SERVER_ADDR)" -passin pass:$(TLS_PASSWD)
	// # Signing the certificate signing request (csr) with the self-created Certificate Authority (CA):
	// openssl x509 -req -days 365 -in kubekit.csr -CA kubekit-ca.crt -CAkey kubekit-ca.key -set_serial 01 -out kubekit.crt -passin pass:$(TLS_PASSWD)
	// # Create insecure (no password) version of KubeKit server key
	// openssl rsa -in kubekit.key -out kubekit.key.insecure -passin pass:$(TLS_PASSWD)
	// mv kubekit.key kubekit.key.secure
	// mv kubekit.key.insecure kubekit.key
	// # Output: kubekit.key, kubekit.csr, kubekit.crt & kubekit.key.secure

	serverKeyFilename = filepath.Join(certDir, tlsPrivateKeyFile)
	cmd := exec.Command("openssl", "genrsa", "-des3", "-passout", "pass:"+defPassword, "-out", serverKeyFilename, defBitSize)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate server key. Error: %s\nOutput: %s", err, output)
	}

	tlsCSRFile := strings.Replace(tlsPrivateKeyFile, ".key", ".csr", 1)
	serverCsrFilename := filepath.Join(certDir, tlsCSRFile)
	cmd = exec.Command("openssl", "req", "-new", "-key", serverKeyFilename, "-out", serverCsrFilename, "-subj", "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN="+host, "-passin", "pass:"+defPassword)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate server certificate signing request. Error: %s\nOutput: %s", err, output)
	}

	serverCrtFilename = filepath.Join(certDir, tlsCertFile)
	cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-in", serverCsrFilename, "-CA", caCrtFilename, "-CAkey", caKeyFilename, "-set_serial", "01", "-out", serverCrtFilename, "-passin", "pass:"+defPassword)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate server certificate. Error: %s\nOutput: %s", err, output)
	}

	cmd = exec.Command("openssl", "rsa", "-in", serverKeyFilename, "-out", serverKeyFilename+".insecure", "-passin", "pass:"+defPassword)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate an insecure (password-less) server key. Error: %s\nOutput: %s", err, output)
	}
	// overwrite `serverKeyFilename` the secure version is not needed
	os.Rename(serverKeyFilename+".insecure", serverKeyFilename)

	return serverKeyFilename, serverCrtFilename, nil
}

func generateManualClientCertificates(defCertDir, certDir string, tlsCertFile string, tlsPrivateKeyFile string, caFile string) {
	// # Input: kubekit-ca.key & kubekit-ca.crt
	// # Generate KubeKit client key and request for signing (csr):
	// openssl genrsa -des3 -passout pass:$(TLS_PASSWD) -out kubekit-client.key 4096
	// openssl req -new -key kubekit-client.key -out kubekitctl.csr -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=$(SERVER_ADDR)" -passin pass:$(TLS_PASSWD)
	// # Signing the certificate signing request (csr) with the self-created Certificate Authority (CA):
	// openssl x509 -req -days 365 -in kubekitctl.csr -CA kubekit-ca.crt -CAkey kubekit-ca.key -set_serial 01 -out kubekit-client.crt -passin pass:$(TLS_PASSWD)
	// # Create insecure (no password) version of KubeKit client key
	// openssl rsa -in kubekit-client.key -out kubekit-client.key.insecure -passin pass:$(TLS_PASSWD)
	// mv kubekit-client.key kubekit-client.key.secure
	// mv kubekit-client.key.insecure kubekit-client.key
	// # Output: kubekit-client.key, kubekitctl.csr, kubekit-client.crt & kubekit-client.key.secure
}

func waitForServerRunning(ctx context.Context, s *server.Server, wantErr bool, services ...string) error {
	done := make(chan struct{})
	timeout := time.After(waitForServerTimeout * time.Second)
	ticker := time.NewTicker(waitForServerTicker * time.Second)

	var err error

	for {
		select {
		case <-timeout:
			err = fmt.Errorf("Server.Start() Timeout, the services %v are not all running yet after %d seconds, canceling this test", services, waitForServerTimeout)
			close(done)
		case <-ticker.C:
			if err := s.Err(); (err != nil) != wantErr {
				err = fmt.Errorf("Server.Start() error = %v, wantErr %v", err, wantErr)
				close(done)
				continue
			}
			allRunning := true
			for _, service := range services {
				if !s.IsServing(service) {
					allRunning = false
				}
			}
			if allRunning {
				err = nil
				close(done)
			}
		case <-ctx.Done():
			err = fmt.Errorf("Server.Start() Something/someone request to cancel the server")
			close(done)
		case <-done:
			ticker.Stop()
			return err
		}
	}
}

type clientOpt struct {
	name                string
	insecure            bool
	host                string
	noHTTP              bool
	grpcPort            string
	httpPort            string
	healthPort          string
	certDir             string
	certificate         tls.Certificate
	certPool            *x509.CertPool
	healthzResult       string
	healthzStatusResult string
	versionResult       string
}

func testClient(t *testing.T, opt *clientOpt) {
	if !opt.insecure {
		t.Logf("Loading certificates from %s", opt.certDir)
		if err := loadCerts(opt); err != nil {
			t.Errorf("Server: Client error: %v", err)
		}
	}

	grpcConn, err := getGRPCConn(opt)
	if err != nil {
		t.Errorf("Server: Client error: %v", err)
	}
	defer grpcConn.Close()
	kubekitClient := apiv1.NewKubekitClient(grpcConn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	baseURL, httpClient := getHTTPClient(opt)

	testClientHealthz(ctx, t, grpcConn, opt.host, opt.healthPort, opt.insecure, opt.noHTTP, "v1.Kubekit", opt.healthzStatusResult)
	testClientVersion(ctx, t, kubekitClient, baseURL, httpClient, opt.noHTTP, opt.versionResult)
}

func loadCerts(opt *clientOpt) error {
	var err error
	certFile := filepath.Join(opt.certDir, opt.name+"-client.crt")
	keyFile := filepath.Join(opt.certDir, opt.name+"-client.key")
	opt.certificate, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("could not load key pair. %s", err)
	}

	caFile := filepath.Join(opt.certDir, opt.name+"-ca.crt")
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("could not read ca certificate: %s", err)
	}

	opt.certPool = x509.NewCertPool()
	if ok := opt.certPool.AppendCertsFromPEM(ca); !ok {
		return fmt.Errorf("failed to append CA certs")
	}

	return nil
}

func getGRPCConn(opt *clientOpt) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%s", opt.host, opt.grpcPort)
	if opt.insecure {
		return grpc.Dial(address, grpc.WithInsecure())
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		ServerName:   opt.host,
		Certificates: []tls.Certificate{opt.certificate},
		RootCAs:      opt.certPool,
	}))}

	return grpc.Dial(address, opts...)
}

func getHTTPClient(opt *clientOpt) (*url.URL, *http.Client) {
	client := &http.Client{}
	schema := "http"

	if !opt.insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:               opt.host,
				ClientAuth:               tls.RequireAndVerifyClientCert,
				Certificates:             []tls.Certificate{opt.certificate},
				RootCAs:                  opt.certPool,
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

	baseURL, _ := url.Parse(fmt.Sprintf("%s://%s:%s", schema, opt.host, opt.httpPort))

	return baseURL, client
}

func testClientHealthz(ctx context.Context, t *testing.T, grpcConn *grpc.ClientConn, host, port string, insecure, noHTTP bool, service, want string) {
	t.Logf("Getting gRPC Healthz")
	got, err := healthcheckGRPC(ctx, grpcConn, service)
	if err != nil {
		t.Errorf("Server: Client gRPC Healthz error: %v", err)
	}
	if got != want {
		t.Errorf("Server: Client gRPC Healthz = %v, want %v", got, want)
	}
	t.Logf("gRPC healthz: %v", got)

	if !noHTTP {
		t.Logf("Getting HTTP Healthz")
		got, err := healthcheckHTTP(t, host, port, insecure, service)
		if err != nil {
			t.Errorf("Server: Client HTTP Healthz error: %v", err)
		}
		if got != want {
			t.Errorf("Server: Client HTTP Healthz = %v, want %v", got, want)
		}
		t.Logf("HTTP healthz: %v", got)
	}
}

func healthcheckGRPC(ctx context.Context, grpcConn *grpc.ClientConn, service string) (string, error) {
	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := grpc_health_v1.NewHealthClient(grpcConn).Check(childCtx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		if stat, ok := status.FromError(err); ok && stat.Code() == codes.Unimplemented {
			return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), fmt.Errorf("this server does not implement the grpc health protocol (grpc.health.v1.Health)")
		}
		return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), fmt.Errorf("health rpc failed: %+v", err)
	}
	return resp.GetStatus().String(), nil
}

func healthcheckHTTP(t *testing.T, host, port string, insecure bool, service string) (string, error) {
	t.Logf("healthcheckHTTP host:%s , port:%s, insecure:%t, service:%s ", host, port, insecure, service)

	service = strings.Replace(service, ".", "/", -1)
	httpClient := &http.Client{}
	schema := "http"
	if !insecure {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				// There is no need to enforce security with health check, this make is slower
				InsecureSkipVerify: true,
			},
		}
		schema = "https"
	}
	healthzURL := fmt.Sprintf("%s://%s:%s/healthz/%s", schema, host, port, service)
	resp, err := httpClient.Get(healthzURL)
	if err != nil {
		return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), err
	}
	defer resp.Body.Close()
	t.Logf("response: %s", resp.Status)

	if resp.StatusCode == http.StatusOK {
		return grpc_health_v1.HealthCheckResponse_SERVING.String(), nil
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		return grpc_health_v1.HealthCheckResponse_NOT_SERVING.String(), nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), fmt.Errorf("unknown service status from HTTP status code %d (%q), Body: %q, URL: %s", resp.StatusCode, resp.Status, body, healthzURL)
}

func testClientVersion(ctx context.Context, t *testing.T, grpcClient apiv1.KubekitClient, baseURL *url.URL, httpClient *http.Client, noHTTP bool, want string) {
	t.Logf("Getting Version with gRPC")
	got, err := versionGRPC(ctx, grpcClient)
	if err != nil {
		t.Errorf("Server: Client gRPC Version error: %v", err)
	}
	if got != want {
		t.Errorf("Server: Client gRPC Version = %v, want %v", got, want)
	}
	t.Logf("gRPC version: %v", got)

	if !noHTTP {
		t.Logf("Getting Version with HTTP")
		got, err := versionHTTP(baseURL, httpClient)
		if err != nil {
			t.Errorf("Server: Client HTTP Version error: %v", err)
		}
		if got != want {
			t.Errorf("Server: Client HTTP Version = %v, want %v", got, want)
		}
		t.Logf("HTTP/REST version: %v", got)
	}
}

func versionGRPC(ctx context.Context, grpcClient apiv1.KubekitClient) (string, error) {
	reqVersion := apiv1.VersionRequest{
		Api: "v1",
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resVersion, err := grpcClient.Version(childCtx, &reqVersion)
	if err != nil {
		return "", fmt.Errorf("failed to request version. %s", err)
	}
	versionJSON, err := json.Marshal(resVersion)
	if err != nil {
		return "", fmt.Errorf("failed to marshall the received version response: %+v. %s", resVersion, err)
	}

	return string(versionJSON), nil
}

func versionHTTP(baseURL *url.URL, httpClient *http.Client) (string, error) {
	resp, err := httpClient.Get(baseURL.String() + "/api/v1/version")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	versionJSON, err := ioutil.ReadAll(resp.Body)
	return string(versionJSON), err
}
