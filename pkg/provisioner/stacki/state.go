package stacki

import (
	"bytes"

	"github.com/kraken/terraformer"
	"github.com/kubekit/kubekit/pkg/provisioner/state"
)

// State returns the current Terraform state of the cluster
func (p *Platform) State() *terraformer.State {
	p.ui.Log.Debugf("%s platform do not implements State()", p.name)
	return nil
}

// LoadState loads the given Terraform state in a buffer into the terraformer state
func (p *Platform) LoadState(stateBuffer *bytes.Buffer) error {
	p.ui.Log.Debugf("%s platform do not implements LoadState()", p.name)
	return nil
}

// Address returns the address to access the Kubernetes cluster
func (p *Platform) Address() string {
	if !p.config.DisableMasterHA && p.config.KubeVirtualIPApi != "" {
		return p.config.KubeVirtualIPApi
	}
	if p.config.APIAddress != "" {
		return p.config.APIAddress
	}
	return p.config.NodePools["master"].Nodes[0].PublicIP
}

// Port returns the port to access the Kubernetes cluster
func (p *Platform) Port() int {
	if !p.config.DisableMasterHA && p.config.KubeVirtualIPApi != "" && p.config.KubeVIPAPISSLPort != 0 {
		return p.config.KubeVIPAPISSLPort
	}
	return p.config.KubeAPISSLPort
}

// Output returns a value from the terraform output
func (p *Platform) Output(name string) string {
	// Returns empty because this platform does have a terraform state
	return ""
}

// Nodes return the list of nodes provisioned. It took the value from the
func (p *Platform) Nodes() []*state.Node {

	stateNodes := make([]*state.Node, 0)
	for roleName, nodePool := range p.config.NodePools {
		for _, node := range nodePool.Nodes {
			stateNode := &state.Node{
				PublicIP:   node.PublicIP,
				PrivateIP:  node.PrivateIP,
				PublicDNS:  node.PublicDNS,
				PrivateDNS: node.PrivateDNS,
				RoleName:   roleName,
			}
			stateNodes = append(stateNodes, stateNode)
		}
	}
	return stateNodes
}
