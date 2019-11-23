package kluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// AzureCredentials represents the credentials just for AWS
type AzureCredentials struct {
	Platform       string `json:"platform" yaml:"platform" toml:"platform" mapstructure:"platform" env:"-"`
	SubscriptionID string `json:"subscription_id" yaml:"subscription_id" mapstructure:"subscription_id" env:"SUBSCRIPTION_ID"`
	TenantID       string `json:"tenant_id" yaml:"tenant_id" mapstructure:"tenant_id" env:"TENANT_ID"`
	ClientID       string `json:"client_id" yaml:"client_id" mapstructure:"client_id" env:"CLIENT_ID"`
	ClientSecret   string `json:"client_secret" yaml:"client_secret" mapstructure:"client_secret" env:"CLIENT_SECRET"`
	cluster        string
	path           string
}

// NewAzureCredentials creates an struct ready for Azure credentials
func NewAzureCredentials(clustername, path string) *AzureCredentials {
	return &AzureCredentials{
		Platform: "Azure",
		cluster:  clustername,
		path:     path,
	}
}

func (c *AzureCredentials) clusterName() string {
	return c.cluster
}

func (c *AzureCredentials) clusterPath() string {
	return c.path
}

// SetPath sets the path of the credentials file be stored
func (c *AzureCredentials) SetPath(path string) {
	c.path = path
}

func (c *AzureCredentials) platform() string {
	return c.Platform
}

func (c *AzureCredentials) parameters() []string {
	return []string{c.SubscriptionID, c.TenantID, c.ClientID, c.ClientSecret}
}

func (c *AzureCredentials) asMap() map[string]string {
	return map[string]string{
		"subscription_id": c.SubscriptionID,
		"tenant_id":       c.TenantID,
		"client_id":       c.ClientID,
		"client_secret":   c.ClientSecret,
	}
}

// SetParameters sets the credentials parameters
func (c *AzureCredentials) SetParameters(params ...string) error {
	if len(params) != 4 {
		return fmt.Errorf("incorrect number of parameters, expecting 4 and got %d", len(params))
	}
	c.SubscriptionID = params[0]
	c.TenantID = params[1]
	c.ClientID = params[2]
	c.ClientSecret = params[3]
	return nil
}

// AssignFromMap sets the credentials parameters from a map. The key is the name
// of the parameter as defined in the json metadata of the AzureCredentials structure
func (c *AzureCredentials) AssignFromMap(params map[string]string) error {
	return assignFromMap(params, c)
}

// Empty returns true if there isn't any credentials set
func (c *AzureCredentials) Empty() bool {
	return len(c.SubscriptionID) == 0 && len(c.TenantID) == 0 && len(c.ClientID) == 0 && len(c.ClientSecret) == 0
}

// Complete returns true if all credentials are set. Use it as !Complete() to
// know if there is a missing parameter. !Complete() is not Empty()
func (c *AzureCredentials) Complete() bool {
	return len(c.SubscriptionID) != 0 && len(c.TenantID) != 0 && len(c.ClientID) != 0 && len(c.ClientSecret) != 0
}

// Getenv gets the Azure credentials from environment variables
func (c *AzureCredentials) Getenv(force bool) error {
	return parseFn(c,
		func(key string) string {
			tmp := os.Getenv(strings.ToUpper(c.platform()) + "_" + key)
			if tmp == "" {
				tmp = os.Getenv("ARM_" + key) // Terraform Azure environment variable
			}
			return tmp
		},
		func(k, v string) {},
		force,
	)
}

func (c *AzureCredentials) setenv() error {
	return parseFn(c,
		func(key string) string {
			return ""
		},
		func(key, value string) {
			os.Setenv(strings.ToUpper(c.platform())+"_"+key, value)
		},
		true,
	)
}

// List prints to stdout the AWS credentials in table format
func (c *AzureCredentials) List() {
	header := "Name\tPlatform\tSubscription ID\tTenant ID\tClient ID\tClient Secret" //\tLocation"
	subscriptionID := c.SubscriptionID
	if len(subscriptionID) == 0 {
		subscriptionID = emptyValue
	}
	tenantID := c.TenantID
	if len(tenantID) == 0 {
		tenantID = emptyValue
	}
	clientID := c.ClientID
	if len(clientID) == 0 {
		clientID = emptyValue
	} else {
		clientID = filter(clientID, truncLen, false)
	}
	clientSecret := c.ClientSecret
	if len(clientSecret) == 0 {
		clientSecret = emptyValue
	} else {
		clientSecret = filter(clientSecret, 0, true)
	}
	row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", c.cluster, c.Platform, subscriptionID, tenantID, clientID, clientSecret) //, c.path)
	printCredentials(header, row)
}

// Ask the user from stdin to get the Azure credentials suggesting values from the environment
func (c *AzureCredentials) Ask() error {
	return parseFn(c,
		func(key string) string {
			param := strings.Replace(strings.ToLower(key), "_", " ", -1)
			title := fmt.Sprintf("%s %s", c.Platform, param)
			env := os.Getenv(strings.ToUpper(c.platform()) + "_" + key)
			v, _ := AskDefault(title, env, true)
			return v
		},
		func(k, v string) {},
		true,
	)
}

// Read reads the Azure credentials from the cluster credentials file
func (c *AzureCredentials) Read() error {
	credentialsBytes, err := read(c.path)
	if err != nil || credentialsBytes == nil {
		return err
	}
	return yaml.Unmarshal(credentialsBytes, &c)
}

// Write writes the Azure credentials to the cluster credentials file
func (c *AzureCredentials) Write() error {
	credentialsBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.path, credentialsBytes, 0600)
}
