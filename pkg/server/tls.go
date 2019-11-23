package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	servertls "github.com/liferaft/kubekit/pkg/server/tls"
)

const (
	defCAFilename = "ca.crt"
	defPubKeyExt  = ".crt"
	defPrivKeyExt = ".key"
)

// EnvTLSServerPassword is the name of the environment variable to store/pass the
// password of the CA key
const EnvTLSServerPassword = "KUBEKIT_SERVER_TLS_PASS"

// SetTLS enable TLS with the given certificates if and only if the first
// parmeter (enabled) is `true`
func (s *Server) SetTLS(enabled bool, certDir, pubKeyFile, privKeyFile, caFile string) *Server {
	if enabled {
		return s.WithTLS(certDir, pubKeyFile, privKeyFile, caFile)
	}
	return s
}

// WithTLS enable and configure TLS to the server
func (s *Server) WithTLS(certDir, pubKeyFile, privKeyFile, caFile string) *Server {
	s.insecure = false

	// If certs can be loaded, do not generate them
	err := s.LoadCertificates(certDir, pubKeyFile, privKeyFile, caFile)
	if err == nil {
		return s
	}
	s.ui.Log.Warnf("the certificates will be generated because %s", err)

	// An error means some of the certs was not found or was incorrectly generated. So, generated them
	if err := s.GenerateCertificates(certDir, pubKeyFile, privKeyFile, caFile); err != nil {
		return s.addErr(err)
	}

	return s
}

// LoadCertificates assign the server certificates loading the given files or
// from the certificates directory. Th certificates files (server key, server
// cert & CA cert) are created by the user.
func (s *Server) LoadCertificates(certDir, pubKeyFile, privKeyFile, caFile string) (err error) {
	if s.insecure {
		return nil
	}

	if !existPath(certDir) {
		return fmt.Errorf("nof found or empty certificates directory (%s), it is required to load the keys or store the generated keys", certDir)
	}

	// Generate the certificates with `openssl` and get the certificates from the files
	pubKeyFile = validFile(pubKeyFile, certDir, s.Name+defPubKeyExt)    // a.k.a cert file
	privKeyFile = validFile(privKeyFile, certDir, s.Name+defPrivKeyExt) // a.k.a key file
	caFile = validFile(caFile, certDir, s.Name+"-"+defCAFilename)

	if len(pubKeyFile) == 0 && len(privKeyFile) == 0 {
		return fmt.Errorf("none of the cert files were not found. Provide the cert files location or put them in the certs directory %q", certDir)
	}
	if len(pubKeyFile) == 0 {
		return fmt.Errorf("the public key file was not found. Provide the cert files location or put them in the certs directory %q", certDir)
	}
	if len(privKeyFile) == 0 {
		return fmt.Errorf("the private key file was not found. Provide the cert files location or put them in the certs directory %q", certDir)
	}

	// Generate the certificate (KeyPair) and CertPool
	if s.certificate, err = tls.LoadX509KeyPair(pubKeyFile, privKeyFile); err != nil {
		return err
	}

	s.certPool = x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}

	if ok := s.certPool.AppendCertsFromPEM(ca); !ok {
		return fmt.Errorf("failed to append CA certs")
	}

	return nil
}

// GenerateCertificates assign the server certificates either by loading the given files,
// from the certificates directory or generating them from a self-signed
// generated CA certificate
func (s *Server) GenerateCertificates(certDir, pubKeyFile, privKeyFile, caFile string) (err error) {
	if s.insecure {
		return nil
	}

	var ca *servertls.Certificate
	if passphrase := os.Getenv(EnvTLSServerPassword); passphrase != "" {
		defaultOpts := servertls.DefaultCertificateOpts(s.Name+"-ca", certDir)
		defaultOpts.Passphrase = passphrase
		ca = servertls.NewCertificate(s.Name+"-ca", defaultOpts).GenerateCertificateAuthority()
	} else {
		ca = servertls.GenerateCertificateAuthority(s.Name+"-ca", certDir)
	}
	s.ui.Log.Infof("saving the CA certificates to %s and %s", ca.Opts.PrivateKeyFile, ca.Opts.CertificateFile) // s.caKeyPair.CertFile, s.caKeyPair.KeyFile)
	if err := ca.Persist().Error(); err != nil {
		return fmt.Errorf("failed to generate the CA certificate. %s", err)
	}

	// Append the CA Cert into the cert pool
	if s.certPool == nil {
		s.certPool = x509.NewCertPool()
	}
	if ok := s.certPool.AppendCertsFromPEM(ca.CertificatePEM()); !ok {
		return fmt.Errorf("failed to append CA cert")
	}
	s.caCert = ca

	serverCert := servertls.GenerateSignedCertificate(s.Name, certDir, ca, s.host).Persist()
	s.ui.Log.Infof("saving the server certificates to %s and %s", serverCert.Opts.PrivateKeyFile, serverCert.Opts.CertificateFile) //s.keyPair.CertFile, s.keyPair.KeyFile)
	if err := serverCert.Error(); err != nil {
		return fmt.Errorf("failed to generate the server certificate. %s", err)
	}
	s.serverCert = serverCert

	if s.certificate, err = tls.X509KeyPair(serverCert.CertificatePEM(), serverCert.PrivateKeyPEM()); err != nil {
		return err
	}

	clientCert := servertls.GenerateSignedClientCertificate(s.Name+"-client", certDir, ca, s.host).Persist()
	s.ui.Log.Infof("saving the client certificates to %s and %s", clientCert.Opts.PrivateKeyFile, clientCert.Opts.CertificateFile) //s.keyPair.CertFile, s.keyPair.KeyFile)
	if err := clientCert.Error(); err != nil {
		return fmt.Errorf("failed to generate the server certificate. %s", err)
	}

	return nil
}

func existPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func validFile(path, certDir, filename string) string {
	if !existPath(path) {
		if existPath(filepath.Join(certDir, filename)) {
			return filepath.Join(certDir, filename)
		}
	}
	// if dosen't exists the given file nor the file in certs dir, it will be generated
	return ""
}
