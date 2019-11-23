package kluster

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/liferaft/kubekit/pkg/configurator"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// ExportTF exports to files the Terraform files, the TF code (main.tf) and the
// TF variables (terraform.tfvars).
func (k *Kluster) ExportTF() error {
	pName := k.Platform()
	logPrefix := fmt.Sprintf("KubeKit [ %s@%s ]", k.Name, pName)
	k.ui.SetLogPrefix(logPrefix)

	tfPath, _ := k.makeTFDir(pName)

	p := k.provisioner[pName]
	if err := p.BeProvisioner(nil); err != nil {
		return fmt.Errorf("failed to create the provisioner for %s. %s", pName, err)
	}

	logPrefix = fmt.Sprintf("Export [ %s@%s ]", k.Name, pName)
	k.ui.SetLogPrefix(logPrefix)

	// Save the TF Code to main.tf
	tfCodeFile := filepath.Join(tfPath, "main.tf")
	tfCode := p.Code()

	if err := ioutil.WriteFile(tfCodeFile, tfCode, 0644); err != nil {
		return fmt.Errorf("failed to write Terraform code file for %s. %s", pName, err)
	}
	k.ui.Log.Infof("saved terraform code for %s to %s", pName, tfCodeFile)

	// Save the TF variables to terraform.tfvars
	tfVarsFile := filepath.Join(tfPath, "terraform.tfvars")
	v := p.Variables()

	// Variables only holds sensitive data now, credentials and keys
	// removing sensitive data from render, but allowing for tfvars
	// for key := range v {
	// 	v[key] = ""
	// }

	tfVars, err := utils.HCL(v)
	if err != nil {
		return fmt.Errorf("failed get the variables in HCL format for %s. %s", pName, err)
	}

	if err := ioutil.WriteFile(tfVarsFile, tfVars, 0644); err != nil {
		return fmt.Errorf("failed to write variables file for %s. %s", pName, err)
	}
	k.ui.Log.Infof("saved terraform variables for %s to %s", pName, tfVarsFile)

	return nil
}

// ExportK8s exports the Kubernetes manifests templates (YAML files) to the
// cluster directory
func (k *Kluster) ExportK8s() error {
	platformName := k.Platform()
	logPrefix := fmt.Sprintf("KubeKit [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	k.makeK8sDir()

	pConf := k.provisioner[platformName].Config()
	clusterDir := k.Dir()

	conf, err := configurator.New(k.Name, platformName, k.State[platformName].Address, k.State[platformName].Port, k.State[platformName].Nodes, k.State[platformName].Data, pConf, k.Config, k.Resources, clusterDir, k.ui)
	if err != nil {
		return err
	}

	logPrefix = fmt.Sprintf("Export [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	return conf.ApplyResources(true)
}
