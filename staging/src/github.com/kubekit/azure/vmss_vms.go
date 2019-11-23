package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
)

// VMSSVMsClient returns a compute.VirtualMachineScaleSetsVMsClient
func VMSSVMsClient(environment azure.Environment, session *Session) (*compute.VirtualMachineScaleSetVMsClient, error) {
	client := compute.NewVirtualMachineScaleSetVMsClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// VMSSVMsClientByEnvStr returns a compute.VirtualMachineScaleSetsClient by looking up the environment by name
func VMSSVMsClientByEnvStr(environmentName string, session *Session) (*compute.VirtualMachineScaleSetVMsClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return VMSSVMsClient(env, session)
}

// ListVMSSVMs lists the virtual machine scale sets virtual machines within a resource group
func ListVMSSVMs(vmssVMsClient *compute.VirtualMachineScaleSetVMsClient, resourceGroupName, vmssName, filter, selectParameter, expand string) (compute.VirtualMachineScaleSetVMListResultPage, error) {
	return ListVMSSVMsWithContext(vmssVMsClient, context.Background(), resourceGroupName, vmssName, filter, selectParameter, expand)
}

// ListVMSSVMsWithContext lists the virtual machine scale sets virtual machines within a resource group, but context needs to be provided
func ListVMSSVMsWithContext(vmssVMsClient *compute.VirtualMachineScaleSetVMsClient, ctx context.Context, resourceGroupName, vmssName, filter, selectParameter, expand string) (compute.VirtualMachineScaleSetVMListResultPage, error) {
	// ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string
	return vmssVMsClient.List(ctx, resourceGroupName, vmssName, filter, selectParameter, expand)
}

// ListVMSSVMIDsWithContext lists the virtual machine scale sets virtual machine IDs within a resource group, but context needs to be provided
func ListVMSSVMIDsWithContext(vmssVMsClient *compute.VirtualMachineScaleSetVMsClient, ctx context.Context, resourceGroupName, vmssName, filter, selectParameter, expand string) (*compute.VirtualMachineScaleSetVMInstanceRequiredIDs, error) {
	instances, err := ListVMSSVMsWithContext(vmssVMsClient, ctx, resourceGroupName, vmssName, filter, selectParameter, expand)
	if err != nil {
		return nil, err
	}

	var instanceIDs []string
	for _, i := range instances.Values() {
		instanceIDs = append(instanceIDs, *i.InstanceID)
	}

	return &compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIds: &instanceIDs,
	}, nil
}
