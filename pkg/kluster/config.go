package kluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kubekit/kubekit/version"

	"github.com/johandry/merger"

	"github.com/kubekit/kubekit/pkg/configurator/resources"

	"github.com/nightlyone/lockfile"
	"github.com/pelletier/go-toml"
	"github.com/kraken/ui"
	"github.com/kubekit/kubekit/pkg/configurator"
	"github.com/kubekit/kubekit/pkg/crypto/tls"
	"github.com/kubekit/kubekit/pkg/provisioner"
	"gopkg.in/yaml.v2"
)

// DefaultFormat is the default format for the Kubernetes Cluster (Kluster)
// configuration file. The options are: yaml, json and toml
// This format will be used for the file extension, so use a short or
// extension-like format name
const (
	DefaultFormat         = "yaml"
	DefaultConfigFilename = "cluster"
	CertificatesDirname   = "certificates"
	TerraformDirname      = "terraform"
	KubernetesDirname     = "kubernetes"
	StateDirname          = ".tfstate"
	RegistryDirname       = "registries"
)

/*
When to modify Version and MinVersion:

Modify Version when the cluster configuration file changes. If it's a major change,
changes the first number (MAJOR) of the version, otherwise change the second (MINOR).
The Version has to be SemVer compliance but a 3rd number (PATCH) is not considered
for the cluster configuration version (yet).

Modify the MinVersion when the cluster configuration file change and the new parameters
does not have default values.
Other way to know is, if a cluster configuration of a previous version can be read,
and this is because the missing parameters can have the default value, then this
version is allowed. Then try to read a cluster configuration file of the previous
version.
*/

var (
	// Version is the cluster configuration version this KubeKit creates. Greater
	// versions are not supported.
	Version = "1.1"

	// MinVersion is the cluster configuration file minimum version
	// accepted by this version of KubeKit. If a cluster config file with a lower
	// version is provided by the user, that cluster config file is rejected.
	MinVersion = "1.0"

	// SemVersion is the value of Version in version.SemVer type.
	SemVersion *version.SemVer

	// MinSemVersion is same as MinVersion but as SemVer
	MinSemVersion *version.SemVer
)

func init() {
	SemVersion, _ = version.NewSemVer(Version)
	MinSemVersion, _ = version.NewSemVer(MinVersion)
	// Ignore errors, make sure the value of Version and MinVersion
	// are SemVer compliance: X.Y or X.Y.Z
}

// Kluster encapsulates the Kubernetes Cluster configuration
type Kluster struct {
	Version      string                             `json:"version" yaml:"version" mapstructure:"version"`                  // KubeKit Configuration/API version
	Kind         string                             `json:"kind" yaml:"kind" mapstructure:"kind"`                           // File kind/type. Example: config, template
	Name         string                             `json:"name" yaml:"name" mapstructure:"name"`                           // Cluster Name
	Platforms    map[string]interface{}             `json:"platforms" yaml:"platforms" mapstructure:"platform"`             // Configuration of the platforms where the cluster could be installed
	State        map[string]*State                  `json:"state" yaml:"state" mapstructure:"state"`                        // State of the cluster for each platform
	Config       *configurator.Config               `json:"config,omitempty" yaml:"config,omitempty" mapstructure:"config"` // Kubernetes configuration, no matter what platform
	Resources    []string                           `json:"resources" yaml:"resources" mapstructure:"resources"`
	path         string                             // Path is where the cluster configuration file is
	provisioner  map[string]provisioner.Provisioner // List of provisioners. It's a platform that can be provisioned
	certificates tls.KeyPairs                       // List of TLS key pairs
	ui           *ui.UI                             // UI to print out to console
}

// New creates a new Kluster or load it if the file already exists
func New(name, platformName, path, format string, parentUI *ui.UI, envConfig map[string]string) (*Kluster, error) {
	if len(format) == 0 {
		format = DefaultFormat
	}
	if !validFormat(format) {
		return nil, fmt.Errorf("invalid format %q for the kubernetes cluster config file", format)
	}
	path = filepath.Join(path, DefaultConfigFilename+"."+format)

	if _, err := os.Stat(path); os.IsExist(err) {
		return nil, fmt.Errorf("the Kluster config file %q already exists", path)
	}

	newUI := parentUI.Copy()

	cluster := Kluster{
		Version: Version,
		Kind:    "cluster",
		Name:    name,
		path:    path,
		ui:      newUI,
	}

	// // TODO: Improve this, all platforms are not needed
	// allPlatforms := provisioner.SupportedPlatforms(name, envConfig)
	// platform, ok := allPlatforms[platformName]
	// if !ok {
	// 	return nil, fmt.Errorf("platform %q is not supported", platformName)
	// }

	platform, err := provisioner.New(name, platformName, envConfig, newUI, Version)
	if err != nil {
		return nil, fmt.Errorf("platform %q is not supported. %s", platformName, err)
	}

	logPrefix := fmt.Sprintf("KubeKit [ %s@%s ]", cluster.Name, platformName)
	cluster.ui.SetLogPrefix(logPrefix)

	cluster.Platforms = make(map[string]interface{}, 1)
	cluster.provisioner = make(map[string]provisioner.Provisioner, 1)
	cluster.State = make(map[string]*State, 1)

	cluster.Platforms[platformName] = platform.Config()
	cluster.provisioner[platformName] = platform
	cluster.State[platformName] = &State{
		Status: AbsentStatus.String(),
	}

	cluster.Resources = resources.DefaultResourcesFor(platformName)

	// return if this is a platform with no configuration, such as EKS or AKS
	switch platformName {
	case "eks", "aks":
		return &cluster, nil
	}

	cluster.Config, err = configurator.DefaultConfig(envConfig)

	return &cluster, err
}

// Update updates the cluster with the given configuration
func (k *Kluster) Update(envConfig map[string]string) error {
	platformName := k.Platform()

	platform := k.provisioner[platformName]
	if err := platform.MergeWithEnv(envConfig); err != nil {
		return err
	}
	k.Platforms[platformName] = platform.Config()

	// Some platforms such as EKS and AKS does not have a config section
	if k.Config == nil {
		return nil
	}

	return merger.Merge(k.Config, envConfig)
	// return k.Config.MergeWithEnv(envConfig)
}

// NewTemplate creates a new Kluster template. A template is a Kluster with multiple platforms
func NewTemplate(name string, platforms []string, path, format string, parentUI *ui.UI, envConfig map[string]string) (*Kluster, error) {
	if len(format) == 0 {
		format = DefaultFormat
	}
	if !validFormat(format) {
		return nil, fmt.Errorf("invalid format %q for the kubernetes cluster config file", format)
	}
	path = filepath.Join(path, name+"."+format)

	newUI := parentUI.Copy()

	cluster := Kluster{
		Version: Version,
		Kind:    "template",
		Name:    name,
		path:    path,
		ui:      newUI,
	}

	if _, err := os.Stat(path); os.IsExist(err) {
		return nil, fmt.Errorf("the template file %q already exists", path)
	}

	allPlatforms := provisioner.SupportedPlatforms(name, envConfig, newUI, Version)

	if len(platforms) == 0 {
		for k := range allPlatforms {
			platforms = append(platforms, k)
		}
	}

	cluster.Platforms = make(map[string]interface{}, len(platforms))

	for _, pName := range platforms {
		platform, ok := allPlatforms[pName]
		if !ok {
			return nil, fmt.Errorf("platform %q is not supported", pName)
		}
		cluster.Platforms[pName] = platform.Config()
	}

	var err error
	cluster.Config, err = configurator.DefaultConfig(envConfig)
	cluster.Resources = resources.DefaultResourcesFor()

	return &cluster, err
}

// Copy copies the current cluster configuration into a new one, with a new or existing platform and format
func (k *Kluster) Copy(name, platformName, path, format string, parentUI *ui.UI, envConfig map[string]string) (*Kluster, error) {
	if len(format) == 0 {
		format = k.format()
	}
	if !validFormat(format) {
		return nil, fmt.Errorf("invalid format %q for the new cluster config file", format)
	}
	path = filepath.Join(path, DefaultConfigFilename+"."+format)

	if _, err := os.Stat(path); os.IsExist(err) {
		return nil, fmt.Errorf("the new cluster config file %q already exists", path)
	}

	newUI := parentUI.Copy()

	cluster := Kluster{
		Version:   k.Version,
		Kind:      k.Kind,
		Name:      name,
		Config:    k.Config,
		Resources: k.Resources,
		path:      path,
		ui:        newUI,
	}

	cluster.Platforms = make(map[string]interface{}, 1)
	if len(platformName) != 0 {
		allPlatforms := provisioner.SupportedPlatforms(name, envConfig, newUI, cluster.Version)
		provisioner, ok := allPlatforms[platformName]
		if !ok {
			return nil, fmt.Errorf("platform %q is not supported", platformName)
		}
		cluster.Platforms[platformName] = provisioner.Config()
	} else {
		platformName = k.Platform()
		cluster.Platforms = k.Platforms
	}

	cluster.State = make(map[string]*State, 1)
	cluster.State[platformName] = &State{
		Status: AbsentStatus.String(),
	}

	return &cluster, nil
}

// Platform returns the name of the cluster platform
func (k *Kluster) Platform() string {
	var platformName string
	for k := range k.Platforms {
		platformName = k
		// It's only one platform, so break after getting the first one (just in case)
		break
	}
	return platformName
}

func validFormat(format string) bool {
	switch format {
	case "yaml", "yml":
		return true
	case "json", "js":
		return true
	case "toml", "tml":
		return true
	default:
		return false
	}
}

func (k *Kluster) format() string {
	ext := filepath.Ext(k.path)

	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json", ".js":
		return "json"
	case ".toml", ".tml":
		return "toml"
	default:
		return DefaultFormat
	}
}

// Path returns the path of the configuration file
func (k *Kluster) Path() string {
	return k.path
}

// Dir returns the path of the configuration directory
func (k *Kluster) Dir() string {
	return filepath.Dir(k.path)
}

// CertsDir returns the path to store Certificates and KubeConfig file
func (k *Kluster) CertsDir() string {
	return filepath.Join(k.Dir(), CertificatesDirname)
}

// MakeCertDir creates the certificate directory for the given platforms or the
// base certificates directory if no platform is given
func (k *Kluster) MakeCertDir(platfomName ...string) (string, error) {
	return k.makeCertDir(platfomName...)
}

func (k *Kluster) makeCertDir(platfomName ...string) (string, error) {
	d := k.CertsDir()
	if len(platfomName) == 0 {
		return d, os.MkdirAll(d, 0700)
	}
	d = filepath.Join(d, platfomName[0])
	return d, os.MkdirAll(d, 0700)
}

func (k *Kluster) tfDir() string {
	return filepath.Join(k.Dir(), TerraformDirname)
}

func (k *Kluster) makeTFDir(pName string) (string, error) {
	d := filepath.Join(k.tfDir(), pName)
	return d, os.MkdirAll(d, 0755)
}

func (k *Kluster) makeK8sDir() (string, error) {
	d := filepath.Join(k.Dir(), KubernetesDirname)
	return d, os.MkdirAll(d, 0755)
}

// StateDir resturns the directory where the Terraform state file is
func (k *Kluster) StateDir() string {
	return filepath.Join(k.Dir(), StateDirname)
}

// StateFile resturns the Terraform state file path
func (k *Kluster) StateFile() string {
	platform := k.Platform()
	return filepath.Join(k.Dir(), StateDirname, platform+".tfstate")
}

func (k *Kluster) makeStateDir() (string, error) {
	d := k.StateDir()
	return d, os.MkdirAll(d, 0755)
}

func (k *Kluster) registryDir() string {
	return filepath.Join(k.Dir(), RegistryDirname)
}

func (k *Kluster) makeRegistryDir() (string, error) {
	d := k.registryDir()
	return d, os.MkdirAll(d, 0755)
}

func (k *Kluster) String() string {
	format := k.format()
	var data []byte
	var err error

	switch format {
	case "yaml":
		data, err = k.YAML()
	case "json":
		data, err = k.JSON(false)
	case "toml":
		data, err = k.TOML()
	default:
		err = fmt.Errorf("can't stringify the Kluster, unknown format %q", format)
	}

	if err != nil {
		return err.Error()
	}
	return string(data)
}

// YAML returns the Kluster in YAML format
func (k *Kluster) YAML() ([]byte, error) {
	return yaml.Marshal(k)
}

// JSON returns the Kluster in JSON format
func (k *Kluster) JSON(pp bool) ([]byte, error) {
	if pp {
		return json.MarshalIndent(k, "", "  ")
	}
	return json.Marshal(k)
}

// TOML returns the Kluster in TOML format
func (k *Kluster) TOML() ([]byte, error) {
	return toml.Marshal(k)
}

// ReadYAML load the Kluster from a YAML format text
func (k *Kluster) ReadYAML(b []byte) error {
	return yaml.Unmarshal(b, k)
}

// ReadJSON load the Kluster from a JSON format text
func (k *Kluster) ReadJSON(b []byte) error {
	return json.Unmarshal(b, k)
}

// ReadTOML load the Kluster from a TOML format text
func (k *Kluster) ReadTOML(b []byte) error {
	return toml.Unmarshal(b, k)
}

// Load loads a given Kubernetes Cluster config file and dump the settings to a
// new Kluster which is returned
func Load(path string, parentUI *ui.UI) (*Kluster, error) {
	newUI := parentUI.Copy()
	cluster := &Kluster{
		path: path,
		ui:   newUI,
	}
	if err := cluster.Load(); err != nil {
		return nil, err
	}

	logPrefix := fmt.Sprintf("KubeKit [ %s@%s ]", cluster.Name, cluster.Platform())
	cluster.ui.SetLogPrefix(logPrefix)

	return cluster, nil
}

// LoadSummary loads the most important information about a Kubernetes cluster
// configuration. Notice that the returned Kluster is incomplete and just to get
// basic/important information about it such as the cluster name
func LoadSummary(path string) (*Kluster, error) {
	cluster := &Kluster{
		path: path,
	}
	if err := cluster.LoadSummary(); err != nil {
		return nil, err
	}

	return cluster, nil
}

// LoadSummary loads the most important information about a Kubernetes cluster
// configuration. Notice that the returned Kluster is incomplete and just to get
// basic/important information about it such as the cluster name
func (k *Kluster) LoadSummary() error {
	var err error

	if _, err = os.Stat(k.path); os.IsNotExist(err) {
		return fmt.Errorf("not found Kluster config file %s", k.path)
	}

	lock, err := lockFile(k.path)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	b, err := ioutil.ReadFile(k.path)
	if err != nil {
		return err
	}

	format := k.format()

	switch format {
	case "json":
		err = k.ReadJSON(b)
	case "yaml":
		err = k.ReadYAML(b)
	case "toml":
		err = k.ReadTOML(b)
	default:
		return fmt.Errorf("unknown format %s", format)
	}

	return err
}

// Load loads the Kubernetes Cluster config file and dump the settings to this
// Kluster
func (k *Kluster) Load() error {
	if err := k.LoadSummary(); err != nil {
		return err
	}

	// DEBUG:
	// fmt.Printf("DEBUG: cluster %s config version: %s\tMin: %s\tMax: %s\n", k.Name, k.Version, MinSemVersion, SemVersion)

	ver, err := version.NewSemVer(k.Version)
	if err != nil {
		return fmt.Errorf("the cluster version (%s) is not well formed or not SemVer compliance. %s", k.Version, err)
	}
	if ver.LT(MinSemVersion) {
		return fmt.Errorf("the cluster version %s is not supported by this KubeKit, the minimun version supported is %s", k.Version, MinVersion)
	}
	if ver.GT(SemVersion) {
		return fmt.Errorf("the cluster version %s is greater than the cluster version supported by this KubeKit (%s)", k.Version, Version)
	}

	k.provisioner = make(map[string]provisioner.Provisioner, 1)
	name := k.Platform()
	config := k.Platforms[name]

	cred, err := k.GetCredentials()
	if err != nil {
		return err
	}

	platform, err := provisioner.NewPlatform(name, k.Name, config, cred, k.ui, k.Version)
	if err != nil {
		return err
	}
	k.provisioner[name] = platform

	return nil
}

func (k *Kluster) loadCredentials() (CredentialHandler, error) {
	platform := k.Platform()
	path := filepath.Join(filepath.Dir(k.Path()), CredentialsFileName)

	credentials := NewCredentials(k.Name, platform, path)

	if err := credentials.Read(); err != nil {
		return nil, err
	}

	if err := credentials.Getenv(true); err != nil {
		return nil, err
	}

	return credentials, nil
}

// GetCredentials retrieve the platform credentials from the credentials file
func (k *Kluster) GetCredentials() ([]string, error) {
	credentials, err := k.loadCredentials()
	if err != nil {
		return nil, err
	}

	// k.ui.Log.Debugf("credentials load from environment for platform %s. (%v)", credentials.platform(), credentials.parameters())
	// assign the credentials to the cluster provisioner
	params := credentials.parameters()
	// k.Credentials(params...)

	return params, nil
}

// GetCredentialsAsMap similar to GetCredentials but returns the credentials in a map of string
func (k *Kluster) GetCredentialsAsMap() (map[string]string, error) {
	credentials, err := k.loadCredentials()
	if err != nil {
		return nil, err
	}

	return credentials.asMap(), nil
}

// Credentials is used to pass the credential parameters to the provisioner
func (k *Kluster) Credentials(params ...string) {
	// DEBUG:
	// k.ui.Log.Debugf("credentials to assign to platform %s. (%v)", k.Platform(), params)
	k.provisioner[k.Platform()].Credentials(params...)
}

// SaveCredentials saves the Cluster credentials file
func (k *Kluster) SaveCredentials(params ...string) error {
	path := filepath.Join(k.Dir(), CredentialsFileName)
	cred := NewCredentials(k.Name, k.Platform(), path)

	if len(params) != 3 {
		if err := cred.Getenv(true); err != nil {
			return err
		}
		params = cred.parameters()
	}

	cred.SetParameters(params...)

	k.ui.Log.Debugf("saving credentials to cluster %s", k.Name)
	return cred.Write()
}

// Save saves the Kubernetes Cluster configuration into a file
func (k *Kluster) Save() error {
	// we load the nil defaults so that future versions
	// don't have to deal with the backwards compatibility with omitted values
	if k.Config != nil {
		k.Config.LoadNilDefault()
	}

	dir := k.Dir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("base directory %s does not exists", dir)
		// Or?:
		// os.MkdirAll(dir, 0755)
	}

	format := k.format()
	var data []byte
	var err error

	// Get platform configurations

	// Update configuration
	// pConfig := make(map[string]interface{}, len(k.Platforms))
	name := k.Platform()
	if p, ok := k.provisioner[name]; ok {
		platform := p
		k.Platforms[name] = platform.Config()
		// c := platform.Config()
		// // TODO: Workaround to not save the vSphere credentials, cannot have metadata to '-' because the Configurator takes the credentials from there.
		// if name == "vsphere" {
		// 	cVsphere := c.(*vsphere.Config)
		// 	cVsphere.VspherePassword = ""
		// 	cVsphere.VsphereUsername = ""
		// 	cVsphere.VsphereServer = ""
		// 	k.Platforms[name] = cVsphere
		// } else {
		// 	k.Platforms[name] = c
		// }

		err := k.LoadState()
		if err != nil {
			return err
		}
		k.UpdateState(name)

		k.ui.Log.Debugf("update state for %s: %v", name, k.State[name])
	}

	// k.Platforms = pConfig

	// Do not use String() because:
	// (1) returns string and []byte is needed, and
	// (2) pretty print (pp=true) is needed with JSON format
	switch format {
	case "yaml":
		data, err = k.YAML()
	case "json":
		data, err = k.JSON(true)
	case "toml":
		data, err = k.TOML()
	default:
		err = fmt.Errorf("can't stringify the Kluster, unknown format %q", format)
	}

	lock, err := lockFile(k.path)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	k.ui.Log.Debugf("updating cluster configuration file %s", k.path)
	return ioutil.WriteFile(k.path, data, 0644)
}

// UpdateState creates a new State structure from the given provisioner TF state
func (k *Kluster) UpdateState(platform string) {
	if _, ok := k.State[platform]; !ok {
		k.State[platform] = &State{
			Status: AbsentStatus.String(),
		}
	}

	if platform == "raw" {
		return
	}

	provisioner := k.provisioner[platform]

	k.State[platform].Address = provisioner.Address()
	k.State[platform].Port = provisioner.Port()

	nodes := provisioner.Nodes()
	hosts := configurator.Hosts{}
	for _, node := range nodes {
		host := configurator.Host{
			PublicIP:   node.PublicIP,
			PrivateIP:  node.PrivateIP,
			PublicDNS:  node.PublicDNS,
			PrivateDNS: node.PrivateDNS,
			RoleName:   node.RoleName,
			Pool:       node.Pool,
		}
		hosts = append(hosts, host)
	}
	k.State[platform].Nodes = hosts
}

// Lock locks the cluster so no action can be done until it's unlocked with lock.Unlock()
func (k *Kluster) Lock(name string) (lockfile.Lockfile, error) {
	if name == "" {
		name = k.Name
	}
	f := filepath.Join(k.Dir(), "."+name)
	return lockFile(f)
}

func lockFile(filename string) (lockfile.Lockfile, error) {
	// Lockfile cannot work with absolute paths  ¯\_(ツ)_/¯
	// So, make it absolute
	if !filepath.IsAbs(filename) {
		wd, err := os.Getwd()
		if err != nil {
			return "", nil
		}
		filename = filepath.Join(wd, filename)
	}

	lock, err := lockfile.New(filename + ".lock")
	if err != nil {
		return "", fmt.Errorf("cannot init the lock for file %q: %v", filename, err)
	}
	err = lock.TryLock()
	if err != nil {
		return "", fmt.Errorf("cannot lock the file %q: %v", filename, err)
	}
	return lock, nil
}
