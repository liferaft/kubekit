package state

// Node represent a node created by the provisioner and should be in the
// terraform output
type Node struct {
	PublicIP   string `json:"public_ip" yaml:"public_ip" mapstructure:"public_ip"`
	PrivateIP  string `json:"private_ip" yaml:"private_ip" mapstructure:"private_ip"`
	PublicDNS  string `json:"public_dns" yaml:"public_dns" mapstructure:"public_dns"`
	PrivateDNS string `json:"private_dns" yaml:"private_dns" mapstructure:"private_dns"`
	RoleName   string `json:"role" yaml:"role" mapstructure:"role"`
	Pool       string `json:"pool" yaml:"pool" mapstructure:"pool"`
}

// NewNodeFromAttr creates a node from the sttributes found in the TF state file
func NewNodeFromAttr(attr map[string]string) *Node {
	node := Node{}
	if val, ok := attr["private_ip"]; ok {
		node.PrivateIP = val
	}
	if val, ok := attr["public_ip"]; ok {
		node.PublicIP = val
	}
	if val, ok := attr["private_dns"]; ok {
		node.PrivateDNS = val
	}
	if val, ok := attr["public_dns"]; ok {
		node.PublicDNS = val
	}
	if val, ok := attr["role"]; ok {
		node.RoleName = val
	}
	if val, ok := attr["pool"]; ok {
		node.Pool = val
	}
	return &node
}
