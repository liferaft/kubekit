package tls

// GenerateCertificateAuthority generates a CA certificate and private key from
// the given options (filenames, passphrase, etc..)
func (crt *Certificate) GenerateCertificateAuthority() *Certificate {
	if crt.PrivateKey == nil {
		if err := crt.GetPrivateKey().Error(); err != nil {
			return crt.withErrf("failed to generate the private key. %s", err)
		}
	}
	crt.Opts.UseAs = CertificateUsedForCA

	return crt.GenerateCertificate()
}

// GenerateCertificateAuthority generates a CA contained in a Certificate
func GenerateCertificateAuthority(cn, certsDir string) *Certificate {
	defaultOpts := DefaultCertificateOpts(cn, certsDir)
	return NewCertificate(cn, defaultOpts).GenerateCertificateAuthority()
}
