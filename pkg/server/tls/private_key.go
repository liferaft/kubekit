package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
)

// GenerateRSAKey generates an RSA Private Key for the given size in the options
func (crt *Certificate) GenerateRSAKey() *Certificate {
	if crt.err != nil {
		return crt
	}

	privKey, err := rsa.GenerateKey(rand.Reader, crt.Opts.Bits)
	if err != nil {
		return crt.withErrf("failed generating the RSA Key. %s", err)
	}
	crt.PrivateKey = privKey

	// pubKey := privKey.Public()
	// crt.PublicKey = &pubKey

	return crt
}

// GeneratePrivateKeyFromPEM generates a Key Pair from the given private key PEM data
func (crt *Certificate) GeneratePrivateKeyFromPEM(data []byte) *Certificate {
	if crt.err != nil {
		return crt
	}

	pemPrivBlock, _ := pem.Decode(data)
	if pemPrivBlock == nil {
		return crt.withErrf("cannot find the next PEM formatted block")
	}
	der := pemPrivBlock.Bytes
	if x509.IsEncryptedPEMBlock(pemPrivBlock) {
		var err error
		if der, err = x509.DecryptPEMBlock(pemPrivBlock, []byte(crt.Opts.Passphrase)); err != nil {
			return crt.withErr(err)
		}
	}

	if pemPrivBlock.Type != "RSA PRIVATE KEY" {
		return crt.withErrf("unmatched type (%q)", pemPrivBlock.Type)
	}

	privKey, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		return crt.withErrf("failed to parse DER encoded private key. %s", err)
	}
	crt.PrivateKey = privKey

	// pubKey := privKey.Public()
	// crt.PublicKey = &pubKey

	return crt
}

// PrivateKeyPEM returns the private key in PEM format. It will be encrypted if
// the passphrase is defined in the options
func (crt *Certificate) PrivateKeyPEM() []byte {
	if crt.PrivateKey == nil {
		return nil
	}

	privateKeyPEMBytes := x509.MarshalPKCS1PrivateKey(crt.PrivateKey)
	var block *pem.Block

	if len(crt.Opts.Passphrase) != 0 {
		var err error
		block, err = x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", privateKeyPEMBytes, []byte(crt.Opts.Passphrase), x509.PEMCipher3DES)
		if err != nil {
			crt.err = err
			return nil
		}
	} else {
		block = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyPEMBytes,
		}
	}

	return pem.EncodeToMemory(block)
}

// WritePrivateKeyToFile saves the private key to the given file
func (crt *Certificate) WritePrivateKeyToFile(filename string) *Certificate {
	if crt.err != nil {
		return crt
	}

	data := crt.PrivateKeyPEM()
	if data == nil {
		if crt.err != nil {
			return crt
		}
		return crt.withErrf("no private key to save")
	}

	crt.err = ioutil.WriteFile(filename, data, 0600)

	return crt
}

// WritePrivateKeyFile saves the private key to the key file defined in the
// options
func (crt *Certificate) WritePrivateKeyFile() *Certificate {
	return crt.WriteCertificateToFile(crt.Opts.PrivateKeyFile)
}

// ReadPrivateKeyFromFile generates a Key Pair from the given private key file
func (crt *Certificate) ReadPrivateKeyFromFile(filename string) *Certificate {
	if crt.err != nil {
		return crt
	}

	privKeyBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return crt.withErrf("failed reading the private key file %s. %s", filename, err)
	}

	return crt.GeneratePrivateKeyFromPEM(privKeyBytes)
}

// ReadPrivateKeyFile loads the private key from the key file defined
// in the options
func (crt *Certificate) ReadPrivateKeyFile() *Certificate {
	return crt.ReadPrivateKeyFromFile(crt.Opts.PrivateKeyFile)
}

// GetPrivateKey loads the private key from the given file if exists,
// otherwise generates a RSA key
func (crt *Certificate) GetPrivateKey() *Certificate {
	if crt.Opts.IsKeyFileFound() {
		return crt.ReadPrivateKeyFile()
	}
	return crt.GenerateRSAKey()
}
