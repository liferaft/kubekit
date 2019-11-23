package kluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"

	toml "github.com/pelletier/go-toml"
	"github.com/kubekit/kubekit/version"
	yaml "gopkg.in/yaml.v2"
)

const tableTemplateKeyword = "table"

var validClusterInfoFieldsToFilterBy = []string{"name", "nodes", "platform", "status", "version", "path", "url", "entrypoint", "kubeconfig"}

// ClusterInfo basic cluster information
type ClusterInfo struct {
	Name       string `json:"name" yaml:"name" toml:"name"`
	Nodes      int    `json:"nodes" yaml:"nodes" toml:"nodes"`
	Platform   string `json:"platform" yaml:"platform" toml:"platform"`
	Status     string `json:"status" yaml:"status" toml:"status"`
	Version    string `json:"version" yaml:"version" toml:"version"`
	Path       string `json:"path" yaml:"path" toml:"path"`
	URL        string `json:"url" yaml:"url" toml:"url"`
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig" toml:"kubeconfig"`
}

func isValidFilterParam(param string) bool {
	for _, p := range validClusterInfoFieldsToFilterBy {
		if p == strings.ToLower(param) {
			return true
		}
	}
	return false
}

// IsValidFilter returns true if one of the filter parameters for clusters info
// is not valid
func IsValidFilter(params map[string]string) bool {
	for p := range params {
		if !isValidFilterParam(p) {
			return false
		}
	}
	return true
}

// InvalidFilterParams return a list of invalid filter parameters for clusters info
func InvalidFilterParams(params map[string]string) []string {
	invalid := []string{}
	for p := range params {
		if !isValidFilterParam(p) {
			invalid = append(invalid, p)
		}
	}
	return invalid
}

// ClustersInfo list of clusters with its information
type ClustersInfo []ClusterInfo

// JSON returns the cluster information in JSON format
func (ci ClustersInfo) JSON(pp bool) (string, error) {
	var output []byte
	var err error
	if pp {
		output, err = json.MarshalIndent(ci, "", "  ")
	} else {
		output, err = json.Marshal(ci)
	}
	return string(output), err
}

// YAML returns the cluster information in YAML format
func (ci ClustersInfo) YAML() (string, error) {
	output, err := yaml.Marshal(ci)
	return string(output), err
}

// TOML returns the cluster information in TOML format
func (ci ClustersInfo) TOML() (string, error) {
	var tomlStruct struct {
		Clusters map[string]ClusterInfo `toml:"clusters"`
	}
	var ciMap map[string]ClusterInfo
	ciMap = make(map[string]ClusterInfo, len(ci))
	for _, ki := range ci {
		ciMap[ki.Name] = ki
	}
	tomlStruct.Clusters = ciMap

	output, err := toml.Marshal(tomlStruct)
	return string(output), err
}

// Table returns the cluster information in a plain text table
func (ci ClustersInfo) Table(wide bool) string {
	var tableTmpl bytes.Buffer

	header := "Name\tNodes\tPlatform\tStatus\tVersion"
	if wide {
		header = header + "\tEntrypoint\tKubeconfig"
	}
	tableTmpl.WriteString(header + "\n")

	for _, k := range ci {
		ver, _ := version.NewSemVer(k.Version)
		// ignore error, the error was validated when the cluster was loaded with LoadSummary()
		var supported string
		if ver.LT(MinSemVersion) {
			supported = fmt.Sprintf(" (<%s)", MinSemVersion)
		}
		if ver.GT(SemVersion) {
			supported = fmt.Sprintf(" (>%s)", SemVersion)
		}
		row := fmt.Sprintf("%s\t%d\t%s\t%s\t%s", k.Name, k.Nodes, k.Platform, k.Status, k.Version+supported)
		if wide {
			kubeconfigpath := "None"
			if _, err := os.Stat(k.Kubeconfig); err == nil {
				kubeconfigpath = k.Kubeconfig
			}
			row = fmt.Sprintf("%s\t%s\t%s", row, k.URL, kubeconfigpath)
		}
		tableTmpl.WriteString(row + "\n")
	}

	return tabletize(tableTmpl.String())
}

// Names returns a list of clusters name
func (ci ClustersInfo) Names() []string {
	names := []string{}
	for _, k := range ci {
		names = append(names, k.Name)
	}
	return names
}

// Stringf returns the clusters info in the requested format to be printed
func (ci ClustersInfo) Stringf(format string, ppArr ...bool) (result string, err error) {
	switch format {
	case "", "wide", "w":
		result = ci.Table((format == "wide") || (format == "w"))
	case "quiet", "q", "names":
		names := ci.Names()
		result = strings.Join(names, "\n")
	case "json":
		var pp bool
		if len(ppArr) != 0 {
			pp = ppArr[0]
		}
		result, err = ci.JSON(pp)
	case "yaml":
		result, err = ci.YAML()
	case "toml":
		result, err = ci.TOML()
	default:
		err = fmt.Errorf("unknown format %q", format)
	}

	return result, err
}

// func (ci ClustersInfo) String() string {
// 	result, err := ci.Stringf("")
// 	if err != nil {
// 		return err.Error()
// 	}
// 	return result
// }

func header(format string) (string, error) {
	format = strings.Replace(format, "{{.Nodes}}", "NODES", 1)

	tmpl, err := template.New("Clusters Information Header").Parse(format)
	if err != nil {
		return "", err
	}

	ci := ClusterInfo{
		Name:       "NAME",
		Platform:   "PLATFORM",
		Status:     "STATUS",
		Version:    "VERSION",
		Path:       "PATH",
		URL:        "ENTRYPOINT",
		Kubeconfig: "KUBECONFIG",
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, ci); err != nil {
		return "", err
	}

	return rendered.String() + "\n", nil
}

func tabletize(text string) string {
	var output bytes.Buffer
	w := tabwriter.NewWriter(&output, 0, 0, 3, ' ', 0)
	fmt.Fprint(w, text)
	w.Flush()
	return output.String()
}

// Template renders the clusters information from a given Go template
func (ci ClustersInfo) Template(format string) (string, error) {
	if len(ci) == 0 {
		return "", nil
	}
	// Transformations to template
	format = strings.Replace(format, "\\t", "\t", -1)
	format = strings.Replace(format, "\\n", "\n", -1)
	format = strings.Replace(format, ".Entrypoint", ".URL", -1)

	var head string
	if strings.Contains(format, tableTemplateKeyword) {
		var err error
		// More transformations
		format = strings.Replace(format, tableTemplateKeyword, "", 1)
		format = strings.TrimSpace(format)

		if head, err = header(format); err != nil {
			return "", err
		}
	}
	// If it's not empty and does not have a final newline, add it
	if len(format) != 0 && format[len(format)-1:] != "\n" {
		format = format + "\n"
	}

	// Add the range to loop thru all the cluster information
	tmplFormat := "{{range .}}" + format + "{{end}}"
	tmpl, err := template.New("Clusters Information").Parse(tmplFormat)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, ci); err != nil {
		return "", err
	}

	output := rendered.String()
	if len(head) != 0 {
		output = tabletize(head + output)
	}

	return output, nil
}

// FilterBy filters the clusters information by the given parameters
func (ci *ClustersInfo) FilterBy(params map[string]string) {
	if len(params) == 0 {
		return
	}

	newCi := ClustersInfo{}

	for _, i := range *ci {
		if i.ContainsAll(params) {
			newCi = append(newCi, i)
		}
	}

	*ci = newCi
}

// ContainsAll returns true if the cluster information contains all the given
// parameters. The paramters is a map of key/value pairs, where the keys are the
// fields of the cluster information named as the JSON value
func (i ClusterInfo) ContainsAll(params map[string]string) bool {
	if len(params) == 0 {
		// This is always true and avoid the Marshaling/unmarshaling of `i`
		return true
	}

	infoJSON, err := json.Marshal(&i)
	if err != nil {
		return false
	}
	infoMap := map[string]interface{}{}
	if err := json.Unmarshal(infoJSON, &infoMap); err != nil {
		return false
	}

	for kp, vp := range params {
		kp := strings.ToLower(kp)
		if kp == "entrypoint" {
			kp = "url"
			if vp == "" {
				vp = "None"
			}
		}
		if v, found := infoMap[kp]; !found || vp != fmt.Sprintf("%v", v) {
			return false
		}
	}

	return true
}

// GetClustersInfo gets the list of clusters and its basic information. If
// clustersName is empty will return the information for all the existing clusters
func GetClustersInfo(baseDir string, params map[string]string, clustersName ...string) (ClustersInfo, error) {
	var ci ClustersInfo

	list, err := List(baseDir, clustersName...)
	if err != nil {
		return nil, err
	}

	ci = make(ClustersInfo, 0)
	for _, k := range list {
		platformName := k.Platform()

		if platformName == "" || k.Name == "" {
			// There is a cluster created incorrectly or corrupt.
			// Do not return an error, would be nice to report it but not to stop the collection of the other clusters
			continue
		}

		i := ClusterInfo{
			Name:     k.Name,
			Platform: platformName,
			Status:   AbsentStatus.String(),
			Path:     filepath.Dir(k.Path()),
			Version:  k.Version,
			URL:      "None",
		}

		if k.State[platformName] != nil {
			i.Status = k.State[platformName].Status
			i.Kubeconfig = filepath.Join(filepath.Dir(k.Path()), "certificates", "kubeconfig")
			i.Nodes = len(k.State[platformName].Nodes)
			if entrypoint := k.GetEntrypoint(); len(entrypoint) != 0 {
				i.URL = entrypoint
			}
		}

		ci = append(ci, i)
	}

	ci.FilterBy(params)

	return ci, nil
}

// GetEntrypoint returns the Kubernetes entrypoint or empty string if doesn't
// exists or invalid/malformed
func (k *Kluster) GetEntrypoint() string {
	platform := k.Platform()
	if len(k.State[platform].Address) == 0 {
		return ""
	}

	scheme := "https://"
	if strings.HasPrefix(k.State[platform].Address, scheme) || strings.HasPrefix(k.State[platform].Address, "http://") {
		scheme = ""
	}

	port := fmt.Sprintf(":%d", k.State[platform].Port)
	if k.State[platform].Port == 0 || strings.HasSuffix(k.State[platform].Address, port) {
		port = ""
	}

	if entrypoint, err := url.Parse(scheme + k.State[platform].Address + port); err == nil {
		return entrypoint.String()
	}

	return ""
}

// GetKubeconfig returns the content of the KubeConfig file
func (k *Kluster) GetKubeconfig() (string, error) {
	kubeconfigFile := filepath.Join(filepath.Dir(k.Path()), "certificates", "kubeconfig")

	data, err := ioutil.ReadFile(kubeconfigFile)

	return string(data), err
}
