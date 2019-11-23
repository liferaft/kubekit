package crypto_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/liferaft/kubekit/pkg/crypto"
)

func TestValidKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input  string
		result bool
		err    error
	}{
		{input: "", result: false, err: nil},                    // empty string
		{input: "123As6789O1234567!8", result: false, err: nil}, // no 16, 24, neither 32 chars
		{input: "4bcd3fghi0123456", result: false, err: nil},    // no upper case
		{input: "4BCD3FGHI0123456", result: false, err: nil},    // no lowercase
		{input: "AbcdEfghIjkLmnOp", result: false, err: nil},    // no number
		{input: "ABCDEFGH!JKLMNOP", result: false, err: nil},    // no number neither lowercase
		{input: "abcdefghijklmnop", result: false, err: nil},    // no number neither uppercase
		{input: "1234567890123456", result: false, err: nil},    // no lowecase neither uppercase
		{input: "4bCd3fGhiO!23As6", result: true, err: nil},     // everything ok
	}
	for _, test := range tests {
		result := crypto.ValidKey([]byte(test.input))
		assert.Equal(t, test.result, result)
	}
}

func TestCryptoValue(t *testing.T) {
	t.Parallel()
	key := "S3cr3t_PassW0rd!"
	c, err := crypto.New([]byte(key))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, key, string(c.Key), "key should be the same")
	//S3cr3t_PassW0rd!
	//1234567890123456
	tests := []struct {
		input  string
		result string
		err    error
	}{
		{input: "secret message", result: "", err: nil}, // no action
		{input: "ENC(secret message)", result: "secret message", err: nil},
	}
	for _, test := range tests {
		t.Logf("test input: %s", test.input)
		cryptotext, err := c.Process(test.input)
		assert.IsType(t, test.err, err)
		result, err := c.Process(cryptotext)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, result)
	}
}

func TestEncryptValue(t *testing.T) {
	t.Parallel()
	key := "S3cr3t_PassW0rd!"
	os.Setenv(crypto.EnvKeyName, key)
	c, err := crypto.New(nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, key, string(c.Key), "key should be the same")

	//S3cr3t%PassW0rd!
	//1234567890123456
	tests := []struct {
		input  string
		result string
		err    error
	}{
		{input: "secret message", result: "secret message", err: nil}, // no action
		{input: "ENC(secret message)", result: "secret message", err: nil},
	}
	for _, test := range tests {
		cryptotext, err := c.EncryptValue([]byte(test.input))
		assert.IsType(t, test.err, err)
		// assert.Equal(t, test.result, cryptotext)
		result, err := c.DecryptValue(cryptotext)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, string(result))
	}
}

func TestDecryptValue(t *testing.T) {
	t.Parallel()
	key := "S3cr3t_PassW0rd!"
	c, err := crypto.New([]byte(key))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, key, string(c.Key), "key should be the same")
	// fmt.Println(c.Encrypt([]byte("secret message")))
	//S3cr3t%PassW0rd!
	//1234567890123456
	tests := []struct {
		input  string
		result string
		err    error
	}{
		{input: "02NABfyt8kUhpK4fd9h8u6jP_ikMZwjI9sg9vFyz", result: "secret message", err: nil}, // no action
		{input: "DEC(fPs_Zj1srOXMqedsKGEZlB3nhfJr4GdPun6fLMqt)", result: "secret message", err: nil},
	}
	for _, test := range tests {
		result, err := c.DecryptValue(test.input)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, string(result))
	}
}
