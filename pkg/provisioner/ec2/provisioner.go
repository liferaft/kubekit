package ec2

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform/builtin/provisioners/file"
	"github.com/kraken/terraformer"
	"github.com/liferaft/kubekit/pkg/crypto"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

// ResourceTemplates maps resource names to content of resources
// implementation specified in code.go
var ResourceTemplates map[string]string

// BeProvisioner setup the Plaftorm to be a Provisioner
func (p *Platform) BeProvisioner(state *terraformer.State) error {
	// If I'm already a provisioner, return
	if p.t != nil {
		return nil
	}

	variables := p.Variables()
	rendered := p.Code()

	// DEBUG
	//fmt.Println(string(rendered))
	//os.Exit(0)

	t, err := utils.NewTerraformer(rendered, variables, state, p.config.ClusterName, "AWS", p.ui)
	if err != nil {
		return err
	}

	t.AddProvider("aws", aws.Provider())
	t.AddProvisioner("file", file.Provisioner())

	p.t = t

	return nil
}

// Plan do the planning of the changes either to create or destroy the cluster on this platform.
func (p *Platform) Plan(destroy bool) (plan *terraformer.Plan, err error) {
	if p.t == nil {
		return nil, fmt.Errorf("cannot get the plan, the %s plaftorm is not a provisioner yet", p.name)
	}

	p.ui.Log.Debug("getting the cluster plan before apply it")
	return p.t.Plan(destroy)
}

// Apply apply the changes either to create or destroy the cluster on this platform
func (p *Platform) Apply(destroy bool) error {
	if p.t == nil {
		return fmt.Errorf("cannot apply the changes, the %s plaftorm is not a provisioner yet", p.name)
	}

	if !destroy {
		p.ui.Log.Debug("starting to provision the cluster")
	} else {
		p.ui.Log.Debug("starting to terminate the cluster")
	}

	return p.t.Apply(destroy)
}

// Provision provisions or creates a cluster on this platform
func (p *Platform) Provision() error {
	if p.t == nil {
		return fmt.Errorf("cannot provision the cluster, the %s plaftorm is not a provisioner yet", p.name)
	}
	return p.t.Apply(false)
}

// Terminate terminates or destroys a cluster on this platform
func (p *Platform) Terminate() error {
	if p.t == nil {
		return fmt.Errorf("cannot terminate the cluster, the %s plaftorm is not a provisioner yet", p.name)
	}

	return p.t.Apply(true)
}

// Code returns the Terraform code to execute
func (p *Platform) Code() []byte {

	var templateContent bytes.Buffer
	var renderedContent bytes.Buffer

	for k, v := range ResourceTemplates {
		templateContent.WriteString(fmt.Sprintf("# section created from template %s\n\n%s\n", k, v))
	}
	tmplFuncMap := template.FuncMap{
		"Join":     strings.Join,
		"Contains": strings.Contains,
		"Dash":     func(s string) string { return strings.NewReplacer("_", "-", ".", "-").Replace(s) },
		"Lower":    func(s string) string { return strings.ToLower(s) },
		"QuoteList": func(s []string) string {
			return fmt.Sprintf(`"%s"`, strings.Join(s, `","`))
		},
		"Trim": strings.TrimSpace,
		"MasterPool": func(pools map[string]NodePool) NodePool {

			// master lookup by label
			for _, pool := range pools {
				for _, label := range pool.KubeletNodeLabels {
					if label == `node-role.kubernetes.io/master=""` {
						return pool
					}
				}
			}

			// fall back, check for named "master" even if label is incorrect
			if master, ok := pools["master"]; ok {
				return master
			}

			// return a default master pool, as it will likely be used just for the name
			return NodePool{
				Name:  "master",
				Count: 1,
			}
		},
		"AllSecGroups": func() []string {
			groupSet := map[string]struct{}{}
			for _, node := range p.config.NodePools {
				for _, group := range node.SecurityGroups {
					groupSet[group] = struct{}{}
				}
			}
			for _, net := range p.config.DefaultNodePool.SecurityGroups {
				groupSet[net] = struct{}{}
			}
			groups := []string{}
			for group := range groupSet {
				groups = append(groups, group)
			}
			return groups
		},
		"Count": func(count int) []int {
			var i int
			var counter []int
			for i = 0; i < (count); i++ {
				counter = append(counter, i)
			}
			return counter
		},
		"isPGStrategy": func(strategy string) bool {
			switch strings.ToLower(strategy) {
			case "cluster", "spread", "partition":
				return true
			}
			return false
		},
		"AllSubNets": func() []string {
			netSet := map[string]struct{}{}
			for _, node := range p.config.NodePools {
				for _, net := range node.Subnets {
					netSet[net] = struct{}{}
				}
			}
			for _, net := range p.config.DefaultNodePool.Subnets {
				netSet[net] = struct{}{}
			}
			nets := []string{}
			for net := range netSet {
				nets = append(nets, net)
			}
			return nets
		},
		"IsFastEphemeral": func(n NodePool) bool {
			for _, label := range n.KubeletNodeLabels {
				if label == "ephemeral-volumes=fast" {
					return true
				}
			}
			return false
		},
	}
	resourceTpl, err := template.
		New("ec2").
		Option("missingkey=error").
		Funcs(tmplFuncMap).
		Parse(templateContent.String())

	if err != nil {
		return []byte(fmt.Sprintf("failed at resourceTpl.New() with %s", err))
	}

	// reload config with default node pool merged in
	// must not altering original config due to write back on config.yaml
	copied := p.config.copyWithDefaults()
	p.reconcileVersion(&copied)

	// future version switch placeholder
	err = resourceTpl.Execute(&renderedContent, copied)

	if err != nil {
		return []byte(fmt.Sprintf("failed at resourceTpl.Execute() with %s\nmap contained: %v", err, p.config))
	}

	if p.t != nil {
		p.t.Code = renderedContent.Bytes()
	}

	return renderedContent.Bytes()

}

// Variables returns the variables as a map where the key is the variable name
// Note: Variables has been reduced to sensative data fields such as credentials
// and private keys. All other values are rendered directly from Config.
func (p *Platform) Variables() map[string]interface{} {
	return map[string]interface{}{
		"aws_access_key": p.config.AwsAccessKey,
		"aws_secret_key": p.config.AwsSecretKey,
		"aws_region":     p.config.AwsRegion,
		"aws_token":      p.config.AwsSessionToken,
		"private_key":    cryptoKey(p.config.PrivateKey),
	}
}

func (p *Platform) reconcileVersion(config *Config) {
	switch p.version {
	case "1.0":
		nodePools := make(map[string]NodePool, len(config.NodePools))
		for key, pool := range config.NodePools {
			switch key {
			case "master", "worker":
				pool.Name = fmt.Sprintf("dumb-%s", key)
				nodePools[key] = pool
			default:
				// currently not honoring new nodepools in 1.0
				// breaks configurator
				// will update in subsequent pr
			}
		}
		config.NodePools = nodePools
	default:
		// do nothing for now
	}
}

func cryptoKey(key string) string {
	if crypto.IsEncrypted(key) {
		if c, err := crypto.New(nil); err == nil {
			if decrypted, err := c.DecryptValue(key); err == nil {
				return string(decrypted)
			}
		}
	}
	return key
}
