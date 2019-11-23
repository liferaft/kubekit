package cli

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
)

// DefaultShell is the default shell used to print the export and unset commands
const DefaultShell = "bash"

// GetEnvOpts encapsulate all the CLI parameters received from the `get env` command
type GetEnvOpts struct {
	ClusterName    string
	Shell          string
	Unset          bool
	KubeconfigFile string
}

// GetEnvGetOpts get the `get env` command parameters from the cobra commands and arguments
func GetEnvGetOpts(cmd *cobra.Command, args []string) (opts *GetEnvOpts, warns []string, err error) {
	warns = make([]string, 0)

	// cluster_name
	clusterName, err := GetOneClusterName(cmd, args, false)
	if err != nil {
		return nil, warns, err
	}

	// Get the flags `--unset` and `--shell`
	unset := false
	unsetFlag := cmd.Flags().Lookup("unset")
	if unsetFlag != nil {
		unset = unsetFlag.Value.String() == "true"
	}

	if len(clusterName) == 0 && !unset {
		return nil, warns, fmt.Errorf("cluster name is required unless --unset or -u is used")
	}

	shell := DefaultShell
	shellFlag := cmd.Flags().Lookup("shell")
	if shellFlag != nil {
		shell = shellFlag.Value.String()
	}

	var kubeconfigFile string
	kubeconfigFileFlag := cmd.Flags().Lookup("kubeconfig-file")
	if kubeconfigFileFlag != nil {
		kubeconfigFile = kubeconfigFileFlag.Value.String()
	}

	return &GetEnvOpts{
		ClusterName:    clusterName,
		Shell:          shell,
		Unset:          unset,
		KubeconfigFile: kubeconfigFile,
	}, warns, nil
}

// SprintEnv returns the string to print to export the cluster environment
// variable. This is to be executed with `eval`
func (o *GetEnvOpts) SprintEnv(env map[string]string) string {
	b := &bytes.Buffer{}

	var shellSetCmd, shellUnsetCmd string
	switch o.Shell {
	case "bash":
		shellSetCmd = "export %s=%s"
		shellUnsetCmd = "unset %s"
	case "fish":
		shellSetCmd = "set -x %s %s;"
		shellUnsetCmd = "unset %s;"
	case "cmd":
		shellSetCmd = "set %s=%s"
		shellUnsetCmd = "unset %s"
	case "powershell":
		shellSetCmd = "$Env:%s = %q"
		shellUnsetCmd = "unset %s"
	default:
		shellSetCmd = "export %s=%s"
		shellUnsetCmd = "unset %s"
	}

	var unsetFlag string
	if o.Unset {
		unsetFlag = " -u"
	}
	var shellFlag string
	if len(o.Shell) != 0 && o.Shell != DefaultShell {
		shellFlag = fmt.Sprintf(" --shell %s", o.Shell)
	}

	for name, value := range env {
		if o.Unset {
			fmt.Fprintf(b, shellUnsetCmd+"\n", name)
			continue
		}
		fmt.Fprintf(b, shellSetCmd+"\n", name, value)
	}

	fmt.Fprintf(b, "# Run this command to configure your shell:\n")
	fmt.Fprintf(b, "# eval \"$(kubekit get env %s%s%s)\"\n", o.ClusterName, unsetFlag, shellFlag)

	return b.String()
}
