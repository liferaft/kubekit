package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/liferaft/kubekit/pkg/configurator"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

// GetNodesOpts encapsulate all the CLI parameters received from the `get nodes` command
type GetNodesOpts struct {
	ClustersName []string
	Output       string
	Pp           bool
	Nodes        []string
	Pools        []string
}

// GetNodesGetOpts get the `get env` command parameters from the cobra commands and arguments
func GetNodesGetOpts(cmd *cobra.Command, args []string) (opts *GetNodesOpts, warns []string, err error) {
	warns = make([]string, 0)

	clustersName, err := GetMultipleClustersName(cmd, args)
	if err != nil {
		return nil, nil, err
	}

	// Get the flags `--output` and `--pp`
	var output string
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag != nil {
		output = outputFlag.Value.String()
	}
	pp := false
	ppFlag := cmd.Flags().Lookup("pp")
	if ppFlag != nil {
		pp = ppFlag.Value.String() == "true"
	}

	// Nodes:
	nodesStr := cmd.Flags().Lookup("nodes").Value.String()
	nodes, err := StringToArray(nodesStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse the list of nodes")
	}

	// Pools:
	poolsStr := cmd.Flags().Lookup("pools").Value.String()
	pools, err := StringToArray(poolsStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse the list of pools")
	}

	if len(nodes) != 0 && len(pools) != 0 {
		return nil, nil, fmt.Errorf("'nodes' and 'pools' flags are mutually exclusive, use --nodes or --pools but not both in the same command")
	}

	return &GetNodesOpts{
		ClustersName: clustersName,
		Output:       output,
		Pp:           pp,
		Nodes:        nodes,
		Pools:        pools,
	}, warns, nil
}

// ClusterNodeInfo is an list of hosts with their important information
type ClusterNodeInfo map[string]configurator.Hosts

// Sprintf returns a string to print in the given format. Pretty Print (`pp`)
// applies only for JSON
func (cni ClusterNodeInfo) Sprintf(format string, pp bool) (string, error) {
	switch format {
	case "", "wide", "w":
		return "", cni.Table((format == "wide") || (format == "w"))
	case "json":
		return cni.JSON(pp)
	case "yaml":
		return cni.YAML()
	case "toml":
		return cni.TOML()
	case "quiet":
		return cni.IPs(), nil
	default:
		return "", fmt.Errorf("unknown format %q", format)
	}
}

// JSON returns the nodes information in JSON format
func (cni ClusterNodeInfo) JSON(pp bool) (string, error) {
	var (
		output []byte
		err    error
	)

	if pp {
		output, err = json.MarshalIndent(cni, "", "  ")
	} else {
		output, err = json.Marshal(cni)
	}

	return string(output), err
}

// YAML returns the nodes information in YAML format
func (cni ClusterNodeInfo) YAML() (string, error) {
	output, err := yaml.Marshal(cni)
	return string(output), err
}

// TOML returns the nodes information in TOML format
func (cni ClusterNodeInfo) TOML() (string, error) {
	var tomlStruct struct {
		Clusters ClusterNodeInfo `toml:"clusters"`
	}
	tomlStruct.Clusters = cni

	output, err := toml.Marshal(tomlStruct)
	return string(output), err
}

// Table returns the nodes information as a table
func (cni ClusterNodeInfo) Table(wide bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	header := "Name\tPublic IP\tPublic DNS\tRole"
	if wide {
		header = header + "\tPrivate IP\tPrivate DNS\tPool"
	}
	fmt.Fprintf(w, header+"\n")

	for cName, ni := range cni {
		for _, node := range ni {
			row := fmt.Sprintf("%s\t%s\t%s\t%s", cName, node.PublicIP, node.PublicDNS, node.RoleName)
			if wide {
				row = fmt.Sprintf("%s\t%s\t%s\t%s", row, node.PrivateIP, node.PrivateDNS, node.Pool)
			}
			fmt.Fprintf(w, "%s\n", row)
		}
	}

	w.Flush()
	return nil
}

// IPs returns only the IP address of the nodes
func (cni ClusterNodeInfo) IPs() string {
	b := &bytes.Buffer{}
	for _, hosts := range cni {
		for _, host := range hosts {
			fmt.Fprintln(b, host.PublicIP)
		}
	}
	return b.String()
}
