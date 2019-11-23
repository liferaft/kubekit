package manifest_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kubekit/kubekit/pkg/manifest"
)

// ManifestFilename is the name of the Manifest file
const ManifestFilename = "MANIFEST"

// Consider Tab is 2 spaces
const Tab2Space = "  "

func TestVersionIsInKubeManifest(t *testing.T) {
	if _, ok := manifest.KubeManifest.Releases[manifest.Version]; !ok {
		t.Errorf("version %s is not in the Kube Manifest yet", manifest.Version)
	}
}

func TestManifestYaml(t *testing.T) {
	actualYamlBytes, err := manifestDataTest.Yaml()
	if err != nil {
		t.Errorf("failed to get manifest in YAML format. %s", err)
	}
	expectedYaml := strings.TrimPrefix(string(manifestYamlDataTest), "\n")
	expectedYaml = strings.TrimSuffix(expectedYaml, "\n")
	expectedYaml = strings.Replace(expectedYaml, "\t", "  ", -1)

	actualYaml := strings.TrimPrefix(string(actualYamlBytes), "\n")
	actualYaml = strings.TrimSuffix(actualYaml, "\n")
	actualYaml = strings.Replace(actualYaml, "\t", "  ", -1)

	if actualYaml != expectedYaml {
		// fmt.Printf("Expected:\n%q\n", expectedYaml)
		// fmt.Printf("Actual:\n%q\n", string(actualYaml))
		t.Error("failed to get the expected YAML manifest")
	}
}

func TestManifestJson(t *testing.T) {
	actualJson, err := manifestDataTest.JSON()
	if err != nil {
		t.Errorf("failed to get manifest in JSON format. %s", err)
	}

	var testData bytes.Buffer
	json.Compact(&testData, manifestJsonDataTest)

	if string(actualJson) != testData.String() {
		t.Error("failed to get the expected JSON manifest")
	}
}

func TestManifestWriteFile(t *testing.T) {
	manifest.WriteFile(ManifestFilename, "")

	fi, err := os.Stat(ManifestFilename)
	if os.IsNotExist(err) {
		t.Errorf("manifest file %q was not created", ManifestFilename)
	}
	if fi.Size() == 0 {
		t.Errorf("manifest file %q is empty", ManifestFilename)
	}
}

var manifestDataTest = manifest.Manifest{
	Releases: map[string]manifest.Release{
		"1.1.1": manifest.Release{
			PreviousVersion:   "1.1.0",
			KubernetesVersion: "1.9.1",
			DockerVersion:     "17.09.1-ce",
			EtcdVersion:       "3.3.9",
			Dependencies: manifest.Dependencies{
				ControlPlane: map[string]manifest.Dependency{
					"etcd": manifest.Dependency{
						Version:      "v3.3.9",
						Name:         "etcd",
						Src:          "quay.io/coreos/etcd:v3.3.9",
						PrebakePath:  "/opt/liferaft/kubekit-control-plane/quay.io/coreos/etcd-v3.3.9.tar.xz",
						Checksum:     "ea49a3d44a50a50770bff84eab87bac2542c7171254c4d84c609b8c66aefc211",
						ChecksumType: "sha256",
						LicenseURL:   "https://github.com/etcd-io/etcd/blob/master/LICENSE",
					},
				},
				Core: map[string]manifest.Dependency{
					"hyperkube": manifest.Dependency{
						Version:      "v1.9.1",
						Name:         "hyperkube",
						Src:          "quay.io/coreos/hyperkube:v1.9.1_coreos.0",
						PrebakePath:  "/opt/liferaft/kubekit-core/quay.io/coreos/hyperkube-v1.9.1_coreos.0.tar.xz",
						Checksum:     "9ed9046760a30d1cafbf9c7dec431428ce4838ddde62eefd7fafb268aa902e2d",
						ChecksumType: "sha256",
					},
					"healthz": manifest.Dependency{
						Version:      "1.2",
						Name:         "exechealthz",
						Src:          "gcr.io/google_containers/exechealthz-amd64:1.2",
						PrebakePath:  "/opt/liferaft/kubekit-core/gcr.io/google_containers/exechealthz-amd64-1.2.tar.xz",
						Checksum:     "503e158c3f65ed7399f54010571c7c977ade7fe59010695f48d9650d83488c0a",
						ChecksumType: "sha256",
					},
				},
			},
		},
		"1.1.0": manifest.Release{
			PreviousVersion:   "1.0.0",
			KubernetesVersion: "1.9.1",
			DockerVersion:     "17.09.1-ce",
			EtcdVersion:       "3.3.5",
			Dependencies: manifest.Dependencies{
				Core: map[string]manifest.Dependency{
					"dnsmasq-metrics": manifest.Dependency{
						Version:      "1.0.1",
						Name:         "dnsmasq-metrics",
						Src:          "gcr.io/google_containers/dnsmasq-metrics-amd64:1.0.1",
						PrebakePath:  "/opt/liferaft/kubekit-core/gcr.io/google_containers/dnsmasq-metrics-amd64-1.0.1.tar.xz",
						Checksum:     "6453b3ee4f5455657133ada25c858ba695d2f90db69f3e8e69b3d9a2f6392a66",
						ChecksumType: "sha256",
					},
					"pause": manifest.Dependency{
						Version:      "1.0",
						Name:         "pause",
						Src:          "gcr.io/google_containers/pause",
						PrebakePath:  "/opt/liferaft/kubekit-core/gcr.io/google_containers/pause-1.0.tar.xz",
						Checksum:     "a78c2d6208eff9b672de43f880093100050983047b7b0afe0217d3656e1b0d5f",
						ChecksumType: "sha256",
					},
				},
			},
		},
	},
}

var manifestYamlDataTest = []byte(`releases:
  1.1.0:
    previous-version: 1.0.0
		kubernetes-version: 1.9.1
		docker-version: 17.09.1-ce
    etcd-version: 3.3.5
    dependencies:
      core:
        dnsmasq-metrics:
          version: 1.0.1
          name: dnsmasq-metrics
          src: gcr.io/google_containers/dnsmasq-metrics-amd64:1.0.1
          prebake-path: /opt/liferaft/kubekit-core/gcr.io/google_containers/dnsmasq-metrics-amd64-1.0.1.tar.xz
          checksum: 6453b3ee4f5455657133ada25c858ba695d2f90db69f3e8e69b3d9a2f6392a66
          checksum_type: sha256
          license_url: ""
        pause:
          version: "1.0"
          name: pause
          src: gcr.io/google_containers/pause
          prebake-path: /opt/liferaft/kubekit-core/gcr.io/google_containers/pause-1.0.tar.xz
          checksum: a78c2d6208eff9b672de43f880093100050983047b7b0afe0217d3656e1b0d5f
          checksum_type: sha256
          license_url: ""
  1.1.1:
    previous-version: 1.1.0
		kubernetes-version: 1.9.1
		docker-version: 17.09.1-ce
    etcd-version: 3.3.9
    dependencies:
      control-plane:
        etcd:
          version: v3.3.9
          name: etcd
          src: quay.io/coreos/etcd:v3.3.9
          prebake-path: /opt/liferaft/kubekit-control-plane/quay.io/coreos/etcd-v3.3.9.tar.xz
          checksum: ea49a3d44a50a50770bff84eab87bac2542c7171254c4d84c609b8c66aefc211
          checksum_type: sha256
          license_url: https://github.com/etcd-io/etcd/blob/master/LICENSE
      core:
        healthz:
          version: "1.2"
          name: exechealthz
          src: gcr.io/google_containers/exechealthz-amd64:1.2
          prebake-path: /opt/liferaft/kubekit-core/gcr.io/google_containers/exechealthz-amd64-1.2.tar.xz
          checksum: 503e158c3f65ed7399f54010571c7c977ade7fe59010695f48d9650d83488c0a
          checksum_type: sha256
          license_url: ""
        hyperkube:
          version: v1.9.1
          name: hyperkube
          src: quay.io/coreos/hyperkube:v1.9.1_coreos.0
          prebake-path: /opt/liferaft/kubekit-core/quay.io/coreos/hyperkube-v1.9.1_coreos.0.tar.xz
          checksum: 9ed9046760a30d1cafbf9c7dec431428ce4838ddde62eefd7fafb268aa902e2d
          checksum_type: sha256
          license_url: ""
`)

var manifestJsonDataTest = []byte(`{
  "releases": {
    "1.1.0": {
      "previous-version": "1.0.0",
      "kubernetes-version": "1.9.1",
      "docker-version": "17.09.1-ce",
      "etcd-version": "3.3.5",
      "dependencies": {
        "core": {
          "dnsmasq-metrics": {
            "version": "1.0.1",
            "name": "dnsmasq-metrics",
            "src": "gcr.io/google_containers/dnsmasq-metrics-amd64:1.0.1",
            "prebake-path": "/opt/liferaft/kubekit-core/gcr.io/google_containers/dnsmasq-metrics-amd64-1.0.1.tar.xz",
            "checksum": "6453b3ee4f5455657133ada25c858ba695d2f90db69f3e8e69b3d9a2f6392a66",
            "checksum_type": "sha256",
            "license_url": ""
          },
          "pause": {
            "version": "1.0",
            "name": "pause",
            "src": "gcr.io/google_containers/pause",
            "prebake-path": "/opt/liferaft/kubekit-core/gcr.io/google_containers/pause-1.0.tar.xz",
            "checksum": "a78c2d6208eff9b672de43f880093100050983047b7b0afe0217d3656e1b0d5f",
            "checksum_type": "sha256",
            "license_url": ""
          }
        }
      }
    },
    "1.1.1": {
      "previous-version": "1.1.0",
      "kubernetes-version": "1.9.1",
      "docker-version": "17.09.1-ce",
      "etcd-version": "3.3.9",
      "dependencies": {
        "control-plane": {
            "etcd": {
                "version": "v3.3.9",
                "name": "etcd",
                "src": "quay.io/coreos/etcd:v3.3.9",
                "prebake-path": "/opt/liferaft/kubekit-control-plane/quay.io/coreos/etcd-v3.3.9.tar.xz",
                "checksum": "ea49a3d44a50a50770bff84eab87bac2542c7171254c4d84c609b8c66aefc211",
                "checksum_type": "sha256",
                "license_url": "https://github.com/etcd-io/etcd/blob/master/LICENSE"
            }
        },
        "core": {
          "healthz": {
            "version": "1.2",
            "name": "exechealthz",
            "src": "gcr.io/google_containers/exechealthz-amd64:1.2",
            "prebake-path": "/opt/liferaft/kubekit-core/gcr.io/google_containers/exechealthz-amd64-1.2.tar.xz",
            "checksum": "503e158c3f65ed7399f54010571c7c977ade7fe59010695f48d9650d83488c0a",
            "checksum_type": "sha256",
            "license_url": ""
          },
          "hyperkube": {
            "version": "v1.9.1",
            "name": "hyperkube",
            "src": "quay.io/coreos/hyperkube:v1.9.1_coreos.0",
            "prebake-path": "/opt/liferaft/kubekit-core/quay.io/coreos/hyperkube-v1.9.1_coreos.0.tar.xz",
            "checksum": "9ed9046760a30d1cafbf9c7dec431428ce4838ddde62eefd7fafb268aa902e2d",
            "checksum_type": "sha256",
            "license_url": ""
          }
        }
      }
    }
  }
}`)
