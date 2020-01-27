#Azure library

This is set to be versioned the same as the Terraform Azure Provider to avoid issues when used in conjunction with: https://github.com/terraform-providers/terraform-provider-azurerm/blob/master/go.mod

At the moment it has limited functionality to only support what is needed, which is the Get List of NICs call.



Standalone build (instead of library):
```
GO111MODULE=on go get
GO111MODULE=on go build
```


Example usage:
```
package main

import (
	"fmt"
	
	"github.com/davecgh/go-spew/spew"
	"github.com/liferaft/azure"
)


var (
	nodeResourceGroup = "my-resource-group-name"
	subscriptionID    = "my-subscription-id"
	tenantID          = "my-tenant-id"
	clientID          = "my-client-id"
	clientSecret      = "my-client-secret"
)

type Node struct {
	PublicIP   string
	PrivateIP  string
	PublicDNS  string
	PrivateDNS string
}

func GetNodes() ([]*Node, error) {
	nodes := []*Node{}
	envName := "public"

	authInfo := &azure.AuthInfo{
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
	}
	session, err := azure.NewSession(authInfo, false)
	if err != nil {
		return nodes, fmt.Errorf("Issues connecting to Azure: %s", err)
	}
	nicsClient, err := azure.NicsClientByEnvStr(envName, session)
	if err != nil {
		return nodes, fmt.Errorf("Issues connecting to Azure via Interface Client: %s", err)
	}
	nicsList, err := azure.ListPrimaryNicInfo(nicsClient, nodeResourceGroup)
	if err != nil {
		return nodes, fmt.Errorf("Issues retrieving network interface info from Azure via Interface Client: %s", err)
	}

	// populate node state
	for _, nic := range nicsList {
		ipConfig := *nic.IPConfigurations

		privateIP := *ipConfig[0].PrivateIPAddress
		privateDNS := ""
		if nic.DNSSettings.InternalFqdn != nil {
			privateDNS = *nic.DNSSettings.InternalFqdn
		}

		publicIP := ""
		publicDNS := ""
		if ipConfig[0].PublicIPAddress != nil {
			publicIPConfig := *ipConfig[0].PublicIPAddress
			if publicIPConfig.IPAddress != nil {
				publicIP = *publicIPConfig.IPAddress
			}
			if publicIPConfig.DNSSettings.Fqdn != nil {
				publicDNS = *publicIPConfig.DNSSettings.Fqdn
			}
		}

		node := &Node{
			PublicIP:   publicIP,
			PrivateIP:  privateIP,
			PublicDNS:  publicDNS,
			PrivateDNS: privateDNS,
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func main() {
	nodes, err := GetNodes()
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}
	spew.Dump(nodes)
}
```
