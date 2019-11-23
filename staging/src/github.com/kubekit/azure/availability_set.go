package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
)

// AvailabilitySetClient returns a compute.AvailabilitySetsClient
func AvailabilitySetClient(environment azure.Environment, session *Session) (*compute.AvailabilitySetsClient, error) {
	client := compute.NewAvailabilitySetsClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// AvailabilitySetClientByEnvStr returns a compute.AvailabilitySetsClient by looking up the environment by name
func AvailabilitySetClientByEnvStr(environmentName string, session *Session) (*compute.AvailabilitySetsClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return AvailabilitySetClient(env, session)
}

// GetVMIDsFromAvailabilitySets returns VM IDs from the filtered availability sets in a resource group
func GetVMIDsFromAvailabilitySets(availSetClient *compute.AvailabilitySetsClient, resourceGroupName string, availSetNameFilterList []string) (map[string][]string, error) {
	vms := map[string][]string{}

	ctx := context.Background()
	availSets, err := availSetClient.ListComplete(ctx, resourceGroupName)
	if err != nil {
		return vms, err
	}

	availSetNameFilter := map[string]struct{}{}
	for _, k := range availSetNameFilterList {
		availSetNameFilter[k] = struct{}{}
	}

	for {
		val := availSets.Value() // Interface
		if val.Name == nil {
			break
		}
		_, ok := availSetNameFilter[*val.Name]
		if len(availSetNameFilter) > 0 && !ok {
			continue
		}
		if val.AvailabilitySetProperties.VirtualMachines == nil {
			continue
		}

		vms[*val.Name] = []string{}
		for _, subresource := range *val.AvailabilitySetProperties.VirtualMachines {
			vms[*val.Name] = append(vms[*val.Name], *subresource.ID)
		}

		err = availSets.NextWithContext(ctx)
		if err != nil {
			// if the list has issues getting the next values than we cant advance, so just return with error
			return vms, err
		}
	}

	return vms, nil
}
