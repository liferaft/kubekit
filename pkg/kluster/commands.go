package kluster

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/kubekit/kubekit/pkg/configurator"
	"github.com/kubekit/kubekit/pkg/configurator/ssh"
)

// HostsFilterBy returns the cluster hosts or nodes in the given pools or
// matching the given node name patterns
func (k *Kluster) HostsFilterBy(nodes []string, pools []string) configurator.Hosts {
	platform := k.Platform()

	allHosts := k.State[platform].Nodes
	var hosts configurator.Hosts
	if len(nodes) != 0 {
		hosts = allHosts.FilterByNode(nodes...)
	} else if len(pools) != 0 {
		hosts = allHosts.FilterByRole(pools...)
	} else {
		hosts = allHosts
	}

	return hosts
}

func (k *Kluster) newCommandFor(nodes []string, pools []string) (*configurator.Command, error) {
	platform := k.Platform()
	hosts := k.HostsFilterBy(nodes, pools)
	platformConfig := k.provisioner[platform].Config()

	return configurator.NewCommand(hosts, platformConfig, k.ui)
}

// CopyFile is to copy files to/form cluster nodes
func (k *Kluster) CopyFile(from, to string, nodes []string, pools []string, forceFiles, backupFiles, sudoFiles bool, owner, group, mode string) error {
	c, err := k.newCommandFor(nodes, pools)
	if err != nil {
		return err
	}

	return c.Copy(from, to, forceFiles, backupFiles, sudoFiles, owner, group, mode)
}

// Exec execute a script file or command line on every node of the cluster or
// the selected nodes
func (k *Kluster) Exec(command, script string, nodes []string, pools []string, sudoExec bool) (*ssh.CommandResult, error) {
	c, err := k.newCommandFor(nodes, pools)
	if err != nil {
		return nil, err
	}

	return c.Exec(command, script, sudoExec)
}

// StartShellTo opens an interactive shell to the given host name
func (k *Kluster) StartShellTo(nodeName string, in io.Reader, out, e io.Writer) error {
	nodes := k.HostsFilterBy([]string{nodeName}, nil)
	if len(nodes) != 1 {
		return fmt.Errorf("node %s not found", nodeName)
	}
	node := nodes[0]

	platform := k.Platform()
	pConfig := k.provisioner[platform].Config()
	platformConfig := make(map[string]interface{})
	pConfigB, err := json.Marshal(pConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal the platform configuration. %s", err)
	}
	json.Unmarshal(pConfigB, &platformConfig)

	username, ok := platformConfig["username"]
	if !ok {
		return fmt.Errorf("username not found in %q platform configuration", platform)
	}
	privKey, err := configurator.GetPrivateKey(platformConfig)
	if err != nil {
		return err
	}
	var password string
	if p, ok := platformConfig["password"]; ok {
		password = p.(string)
	}

	if err := node.Config(node.RoleName, username.(string), privKey, password); err != nil {
		return err
	}

	return node.StartShell(in, out, e)
}

// CopyPackage copies a system package (rpm or deb) to every cluster node to a
// default location `/tmp`
func (k *Kluster) CopyPackage(source, target string, backupPkg bool) error {
	if target == "" {
		target = "/tmp/"
	}
	target = ":" + target
	return k.CopyFile(source, target, nil, nil, true, backupPkg, false, "", "", "0644")
}

// InstallPackage installs a system package (rpm or deb) already located in the
// cluster nodes. To copy the package use the method `CopyPackage()`
func (k *Kluster) InstallPackage(filename string, forcePkg bool) (result *ssh.CommandResult, failedNodes []string, err error) {
	if filename == "" {
		return nil, nil, fmt.Errorf("the package filename is required, cannot be empty")
	}

	var forceStr, command string

	fileExt := filepath.Ext(filename)

	switch fileExt {
	case ".rpm":
		if forcePkg {
			forceStr = fmt.Sprintf("-Uvh --force %s'", filename)
		} else {
			forceStr = fmt.Sprintf("-q $(rpm -qp %s) || rpm -Uvh %s'", filename, filename)
		}
		command = fmt.Sprintf("sudo /bin/sh -c 'rm -rf /root/manifest.src && rpm %s", forceStr)
	case ".deb":
		if forcePkg {
			forceStr = "--force-all "
		}
		command = fmt.Sprintf("sudo /bin/sh -c 'rm -rf /root/manifest.src && dpkg --skip-same-version %s-i %s'", forceStr, filename)
	default:
		// we assume the "package" can be natively loaded by docker
		command = fmt.Sprintf("sudo /bin/sh -c 'echo \"%s\"' > /root/manifest.src", filename)
	}

	if result, err = k.Exec(command, "", nil, nil, true); err != nil {
		return nil, nil, err
	}

	if result.Failures == 0 {
		return result, nil, nil
	}

	failedNodes = []string{}
	for host, res := range result.Hosts.GetSnapshot() {
		if res.ExitStatus != 0 {
			failedNodes = append(failedNodes, host)
		}
	}

	return result, failedNodes, err
}
