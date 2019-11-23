package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultRSABits            = 4096
	defaultDuration           = 365 // 1 year
	defaultOrganizationalUnit = ""  // "KubeKit"
	defaultOrganization       = ""  // "LifeRaft"
	defaultLocality           = ""  // "San Diego"
	defaultProvince           = ""  // "California"
	defaultCountry            = ""  // "US"
	defaultPassphrase         = ""
)

var defaultCertificateOpts = CertificateOpts{
	Bits:               defaultRSABits,
	Passphrase:         defaultPassphrase,
	Organization:       defaultOrganization,
	OrganizationalUnit: defaultOrganizationalUnit,
	Locality:           defaultLocality,
	Province:           defaultProvince,
	Country:            defaultCountry,
	Duration:           defaultDuration,
}

// CertificateUse is to specify what's the use of the certificate
type CertificateUse int

// CA, Server, Client are the different uses for a certificate
const (
	CertificateUsedForCA     CertificateUse = 1 << iota // 0001 = 1
	CertificateUsedForClient                            // 0010 = 2
)

// CertificateOpts contain the options to create the Certificate struct
type CertificateOpts struct {
	CommonName         string
	CertsDir           string
	CertificateFile    string
	PrivateKeyFile     string
	Passphrase         string
	Bits               int
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
	Country            string
	Duration           time.Duration
	DNSNames           []string
	IPAddresses        []string
	URIs               []string
	UseAs              CertificateUse
}

// DefaultCertificateOpts creates a CertificateOpts with default values
func DefaultCertificateOpts(cn, certsDir string) *CertificateOpts {
	if cn == "" {
		return nil
	}

	// Works as a copy, defOpts is not a reference to defaultCertificateOpts
	defOpts := defaultCertificateOpts

	opts := &defOpts
	opts.CommonName = cn

	opts.WithPath(certsDir)

	return opts
}

// WithPath sets the certificate directory and assign the private key file and
// certificate file if not valid
func (opts *CertificateOpts) WithPath(certsDir string) *CertificateOpts {
	if certsDir == "" {
		return opts
	}

	opts.CertsDir = certsDir

	if len(opts.PrivateKeyFile) == 0 {
		opts.PrivateKeyFile = opts.defaultPrivateKeyFilename()
	}

	if len(opts.CertificateFile) == 0 {
		opts.CertificateFile = opts.defaultCertificateFilename()
	}

	return opts
}

func (opts *CertificateOpts) defaultPrivateKeyFilename() string {
	return filepath.Join(opts.CertsDir, opts.CommonName+".key")
}
func (opts *CertificateOpts) defaultCertificateFilename() string {
	return filepath.Join(opts.CertsDir, opts.CommonName+".crt")
}

// defaultFilenames set the default filenames for Private Key and Certificate if
// they are not set
func (opts *CertificateOpts) defaultFilenames() error {
	if len(opts.PrivateKeyFile) == 0 {
		if len(opts.CertsDir) == 0 {
			return fmt.Errorf("the Private Key filename cannot be made, it is not defined in the options neither the certificates directory")
		}
		opts.PrivateKeyFile = opts.defaultPrivateKeyFilename()
	}
	if len(opts.CertificateFile) == 0 {
		if len(opts.CertsDir) == 0 {
			return fmt.Errorf("the Certificate filename cannot be made, it is not defined in the options neither the certificates directory")
		}
		opts.CertificateFile = opts.defaultCertificateFilename()
	}
	return nil
}

func isFileFound(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// IsCertFileFound returns true if the certificate file is found or accesible
func (opts *CertificateOpts) IsCertFileFound() bool {
	return isFileFound(opts.CertificateFile)
}

// IsKeyFileFound returns true if the private key file is found or accesible
func (opts *CertificateOpts) IsKeyFileFound() bool {
	return isFileFound(opts.PrivateKeyFile)
}

// AsCA changes the use of the Certificate to CA
func (opts *CertificateOpts) AsCA() *CertificateOpts {
	opts.UseAs = CertificateUsedForCA
	return opts
}

// AsServer changes the use of the Certificate to server
func (opts *CertificateOpts) AsServer() *CertificateOpts {
	opts.UseAs = CertificateUse(0)
	return opts
}

// AsClient changes the use of the Certificate to client
func (opts *CertificateOpts) AsClient() *CertificateOpts {
	opts.UseAs = CertificateUsedForClient
	return opts
}

// GenerateSubject generates the subject (pkix.Name) required to generate
// certificate templates
func (opts *CertificateOpts) GenerateSubject() pkix.Name {
	subject := pkix.Name{
		CommonName: opts.CommonName,
	}

	if len(opts.OrganizationalUnit) != 0 {
		subject.OrganizationalUnit = []string{opts.OrganizationalUnit}
	}
	if len(opts.Organization) != 0 {
		subject.Organization = []string{opts.Organization}
	}
	if len(opts.Locality) != 0 {
		subject.Locality = []string{opts.Locality}
	}
	if len(opts.Province) != 0 {
		subject.Province = []string{opts.Province}
	}
	if len(opts.Country) != 0 {
		subject.Country = []string{opts.Country}
	}

	return subject
}

// GenerateTemplate generates a x509 Certificate template from the given options
func (opts *CertificateOpts) GenerateTemplate() (x509.Certificate, error) {
	if len(opts.CommonName) == 0 {
		return x509.Certificate{}, fmt.Errorf("CommonName is required to generate a certificate")
	}

	subject := opts.GenerateSubject()

	if opts.Duration == 0 {
		opts.Duration = defaultDuration
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return x509.Certificate{}, err
	}

	ipAddresses := generateIPs(opts.IPAddresses)
	uriList := generateURLs(opts.URIs)

	// TODO: Include extensions
	// ext, err := opts.GenerateSubjectAltNameExtension()
	// if err != nil {
	// 	return x509.Certificate{}, err
	// }

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now().UTC(),
		NotAfter:     time.Now().Add(opts.Duration * 24 * time.Hour).UTC(),
		DNSNames:     opts.DNSNames,
		IPAddresses:  ipAddresses,
		URIs:         uriList,
		// ExtraExtensions: []pkix.Extension{*ext},
	}

	switch opts.UseAs {
	case CertificateUsedForCA:
		template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		template.ExtKeyUsage = nil
		template.IsCA = true
		template.MaxPathLenZero = true
		template.BasicConstraintsValid = true
	case CertificateUsedForClient:
		template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	default:
		template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}

	return template, nil
}

func generateSubjectKeyID(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := asn1.Marshal(rsa.PublicKey{
		N: publicKey.N,
		E: publicKey.E,
	})
	if err != nil {
		return nil, err
	}

	subjectKeyID := sha1.Sum(publicKeyBytes)

	return subjectKeyID[:], nil
}

// GenerateSubjectAltNameExtension generates the SAN extension for the template certificate
func (opts *CertificateOpts) GenerateSubjectAltNameExtension() (*pkix.Extension, error) {
	rawValues := []asn1.RawValue{}

	// See https://tools.ietf.org/html/rfc5280#appendix-A.2
	for _, ip := range opts.IPAddresses {
		rawValues = append(rawValues, asn1.RawValue{
			Bytes: []byte(ip),
			Class: asn1.ClassContextSpecific,
			Tag:   7,
		})
	}

	for _, uri := range opts.URIs {
		rawValues = append(rawValues, asn1.RawValue{
			Bytes: []byte(uri),
			Class: asn1.ClassContextSpecific,
			Tag:   6,
		})
	}

	for _, dn := range opts.DNSNames {
		rawValues = append(rawValues, asn1.RawValue{
			Bytes: []byte(dn),
			Class: asn1.ClassContextSpecific,
			Tag:   2,
		})
	}

	encValues, err := asn1.Marshal(rawValues)
	if err != nil {
		return nil, err
	}

	return &pkix.Extension{
		Id:       asn1.ObjectIdentifier{2, 5, 29, 17}, // http://www.alvestrand.no/objectid/2.5.29.17.html
		Critical: true,
		Value:    encValues,
	}, nil
}

// generateIPs generates a list of IPs from a list of IPs as string
func generateIPs(ips []string) []net.IP {
	ipAddresses := []net.IP{}
	for _, ip := range ips {
		ipAddresses = append(ipAddresses, net.ParseIP(ip))
	}
	return ipAddresses
}

// generateURLs generates a list of URLs from a list of URLs as string
func generateURLs(urlList []string) []*url.URL {
	urls := []*url.URL{}
	for _, u := range urlList {
		parsedURL, err := url.Parse(u)
		if err != nil || parsedURL == nil {
			continue
		}
		urls = append(urls, parsedURL)
	}
	return urls
}

func generateAltNames(altNames ...string) (ipAddresses []net.IP, domainNames []string, urlList []*url.URL) {
	ipAddresses = []net.IP{}
	domainNames = []string{}
	urlList = []*url.URL{}
	for _, an := range altNames {
		if ip := net.ParseIP(an); ip != nil {
			ipAddresses = append(ipAddresses, ip)
			continue
		}
		if url, err := url.Parse(an); err == nil && url != nil && url.IsAbs() {
			urlList = append(urlList, url)
			continue
		}
		domainNames = append(domainNames, an)
	}
	return ipAddresses, domainNames, urlList
}

func generateStringAltNames(altNames ...string) (ipAddresses []string, domainNames []string, urlList []string) {
	ipAddresses = []string{}
	domainNames = []string{}
	urlList = []string{}
	for _, an := range altNames {
		if ip := net.ParseIP(an); ip != nil {
			ipAddresses = append(ipAddresses, ip.String())
			continue
		}
		if url, err := url.Parse(an); err == nil && url != nil && url.IsAbs() {
			urlList = append(urlList, url.String())
			continue
		}
		domainNames = append(domainNames, an)
	}
	return ipAddresses, domainNames, urlList
}

// GenerateCSRTemplate generates a x509 CertificateRequest template from the given options
func (opts *CertificateOpts) GenerateCSRTemplate() (x509.CertificateRequest, error) {
	if len(opts.CommonName) == 0 {
		return x509.CertificateRequest{}, fmt.Errorf("CommonName is required to generate a certificate sign request")
	}

	subject := opts.GenerateSubject()

	ipAddresses := generateIPs(opts.IPAddresses)
	uriList := generateURLs(opts.URIs)

	if opts.Duration == 0 {
		opts.Duration = defaultDuration
	}

	tmpl := x509.CertificateRequest{
		Subject:     subject,
		IPAddresses: ipAddresses,
		DNSNames:    opts.DNSNames,
		URIs:        uriList,
	}

	return tmpl, nil
}
