package tls_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kubekit/kubekit/pkg/server/tls"
)

func Example() {
	certsDir := "./certificates"
	// Generate the certificates in memory for the CN `kubekit.io`
	ca := tls.GenerateCertificateAuthority("kubekit.io", certsDir)
	// Save them to ./certificates/kubekit.io.{key,crt}
	if err := ca.Persist().Error(); err != nil {
		log.Fatalf("failed to generate the CA certificate. %s", err)
	}

	// Generate a server certificate for the server on `www.server.io` and
	// `192.168.100.10` in memory, sign it with the CA generated and save them
	// to ./certificates/server.{key,crt}
	serverCert := tls.GenerateSignedCertificate("server", certsDir, ca, "www.server.io", "192.168.100.10").Persist()
	if err := serverCert.Error(); err != nil {
		log.Fatalf("failed to generate the server certificate. %s", err)
	}

	// Generate a client certificate for `localhost` in memory, sign it with the CA
	// generated and save them to ./certificates/client.{key,crt}
	clientCert := tls.GenerateSignedClientCertificate("client", certsDir, ca, "www.server.io", "192.168.100.10").Persist()
	if err := clientCert.Error(); err != nil {
		log.Fatalf("failed to generate the server certificate. %s", err)
	}

	// Generate a self signed sertificate in memory and save it to
	// ./certificates/sscert.{key,crt}
	// 	selfSignCert := tls.GenerateSelfSignedCertificate("sscert", certsDir).Persist()
	// 	if err := selfSignCert.Error(); err != nil {
	// 		log.Fatalf("failed to generate the self signed certificate. %s", err)
	// 	}
}

func Example_options() {
	// Loads the existing key, generates the CA certificate from it and save the
	// generated certificate to `./certificates/ca.crt`
	caOpts := tls.CertificateOpts{
		CertsDir: "./certificates",
	}
	ca := tls.NewCertificate("corp.org", &caOpts).ReadPrivateKeyFromFile("/var/pki/ca.key").GenerateCertificateAuthority().WriteCertificateToFile("./certificates/ca.crt")
	if err := ca.Error(); err != nil {
		log.Fatalf("filed to generate the CA certificate. %s", err)
	}

	// Generates and save the server certificate from an existing key to be used on the following domain names
	serverOpts := tls.CertificateOpts{
		CertsDir: "./certificates",
		Bits:     1024,
		DNSNames: []string{"www.server.com", "www.server.io"},
	}
	serverCert := tls.NewCertificate("server", &serverOpts).GenerateCertificate().Persist()
	if err := serverCert.Error(); err != nil {
		log.Fatalf("filed to generate the server certificate. %s", err)
	}

	clientOpts := tls.CertificateOpts{
		CertsDir: "./certificates",
		Bits:     2048,
		UseAs:    tls.CertificateUsedForClient,
	}
	clientCert := tls.NewCertificate("client", &clientOpts).GenerateCertificate().Persist()
	if err := clientCert.Error(); err != nil {
		log.Fatalf("filed to generate the client certificate. %s", err)
	}
}

func ExampleGenerateCertificateAuthority() {
	kpCA := tls.GenerateCertificateAuthority("ca", "/tmp/example").Persist()

	if err := kpCA.Error(); err != nil {
		log.Fatalf("failed to save the CA key/cert files. %s", err)
	}
}

func ExampleCertificate_GenerateCertificateAuthority_privateKeyFromPEM() {
	// generate this CA PEM with:
	// openssl genrsa -des3 -passout pass:Test1ng 1024 2>/dev/null
	caPEM := []byte(`-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,32DE7AED71E4F727

ZqmUAYaFOmSbT/yr8rvHI3hQVQ9oqtHyV5mLNfYYRkOCwYgo3ZbBnIMNTPHMrsb7
8X8YAkpan/g0K0ftN7j0ajXkOlDN6BBOKqDoXPBu+pxNixgGpC63wJSAu+undzW8
Oigj/Vlv//yM5X699UNawGMD++geOAyvwO5QFX4OvvnCi+eCmCuAp28j6xHk8BMV
MiXznogl2KLUBYDINKajBwIU0n2todxJwDWmADaCK29hDg26pMFrDrVE67LzyHMH
NZaKjWaVwLzfW1Xf8vuO4vNMJjlTvhLGuMX+O5fln1nt9Af/HsIqod2BGvZKJGGn
HMxuj0I/rKJE/brpYW0dOENKKMfeoeZ58Zv1xTffOj78NAMwceNyIWViVR/iu/6p
7eFR/e/Ia2iH1oKyBKaGhIuoGx+Sz7bwD2OmKGnFS8km+Pq2V3uQnWFvhSMIZGd+
BOvU9re9H++j9aMQPUQCU0lH6AaZlB4wdPNqC+KJ3NoSz77tYLwvkfCFnn9uksuF
fZl/RUsK8zAGqULeT49IFG96Jqbv3rgeRLbmnD/QFtUUBnVX3gsGgBxdzW9zQ3/1
vg6Q1XW1LL3I+86gCCRimmMRSyeSsximN7jLzJqwJf0WYFrvl55uO9GqvVfGnWIF
M3PuQprkT/NUkZZ0AzuZRnHL42fuLuIIoeQg+TaC/wiMpWQLT3x1Mi2p50RN9Bjg
fuFAZNrlvx4IzbUxttFlaX7pfdBvPm+7T8DY9Rm6+IUI1Sdq09q/jtaXGynJM1tM
t4gsOW4mBarSMTyglY7cW1lN6nFyyuhT+bzqRHhu8DxvZulFSpYhgA==
-----END RSA PRIVATE KEY-----`)

	kpCA := tls.NewCertificate("example", nil).GeneratePrivateKeyFromPEM(caPEM).GenerateCertificateAuthority()

	tmpCACertFile, _ := ioutil.TempFile("", "example")
	caCertFilename := tmpCACertFile.Name() + ".crt"
	defer os.Remove(caCertFilename)

	if err := kpCA.WriteCertificateToFile(caCertFilename).Error(); err != nil {
		log.Fatalf("failed to save the CA cert file to %q. %s", caCertFilename, err)
	}
}

func ExampleCertificate_GenerateCertificateAuthority_readPrivateKeyFromFile() {
	// generate the CA key with:
	// openssl genrsa -des3 -passout pass:Test1ng -out /tmp/example/ca.key 1024
	kpCA := tls.NewCertificate("example", nil).WithPassphrase("Test1ng").ReadPrivateKeyFromFile("/tmp/example/ca.key").GenerateCertificateAuthority()

	certPEMBytes := kpCA.CertificatePEM()
	fmt.Println(string(certPEMBytes))
}

func ExampleCertificate_GenerateCertificateAuthority_withBits() {
	kpCA := tls.NewCertificate("example", nil).WithBits(2048).GenerateCertificateAuthority()
	if err := kpCA.Error(); err != nil {
		log.Fatalf("failed to generate the CA with 2046 bits. %s", err)
	}

	tmpCAKeyFile, _ := ioutil.TempFile("", "example")
	caKeyFilename := tmpCAKeyFile.Name() + ".key"
	defer os.Remove(caKeyFilename)
	tmpCACertFile, _ := ioutil.TempFile("", "example")
	caCertFilename := tmpCACertFile.Name() + ".crt"
	defer os.Remove(caCertFilename)

	kpCA.WriteToFiles(caKeyFilename, caCertFilename)
	if err := kpCA.Error(); err != nil {
		log.Fatalf("failed to save the CA key/cert files to %q and %q. %s", caKeyFilename, caCertFilename, err)
	}
}
