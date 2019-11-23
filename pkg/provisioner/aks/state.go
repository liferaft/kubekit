package aks

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-08-01/network"
	"github.com/hashicorp/terraform/states"

	"github.com/kraken/terraformer"
	"github.com/kubekit/azure"
	"github.com/liferaft/kubekit/pkg/provisioner/state"
)

// State returns the current Terraform state of the cluster
func (p *Platform) State() *terraformer.State {
	if p.t == nil {
		return nil
	}
	return p.t.State
}

// LoadState loads the given Terraform state in a buffer into the terraformer state
func (p *Platform) LoadState(stateBuffer *bytes.Buffer) error {
	if p.t == nil {
		return fmt.Errorf("the %s plaftorm is not a provisioner yet", p.name)
	}

	state, err := terraformer.LoadState(stateBuffer)
	if err != nil {
		return err
	}
	p.t.State = state

	return nil
}

// Output returns a value from the terraform output
func (p *Platform) Output(name string) string {
	// fmt.Printf("[DEBUG] Requesting output value %q\n", name)
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return ""
	}

	output := p.t.State.RootModule().OutputValues
	return state.OutputKeysValueAsStringDefault(output, name, "")
}

// Address returns the address to access the Kubernetes cluster
func (p *Platform) Address() string {
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return ""
	}

	output := p.t.State.RootModule().OutputValues

	return address(output)
}

// return the first host from the state output
func address(output map[string]*states.OutputValue) string {
	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return ""
	}

	// if "host" doesnt work try "fqdn"
	if address, err := state.OutputKeysValueAsString(output, "host"); err == nil {
		if len(address) != 0 {
			u, err := url.Parse(address)
			if err != nil {
				return ""
			}
			return u.Hostname()
		}
	}

	return ""
}

// Port returns the port to access the Kubernetes cluster
func (p *Platform) Port() int {
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return 0
	}

	output := p.t.State.RootModule().OutputValues

	return port(output)
}

func port(output map[string]*states.OutputValue) int {
	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return 0
	}

	// if "host" doesnt work well, try "fqdn"
	if address, err := state.OutputKeysValueAsString(output, "host"); err == nil {
		u, err := url.Parse(address)
		if err != nil {
			return 0
		}
		p := u.Port()
		if len(p) != 0 {
			port, _ := strconv.Atoi(p)
			return port
		}
		switch u.Scheme {
		case "https":
			return 443
		case "http":
			return 8000
		}
	}

	return 0
}

func areMapStringValuesAllNonEmpty(m map[string]string) bool {
	for _, v := range m {
		if v == "" {
			return false
		}
	}
	return true
}

func reformatRGLocation(rgName string) string {
	return strings.ToLower(strings.Replace(rgName, " ", "", -1))
}

func (p *Platform) getNodeStates(nicsList []network.Interface, pool, roleName string) []*state.Node {
	var nodes []*state.Node

	// populate node state
	for _, nic := range nicsList {
		ipConfig := *nic.IPConfigurations

		privateIP := *ipConfig[0].PrivateIPAddress
		privateDNS := ""
		if nic.DNSSettings.InternalFqdn != nil {
			privateDNS = *nic.DNSSettings.InternalFqdn
		}
		if privateDNS == "" {
			var host string
			if nic.VirtualMachine != nil {
				vmID := strings.Split(*nic.VirtualMachine.ID, "/")
				host = vmID[len(vmID)-1]
			} else {
				host = *nic.Name
			}

			domain := *nic.DNSSettings.InternalDomainNameSuffix
			privateDNS = host + "." + domain
		}

		// public IP and DNS are not assigned by default and must be set explicitly
		publicIP := ""
		publicDNS := ""
		if ipConfig[0].PublicIPAddress != nil && ipConfig[0].PublicIPAddress.PublicIPAddressPropertiesFormat != nil {
			publicIPConfig := *ipConfig[0].PublicIPAddress.PublicIPAddressPropertiesFormat
			if publicIPConfig.IPAddress != nil {
				publicIP = *publicIPConfig.IPAddress
			}
			if publicIPConfig.DNSSettings != nil && publicIPConfig.DNSSettings.Fqdn != nil {
				publicDNS = *publicIPConfig.DNSSettings.Fqdn
			}
		}

		p.ui.Log.Debugf("publicIP: " + publicIP)
		p.ui.Log.Debugf("privateIP: " + privateIP)
		p.ui.Log.Debugf("publicDNS: " + publicDNS)
		p.ui.Log.Debugf("privateDNS: " + privateDNS)
		p.ui.Log.Debugf("pool: " + pool)
		p.ui.Log.Debugf("role: " + roleName)

		node := &state.Node{
			PublicIP:   publicIP,
			PrivateIP:  privateIP,
			PublicDNS:  publicDNS,
			PrivateDNS: privateDNS,
			Pool:       pool,
			RoleName:   roleName,
		}
		nodes = append(nodes, node)
	}

	return nodes
}

// Nodes return the list of nodes provisioned. It took the value from the
func (p *Platform) Nodes() []*state.Node {
	var nodes []*state.Node

	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return nodes
	}

	output := p.t.State.RootModule().OutputValues

	defaultNodeResourceGroup := "MC_" + p.config.ClusterName + "_" + p.config.ClusterName + "_" + reformatRGLocation(p.config.ResourceGroupLocation)
	nodeResourceGroup := state.OutputKeysValueAsStringDefault(output, "node_resource_group", defaultNodeResourceGroup)

	p.config.KubernetesVersion = state.OutputKeysValueAsStringDefault(output, "kubernetes_version", "")
	p.config.VnetName = state.OutputKeysValueAsStringDefault(output, "vnet", "")

	if len(p.config.ClusterClientID) == 0 {
		p.config.ClusterClientID = p.config.ClientID
	}
	if len(p.config.ClusterClientSecret) == 0 {
		p.config.ClusterClientSecret = p.config.ClientSecret
	}

	if registryHost, err := state.OutputKeysValueAsString(output, "container_registry_server"); err == nil {
		p.config.ContainerRegistryHost = Registry{
			Host:     registryHost,
			Username: p.config.ClientID,
			Password: p.config.ClientSecret,
		}
	}

	// connect to Azure to get IP and DNS info since Terraform doesnt provide it for us
	authInfo := &azure.AuthInfo{
		SubscriptionID: p.config.SubscriptionID,
		TenantID:       p.config.TenantID,
		ClientID:       p.config.ClusterClientID,
		ClientSecret:   p.config.ClusterClientSecret,
	}
	session, err := azure.NewSession(authInfo, false)
	if err != nil {
		panic(fmt.Errorf("issues connecting to Azure: %s", err))
	}
	nicsClient, err := azure.NicsClientByEnvStr(p.config.Environment, session)
	if err != nil {
		panic(fmt.Errorf("issues connecting to Azure via Interface Client: %s", err))
	}
	publicIPsClient, err := azure.PublicIPAddressesClientByEnvStr(p.config.Environment, session)
	if err != nil {
		panic(fmt.Errorf("issues connecting to Azure via Public IP Addresses Client: %s", err))
	}
	vmssClient, err := azure.VMSSClientByEnvStr(p.config.Environment, session)
	if err != nil {
		panic(fmt.Errorf("issues connecting to Azure via VMSS Client: %s", err))
	}

	if jumpbox, err := state.OutputKeysValueAsString(output, "jumpbox"); err == nil {
		nicsList, err := azure.ListPrimaryIPsInfo(nicsClient, publicIPsClient, p.config.ClusterName)
		if err != nil {
			if strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
				p.ui.Log.Infof("The '%s' resource group was not found. Skipping listing of NICs.", p.config.ClusterName)
			} else {
				panic(fmt.Errorf("issues retrieving network interface info from Azure via Interface Client: %s", err))
			}
		}

		// filter jumpbox in case there are other VMs in the resource group
		for _, n := range p.getNodeStates(nicsList, "", "jumpbox") {
			if n.PrivateIP == jumpbox {
				nodes = append(nodes, n)
				break
			}
		}
	}

	// workers can be from availability sets or scale sets, so we check for both

	availabilitySetWorkersNICList, err := azure.ListPrimaryIPsInfo(nicsClient, publicIPsClient, nodeResourceGroup)
	if err != nil {
		if strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
			p.ui.Log.Infof("The '%s' resource group was not found. Skipping listing of NICs.", nodeResourceGroup)
			return nodes
		}
		panic(fmt.Errorf("issues retrieving network interface info from Azure via Interface Client: %s", err))
	}
	nodes = append(nodes, p.getNodeStates(availabilitySetWorkersNICList, "", "worker")...)

	vmssNames, err := azure.ListVMSSNames(vmssClient, nodeResourceGroup)
	if err != nil {
		if strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
			p.ui.Log.Infof("The '%s' resource group was not found. Skipping listing of VMSSes.", nodeResourceGroup)
			return nodes
		}
		panic(fmt.Errorf("issues retrieving virtual machine scale set names from Azure via VMSS Client: %s", err))
	}

	for _, vmssName := range vmssNames {
		vmssWorkersNICList, err := azure.ListVMSSPrimaryIPs(nicsClient, publicIPsClient, nodeResourceGroup, vmssName)
		if err != nil {
			if strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
				p.ui.Log.Infof("The '%s' resource group was not found. Skipping listing of NICs.", nodeResourceGroup)
				return nodes
			}
			panic(fmt.Errorf("issues retrieving network interface info from Azure via Interface Client: %s", err))
		}
		nodes = append(nodes, p.getNodeStates(vmssWorkersNICList, vmssName, "worker")...)
	}

	return nodes
}
