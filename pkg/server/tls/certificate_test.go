package tls

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestCertificate_GenerateCertificateAuthority(t *testing.T) {
	emptyKP := NewCertificate("empty", nil)
	// emptyKPwithErr := NewCertificate().withErrf("some error here")

	tests := []struct {
		name    string
		crt     *Certificate
		wantErr bool
	}{
		{"generate RSA", emptyKP, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.crt.GenerateCertificateAuthority()

			if err := tt.crt.Error(); (err != nil) != tt.wantErr {
				t.Errorf("Certificate.GenerateCertificateAuthority() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if bits := tt.crt.PrivateKey.N.BitLen(); bits != defaultRSABits {
				t.Errorf("Certificate.GenerateCertificateAuthority() failed, key too short. expected = %v, received = %v", defaultRSABits, bits)
			}
			if err := tt.crt.PrivateKey.Validate(); err != nil {
				t.Errorf("Certificate.GenerateCertificateAuthority() failed, it's not valid: %v", err)
			}
		})
	}
}

func TestCertificate_GenerateCertificate(t *testing.T) {
	tests := []struct {
		name       string
		kpName     string
		bits       int
		privKeyPEM []byte
		passphrase string
		wantErr    bool
	}{
		{"generate 1024 RSA", "ca", 1024, nil, "", false},
		{"generate 2048 RSA", "ca", 2048, nil, "", false},
		{"generate 3072 RSA", "ca", 3072, nil, "", false},
		{"generate 4096 RSA", "ca", 4096, nil, "", false},
		{"generate 5120 RSA", "ca", 5120, nil, "", false},
		{"generate 6144 RSA", "ca", 6144, nil, "", false},
		//{"generate 7168 RSA", "ca", 7168, nil, "", false},
		//{"generate 8192 RSA", "ca", 8192, nil, "", false},
		//		{"generate 15360 RSA", "ca", 15360, nil, "", false},
		{"from encrypted 1024 RSA", "ca", 1024, mustLoad(t, 1024), "Test1ng", false},
		{"from encrypted 2048 RSA", "ca", 2048, mustLoad(t, 2048), "Test1ng", false},
		{"from encrypted 3072 RSA", "ca", 3072, mustLoad(t, 3072), "Test1ng", false},
		{"from encrypted 4096 RSA", "ca", 4096, mustLoad(t, 4096), "Test1ng", false},
		{"from encrypted 5120 RSA", "ca", 5120, mustLoad(t, 5120), "Test1ng", false},
		//{"from encrypted 7168 RSA", "ca", 7168, mustLoad(t, 7168), "Test1ng", false},
		//{"from encrypted 8192 RSA", "ca", 8192, mustLoad(t, 8192), "Test1ng", false},
		//		{"from encrypted 15360 RSA", "ca", 15360, mustLoad(t, 15360), "Test1ng", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crt := NewCertificate(tt.kpName, nil)
			t.Logf("Testing bits: %d", tt.bits)
			if tt.privKeyPEM == nil {
				crt.WithBits(tt.bits).GenerateRSAKey()
			} else {
				crt.WithPassphrase(tt.passphrase).GeneratePrivateKeyFromPEM(tt.privKeyPEM)
			}

			if err := crt.GenerateCertificate().Error(); (err != nil) != tt.wantErr {
				t.Errorf("Certificate.GenerateCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				validateCertificate(t, crt, tt.bits, "GenerateCertificate")
				// certPEM := tt.crt.CertificatePEM()
			}
		})
	}
}

func BenchmarkCertificate_GenerateCertificate(b *testing.B) {
	tests := []struct {
		name       string
		kpName     string
		bits       int
		privKeyPEM []byte
		passphrase string
		wantErr    bool
	}{
		{"generate 1024 RSA", "ca", 1024, nil, "Test1ng", false},
		{"generate 2048 RSA", "ca", 2048, nil, "Test1ng", false},
		{"generate 3072 RSA", "ca", 3072, nil, "Test1ng", false},
		{"generate 4096 RSA", "ca", 4096, nil, "Test1ng", false},
		{"generate 5120 RSA", "ca", 5120, nil, "Test1ng", false},
		{"generate 6144 RSA", "ca", 6144, nil, "Test1ng", false},
		{"generate 7168 RSA", "ca", 7168, nil, "Test1ng", false},
		{"generate 8192 RSA", "ca", 8192, nil, "Test1ng", false},
		{"generate 15360 RSA", "ca", 15360, nil, "Test1ng", false},
		{"from encrypted 1024 RSA", "ca", 1024, mustLoad(b, 1024), "Test1ng", false},
		{"from encrypted 2048 RSA", "ca", 2048, mustLoad(b, 2048), "Test1ng", false},
		{"from encrypted 3072 RSA", "ca", 3072, mustLoad(b, 3072), "Test1ng", false},
		{"from encrypted 4096 RSA", "ca", 4096, mustLoad(b, 4096), "Test1ng", false},
		{"from encrypted 5120 RSA", "ca", 5120, mustLoad(b, 5120), "Test1ng", false},
		{"from encrypted 7168 RSA", "ca", 7168, mustLoad(b, 7168), "Test1ng", false},
		{"from encrypted 8192 RSA", "ca", 8192, mustLoad(b, 8192), "Test1ng", false},
		//		{"from encrypted 15360 RSA", "ca", 15360, mustLoad(b, 15360), "Test1ng", false},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			crt := NewCertificate(tt.kpName, nil)
			if tt.privKeyPEM == nil {
				crt.WithBits(tt.bits).GenerateRSAKey()
			} else {
				crt.WithPassphrase(tt.passphrase).GeneratePrivateKeyFromPEM(tt.privKeyPEM)
			}

			if err := crt.GenerateCertificate().Error(); (err != nil) != tt.wantErr {
				b.Errorf("Certificate.GenerateCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mustLoad(tb testing.TB, bits int) []byte {
	filename := fmt.Sprintf("testdata/test%d.key", bits)
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		tb.Fatalf("unable to open %s : %s", filename, err)
	}
	return out
}
