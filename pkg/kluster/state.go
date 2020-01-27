package kluster

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform/states"
	"github.com/kraken/terraformer"
	"github.com/liferaft/kubekit/pkg/configurator"
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

	var state *states.State

	if _, err := os.Stat(stateFilename); !os.IsNotExist(err) {
		k.ui.Log.Debugf("Loading state from state file %q", stateFilename)
		// If there is a state file, read it to assign it to the provisioner
		stateBytes, err := ioutil.ReadFile(stateFilename)
		if err != nil {
			return err
		}
		stateBuffer := bytes.NewBuffer(stateBytes)

		state, err = terraformer.LoadState(stateBuffer)
		if err != nil {
			return fmt.Errorf("can't load the state from %q. %s", stateFilename, err)
		}
	}

	p := k.provisioner[platform]
	if err := p.BeProvisioner(state); err != nil {
		return err
	}

	// If there is a state file:
	// 		- the function PersistStateToFile() create a backup of the state file
	// 		- saves the current state (previously loaded from the file) to the state file
	// 		- make sure the state is always up dated in the file
	// If there isn't a state file:
	// 		- the platform was created without state, so it has the empty state
	// 		- the function PersistStateToFile() creates the state file with the empty state
	// 		- make sure the state is always up dated in the file
	p.PersistStateToFile(stateFilename)

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
		// case "ec2":
		// 	dataKeys = append(dataKeys, "elastic-fileshares")
	case "vsphere":
		var err error
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
// If the state is persisted to a file, there is no need to use this func
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
