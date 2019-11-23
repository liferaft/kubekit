package eks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
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
	// fmt.Printf("[DEBUG] Requesting output value %q\n", name)
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
	// fmt.Printf("[DEBUG] Output value of %q = %s\n", name, value)

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

	if endpointOutput, ok := output["endpoint"]; ok {
		address, _ := state.ValueAsString(endpointOutput)
		if len(address) != 0 {
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

	if endpointOutput, ok := output["endpoint"]; ok {
		address, err := state.ValueAsString(endpointOutput)
		if err != nil {
			return 0
		}
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
			return 80
		}
	}

	return 0
}

// Nodes return the list of nodes provisioned. It took the value from the
func (p *Platform) Nodes() []*state.Node {
	if p.t == nil || p.t.State == nil || p.t.State.Empty() {
		// If I'm not a provisioner yet, or the state is null/empty, return no address
		return []*state.Node{}
	}

	output := p.t.State.RootModule().OutputValues

	// if all pools using the same image, set default for config
	amiSet := map[string]struct{}{}
	for k, v := range p.config.NodePools {
		name := fmt.Sprintf("%s-ami", strings.NewReplacer("_", "-", ".", "-").Replace(strings.ToLower(k)))
		if ami, ok := output[name]; ok {
			if amiStr, _ := state.ValueAsString(ami); len(amiStr) != 0 {
				v.AwsAmi = amiStr
			}
		}
		amiSet[v.AwsAmi] = struct{}{}
		p.config.NodePools[k] = v
	}
	if len(amiSet) == 1 {
		// set defaultNodePool AMI in config
		for k := range amiSet {
			p.config.DefaultNodePool.AwsAmi = k
		}
	}

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
