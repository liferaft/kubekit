package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/states"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"github.com/kraken/terraformer"
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
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return ""
	}

	mod := p.t.State.RootModule()
	// mod shouldn't be null, there's no need to check it's nil
	output := mod.OutputValues

	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return ""
	}

	if _, ok := output[name]; !ok {
		return ""
	}

	value, _ := state.ValueAsString(output[name])
	if len(value) == 0 {
		return ""
	}

	return value
}

// Address returns the address to access the Kubernetes cluster
func (p *Platform) Address() string {
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return ""
	}

	mod := p.t.State.RootModule()
	// mod shouldn't be null, there's no need to check it's nil
	output := mod.OutputValues

	return address(output)
}

// return the service IP (1st master IP) from the state output
func address(output map[string]*states.OutputValue) string {
	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return ""
	}

	if addressOutput, ok := output["service_ip"]; ok {
		address, _ := state.ValueAsString(addressOutput)
		if len(address) != 0 {
			// return the service_ip (master[0] public ip) if it is there
			return address
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

	return port(p.t.State)
}

func port(st *terraformer.State) int {
	mod := st.RootModule()
	// mod shouldn't be null, there's no need to check it's nil
	output := mod.OutputValues

	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return 0
	}

	// TODO: Verify this is correct

	if portOutput, ok := output["service_port"]; ok {
		port, _ := state.ValueAsString(portOutput)
		if len(port) != 0 {
			// return the VIP API Port if it is there
			p, _ := strconv.Atoi(port)
			return p
		}
	}

	return 0
}

// Nodes returns the list of provisioned nodes in the current terraform state
func (p *Platform) Nodes() []*state.Node {
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		p.ui.Log.Debugf("the provisioner or state for %s doesn't exists", p.Name())
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return []*state.Node{}
	}

	output := p.t.State.RootModule().OutputValues
	nodes := []*state.Node{}

	if marshalledNodes, ok := output["nodes"]; ok {
		for _, nodeValue := range marshalledNodes.Value.AsValueSlice() {
			node := &state.Node{}
			jsonVal, err := ctyjson.Marshal(nodeValue, nodeValue.Type())
			jsonValStr := strings.Replace(string(jsonVal), `\`, "", -1)
			jsonValStr = strings.Trim(jsonValStr, `"`)
			// p.ui.Log.Debugf("Node from state: %s", jsonValStr)
			if err != nil || len(jsonVal) == 0 {
				continue
			}
			json.Unmarshal([]byte(jsonValStr), &node)
			if node == nil {
				continue
			}
			nodes = append(nodes, node)

			p.ui.Log.Debugf(fmt.Sprintf("publicIP: %s", node.PublicIP))
			p.ui.Log.Debugf(fmt.Sprintf("privateIP: %s", node.PrivateIP))
			p.ui.Log.Debugf(fmt.Sprintf("publicDNS: %s", node.PublicDNS))
			p.ui.Log.Debugf(fmt.Sprintf("privateDNS: %s", node.PrivateDNS))
			p.ui.Log.Debugf(fmt.Sprintf("pool: %s", node.Pool))
			p.ui.Log.Debugf(fmt.Sprintf("role: %s", node.RoleName))
		}
	}
	return nodes
}
