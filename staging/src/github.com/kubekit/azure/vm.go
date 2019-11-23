package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
)

// VirtualMachinesClient returns a compute.VirtualMachinesClient
func VirtualMachinesClient(environment azure.Environment, session *Session) (*compute.VirtualMachinesClient, error) {
	client := compute.NewVirtualMachinesClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// VirtualMachinesClientByEnvStr returns a compute.VirtualMachinesClient by looking up the environment by name
func VirtualMachinesClientByEnvStr(environmentName string, session *Session) (*compute.VirtualMachinesClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return VirtualMachinesClient(env, session)
}

// GetVMs returns VMs from the filtered list of IDs in a resource group
func GetVMs(vmClient *compute.VirtualMachinesClient, resourceGroupName string, vmIDFilterList []string) ([]*compute.VirtualMachine, error) {
	var vms []*compute.VirtualMachine

	ctx := context.Background()
	vmList, err := vmClient.ListComplete(ctx, resourceGroupName)
	if err != nil {
		return vms, err
	}

	vmIDFilter := map[string]struct{}{}
	for _, k := range vmIDFilterList {
		vmIDFilter[k] = struct{}{}
	}

	for {
		val := vmList.Value() // Interface
		if val.ID == nil {
			break
		}
		_, ok := vmIDFilter[*val.ID]
		if len(vmIDFilter) > 0 && !ok {
			continue
		}
		vms = append(vms, &val)

		err = vmList.NextWithContext(ctx)
		if err != nil {
			// if the list has issues getting the next values than we cant advance, so just return with error
			return vms, err
		}
	}

	return vms, nil
}

// UpdateVM updates the specified VM
func UpdateVM(vmClient *compute.VirtualMachinesClient, resourceGroupName, vmName string, update compute.VirtualMachineUpdate) error {
	ctx := context.Background()

	_, err := vmClient.Update(ctx, resourceGroupName, vmName, update)
	if err != nil {
		return err
	}

	return nil
}
