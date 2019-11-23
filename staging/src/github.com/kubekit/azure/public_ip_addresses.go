package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-08-01/network"
	"github.com/Azure/go-autorest/autorest/azure"
)

// PublicIPAddressesClient returns a network.PublicIPAddressesClient
func PublicIPAddressesClient(environment azure.Environment, session *Session) (*network.PublicIPAddressesClient, error) {
	client := network.NewPublicIPAddressesClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// PublicIPAddressesClientByEnvStr returns a network.PublicIPAddressesClient by looking up the environment by name
func PublicIPAddressesClientByEnvStr(environmentName string, session *Session) (*network.PublicIPAddressesClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return PublicIPAddressesClient(env, session)
}

// GetPublicIPsMap returns a map of map[string]*network.PublicIPAddress where the keys are the names of the public ips
func GetPublicIPsMap(publicIPsClient *network.PublicIPAddressesClient, ctx context.Context, resourceGroupName string) (map[string]*network.PublicIPAddress, error) {
	publicIPsMap := make(map[string]*network.PublicIPAddress)

	publicIPs, err := publicIPsClient.ListComplete(ctx, resourceGroupName)
	if err != nil {
		return nil, err
	}

	for {
		val := publicIPs.Value()
		if val.Name == nil { // we reached the end
			break
		}

		publicIPsMap[*val.Name] = &val

		err = publicIPs.NextWithContext(ctx)
		if err != nil {
			return publicIPsMap, err
		}
	}

	return publicIPsMap, nil
}

// GetVMSSPublicIPsMap returns a map of map[string]*network.PublicIPAddress where the keys are the names of the public ips
func GetVMSSPublicIPsMap(publicIPsClient *network.PublicIPAddressesClient, ctx context.Context, resourceGroupName, vmssName string) (map[string]*network.PublicIPAddress, error) {
	publicIPsMap := make(map[string]*network.PublicIPAddress)

	publicIPs, err := publicIPsClient.ListVirtualMachineScaleSetPublicIPAddressesComplete(ctx, resourceGroupName, vmssName)
	if err != nil {
		return nil, err
	}

	for {
		val := publicIPs.Value()
		if val.Name == nil { // we reached the end
			break
		}
		publicIPsMap[*val.Name] = &val

		err = publicIPs.NextWithContext(ctx)
		if err != nil {
			return publicIPsMap, err
		}
	}

	return publicIPsMap, nil
}
