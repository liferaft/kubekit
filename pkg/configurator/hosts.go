package configurator

import (
	"fmt"
	"io"
	"strings"

	"github.com/liferaft/kubekit/pkg/configurator/ssh"
)

// Host is the host configuration
type Host struct {
	PublicIP   string `json:"public_ip" yaml:"public_ip" mapstructure:"public_ip"`
	PrivateIP  string `json:"private_ip" yaml:"private_ip" mapstructure:"private_ip"`
	PublicDNS  string `json:"public_dns" yaml:"public_dns" mapstructure:"public_dns"`
	PrivateDNS string `json:"private_dns" yaml:"private_dns" mapstructure:"private_dns"`
	RoleName   string `json:"role" yaml:"role" mapstructure:"role"`
	Pool       string `json:"pool" yaml:"pool" mapstructure:"pool"`
	ssh        *ssh.Config
}

// GetSSHConfig returns the ssh config of the host
func (h *Host) GetSSHConfig() *ssh.Config {
	return h.ssh
}

// Config configures a host
func (h *Host) Config(roleName, username, privateKey, password string) error {
	h.RoleName = roleName
	sshConf, err := ssh.New(username, h.PublicIP, privateKey, password)
	if err != nil {
		return err
	}
	h.ssh = sshConf

	return nil
}

// StartShell initiate an interactive remote shell to this host
func (h *Host) StartShell(in io.Reader, out, err io.Writer) error {
	if h.ssh == nil {
		return fmt.Errorf("SSH session not configured")
	}
	return h.ssh.Shell(in, out, err)
}

// Hosts is a slice of Host
type Hosts []Host

// Config configures the hosts
func (hs Hosts) Config(username, privateKey, password string, applyRoleNameFormat bool) error {
	var newRoleName string
	roleNameFormat := fmt.Sprintf("%%s%%0%dd", ZeroPadLen)
	counter := map[string]int{}
	for i, host := range hs {
		if _, ok := counter[host.RoleName]; ok {
			counter[host.RoleName]++
		} else {
			counter[host.RoleName] = 0
		}
		if applyRoleNameFormat {
			newRoleName = fmt.Sprintf(roleNameFormat, host.RoleName, counter[host.RoleName])
		} else {
			newRoleName = host.RoleName
		}
		if err := host.Config(newRoleName, username, privateKey, password); err != nil {
			return err
		}
		hs[i] = host
	}
	return nil
}

// FilterByRole returns all the hosts with the given roles
func (hs Hosts) FilterByRole(roles ...string) Hosts {
	newHosts := Hosts{}
	for _, h := range hs {
		for _, role := range roles {
			if h.RoleName == role {
				newHosts = append(newHosts, h)
			}
		}
	}

	return newHosts
}

// FilterByRolePrefix returns all the hosts with the given role prefix
// this is to be used after Config() is ran since it reassigns the role name by appending an index
func (hs Hosts) FilterByRolePrefix(rolePrefixes ...string) Hosts {
	newHosts := Hosts{}
	for _, h := range hs {
		for _, rolePrefix := range rolePrefixes {
			if strings.HasPrefix(h.RoleName, rolePrefix) {
				newHosts = append(newHosts, h)
			}
		}
	}

	return newHosts
}

// FilterByNode returns all the hosts that contain an IP or DNS from the given list
func (hs Hosts) FilterByNode(ipOrDNS ...string) Hosts {
	newHosts := Hosts{}
	for _, h := range hs {
		for _, filter := range ipOrDNS {
			if h.PrivateDNS == filter || h.PublicDNS == filter || h.PrivateIP == filter || h.PublicIP == filter {
				newHosts = append(newHosts, h)
			}
		}
	}

	return newHosts
}
