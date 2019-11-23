// +build integration

package provisioner_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"testing"

	"github.com/johandry/log"
	"github.com/kraken/ui"
	"github.com/kubekit/kubekit/pkg/provisioner"
	"github.com/kubekit/kubekit/pkg/provisioner/openstack"
	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v2"
)

var (
	destroy = flag.Bool("destroy", true, "should provisioned resources be destroyed?")
)

func TestProvisionerIntegration(t *testing.T) {
	var (
		tUI = ui.New(false, log.NewDefault())
	)

	clusterName, ok := os.LookupEnv("OPENSTACK_CLUSTER_NAME")
	if !ok {
		clusterName = "test-cluster"
	}

	platforms := provisioner.SupportedPlatforms(
		clusterName,
		map[string]string{
			"openstack_domain_name": "Default",
			"openstack_net_name":    "kubekit-net",
			"openstack_private_key": "~/.ssh/id_rsa",
			"openstack_public_key":  "~/.ssh/id_rsa.pub",
		},
		tUI,
	)

	for _, p := range []struct {
		name  string
		creds []string
	}{
		{
			name: "openstack",
			creds: []string{
				os.Getenv("OPENSTACK_AUTH_URL"),
				os.Getenv("OPENSTACK_USER_NAME"),
				os.Getenv("OPENSTACK_PASSWORD"),
			},
		},
	} {
		t.Run(p.name, func(t *testing.T) {
			platform, ok := platforms[p.name]
			if !ok {
				t.Fatalf("unable to find platform %s", p.name)
			}

			mapConfig := map[interface{}]interface{}{}

			defaultConfig := platform.Config()

			t.Logf("Default config: %+v", defaultConfig)

			osCfg, ok := defaultConfig.(*openstack.Config)
			if !ok {
				t.Fatalf("expected openstack config but got: %#v", osCfg)
			}

			// Customize platform config
			privKey, pubKey := sshRSAKeyPair(t)
			osCfg.PublicKey = pubKeySSHStr(t, pubKey)
			osCfg.PrivateKey = privKeyPEMStr(t, privKey)
			osCfg.OpenstackTenantName = "kubekit"

			data, err := yaml.Marshal(defaultConfig)
			if err != nil {
				t.Fatalf("unable to marshal config: %s", err)
			}

			err = yaml.Unmarshal(data, mapConfig)
			if err != nil {
				t.Fatalf("unable to reconstitute YAML config: %s", err)
			}

			prov, err := provisioner.NewPlatform(p.name, clusterName, mapConfig, p.creds, tUI)

			err = prov.BeProvisioner(nil)
			if err != nil {
				t.Fatalf("unable to initialize provisioner: %s", err)
			}

			if *destroy {
				defer func() {
					t.Log("destroying provisioned resources")
					err := prov.Apply(true)
					if err != nil {
						t.Fatalf("unable to destroy provisioned resources: %s", err)
					}
				}()
			}

			err = prov.Apply(false)
			if err != nil {
				t.Fatalf("unable to provision openstack: %s", err)
			}

			// TODO: provide hook script to run extensive checks against cluster
		})
	}
}

// mustGetEnv will panic if the environment variable contains the empty string
func mustGetEnv(t *testing.T, envVar string) {
	val := os.Getenv(envVar)
	if val == "" {
		t.Fatalf("Required environment variable was empty: %s", envVar)
	}
}

func sshRSAKeyPair(t *testing.T) (*rsa.PrivateKey, ssh.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("unable to generate RSA key: %s", err)
	}

	// generate and write public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("unable to create public key from private key: %s", err)
	}

	return privateKey, publicKey
}

func privKeyPEMStr(t *testing.T, privateKey *rsa.PrivateKey) string {
	privBuf := new(bytes.Buffer)

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(privBuf, privateKeyPEM); err != nil {
		t.Fatalf("unable to encode private key PEM block: %s", err)
	}

	return privBuf.String()
}

func pubKeySSHStr(t *testing.T, publicKey ssh.PublicKey) string {
	return string(ssh.MarshalAuthorizedKey(publicKey))
}
