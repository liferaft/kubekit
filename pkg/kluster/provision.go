package kluster

import (
	"fmt"

	"github.com/kraken/terraformer"
)

func (k *Kluster) provision(destroy bool) error {
	platformName := k.Platform()
	logPrefix := fmt.Sprintf("KubeKit [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	k.LoadState()
	k.ui.Log.Debug("state(s) loaded")

	// LoadState makes the platforms provisioners
	p := k.provisioner[platformName]

	k.ui.Log.Debugf("starting process to provisioning/terminating the cluster %q", platformName)

	logPrefix = fmt.Sprintf("Provisioner [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)
	err := p.Apply(destroy)
	k.ui.TerminateAllNotifications("")

	logPrefix = fmt.Sprintf("KubeKit [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	// Save the state, no matter if failed or not
	k.SaveState()

	if err != nil {
		if destroy {
			k.State[platformName].Status = FailedTerminationStatus.String()
		} else {
			k.State[platformName].Status = FailedProvisioningStatus.String()
		}
	} else {
		if destroy {
			k.State[platformName].Status = TerminatedStatus.String()
		} else {
			k.State[platformName].Status = ProvisionedStatus.String()
		}
	}
	return err
}

// Create provision the cluster on all the required platforms
func (k *Kluster) Create() error {
	return k.provision(false)
}

// Terminate destroy the cluster on all the required platforms
func (k *Kluster) Terminate() error {
	return k.provision(true)
}

// Plan prints to the UI and Log the plan before applying the changes
func (k *Kluster) Plan(destroy bool) error {
	platformName := k.Platform()

	k.LoadState()
	k.ui.Log.Debug("state(s) loaded")

	p := k.provisioner[platformName]

	logPrefix := fmt.Sprintf("Provisioner [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	plan, err := p.Plan(destroy)
	if err != nil {
		return err
	}
	k.ui.Log.Debugf("plan to apply: %v", plan)

	stats := terraformer.NewStats(plan)
	statStr := fmt.Sprintf("Actions: %d to add, %d to change, %d to destroy", stats.Add, stats.Change, stats.Destroy)
	// k.ui.Log.Infof(statStr)
	fmt.Println(statStr)

	logPrefix = fmt.Sprintf("KubeKit [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	return err
}
