package tls

import (
	"testing"
)

func validateCertificate(t *testing.T, crt *Certificate, size int, funcName string) {
	if crt == nil {
		t.Errorf("Certificate.%s() failed, validating a non existing Certificate (nil)", funcName)
	}

	if crt.PrivateKey == nil {
		t.Errorf("Certificate.%s() failed, the Private Key does not exists (nil)", funcName)
	}

	if err := crt.PrivateKey.Validate(); err != nil {
		t.Errorf("Certificate.%s() failed, it's not valid: %v", funcName, err)
	}

	if size == 0 {
		size = defaultRSABits
	}

	if bits := crt.PrivateKey.N.BitLen(); bits != size {
		t.Errorf("Certificate.%s() failed, key too short. expected = %v, received = %v", funcName, size, bits)
	}
}
