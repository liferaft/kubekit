package kluster

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kraken/terraformer"
	"github.com/kubekit/kubekit/pkg/configurator"
)

// State represent the final state of one platform. It's basically the
// address:port to access the cluster and the list of nodes
type State struct {
	Status  string                 `json:"status" yaml:"status" mapstructure:"status"`
	Address string                 `json:"address,omitempty" yaml:"address,omitempty" mapstructure:"address,omitempty"`
	Port    int                    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port,omitempty"`
	Nodes   configurator.Hosts     `json:"nodes,omitempty" yaml:"nodes,omitempty" mapstructure:"nodes,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty" mapstructure:"data,omitempty"`
}

// LoadState load the state from the state file for the given platform
func (k *Kluster) LoadState() error {
	platform := k.Platform()

	if _, err := k.makeStateDir(); err != nil {
		return err
	}
	stateFilename := k.StateFile()

	// If there is no state file, assign an empty state. The provisioner should save it later
	if _, err := os.Stat(stateFilename); os.IsNotExist(err) {
		k.ui.Log.Debugf("not found state for cluster %q in %s, using empty state", k.Name, stateFilename)
		// TODO: Should state be set to nil or should be as it is?
		// i.e. if there is an state, should it be nil or keep that state?
		p := k.provisioner[platform]
		return p.BeProvisioner(nil)
	}

	lock, err := lockFile(stateFilename)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	stateBytes, err := ioutil.ReadFile(stateFilename)
	if err != nil {
		return err
	}
	stateBuffer := bytes.NewBuffer(stateBytes)

	state, err := terraformer.LoadState(stateBuffer)
	if err != nil {
		return err
	}

	p := k.provisioner[platform]
	if err := p.BeProvisioner(state); err != nil {
		return err
	}

	if _, ok := k.State[platform]; !ok {
		k.State[platform] = &State{
			Status: AbsentStatus.String(),
		}
	}

	k.State[platform].Address = p.Address()
	k.State[platform].Port = p.Port()
	nodes := configurator.Hosts{}
	for _, h := range p.Nodes() {
		host := configurator.Host{
			PublicDNS:  h.PublicDNS,
			PrivateDNS: h.PrivateDNS,
			PublicIP:   h.PublicIP,
			PrivateIP:  h.PrivateIP,
			RoleName:   h.RoleName,
			Pool:       h.Pool,
		}
		nodes = append(nodes, host)
	}
	k.State[platform].Nodes = nodes

	dataKeys := []string{}
	var creds map[string]string
	switch platform {
	case "eks":
		dataKeys = append(dataKeys, "role-arn", "elastic-fileshares")
		// case "aws":
		// 	dataKeys = append(dataKeys, "elastic-fileshares")
	case "vsphere":
		dataKeys = append(dataKeys, "server", "username", "password", "some-other-shit")
		creds, err = k.GetCredentialsAsMap()
		if err != nil {
			return err
		}
	}
	if len(dataKeys) != 0 {
		k.State[platform].Data = make(map[string]interface{}, len(dataKeys))
		for _, key := range dataKeys {
			switch key {
			case "username", "password", "server":
				k.State[platform].Data[key] = creds[key]
			default:
				k.State[platform].Data[key] = p.Output(key)
			}
		}
	}

	k.ui.Log.Debugf("loaded state for cluster on %s from %s", platform, stateFilename)
	return nil
}

// SaveState saves the state to the state file for the given platform
func (k *Kluster) SaveState() error {
	platform := k.Platform()

	stateDirname, err := k.makeStateDir()
	if err != nil {
		return err
	}

	stateFilename := filepath.Join(stateDirname, platform+".tfstate")

	p := k.provisioner[platform]
	state := p.State()

	if state == nil || state.Empty() {
		k.ui.Log.Debugf("the state of cluster %q is empty, the tfstate file won't be saved", k.Name)
		return nil
	}

	lock, err := lockFile(stateFilename)
	defer lock.Unlock()
	if err != nil {
		return err
	}

	if _, err := os.Stat(stateFilename); os.IsExist(err) {
		os.Rename(stateFilename, stateFilename+".bkp")
		k.ui.Log.Debugf("previous state of cluster %q was backup to %s", k.Name, stateFilename+".bkp")
	}

	var stateBytes bytes.Buffer
	if err := terraformer.SaveState(&stateBytes, state); err != nil {
		return err
	}

	if err := ioutil.WriteFile(stateFilename, stateBytes.Bytes(), 0644); err != nil {
		return err
	}

	k.ui.Log.Debugf("saved state of cluster %q to %s", k.Name, stateFilename)
	return nil
}
