package kluster

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/liferaft/kubekit/pkg/crypto"
	"github.com/liferaft/kubekit/pkg/provisioner"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const (
	bitSize = 4096

	privateKeyFileName = "id_rsa"
	publicKeyFileName  = "id_rsa.pub"
)

var (
	platformKeyGenExceptions = map[string]struct{}{
		"raw":    struct{}{},
		"stacki": struct{}{},
		"vra":    struct{}{},
	}
)

// HandleKeys create or load the public/private key required to provision the
// cluster nodes
func (k *Kluster) HandleKeys() error {
	pName := k.Platform()
	logPrefix := fmt.Sprintf("Certificates [ %s@%s ]", k.Name, pName)
	k.ui.SetLogPrefix(logPrefix)

	platform, ok := k.provisioner[pName]
	if !ok {
		return fmt.Errorf("not found platform named %s", pName)
	}
	if err := k.handlePrivateKey(platform); err != nil {
		return err
	}

	if err := k.handlePublicKey(platform); err != nil {
		return err
	}

	return k.Save()
}

func (k *Kluster) handlePrivateKey(platform provisioner.Provisioner) error {
	privKeyFile, privKey, requiredPrivKey := platform.GetPrivateKey()

	if len(privKey) == 0 && !requiredPrivKey {
		return nil
	}

	if len(privKey) != 0 {
		// This platform already have a private key ...
		if crypto.IsEncrypted(string(privKey)) {
			// ... and it's encrypted, so leave it as it is
			return nil
		}
		// ... it's not encrypted, it's plain text or requesting to encrypt, so encrypt it
		c, err := crypto.New(nil)
		if err != nil {
			return fmt.Errorf("generic error during encryption setup for cluster %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
		}
		encPrivKey, err := c.EncryptValue(privKey)
		if err != nil {
			return fmt.Errorf("generic error during encryption of the private key for cluster %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
		}

		// ... and finally, assign it
		platform.PrivateKey(privKeyFile, []byte(encPrivKey), privKey)
		return nil
	}

	// If this platform does not have private key, get one ...
	var err error
	privKeyFile, err = homedir.Expand(privKeyFile)
	if err != nil {
		return fmt.Errorf("failed to expand the path to the private key for cluster %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
	}

	_, err = os.Stat(privKeyFile)
	_, isException := platformKeyGenExceptions[k.Platform()]
	if err == nil {
		// ... or reading the private key file, if it exists
		privKey, err = ioutil.ReadFile(privKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read private key for %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
		}
	} else if os.IsNotExist(err) && requiredPrivKey && !isException {
		// ... either generating it, if the private key file does not exists
		privKeyFile, privKey, err = k.GenPrivKeyFile(true)
		if err != nil {
			return fmt.Errorf("failed to generate private key for %s. error: %s", platform.Name(), err)
		}
	} else if os.IsNotExist(err) && !requiredPrivKey {
		//No priv key required as we have a password
		return nil
	} else {
		return fmt.Errorf("generic error in private key for cluster %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
	}

	// private keys should not be plain in config file, so encrypt it ..
	c, err := crypto.New(nil)
	if err != nil {
		return fmt.Errorf("generic error during encryption setup for cluster %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
	}
	cryptoKey, err := c.EncryptValue(privKey)
	if err != nil {
		return fmt.Errorf("generic error during encryption of the private key for cluster %s, filename: %q. error: %s", platform.Name(), privKeyFile, err)
	}

	// ... and finally, assign it
	platform.PrivateKey(privKeyFile, []byte(cryptoKey), privKey)
	return nil
}

func (k *Kluster) handlePublicKey(platform provisioner.Provisioner) error {
	pubKeyFile, pubKey, requiredPublicKey := platform.GetPublicKey()

	if len(pubKey) != 0 {
		// This platform does not require a public key or already have one
		return nil
	}

	// If does not exists, generate it
	var err error
	pubKeyFile, err = homedir.Expand(pubKeyFile)
	if err != nil {
		return fmt.Errorf("failed to expand the path to the public key for cluster %s, filename: %q. error: %s", platform.Name(), pubKeyFile, err)
	}

	_, err = os.Stat(pubKeyFile)
	_, isException := platformKeyGenExceptions[k.Platform()]

	if os.IsNotExist(err) && requiredPublicKey && !isException {
		//we need a public key
		pubKeyFile, pubKey, err = k.GenPubKeyFile(platform)
		if err != nil {
			return fmt.Errorf("failed to generate public key for cluster %s. error: %s", platform.Name(), err)
		}
	} else if err == nil {
		// If exists, ready it
		pubKey, err = ioutil.ReadFile(pubKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read public key for cluster %s, filename: %q. error: '%s'", platform.Name(), pubKeyFile, err)
		}
	} else if os.IsNotExist(err) && !requiredPublicKey {
		//No pub key required as we have a password
		return nil
	} else {
		return fmt.Errorf("generic error in public key for cluster %s, filename: %q. error: %s", platform.Name(), pubKeyFile, err)
	}

	k.ui.Log.Debugf("creating public key file %s", pubKeyFile)
	platform.PublicKey(pubKeyFile, pubKey)
	return nil
}

func (k *Kluster) certFilePath(filename string) (string, error) {
	certDirPath, err := k.makeCertDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(certDirPath, filename), nil
}

// GenPubKeyFile generates a public key file in the certificates directory
func (k *Kluster) GenPubKeyFile(platform provisioner.Provisioner) (string, []byte, error) {
	if platform == nil {
		return "", nil, fmt.Errorf("require a platform to generate a public key")
	}
	privKeyByte, err := k.getOrGenPrivateKey(platform)
	if err != nil {
		return "", nil, err
	}

	block, _ := pem.Decode(privKeyByte)
	if block == nil {
		return "", nil, fmt.Errorf("failed to decode private key")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", nil, err
	}

	pubKey, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", nil, err
	}

	pubKeyByte := ssh.MarshalAuthorizedKey(pubKey)

	fileName, err := k.certFilePath(publicKeyFileName)
	if err != nil {
		return fileName, pubKeyByte, err
	}

	k.ui.Log.Infof("saving the TLS Public Key to file %s", fileName)

	return fileName, pubKeyByte, writeKeyFile(fileName, pubKeyByte)
}

func (k *Kluster) getOrGenPrivateKey(platform provisioner.Provisioner) ([]byte, error) {
	_, privKeyBytes, requiredPrivKey := platform.GetPrivateKey()

	var err error
	var filePrivKey string

	if len(privKeyBytes) == 0 {
		// There is no private key, generate one and save it if the platform needs it
		filePrivKey, privKeyBytes, err = k.GenPrivKeyFile(requiredPrivKey)
		if err != nil {
			return nil, err
		}
		if requiredPrivKey {
			// if needed, set it in the platform
			err = setPrivateKey(platform, filePrivKey, privKeyBytes)
		}

		return privKeyBytes, err
	}

	if crypto.IsEncrypted(string(privKeyBytes)) {
		// If it's encrypted (should be), decrypt it
		c, err := crypto.New(nil)
		if err != nil {
			return nil, err
		}
		return c.DecryptValue(string(privKeyBytes))
	}

	// If not encrypted, use it as it is
	// but set it in the platform if required
	if requiredPrivKey {
		err = setPrivateKey(platform, filePrivKey, privKeyBytes)
	}
	return privKeyBytes, err
}

// GenPrivKeyFile generates a private key file in the certificates directory
func (k *Kluster) GenPrivKeyFile(writeFile bool) (string, []byte, error) {
	key, err := genPrivKey()
	if err != nil {
		return "", nil, err
	}

	keyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if !writeFile {
		return "", keyBytes, nil
	}

	fileName, err := k.certFilePath(privateKeyFileName)
	if err != nil {
		return fileName, keyBytes, err
	}

	k.ui.Log.Infof("saving the TLS Private Key to file %s", fileName)

	return fileName, keyBytes, writeKeyFile(fileName, keyBytes)
}

func genPrivKey() (*rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}
	err = privKey.Validate()
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func writeKeyFile(filename string, key []byte) error {
	return ioutil.WriteFile(filename, key, 0600)
}

func setPrivateKey(platform provisioner.Provisioner, filename string, key []byte) error {
	c, err := crypto.New(nil)
	if err != nil {
		return err
	}

	cryptoKey, err := c.EncryptValue(key)
	if err != nil {
		return err
	}

	platform.PrivateKey(filename, []byte(cryptoKey), key)
	return nil
}
