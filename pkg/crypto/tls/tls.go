package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// KeyPair encapsulate the Private Key and Certificate key pair
type KeyPair struct {
	Name           string
	KeyFile        string
	PrivateKey     *rsa.PrivateKey
	PrivateKeyPEM  []byte
	CN             string
	O              string
	DNSNames       []string
	IPAddresses    []string
	CertFile       string
	Certificate    *x509.Certificate
	CertificatePEM []byte
	IsCA           bool
	ExtKeyUsage    []x509.ExtKeyUsage
}

// KeyPairs is a list of KeyPair
type KeyPairs map[string]*KeyPair

// Constants used to create the certificate
const (
	Duration           = 3650 // 10 years = 3650 days
	OrganizationalUnit = "KubeKit"
	Organization       = "LifeRaft"
	Locality           = "San Diego"
	Province           = "California"
	Country            = "US"
)

// Temporal: DNS and IPs to include in the certificates
var (
	GenericDNSNames = []string{
		"localhost",
	}
	GenericIPAddresses = []string{
		"127.0.0.1",
	}
)

// EnvCAKeyPassword is the environment variable to store the password used to
// encrypt the provided CA Key file
const EnvCAKeyPassword = "KUBEKIT_CA_KEY_PASSWORD"

// NewEmptyKeyPair creates a KeyPair with everything but the key and cert
func NewEmptyKeyPair(baseCertsDir, name, cn, o string, dns, ips []string, extKeyUsage []x509.ExtKeyUsage) *KeyPair {
	keyfile := filepath.Join(baseCertsDir, name+".key")
	certfile := filepath.Join(baseCertsDir, name+".crt")
	return &KeyPair{
		Name:        name,
		KeyFile:     keyfile,
		CN:          cn,
		O:           o,
		DNSNames:    dns,
		IPAddresses: ips,
		CertFile:    certfile,
		ExtKeyUsage: extKeyUsage,
	}
}

// NewKeyPair creates a new KeyPair with the key and cert
func NewKeyPair(baseCertsDir, name, cn, o string, dns, ips []string, caKeyPair *KeyPair, extKeyUsage []x509.ExtKeyUsage) (*KeyPair, error) {
	kp := NewEmptyKeyPair(baseCertsDir, name, cn, o, dns, ips, extKeyUsage)
	err := kp.GenKeyPair(caKeyPair)
	return kp, err
}

// NewCAKeyPair creates a new CA Key Pair from the given filenames or generates
// them if the files does not exists
func NewCAKeyPair(fromCAKeyPair *KeyPair, baseCertsDir, name, cn string) (*KeyPair, error) {
	kp := NewEmptyKeyPair(baseCertsDir, name, cn, "", []string{}, []string{}, nil)
	err := kp.GenCAKeyPair(fromCAKeyPair)
	return kp, err
}

// NewFilenames updates the Key Pair filenames (key and cert) using a new base
// directory and name. If name is empty will use the current Key Pair name.
// It returns the previous filenames
func (kp *KeyPair) NewFilenames(baseCertsDir, name string) (prevKeyFile string, prevCertFile string) {
	prevKeyFile = kp.KeyFile
	prevCertFile = kp.CertFile

	if len(name) == 0 {
		name = kp.Name
	}
	kp.KeyFile = filepath.Join(baseCertsDir, name+".key")
	kp.CertFile = filepath.Join(baseCertsDir, name+".crt")

	return prevKeyFile, prevCertFile
}

// Load creates and loads the key pair from the key and cert files located in the
// given directory
func Load(baseCertsDir, name, cn string) (*KeyPair, error) {
	kp := NewEmptyKeyPair(baseCertsDir, name, cn, "", []string{}, []string{}, nil)
	if err := kp.Load(); err != nil {
		return nil, err
	}
	return kp, nil
}

// Load loads the keypair from the files in the given directory, if they exists
func (kp *KeyPair) Load() error {
	if _, err := os.Stat(kp.KeyFile); err != nil {
		return err
	}
	if _, err := os.Stat(kp.CertFile); err != nil {
		return err
	}

	// TODO: Load all the key pair parameters

	// Read Private Key file content (PrivateKeyPEM)
	privKeyBytes, err := ioutil.ReadFile(kp.KeyFile)
	if err != nil {
		return fmt.Errorf("failed reading the Key file %s. %s", kp.KeyFile, err)
	}
	kp.PrivateKeyPEM = privKeyBytes

	// Read Private Key (PrivateKey)
	pemPrivBlock, _ := pem.Decode(privKeyBytes)
	var der []byte
	if x509.IsEncryptedPEMBlock(pemPrivBlock) {
		return fmt.Errorf("The Key file %s is encrypted with a passphrase. This is not supported at this time by KubeKit", kp.KeyFile)

		// TODO: Pass the password for this private key in some environment variable TBD
		// password := os.Getenv("")
		// if password == "" {
		// 	return fmt.Errorf("The Key file %s is encrypted with a passphrase. Provide such passphrase in the environment variable TBD", kp.KeyFile)
		// }
		// der, err = x509.DecryptPEMBlock(pemPrivBlock, []byte(password))
		// if err != nil {
		// 	return err
		// }

	}

	der = pemPrivBlock.Bytes

	privKey, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		return fmt.Errorf("failed to parse DER encoded private key. %s", err)
	}
	kp.PrivateKey = privKey

	// Read Certificate or public key file content (CertificatePEM)
	certPEM, err := ioutil.ReadFile(kp.CertFile)
	if err != nil {
		return fmt.Errorf("failed reading the certificate file %s. %s", kp.CertFile, err)
	}
	kp.CertificatePEM = certPEM

	// Read the Certificate or Public Key (Certificate)
	pemPubBlock, _ := pem.Decode(certPEM)
	der = pemPubBlock.Bytes
	// pub, err := x509.ParsePKIXPublicKey(der)
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return fmt.Errorf("failed to parse DER encoded public key. %s ", err)
	}
	kp.Certificate = cert

	if len(cert.Subject.Organization) != 0 {
		kp.O = cert.Subject.Organization[0]
	}

	ipAddresses := []string{}
	for _, ip := range cert.IPAddresses {
		ipAddresses = append(ipAddresses, ip.String())
	}
	kp.IPAddresses = ipAddresses

	kp.CN = cert.Subject.CommonName
	kp.DNSNames = cert.DNSNames
	kp.IsCA = cert.IsCA
	kp.ExtKeyUsage = cert.ExtKeyUsage

	return nil
}

// GenKeyPair creates the key and cert
func (kp *KeyPair) GenKeyPair(caKeyPair *KeyPair) error {
	kp.IsCA = false

	privKey, err := GenRSAPrivateKey()
	if err != nil {
		return err
	}
	privKeyBytes := pemEncodePrivateKey(privKey)

	kp.PrivateKey = privKey
	kp.PrivateKeyPEM = privKeyBytes

	// If a CA Key Pair is not provided, then only the private key is required
	if caKeyPair == nil {
		return nil
	}

	cert, certPEM, err := SignedCert(privKey, caKeyPair, kp.CN, kp.O, kp.DNSNames, kp.IPAddresses, kp.ExtKeyUsage)
	if err != nil {
		return err
	}

	kp.Certificate = cert
	kp.CertificatePEM = certPEM

	return nil
}

// GenCAKeyPair reads the CA Key and Certificate from the given files or generates
// them if the file names are empty or does not exists
func (kp *KeyPair) GenCAKeyPair(fromCAKeyPair *KeyPair) error {
	kp.IsCA = true

	caKey, caKeyPEM, err := GenCAPrivateKey(fromCAKeyPair.KeyFile)
	if err != nil {
		return err
	}

	kp.PrivateKey = caKey
	kp.PrivateKeyPEM = caKeyPEM

	caCert, caCertPEM, err := SelfSignedCACert(fromCAKeyPair.CertFile, caKey, kp.CN)
	if err != nil {
		return err
	}

	kp.Certificate = caCert
	kp.CertificatePEM = caCertPEM

	return nil
}

// SavePrivateKey saves the private key file from this key pair
func (kp *KeyPair) SavePrivateKey(overwrite bool) error {
	if len(kp.PrivateKeyPEM) == 0 {
		return fmt.Errorf("empty private key for %s", kp.Name)
	}

	if _, err := os.Stat(kp.KeyFile); err == nil && !overwrite {
		// the private key already exists, don't overwrite
		// replace the key in memory with the one on the file system
		kp.PrivateKeyPEM, _ = ioutil.ReadFile(kp.KeyFile)
		return nil
	}

	err := ioutil.WriteFile(kp.KeyFile, kp.PrivateKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("failed to save the private key to %s. %s", kp.KeyFile, err)
	}

	return nil
}

// SaveCertificate saves the certificate file from this key pair
func (kp *KeyPair) SaveCertificate(overwrite bool) error {
	// If CertificatePEM is empty, then only the Private Key is required
	if len(kp.CertificatePEM) == 0 {
		return fmt.Errorf("empty certificate for %s", kp.Name)
		// kp.PEMEncodeCert()
	}

	if _, err := os.Stat(kp.CertFile); err == nil && !overwrite {
		// the certificate already exists, don't overwrite
		// replace the cert in memory with the one on the file system
		kp.CertificatePEM, _ = ioutil.ReadFile(kp.CertFile)
		return nil
	}

	err := ioutil.WriteFile(kp.CertFile, kp.CertificatePEM, 0600)
	if err != nil {
		return fmt.Errorf("failed to save the certificate to %s. %s", kp.CertFile, err)
	}

	return nil
}

// GenCAPrivateKey generates a CA RSA Key or returns the CA Key from the given filename.
// The file (if provided) should contain a PEM encoded CA RSA Key
func GenCAPrivateKey(filename string) (caKey *rsa.PrivateKey, caKeyBytes []byte, err error) {
	if filename == "" {
		caKey, err = GenRSAPrivateKey()
		caKeyBytes := pemEncodePrivateKey(caKey)
		return caKey, caKeyBytes, err
	}

	caKeyBytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, []byte{}, fmt.Errorf("failed reading the CA Key file %s. %s", filename, err)
	}

	pemBlock, _ := pem.Decode(caKeyBytes)
	var der []byte
	if x509.IsEncryptedPEMBlock(pemBlock) {
		password := os.Getenv("")
		if password == "" {
			return nil, caKeyBytes, fmt.Errorf("The CA Key file %s is encrypted with a passphrase. Provide such passphrase in the environment variable %q", filename, EnvCAKeyPassword)
		}
		der, err = x509.DecryptPEMBlock(pemBlock, []byte(password))
		if err != nil {
			return nil, caKeyBytes, err
		}
	} else {
		der = pemBlock.Bytes
	}

	caKey, err = x509.ParsePKCS1PrivateKey(der)
	return caKey, caKeyBytes, err
}

// GenRSAPrivateKey generates a RSA private key
func GenRSAPrivateKey() (key *rsa.PrivateKey, err error) {
	key, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("failed generating the RSA Key. %s", err)
	}
	return key, err
}

// SelfSignedCACert generates a CA x509 Certificate from a CA RSA Key or returns the CA
// x509 Certificate from the given filename. The file (if provided) should
// contain a PEM encoded CA x509 Certificate
func SelfSignedCACert(filename string, caKey *rsa.PrivateKey, cn string) (*x509.Certificate, []byte, error) {
	if filename == "" {
		caCert, err := NewSelfSignedCACert(caKey, cn)
		caCertBytes := pemEncodeCert(caCert)
		return caCert, caCertBytes, err
	}

	caCertBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, []byte{}, fmt.Errorf("failed reading or finding CA Certificate file %s. %s", filename, err)
	}

	// TODO: Validate this Cert came from caKey

	pemBlock, _ := pem.Decode(caCertBytes)
	caCert, err := x509.ParseCertificate(pemBlock.Bytes)
	return caCert, caCertBytes, err
}

// NewSelfSignedCACert creates a Self Signed CA Certificate with a given
// CA Private Key and a Common Name
func NewSelfSignedCACert(caKey *rsa.PrivateKey, cn string) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number to generate the CA x509 Certificate: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: cn,
			// OrganizationalUnit: []string{OrganizationalUnit},
			// Organization:       []string{Organization},
			// Locality:           []string{Locality},
			// Province:           []string{Province},
			// Country:            []string{Country},
		},
		NotBefore: time.Now().UTC(),
		NotAfter:  time.Now().Add(Duration * 24 * time.Hour).UTC(),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, caKey.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(derBytes)
}

// SignedCert generates a Self Signed Certificate and returns also the pem
// decoded bytes
func SignedCert(privKey *rsa.PrivateKey, caKeyPair *KeyPair, cn, o string, dns, ips []string, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, []byte, error) {
	cert, err := NewSignedCert(privKey, caKeyPair, cn, o, dns, ips, extKeyUsage)
	if err != nil {
		return nil, []byte{}, err
	}

	certBytes := pemEncodeCert(cert)

	return cert, certBytes, nil
}

// NewSignedCert creates a Self Signed Certificate with a given private key,
// the CA key pair (key and cert) and a Common Name
func NewSignedCert(privKey *rsa.PrivateKey, caKeyPair *KeyPair, cn, o string, dns, ips []string, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number to generate the CA x509 Certificate: %s", err)
	}

	// fmt.Printf("Creating a signed certificate with:\n\t\tCN = %s\n\t\tDNS: %v\n\t\tIP: %v\n", cn, dns, ips)

	ipAddresses := []net.IP{}
	for _, ip := range ips {
		ipAddresses = append(ipAddresses, net.ParseIP(ip))
	}

	var organizations []string
	if len(o) != 0 {
		organizations = []string{o}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: organizations, //   []string{Organization},
			// OrganizationalUnit: []string{OrganizationalUnit},
			// Locality:           []string{Locality},
			// Province:           []string{Province},
			// Country:            []string{Country},
		},
		NotBefore:             caKeyPair.Certificate.NotBefore,
		NotAfter:              time.Now().Add(Duration * 24 * time.Hour).UTC(),
		DNSNames:              dns,
		IPAddresses:           ipAddresses,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caKeyPair.Certificate, privKey.Public(), caKeyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(derBytes)
}

// PEMEncodeCert returns the PEM encode of the given certificate
func (kp *KeyPair) PEMEncodeCert() []byte {
	kp.CertificatePEM = pemEncodeCert(kp.Certificate)
	return kp.CertificatePEM
}

func pemEncodeCert(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// PEMEncodePrivateKey returns the PEM encode of the given private key
func (kp *KeyPair) PEMEncodePrivateKey() []byte {
	kp.PrivateKeyPEM = pemEncodePrivateKey(kp.PrivateKey)
	return kp.PrivateKeyPEM
}

func pemEncodePrivateKey(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

// Save saves all the private keys and certificates (if not empty) for all the
// Key Pairs
func (kps KeyPairs) Save(overwrite bool) error {
	for name := range kps {
		if err := kps[name].SavePrivateKey(overwrite); err != nil {
			if os.IsExist(err) {
				continue
			}
			return err
		}

		// A Key Pairs may only have a Private Key. If so, the cert is empty and can't be saved
		if len(kps[name].CertificatePEM) > 0 {
			kps[name].SaveCertificate(overwrite)
		}
	}

	return nil
}
