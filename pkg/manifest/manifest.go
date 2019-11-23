package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// DEFAULTFORMAT is the default format to write the Manifest. Could be Yaml or Json
const DEFAULTFORMAT = "yaml"

// Manifest is the main type that encapsulate a manifest
type Manifest struct {
	Releases map[string]Release `json:"releases" yaml:"releases" mapstructure:"releases"`
}

// KubeManifest is a global variable that defines the KubekitOS Manifest and all
// the Releases in it
var KubeManifest = Manifest{
	Releases: map[string]Release{
		"0.1.0": release,
	},
}

// Release contain a Kubernetes version and dependencies
type Release struct {
	PreviousVersion   string       `json:"previous-version" yaml:"previous-version" mapstructure:"previous-version"`
	KubernetesVersion string       `json:"kubernetes-version" yaml:"kubernetes-version" mapstructure:"kubernetes-version"`
	DockerVersion     string       `json:"docker-version" yaml:"docker-version" mapstructure:"docker-version"`
	EtcdVersion       string       `json:"etcd-version" yaml:"etcd-version" mapstructure:"etcd-version"`
	Dependencies      Dependencies `json:"dependencies" yaml:"dependencies" mapstructure:"dependencies"`
}

// Dependencies has 1 kind of dependencies list: core
type Dependencies struct {
	ControlPlane map[string]Dependency `json:"control-plane,omitempty" yaml:"control-plane,omitempty" mapstructure:"control-plane"`
	Core         map[string]Dependency `json:"core,omitempty" yaml:"core,omitempty" mapstructure:"core"`
}

// Dependency defines a software, package or application required on KubekitOS
type Dependency struct {
	Version      string       `json:"version" yaml:"version" mapstructure:"version"`
	Name         string       `json:"name" yaml:"name" mapstructure:"name"`
	Src          string       `json:"src" yaml:"src" mapstructure:"src"`
	PrebakePath  string       `json:"prebake-path" yaml:"prebake-path" mapstructure:"prebake-path"`
	Checksum     string       `json:"checksum" yaml:"checksum" mapstructure:"checksum"`
	ChecksumType ChecksumType `json:"checksum_type" yaml:"checksum_type" mapstructure:"checksum_type"`
	LicenseURL   string       `json:"license_url" yaml:"license_url" mapstructure:"license_url"`
}

// ChecksumType is a type to group the few supported checksum types
type ChecksumType string

// ChecksumTypes is a list of supported checksum types
var ChecksumTypes = []ChecksumType{"sha256"}

// Yaml returns the Manifest structure in YAML format
func (m *Manifest) Yaml() ([]byte, error) {
	return yaml.Marshal(m)
}

// JSON returns the Manifest structure in JSON format
func (m *Manifest) JSON() ([]byte, error) {
	return json.Marshal(m)
}

// Yaml returns the KubeManifest in YAML format
func Yaml() ([]byte, error) {
	return KubeManifest.Yaml()
}

// JSON returns the KubeManifest in JSON format
func JSON() ([]byte, error) {
	return KubeManifest.JSON()
}

// WriteFile writes the manifest to a file in a given format
func (m *Manifest) WriteFile(filename, format string) (err error) {
	var data []byte

	if len(format) == 0 {
		format = strings.TrimPrefix(filepath.Ext(filename), ".")
	}
	if len(format) == 0 {
		format = DEFAULTFORMAT
	}

	switch strings.ToLower(format) {
	case "yaml", "yml":
		data, err = m.Yaml()
	case "json", "js":
		data, err = m.JSON()
	default:
		err = fmt.Errorf("Unknown format %s to convert the manifest", format)
	}
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

// WriteFile writes the KubeManifest to a file in a given format
func WriteFile(filename, format string) error {
	return KubeManifest.WriteFile(filename, format)
}
