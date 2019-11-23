package raw

import (
	"github.com/kraken/terraformer"
)

// BeProvisioner setup the Plaftorm to be a Provisioner
func (p *Platform) BeProvisioner(state *terraformer.State) error {
	p.ui.Log.Debugf("%s platform do not implements BeProvisioner()", p.name)
	return nil
}

// Plan do the planning of the changes either to create or destroy the cluster on this platform.
func (p *Platform) Plan(destroy bool) (plan *terraformer.Plan, err error) {
	p.ui.Log.Debugf("%s platform do not implements Plan()", p.name)
	return nil, nil
}

// Apply apply the changes either to create or destroy the cluster on this platform
func (p *Platform) Apply(destroy bool) error {
	p.ui.Log.Debugf("%s platform do not implements Apply()", p.name)
	return nil
}

// Provision provisions or creates a cluster on this platform
func (p *Platform) Provision() error {
	p.ui.Log.Debugf("%s platform do not implements Provision()", p.name)
	return nil
}

// Terminate terminates or destroys a cluster on this platform
func (p *Platform) Terminate() error {
	p.ui.Log.Debugf("%s platform do not implements Terminate()", p.name)
	return nil
}

// Code returns the Terraform code to execute
func (p *Platform) Code() []byte {
	p.ui.Log.Debugf("%s platform do not implements Code()", p.name)
	return []byte{}
}

// Variables returns the variables as a map where the key is the variable name
func (p *Platform) Variables() map[string]interface{} {
	p.ui.Log.Debugf("%s platform do not implements Variables()", p.name)
	return nil
}
