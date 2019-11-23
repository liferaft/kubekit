package resources

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kraken/ui"

	"github.com/kubekit/kubekit/pkg/configurator/kube"
	"k8s.io/client-go/kubernetes"
)

// DefaultResourcesPerPlatform store the resources to create in a K8s cluster after
// it's up and running. Some resources are applied on any platform, those are
// under the name of "default", and some platforms require specific resources
var DefaultResourcesPerPlatform = map[string][]string{
	"default": []string{
		"open-policy-agent",
	},
	"aks": []string{
		"azure-storage-classes",
		"pod-security-policies",
		"priority-classes",
		"resource-quotas",
		"kube-state-metrics",
		//"aks-network-policies",
		"aks-aad-pod-identity-nmi-mic",
		"aks-acr-docker-secret",
	},
	"eks": []string{
		"aws-auth",
		"eks-calico",
		"rook-cluster",
		"rook-common",
		"rook-operator",
		"rook-blockstore",
		"rook-filestore",
		"ebs-blockstore",
		"efs-filestore",
		"pod-security-policies",
		"priority-classes",
		"resource-quotas",
		"eks-heapster",
		"kube-state-metrics",
		"eks-network-policies",
	},
	"aws": []string{
		"ebs-blockstore",
		// "efs-filestore",
	},
	"vsphere": []string{
		"vsphere-volumes",
	},
	"raw":       []string{},
	"vra":       []string{},
	"stacki":    []string{},
	"openstack": []string{},
}

// // DefaultDataKeyMapping is a default map of state keys and data template
// // variables. If the data template variable is not defined here, it will use the
// // same name in the state keys. It is optional to do a mapping for every variable
// // but recommended just in case in the future we use a struct instead of a map
// // to provide values to template variables, and required if the key use special
// // characters such as `-`
// var DefaultDataKeyMapping = map[string]string{
// 	"role-arn": "RoleARN",
// }

// Resources is a list of resources with the kubernetes client used to access
// the cluster where these resources exists
type Resources struct {
	order      []string
	content    map[string]string
	data       map[string]string
	kubeClient *kube.Client
	ui         *ui.UI
}

// New creates a list of resources
func New(kubeconfigPath string, ui *ui.UI) (*Resources, error) {
	kubeClient, err := kube.NewClientE("", kubeconfigPath, ui)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string, 0)

	return &Resources{
		content:    make(map[string]string, 0),
		kubeClient: kubeClient,
		data:       data,
		ui:         ui,
	}, nil
}

//go:generate go run ../codegen/kubernetes/main.go --src ../templates/resources --dst ./code.go --exclude "*.new,*.not-ready"

// ResourceTemplates contain the code of the resources from the templates/resource
// directory. The value of ResourceTemplates is in the generated file 'code.go'
var ResourceTemplates map[string]string

// DefaultResourcesFor return the list of resources to create in a K8s cluster
// on the given platforms. The list includes the default resources that should
// be in every platform
func DefaultResourcesFor(platforms ...string) []string {
	var resources = []string{}
	platforms = append(platforms, "default")
	for _, p := range platforms {
		resources = append(resources, DefaultResourcesPerPlatform[p]...)
	}
	return resources
}

// Names return the names of the resources loaded
func (r *Resources) Names() []string {
	return r.order
}

// AddResources adds the given list of resources to the list of resources.
// Returns an error if it's a template resource that does not exists
func (r *Resources) AddResources(resources []string) error {
	for _, res := range resources {
		switch {
		case isFile(res):
			r.ui.Log.Debugf("loaded resource in file %q", res)
			r.content[res] = ""
			r.order = append(r.order, res)
		default:
			if _, ok := ResourceTemplates[res]; !ok {
				return fmt.Errorf("unknown resource template with name %q. If it is a file, use the prefix 'file://',  'http://' or 'https://'", res)
			}
			r.ui.Log.Debugf("loaded template resource %q", res)
			r.content[res] = ResourceTemplates[res]
			r.order = append(r.order, res)
		}
	}

	return nil
}

// AddDefaultResourcesFor adds the default list of resources for the given platforms
func (r *Resources) AddDefaultResourcesFor(platforms ...string) {
	resources := DefaultResourcesFor(platforms...)
	r.AddResources(resources)
}

// AddData adds into the map of data the given value on that name
func (r *Resources) AddData(name, value string) {
	// key := lookupKey(name)
	r.data[name] = value
}

// AppendData appends into the map of data the given map of values
func (r *Resources) AppendData(data map[string]string) {
	for k, v := range data {
		// key := lookupKey(k)
		r.data[k] = v
	}
}

// // AppendConfigData appends into the map of data some useful parameters from
// // the cluster configuration
// func (r *Resources) AppendConfigData(
// 	clusterName string,
// 	platform string,
// 	certsPath string,
// ) {
// 	r.data["clusterName"] = clusterName
// 	r.data["platform"] = platform
// 	r.data["certsPath"] = certsPath
// }

// // AppendDataFromMap appends into the map of data the given map of values but
// // using the default mapping to lookup the right key
// func (r *Resources) AppendDataFromMap(data map[string]interface{}) {
// 	for k, v := range data {
// 		key := lookupKey(k)
// 		r.data[key] = fmt.Sprintf("%v", v)
// 	}
// }

// func lookupKey(name string) string {
// 	if key, ok := DefaultDataKeyMapping[name]; ok {
// 		return key
// 	}
// 	return name
// }

// Render creates a resource from a resource name
func (r *Resources) Render(name string, filename string) ([]byte, error) {
	var (
		codeTemplate    string
		ok              bool
		resourceContent bytes.Buffer
	)

	if codeTemplate, ok = r.content[name]; !ok {
		return nil, fmt.Errorf("not found resource named %q", name)
	}
	resourceTpl, err := template.
		New(name).
		Option("missingkey=error").
		Funcs(tmplFuncMap).
		Parse(codeTemplate)
	if err != nil {
		return nil, err
	}
	err = resourceTpl.Execute(&resourceContent, r.data)
	// since missingkey=error is set, return the error
	if err != nil {
		return nil, err
	}

	if len(filename) != 0 {
		r.ui.Log.Infof("saving manifest %s into %s", name, filename)
		if errW := ioutil.WriteFile(filename, resourceContent.Bytes(), 0644); errW != nil {
			r.ui.Log.Errorf("failed to save manifest %s to file %s ", name, filename)
		}
	}

	return resourceContent.Bytes(), err
}

func isFile(name string) bool {
	return strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") || strings.HasPrefix(name, "file://")
}

func isURL(name string) bool {
	return strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://")
}

// KubernetesClientSet returns the Kubernetes Clientset from the kubernetes client config
func (r *Resources) KubernetesClientSet() (*kubernetes.Clientset, error) {
	return r.kubeClient.Config.KubernetesClientSet()
}

// KubernetesClient returns the Kubernetes Client
func (r *Resources) KubernetesClient() *kube.Client {
	return r.kubeClient
}

// Export exports all the Kubernetes manifest templates to the given directory
func (r *Resources) Export(exportDir string) error {
	exportErrors := applyErrors{}

	for _, name := range r.order {
		if _, ok := r.content[name]; !ok {
			r.ui.Log.Warnf("resource content for %q not found", name)
			continue
		}
		if isFile(name) {
			r.ui.Log.Debugf("resource %s is a file (local or remote)", name)
			continue
		}

		filename := filepath.Join(exportDir, name+".yaml")

		if _, err := r.Render(name, filename); err != nil {
			r.ui.Log.Errorf("failed applying resource %s. %v", name, err)
			exportErrors.Add(name, err)
		}
	}

	if exportErrors.Empty() {
		return nil
	}

	r.ui.Log.Errorf(exportErrors.Error())

	return fmt.Errorf(exportErrors.Error())
	// return exportErrors
}
