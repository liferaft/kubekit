package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
)

func (crt *Certificate) generateTemplate() (x509.Certificate, error) {
	tmpl, err := crt.Opts.GenerateTemplate()
	if err != nil {
		return x509.Certificate{}, err
	}

	publicKey := crt.PrivateKey.Public().(*rsa.PublicKey)
	subjectKeyID, err := generateSubjectKeyID(publicKey)
	if err != nil {
		return x509.Certificate{}, err
	}
	tmpl.SubjectKeyId = subjectKeyID

	return tmpl, nil
}

// GenerateCertificate generates the certificate from the existing private key
func (crt *Certificate) GenerateCertificate() *Certificate {
	if crt.err != nil {
		return crt
	}
	if crt.PrivateKey == nil {
		return crt.withErrf("no private key created yet")
	}

	template, err := crt.generateTemplate()
	if err != nil {
		return crt.withErr(err)
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, crt.PrivateKey.Public(), crt.PrivateKey)
	if err != nil {
		return crt.withErr(err)
	}
	cert, err := x509.ParseCertificate(certDERBytes)
	if err != nil {
		return crt.withErr(err)
	}
	crt.Certificate = cert

	return crt
}

// GenerateCertificateFromPEM ?
func (crt *Certificate) GenerateCertificateFromPEM(data []byte) *Certificate {
	pemPubBlock, _ := pem.Decode(data)

	cert, err := x509.ParseCertificate(pemPubBlock.Bytes)
	if err != nil {
		return crt.withErrf("failed to parse DER encoded certificate. %s ", err)
	}
	crt.Certificate = cert

	return crt
}

// CertificatePEM returns the PEM encode of the certificate
func (crt *Certificate) CertificatePEM() []byte {
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt.Certificate.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// WriteCertificateToFile writes the certificate to the given filename
func (crt *Certificate) WriteCertificateToFile(filename string) *Certificate {
	if crt.err != nil {
		return crt
	}

	data := crt.CertificatePEM()
	if data == nil {
		return crt.withErrf("no certificate to save")
	}

	crt.err = ioutil.WriteFile(filename, data, 0600)

	return crt
}

// WriteCertificate saves the certificate to the certificate file defined in the
// options
func (crt *Certificate) WriteCertificate() *Certificate {
	return crt.WriteCertificateToFile(crt.Opts.CertificateFile)
}

// ReadCertificateFromFile generates a Certificate from the given certificate file
func (crt *Certificate) ReadCertificateFromFile(filename string) *Certificate {
	if crt.err != nil {
		return crt
	}

	if !isFileFound(filename) {
		return crt.withErrf("not found file %q", filename)
	}

	certKeyBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return crt.withErrf("failed reading the certificate file %s. %s", filename, err)
	}

	return crt.GenerateCertificateFromPEM(certKeyBytes)
}

// ReadCertificateFile loads the certificate from the certificate file defined
// in the options
func (crt *Certificate) ReadCertificateFile() *Certificate {
	return crt.ReadCertificateFromFile(crt.Opts.CertificateFile)
}

// GenerateCertificate loads the certificate from the given file if exists, otherwise
// generate it
func GenerateCertificate(cn, certsDir string) *Certificate {
	defaultOpts := DefaultCertificateOpts(cn, certsDir)
	return NewCertificate(cn, defaultOpts).GenerateCertificate()
}
