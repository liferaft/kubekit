package ec2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/states"
	"github.com/kraken/terraformer"
	"github.com/liferaft/kubekit/pkg/provisioner/state"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// State returns the current Terraform state of the cluster
func (p *Platform) State() *terraformer.State {
	if p.t == nil {
		return nil
	}
	return p.t.State
}

// PersistStateToFile makes the state to persist in a file and be up to date all
// the time. Every time the state changes Terraformer will update the file
func (p *Platform) PersistStateToFile(filename string) error {
	if p.t == nil {
		return nil
	}
	return p.t.PersistStateToFile(filename)
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

	value, err := state.ValueAsString(output[name])
	if err != nil {
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

// return the ALB DNS Name from the state output
func address(output map[string]*states.OutputValue) string {
	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return ""
	}

	if addressOutput, ok := output["alb_dns"]; ok {
		address, _ := state.ValueAsString(addressOutput)
		if len(address) != 0 {
			// return the ALB DNS if it is there
			return address
		}
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

func port(state *terraformer.State) int {
	mod := state.RootModule()
	// mod shouldn't be null, there's no need to check it's nil
	output := mod.OutputValues

	if output == nil {
		// TODO
		// there is no point to check the resources if output is nil, right?
		return 0
	}

	// TODO: Verify this is correct

	if portOutput, ok := output["kube_vip_api_ssl_port"]; ok {
		jsonPortStr, err := ctyjson.Marshal(portOutput.Value, portOutput.Value.Type())
		if err != nil {
			return 0
		}
		p, _ := strconv.Atoi(string(jsonPortStr))
		return p
	}

	if portOutput, ok := output["kube_api_ssl_port"]; ok {
		jsonPortStr, err := ctyjson.Marshal(portOutput.Value, portOutput.Value.Type())
		if err != nil {
			return 0
		}
		p, _ := strconv.Atoi(string(jsonPortStr))
		return p
	}

	// // if the parameters are not in the output (they aren't now), get them from the resources
	// resources := mod.Resources

	// if listener, ok := resources["aws_alb_listener.kube_vip_api_ssl_port"]; ok {
	// 	// return the VIP API Port attribute in the resources if it is there
	// 	if port, okattr := listener.Primary.Attributes["port"]; okattr {
	// 		p, _ := strconv.Atoi(port)
	// 		return p
	// 	}
	// }

	// if listener, ok := resources["aws_alb_listener.kube_api_ssl_port"]; ok {
	// 	// return the API Port attribute in the resources if it is there
	// 	if port, okattr := listener.Primary.Attributes["port"]; okattr {
	// 		p, _ := strconv.Atoi(port)
	// 		return p
	// 	}
	// }

	return 0
}

// Nodes return the list of nodes provisioned. It took the value from the
func (p *Platform) Nodes() []*state.Node {
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
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
