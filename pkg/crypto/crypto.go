package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Here we provide some tools to deal with encryption. All the methods here can
// and should be improved to provide more security, here are just the simplest
// and secure methods

// EnvKeyName is the name of the environment variable to store the key
// KeyValidLen is the valid length for a key. A valid key should have an equal or larger length than KeyValidLen
const (
	EnvKeyName           = "KUBEKIT_KEY"
	KeyValidRules        = "A key should have 16, 24 or 32 characters, should contain at least one number and at least one uppercase and lowercase letter"
	DefaultKeyPassphrase = "S3cr3t%PassW0rd!"
)

// Action defines an action like encrypt or decrypt
type Action int

// Defines all the possible CryptoActions
const (
	ActionDec Action = 1 << iota
	ActionEnc
	ActionUnknown
)

// String returns the text to identify a CryptoAction
func (a Action) String() string {
	switch a {
	case ActionDec:
		return "DEC"
	case ActionEnc:
		return "ENC"
	default:
		return ""
	}
}

// ParseAction returns the action to execute with the given text
func ParseAction(text string) Action {
	re, _ := regexp.Compile(ActionDec.String() + `\((.*)\)`)
	if re.MatchString(text) {
		return ActionDec
	}
	re, _ = regexp.Compile(ActionEnc.String() + `\((.*)\)`)
	if re.MatchString(text) {
		return ActionEnc
	}
	return ActionUnknown
}

// IsEncrypted returns true if the given text has the action to decrypt, returns
// true if it's encrypted
func IsEncrypted(text string) bool {
	return ParseAction(text) == ActionDec
}

// IsDecrypted returns true if the given text has the action to encrypt, returns
// true if it's decrypted/plaintext
func IsDecrypted(text string) bool {
	return ParseAction(text) == ActionEnc
}

// Crypto is a struct to handle cryptography (encryption/decryption)
type Crypto struct {
	Key []byte
}

// KeyError is the error for an invalid key
type KeyError []byte

func (ke KeyError) Error() string {
	return fmt.Sprintf("invalid key %q. %s", string(ke), KeyValidRules)
}

// New creates a crypto struct to manage cryptography
func New(key []byte) (*Crypto, error) {
	var err error
	if len(key) == 0 {
		key, err = getKey()
		if err != nil {
			return nil, err
		}
	}
	if !ValidKey(key) {
		return nil, KeyError(key)
	}

	return &Crypto{
		Key: key,
	}, nil
}

// getKey get the passphrase or key from the environment variables defined on
// 'EnvKeyName' constant
func getKey() ([]byte, error) {
	keyStr := os.Getenv(EnvKeyName)
	if len(keyStr) == 0 {
		// return nil, fmt.Errorf("not found passphrase to generate provate key in the environment variable %s", EnvKeyName)
		keyStr = DefaultKeyPassphrase
	}
	return []byte(keyStr), nil
}

func alphaRange(first byte) string {
	p := make([]byte, 26)
	for i := range p {
		p[i] = first + byte(i)
	}
	return string(p)
}

// ValidKey returns true if this is a valid key
func ValidKey(key []byte) bool {
	l := len(key)
	switch l {
	default:
		return false
	case 16, 24, 32:
		break
	}

	lowerCaseChars := alphaRange('a')
	upperCaseChars := alphaRange('A')
	switch {
	case !strings.ContainsAny(string(key), "1234567890"):
		return false
	case !strings.ContainsAny(string(key), lowerCaseChars):
		return false
	case !strings.ContainsAny(string(key), upperCaseChars):
		return false
	default:
		return true
	}
}

// Encrypt encrypts a plain text using the given key
func (c *Crypto) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(c.Key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an encrypted text using the given key
func (c *Crypto) Decrypt(cryptotext string) ([]byte, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptotext)

	block, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext length (%d) it's too short, should be larger than or equal to %d", len(ciphertext), aes.BlockSize)
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// EncryptValue encrypts a plain text using the given key and returns it in the
// configuration format
func (c *Crypto) EncryptValue(plaintext []byte) (string, error) {
	re, _ := regexp.Compile(ActionEnc.String() + `\((.*)\)`)
	result := re.FindStringSubmatch(string(plaintext))
	if len(result) == 2 {
		plaintext = []byte(result[1])
	}

	cryptotext, err := c.Encrypt(plaintext)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s(%s)", ActionDec, cryptotext), nil
}

// DecryptValue decrypts an encrypted text using the given key and returns it in
// the configuration format
func (c *Crypto) DecryptValue(cryptotext string) ([]byte, error) {
	re, _ := regexp.Compile(ActionDec.String() + `\((.*)\)`)
	result := re.FindStringSubmatch(cryptotext)
	if len(result) == 2 {
		cryptotext = result[1]
	}

	return c.Decrypt(cryptotext)
	// plaintext, err := c.Decrypt(cryptotext)
	// if err != nil {
	// 	return nil, err
	// }
	// return []byte(fmt.Sprintf("%s(%s)", ActionEnc, string(plaintext))), nil
}

// Process either encrypt or decryt based on the action/function as text preffix
func (c *Crypto) Process(text string) (string, error) {
	if ParseAction(text) == ActionDec {
		plaintext, err := c.DecryptValue(text)
		return string(plaintext), err
	}
	if ParseAction(text) == ActionEnc {
		return c.EncryptValue([]byte(text))
	}
	return "", nil
}
