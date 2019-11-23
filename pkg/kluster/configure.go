package kluster

import (
	"fmt"

	"github.com/liferaft/kubekit/pkg/configurator"
)

// Configure configures the cluster to have Kubernetes up and running. It uses
// the configurator to do this task
func (k *Kluster) Configure() error {
	platformName := k.Platform()
	logPrefix := fmt.Sprintf("KubeKit [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	k.ui.Log.Debugf("starting the configuration of cluster %q on %s", k.Name, platformName)

	pConf := k.provisioner[platformName].Config()
	clusterDir := k.Dir()

	conf, err := configurator.New(k.Name, platformName, k.State[platformName].Address, k.State[platformName].Port, k.State[platformName].Nodes, k.State[platformName].Data, pConf, k.Config, k.Resources, clusterDir, k.ui)
	if err != nil {
		return err
	}

	if err := conf.Configure(); err != nil {
		k.State[platformName].Status = FailedConfigurationStatus.String()
		return err
	}

	k.State[platformName].Status = RunningStatus.String()

	return nil
}
