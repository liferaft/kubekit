package provisioner

import (
	"bytes"
	"fmt"

	"github.com/kraken/ui"

	"github.com/kraken/terraformer"
	"github.com/kubekit/kubekit/pkg/provisioner/aks"
	"github.com/kubekit/kubekit/pkg/provisioner/aws"
	"github.com/kubekit/kubekit/pkg/provisioner/eks"
	"github.com/kubekit/kubekit/pkg/provisioner/openstack"
	"github.com/kubekit/kubekit/pkg/provisioner/raw"
	"github.com/kubekit/kubekit/pkg/provisioner/stacki"
	"github.com/kubekit/kubekit/pkg/provisioner/state"
	"github.com/kubekit/kubekit/pkg/provisioner/vra"
	"github.com/kubekit/kubekit/pkg/provisioner/vsphere"
)

// Provisioner represents a platform to provision a cluster.
type Provisioner interface {
	Config() interface{}
	Variables() map[string]interface{}
	Name() string
	BeProvisioner(*terraformer.State) error
	GetPublicKey() (string, []byte, bool)
	PublicKey(string, []byte)
	GetPrivateKey() (string, []byte, bool)
	PrivateKey(string, []byte, []byte)
	Plan(bool) (*terraformer.Plan, error)
	Apply(bool) error
	Provision() error
	Terminate() error
	Code() []byte
	State() *terraformer.State
	LoadState(*bytes.Buffer) error
	Address() string
	Port() int
	Output(string) string
	Nodes() []*state.Node
	Credentials(...string)
	MergeWithEnv(map[string]string) error
}

var allPlatforms = []string{
	"aks",
	"aws",
	"eks",
	"vsphere",
	"openstack",
	"raw",
	"vra",
	"stacki",
}

// SupportedPlatformsName returns all the supported platforms name
func SupportedPlatformsName() []string {
	return allPlatforms
}

// SupportedPlatforms creates and returns all the supported platforms
func SupportedPlatforms(clusterName string, envConfig map[string]string, ui *ui.UI, version string) map[string]Provisioner {
	platforms := make(map[string]Provisioner, len(allPlatforms))

	for _, name := range allPlatforms {
		if p, err := New(clusterName, name, envConfig, ui, version); err == nil {
			platforms[name] = p
		}
	}
	return platforms

	// return map[string]Provisioner{
	// 	"aws":       aws.New(clusterName, envConfig, ui),
	// 	"eks":       eks.New(clusterName, envConfig, ui),
	// 	"vsphere":   vsphere.New(clusterName, envConfig, ui),
	// 	"raw":       raw.New(clusterName, envConfig, ui),
	// 	"vra":       vra.New(clusterName, envConfig, ui),
	// 	"stacki":    stacki.New(clusterName, envConfig, ui),
	// 	"openstack": openstack.New(clusterName, envConfig, ui),
	// }
}

// New creates a new Provisioner for the given platform
func New(clusterName, platformName string, envConfig map[string]string, ui *ui.UI, version string) (Provisioner, error) {
	var p Provisioner
	var err error

	switch platformName {
	case "aks":
		p, err = aks.New(clusterName, envConfig, ui, version)
	case "aws":
		p, err = aws.New(clusterName, envConfig, ui, version)
	case "eks":
		p, err = eks.New(clusterName, envConfig, ui, version)
	case "vsphere":
		p, err = vsphere.New(clusterName, envConfig, ui, version)
	case "raw":
		p, err = raw.New(clusterName, envConfig, ui, version)
	case "vra":
		p, err = vra.New(clusterName, envConfig, ui, version)
	case "stacki":
		p, err = stacki.New(clusterName, envConfig, ui, version)
	case "openstack":
		p, err = openstack.New(clusterName, envConfig, ui, version)
	default:
		return nil, fmt.Errorf("platform %s is not supported", platformName)
	}

	return p, err
}

// NewPlatform create a new Provisioner with the given name and from the provided
// configuration (probably obtained from the cluster config file)
func NewPlatform(name, clusterName string, config interface{}, credentials []string, ui *ui.UI, version string) (Provisioner, error) {
	var c map[interface{}]interface{}
	if config != nil {
		c = config.(map[interface{}]interface{})
	}

	switch name {
	case "aks":
		return aks.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "aws":
		return aws.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "eks":
		return eks.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "vsphere":
		return vsphere.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "raw":
		return raw.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "vra":
		return vra.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "stacki":
		return stacki.CreateFrom(clusterName, c, credentials, ui, version), nil
	case "openstack":
		return openstack.CreateFrom(clusterName, c, credentials, ui, version), nil
	}

	return nil, fmt.Errorf("unknown platform named %q", name)
}
