package aks

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm"

	"github.com/kraken/terraformer"
	"github.com/kubekit/azure"
	"github.com/liferaft/kubekit/pkg/crypto"
	"github.com/liferaft/kubekit/pkg/provisioner/state"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// ResourceTemplates maps resource names to content of resources
// implementation specified in code.go
var ResourceTemplates map[string]string

type publicSettings struct {
	//CommandToExecute string   `json:"commandToExecute"`
	//FileURLs         []string `json:"fileUris"`
	Script    string `json:"script"`
	Timestamp int    `json:"timestamp"`
}

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

	p.ui.Log.Debugf("Variables: %+v", variables)

	t, err := utils.NewTerraformer(rendered, variables, state, p.config.ClusterName, "AKS", p.ui)
	if err != nil {
		return err
	}

	t.AddProvider("azurerm", azurerm.Provider())

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

// PreApply applies changes before the changes done in the Apply method
func (p *Platform) PreApply(destroy bool) error {
	if destroy {
		return nil
	}

	// enable preview features
	return p.setupPreviewFeatures()
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

	if err := p.PreApply(destroy); err != nil {
		return err
	}

	if err := p.t.Apply(destroy); err != nil {
		return err
	}

	if err := p.PostApply(destroy); err != nil {
		return err
	}

	return nil
}

// Provision provisions or creates a cluster on this platform
func (p *Platform) Provision() error {
	if p.t == nil {
		return fmt.Errorf("cannot provision the cluster, the %s plaftorm is not a provisioner yet", p.name)
	}
	return p.t.Apply(false)
}

// PostApply applies changes after the changes done in the Apply method
func (p *Platform) PostApply(destroy bool) error {
	if destroy {
		return nil
	}

	authInfo := &azure.AuthInfo{
		SubscriptionID: p.config.SubscriptionID,
		TenantID:       p.config.TenantID,
		ClientID:       p.config.ClientID,
		ClientSecret:   p.config.ClientSecret,
	}
	session, err := azure.NewSession(authInfo, false)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure: %s", err)
	}

	output := p.t.State.RootModule().OutputValues

	defaultNodeResourceGroup := "MC_" + p.config.ClusterName + "_" + p.config.ClusterName + "_" + reformatRGLocation(p.config.ResourceGroupLocation)
	nodeResourceGroup := state.OutputKeysValueAsStringDefault(output, "node_resource_group", defaultNodeResourceGroup)

	if err := p.setupJumpbox(session, output, nodeResourceGroup); err != nil {
		return err
	}

	if err := p.setupAvailabilitySets(session, output, nodeResourceGroup); err != nil {
		return err
	}

	if err := p.setupScaleSets(session, output, nodeResourceGroup); err != nil {
		return err
	}

	return nil
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

	alphanumericRegex, err := regexp.Compile(`[^A-Za-z0-9]+`)
	if err != nil {
		panic(fmt.Errorf("alphanumeric regular expression failed to compile: %s", err))
	}

	alphanumericHyphenRegex, err := regexp.Compile(`[^A-Za-z0-9\-]+`)
	if err != nil {
		panic(fmt.Errorf("alphanumeric with hyphen regular expression failed to compile: %s", err))
	}

	tmplFuncMap := template.FuncMap{
		"Join": strings.Join,
		"Trim": strings.TrimSpace,
		"Alphanumeric": func(s string) string {
			return alphanumericRegex.ReplaceAllString(s, "")
		},
		"AlphanumericHyphen": func(s string) string {
			return alphanumericHyphenRegex.ReplaceAllString(s, "")
		},
		"BoolToString": strconv.FormatBool,
		"QuoteList": func(s []string) string {
			return fmt.Sprintf(`"%s"`, strings.Join(s, `","`))
		},
		"DefaultString": func(s string, def string) string {
			if s == "" {
				return def
			}
			return s
		},
		"DefaultInt": func(i int, def int) int {
			if i == 0 {
				return def
			}
			return i
		},
		"Multiply": func(x int, y int) int { return x * y },
		"Dash":     func(s string) string { return strings.NewReplacer("_", "-", ".", "-").Replace(s) },
		"Lower":    func(s string) string { return strings.ToLower(s) },
	}

	resourceTpl, err := template.
		New("aks").
		Option("missingkey=error").
		Funcs(tmplFuncMap).
		Parse(templateContent.String())

	if err != nil {
		return []byte(fmt.Sprintf("failed at resourceTpl.New() with %s", err))
	}

	// reload config with default node pool merged in
	// must not altering original config due to write back on config.yaml
	copied := p.config.copyWithDefaults()

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
func (p *Platform) Variables() map[string]interface{} {
	return map[string]interface{}{
		"subscription_id": p.config.SubscriptionID,
		"tenant_id":       p.config.TenantID,
		"client_id":       p.config.ClientID,
		"client_secret":   p.config.ClientSecret,
	}
}

func decryptKey(key string) (string, error) {
	if !crypto.IsEncrypted(string(key)) {
		// If the private key is not encrypted, there's nothing to do
		return key, nil
	}

	// If we are here, means the key is encrypted ...
	c, err := crypto.New(nil)
	if err != nil {
		return "", err
	}
	// ... decrypt it ...
	decryptedKey, err := c.DecryptValue(key)
	if err != nil {
		return "", err
	}

	// ... and assign it
	return string(decryptedKey), nil
}
