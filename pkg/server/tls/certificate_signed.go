package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// GenerateCertificateSigningRequest generates a CSR to be used to sign the
// certificate with a CA
func (crt *Certificate) GenerateCertificateSigningRequest() *Certificate {
	if crt.err != nil {
		return crt
	}
	if crt.PrivateKey == nil {
		return crt.withErrf("no private key created yet")
	}

	template, err := crt.Opts.GenerateCSRTemplate()
	if err != nil {
		return crt.withErr(err)
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, crt.PrivateKey)
	if err != nil {
		return crt.withErr(err)
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return crt.withErr(err)
	}
	crt.CertificateRequest = csr

	return crt
}

// GenerateCertificateSigningRequest generates a CSR to be used to sign the
// certificate with a CA
func GenerateCertificateSigningRequest(cn, certDir string, altNames ...string) *Certificate {
	certOpts := DefaultCertificateOpts(cn, certDir)

	ipAddresses, domainNames, urlList := generateStringAltNames(altNames...)
	certOpts.IPAddresses = ipAddresses
	certOpts.DNSNames = domainNames
	certOpts.URIs = urlList

	return NewCertificate(cn, certOpts).GenerateCertificateSigningRequest()
}

// CertificateSigningRequestPEM returns the PEM encode of the CSR
func (crt *Certificate) CertificateSigningRequestPEM() []byte {
	block := pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: crt.CertificateRequest.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// generateTemplateFromCSR generates a certificate template from a CSR to
// generate a signed certificate
func (crt *Certificate) generateTemplateFromCSR(ca *Certificate) (x509.Certificate, error) {
	template, err := crt.Opts.GenerateTemplate()
	if err != nil {
		return x509.Certificate{}, err
	}

	publicKey := crt.PrivateKey.Public().(*rsa.PublicKey)
	subjectKeyID, err := generateSubjectKeyID(publicKey)
	if err != nil {
		return x509.Certificate{}, err
	}
	template.SubjectKeyId = subjectKeyID

	template.RawSubject = crt.CertificateRequest.RawSubject
	template.NotAfter = ca.Certificate.NotAfter

	return template, nil
}

// Sign signs the current certificate with the given CA certificate
func (crt *Certificate) Sign(ca *Certificate) *Certificate {
	if crt.PrivateKey == nil {
		return crt.withErrf("no private key created yet")
	}

	if crt.CertificateRequest == nil {
		return crt.withErrf("no certificate request signing (CSR) created yet")
	}

	if ca.Certificate == nil {
		return crt.withErrf("the Certificate Autority %s does not contain a certificate", ca.Opts.CommonName)
	}

	if !ca.Certificate.IsCA {
		return crt.withErrf("the certificate %s is not a Certificate Authority", ca.Opts.CommonName)
	}

	if ca.PrivateKey == nil {
		return crt.withErrf("the Certificate Autority %s does not contain a key", ca.Opts.CommonName)
	}

	template, err := crt.generateTemplateFromCSR(ca)
	if err != nil {
		return crt.withErr(err)
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &template, ca.Certificate, crt.CertificateRequest.PublicKey, ca.PrivateKey)
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

// GenerateSignedCertificate create and signs a certificate with the given CA certificate
func (crt *Certificate) GenerateSignedCertificate(ca *Certificate) *Certificate {
	return crt.GetPrivateKey().GenerateCertificateSigningRequest().Sign(ca)
}

// GenerateSignedCertificate create and signs a certificate with the given CA certificate
func GenerateSignedCertificate(cn, certsDir string, ca *Certificate, altNames ...string) *Certificate {
	certOpts := DefaultCertificateOpts(cn, certsDir)

	ipAddresses, domainNames, urlList := generateStringAltNames(altNames...)
	certOpts.IPAddresses = ipAddresses
	certOpts.DNSNames = domainNames
	certOpts.URIs = urlList

	return NewCertificate(cn, certOpts).GenerateSignedCertificate(ca)
}

// GenerateSignedClientCertificate create and signs a certificate to be used by
// a client with the given CA certificate
func GenerateSignedClientCertificate(cn, certsDir string, ca *Certificate, altNames ...string) *Certificate {
	certOpts := DefaultCertificateOpts(cn, certsDir)

	ipAddresses, domainNames, urlList := generateStringAltNames(altNames...)
	certOpts.IPAddresses = ipAddresses
	certOpts.DNSNames = domainNames
	certOpts.URIs = urlList

	certOpts.UseAs = CertificateUsedForClient

	return NewCertificate(cn, certOpts).GenerateSignedCertificate(ca)
}
