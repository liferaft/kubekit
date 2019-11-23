package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-08-01/network"
	"github.com/Azure/go-autorest/autorest/azure"
)

// VMSSClient returns a compute.VirtualMachineScaleSetsClient
func VMSSClient(environment azure.Environment, session *Session) (*compute.VirtualMachineScaleSetsClient, error) {
	client := compute.NewVirtualMachineScaleSetsClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// VMSSClientByEnvStr returns a compute.VirtualMachineScaleSetsClient by looking up the environment by name
func VMSSClientByEnvStr(environmentName string, session *Session) (*compute.VirtualMachineScaleSetsClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return VMSSClient(env, session)
}

// ListVMSSs lists the virtual machine scale sets within a resource group
func ListVMSSs(vmssClient *compute.VirtualMachineScaleSetsClient, resourceGroupName string) (compute.VirtualMachineScaleSetListResultPage, error) {
	return ListVMSSsWithContext(vmssClient, context.Background(), resourceGroupName)
}

// ListVMSSsWithContext lists the virtual machine scale sets within a resource group, but context needs to be provided
func ListVMSSsWithContext(vmssClient *compute.VirtualMachineScaleSetsClient, ctx context.Context, resourceGroupName string) (compute.VirtualMachineScaleSetListResultPage, error) {
	return vmssClient.List(ctx, resourceGroupName)
}

// GetVMSS gets the specified VMSS info
func GetVMSS(vmssClient *compute.VirtualMachineScaleSetsClient, resourceGroupName, vmssName string) (compute.VirtualMachineScaleSet, error) {
	return GetVMSSWithContext(vmssClient, context.Background(), resourceGroupName, vmssName)
}

// GetVMSS gets the specified VMSS info, but context needs to be provided
func GetVMSSWithContext(vmssClient *compute.VirtualMachineScaleSetsClient, ctx context.Context, resourceGroupName, vmssName string) (compute.VirtualMachineScaleSet, error) {
	return vmssClient.Get(ctx, resourceGroupName, vmssName)
}

// UpdateVMSS modifies the VMSS resource by getting it, updating it locally, and putting it back to the server.
func UpdateVMSS(vmssClient *compute.VirtualMachineScaleSetsClient, resourceGroupName, vmssName string, updatedVMSS *compute.VirtualMachineScaleSet) error {
	return UpdateVMSSWithContext(vmssClient, context.Background(), resourceGroupName, vmssName, updatedVMSS)
}

// UpdateVMSSWithContext modifies the VMSS resource by getting it, updating it locally, and putting it back to the server, but context needs to be provided
func UpdateVMSSWithContext(vmssClient *compute.VirtualMachineScaleSetsClient, ctx context.Context, resourceGroupName, vmssName string, updatedVMSS *compute.VirtualMachineScaleSet) error {
	future, err := vmssClient.CreateOrUpdate(ctx, resourceGroupName, vmssName, *updatedVMSS)
	if err != nil {
		return fmt.Errorf("cannot update vmss: %v", err)
	}
	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return fmt.Errorf("cannot get the vmss create or update future response: %v", err)
	}

	if future.Response().StatusCode != http.StatusOK {
		return errors.New("did not receive a successful response from vmss update instances future")
	}

	return nil
}

// UpgradeVMSS applies the updated SKU to the VMSS instances, but context needs to be provided
func UpgradeVMSS(vmssClient *compute.VirtualMachineScaleSetsClient, vmssVMsClient *compute.VirtualMachineScaleSetVMsClient, resourceGroupName, vmssName, filter, selectParameter, expand string) error {
	return UpgradeVMSSWithContext(vmssClient, vmssVMsClient, context.Background(), resourceGroupName, vmssName, filter, selectParameter, expand)
}

// UpgradeVMSSWithContext applies the updated SKU to the VMSS instances, but context needs to be provided
func UpgradeVMSSWithContext(vmssClient *compute.VirtualMachineScaleSetsClient, vmssVMsClient *compute.VirtualMachineScaleSetVMsClient, ctx context.Context, resourceGroupName, vmssName, filter, selectParameter, expand string) error {
	instanceIDs, err := ListVMSSVMIDsWithContext(vmssVMsClient, ctx, resourceGroupName, vmssName, filter, selectParameter, expand)
	if err != nil {
		return fmt.Errorf("issues getting vmss instance ids: %v", err)
	}

	future, err := vmssClient.UpdateInstances(ctx, resourceGroupName, vmssName, *instanceIDs)
	if err != nil {
		return fmt.Errorf("cannot upgrade vmss instances: %v", err)
	}
	err = future.WaitForCompletionRef(ctx, vmssClient.Client)
	if err != nil {
		return fmt.Errorf("cannot get the vmss update instances future response: %v", err)
	}

	if future.Response().StatusCode != http.StatusOK {
		return errors.New("did not receive a successful response from vmss update instances future")
	}

	return nil
}

// ListVMSSPrimaryIPs lists the primary NICs and extracts their primary IPConfiguration and merges it with the public IPs
func ListVMSSPrimaryIPs(nicsClient *network.InterfacesClient, publicIPsClient *network.PublicIPAddressesClient, resourceGroupName, vmssName string) ([]network.Interface, error) {
	var ifaces []network.Interface

	ctx := context.Background()
	nics, err := nicsClient.ListVirtualMachineScaleSetNetworkInterfacesComplete(ctx, resourceGroupName, vmssName) // InterfaceListResultIterator
	if err != nil {
		return ifaces, err
	}

	publicIPs, err := GetVMSSPublicIPsMap(publicIPsClient, ctx, resourceGroupName, vmssName)
	if err != nil {
		return ifaces, err
	}

	for {
		val := nics.Value() // Interface
		if val.Name == nil {
			break
		}
		if !*val.Primary {
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

// ListVMSSNames lists the virtual machine machine scale set names in a resource group
func ListVMSSNames(vmssClient *compute.VirtualMachineScaleSetsClient, resourceGroupName string) ([]string, error) {
	var names []string

	vmsss, err := ListVMSSs(vmssClient, resourceGroupName)
	if err != nil {
		return names, err
	}

	for _, v := range vmsss.Values() {
		names = append(names, *v.Name)
	}

	return names, nil
}
