package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-08-01/network"
	"github.com/Azure/go-autorest/autorest/azure"
)

// NicsClient returns a network.InterfacesClient
func NicsClient(environment azure.Environment, session *Session) (*network.InterfacesClient, error) {
	client := network.NewInterfacesClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// NicsClientByEnvStr returns a network.InterfacesClient by looking up the environment by name
func NicsClientByEnvStr(environmentName string, session *Session) (*network.InterfacesClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return NicsClient(env, session)
}

// ListPrimaryIPsInfo lists the primary NICs and extracts their primary IPConfiguration and merges it with the public IPs
func ListPrimaryIPsInfo(nicsClient *network.InterfacesClient, publicIPsClient *network.PublicIPAddressesClient, resourceGroupName string) ([]network.Interface, error) {
	var ifaces []network.Interface

	ctx := context.Background()
	nics, err := nicsClient.ListComplete(ctx, resourceGroupName) // InterfaceListResultIterator
	if err != nil {
		return ifaces, err
	}

	publicIPs, err := GetPublicIPsMap(publicIPsClient, ctx, resourceGroupName)
	if err != nil {
		return ifaces, err
	}

	for {
		val := nics.Value() // Interface
		if val.Name == nil {
			break
		}
		if val.Primary != nil && !*val.Primary {
			continue
		}

		var primaryIP *network.InterfaceIPConfiguration
		for _, ipConfig := range *val.IPConfigurations { // *[]InterfaceIPConfiguration
			if *ipConfig.Primary {
				primaryIP = &ipConfig
				primaryIP.PublicIPAddress = publicIPs[*val.Name]
				break
			}
		}

		if primaryIP != nil {
			val.IPConfigurations = &[]network.InterfaceIPConfiguration{*primaryIP}

			ifaces = append(ifaces, val)
		}

		err = nics.NextWithContext(ctx)
		if err != nil {
			// if the list has issues getting the next values than we cant advance, so just return with error
			return ifaces, err
		}
	}

	return ifaces, nil
}
