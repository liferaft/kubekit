package tls

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
)

// Certificate is a X.509 Key Pair or Certificate that includes the Private Key,
// Certificate and other useful information about them.
type Certificate struct {
	PrivateKey         *rsa.PrivateKey // Private Key may be included/repeated inside Certificate
	Certificate        *x509.Certificate
	CertificateRequest *x509.CertificateRequest
	err                error
	Opts               *CertificateOpts

	// These fiels may be needed:
	// PublicKey   crypto.PublicKey			// PublicKey is the public key for the private key, do not confuse with the certificate
}

// NewCertificate creates an empty Certificate
func NewCertificate(cn string, opts *CertificateOpts) *Certificate {
	if opts == nil {
		opts = DefaultCertificateOpts(cn, "")
	}
	opts.CommonName = cn
	return &Certificate{
		Opts: opts,
	}
}

// WithPath append the certificates diretory into the Certificate options
func (crt *Certificate) WithPath(certsDir string) *Certificate {
	crt.Opts.WithPath(certsDir)
	return crt
}

// WithBits assign the number of bits to generate the RSA Private Key. If zero
// sets the default value
func (crt *Certificate) WithBits(bits int) *Certificate {
	if bits == 0 {
		bits = defaultRSABits
	}
	crt.Opts.Bits = bits
	return crt
}

// WithPassphrase assigns the given password to the Key Pair that is used to
// open encrypted key files
func (crt *Certificate) WithPassphrase(passphrase string) *Certificate {
	crt.Opts.Passphrase = passphrase
	return crt
}

// Error returns the latest occured error
func (crt *Certificate) Error() error {
	return crt.err
}

// withErrf set the error to the Key Pair and return it. Useful for error handling
func (crt *Certificate) withErrf(format string, a ...interface{}) *Certificate {
	if len(format) == 0 {
		return crt
	}
	crt.err = fmt.Errorf(format, a...)
	return crt
}

// withErr set the error to the Key Pair and return it. Useful for error handling
func (crt *Certificate) withErr(err error) *Certificate {
	crt.err = err
	return crt
}

// WriteToFiles save the Private Key and Certificate to the given files
func (crt *Certificate) WriteToFiles(keyFilename, certFilename string) *Certificate {
	return crt.WritePrivateKeyToFile(keyFilename).WriteCertificateToFile(certFilename)
}

// ReadFromFile reads the Private Key and Certificate to the given files
func (crt *Certificate) ReadFromFile(keyFilename, certFilename string) *Certificate {
	return crt.ReadPrivateKeyFromFile(keyFilename).ReadCertificateFromFile(certFilename)
}

// Persist make the keys persistents, store them into the filenames defined in
// the options or in the certificates directory
func (crt *Certificate) Persist() *Certificate {
	// set the deault filenames if needed
	if err := crt.Opts.defaultFilenames(); err != nil {
		return crt.withErr(err)
	}

	return crt.WriteToFiles(crt.Opts.PrivateKeyFile, crt.Opts.CertificateFile)
}

// Load loads the Key Pair from the files in the opts
func (crt *Certificate) Load() *Certificate {
	// set the deault filenames if needed
	if err := crt.Opts.defaultFilenames(); err != nil {
		return crt.withErr(err)
	}

	return crt.ReadFromFile(crt.Opts.PrivateKeyFile, crt.Opts.CertificateFile)
}

// LoadFromDir loads the private key and certificate from the given certificates
// directory to create a Certificate. The filenames should be `certsDir/name.key`
// and `certsDir/name.crt`
func LoadFromDir(cn, certsDir string) *Certificate {
	return NewCertificate(cn, nil).WithPath(certsDir).Load()
}

// Load loads the private key and certificate from the given files to create a
// Certificate
func Load(cn, keyFilename, certFilename string) *Certificate {
	crt := NewCertificate(cn, nil)
	crt.Opts.PrivateKeyFile = keyFilename
	crt.Opts.CertificateFile = certFilename
	return crt.Load()
}
