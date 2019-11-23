package tls

import (
	"reflect"
	"testing"
)

func TestDefaultCertificateOpts(t *testing.T) {
	type args struct {
		name     string
		certsDir string
	}
	tests := []struct {
		name string
		args args
		want *CertificateOpts
	}{
		{"no name", args{"", ""}, nil},
		{"just name", args{"name", ""}, &CertificateOpts{
			CommonName:         "name",
			Bits:               defaultRSABits,
			Passphrase:         defaultPassphrase,
			Organization:       defaultOrganization,
			OrganizationalUnit: defaultOrganizationalUnit,
			Duration:           defaultDuration,
			Locality:           defaultLocality,
			Province:           defaultProvince,
			Country:            defaultCountry,
		}},
		{"on current path", args{"name", "."}, &CertificateOpts{
			CommonName:         "name",
			CertsDir:           ".",
			PrivateKeyFile:     "name.key",
			CertificateFile:    "name.crt",
			Bits:               defaultRSABits,
			Passphrase:         defaultPassphrase,
			Organization:       defaultOrganization,
			OrganizationalUnit: defaultOrganizationalUnit,
			Duration:           defaultDuration,
			Locality:           defaultLocality,
			Province:           defaultProvince,
			Country:            defaultCountry,
		}},
		{"on some path", args{"name", "/some/path"}, &CertificateOpts{
			CommonName:         "name",
			CertsDir:           "/some/path",
			PrivateKeyFile:     "/some/path/name.key",
			CertificateFile:    "/some/path/name.crt",
			Bits:               defaultRSABits,
			Passphrase:         defaultPassphrase,
			Organization:       defaultOrganization,
			OrganizationalUnit: defaultOrganizationalUnit,
			Duration:           defaultDuration,
			Locality:           defaultLocality,
			Province:           defaultProvince,
			Country:            defaultCountry,
		}},
		{"as Server", args{"server", "."}, &CertificateOpts{
			CommonName:         "server",
			CertsDir:           ".",
			PrivateKeyFile:     "server.key",
			CertificateFile:    "server.crt",
			Bits:               defaultRSABits,
			Passphrase:         defaultPassphrase,
			Organization:       defaultOrganization,
			OrganizationalUnit: defaultOrganizationalUnit,
			Duration:           defaultDuration,
			Locality:           defaultLocality,
			Province:           defaultProvince,
			Country:            defaultCountry,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DefaultCertificateOpts(tt.args.name, tt.args.certsDir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultCertificateOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}
