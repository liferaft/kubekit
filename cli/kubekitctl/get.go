package kubekitctl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/liferaft/kubekit/pkg/configurator"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/cli"
)

// getCmd represents the `get` command
var getCmd = &cobra.Command{
	Use:     "get [cluster] NAME[,NAME ...]",
	Aliases: []string{"g"},
	Short:   "Retrieve information from a given object name",
	Long: `The get command is used to retrieve information from a given object. If no
object name is specified, it will return information from all the occurrences of
such object.`,
	RunE: getClusterRun,
}

// getClusterCmd represents the 'get cluster' command
var getClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Prints information about the clusters configured in the system",
	Long: `Prints information about the clusters configured in my system, like: cluster
name, total number of nodes, status and platform.`,
	RunE: getClusterRun,
}

// getEnvCmd represents the 'get env' command
var getEnvCmd = &cobra.Command{
	Use:     "env NAME",
	Aliases: []string{"e", "environment"},
	Short:   "Prints out export commands which can be run in a subshell",
	Long: `Prints outs the environment variables to export to use the cluster. Run this
command with the shell command 'eval'.`,
	RunE: getEnvRun,
}

// getNodesCmd represents the 'get nodes' command
var getNodesCmd = &cobra.Command{
	Use:     "nodes CLUSTER-NAME NAME[,NAME...]",
	Aliases: []string{"n"},
	Short:   "Prints the list of nodes from the given cluster",
	Long: `Prints the list of node from the given cluster name with the following
information: name, public IP address, pool name and status. If the cluster is
absent, prints the nodes in the configuration.`,
	RunE: getNodesRun,
}

func getAddCommands() {
	// get [cluster] NAME[,NAME ...] --full --show-config --show-nodes
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("quiet", "q", false, "Only display the cluster names")
	getCmd.Flags().StringArray("filter", []string{}, "filter output based on conditions provided. Each filter is a key/value pair")

	getCmd.AddCommand(getClusterCmd)
	getClusterCmd.Flags().BoolP("quiet", "q", false, "Only display the cluster names")
	getClusterCmd.Flags().StringArray("filter", []string{}, "filter output based on conditions provided. Each filter is a key/value pair")

	// get env [NAME] --unset --shell SHELL --file
	// RootCmd.AddCommand(getEnvCmd)
	getCmd.AddCommand(getEnvCmd)
	getEnvCmd.Flags().BoolP("unset", "u", false, "prints unset commands which reverse the command effect")
	getEnvCmd.Flags().String("shell", cli.DefaultShell, "use the variable set command for this shell")
	getEnvCmd.Flags().StringP("kubeconfig-file", "f", "", "file to save the cluster KubeConfig file content")

	// get nodes CLUSTER-NAME NAME[,NAME...] --output (wide|json|yaml|toml) --pp --nodes NODE[,NODE] --pools POOL[,POOL]
	// RootCmd.AddCommand(getNodesCmd)
	getCmd.AddCommand(getNodesCmd)
	getNodesCmd.Flags().BoolP("quiet", "q", false, "Only display the nodes IP address")
	getNodesCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to print information")
	getNodesCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools to print information about the nodes in there")
}

func getClusterRun(cmd *cobra.Command, args []string) error {
	clustersName := args

	quiet := cmd.Flags().Lookup("quiet").Value.String() == "true"

	filter, warns, err := cli.GetFilters(cmd)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	// DEBUG:
	var quietFlag string
	if quiet {
		quietFlag = " --quiet "
	}
	var filterFlag string
	for k, v := range filter {
		filterFlag = fmt.Sprintf("%s --filter %q", filterFlag, k+"="+v)
	}
	config.Logger.Debugf("get cluster %s %s %s\n", strings.Join(clustersName, " "), quietFlag, filterFlag)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	output, err := config.client.GetClusters(ctx, quiet, filter, clustersName...)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}

func getEnvRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.GetEnvGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	// DEBUG:
	var unsetFlag string
	if opts.Unset {
		unsetFlag = " --unset"
	}
	config.Logger.Debugf("get env %s%s --shell %s --kubeconfig-file %s\n", opts.ClusterName, unsetFlag, opts.Shell, opts.KubeconfigFile)

	var kubeconfigPath string

	if !opts.Unset {
		var err error
		if kubeconfigPath, err = getClusterKubeConfigPath(opts.ClusterName, opts.KubeconfigFile); err != nil {
			return err
		}
	} else {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}

	// At this time the only environment variable to print is KUBECONFIG, if/when more include them here
	env := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	output := opts.SprintEnv(env)
	fmt.Println(output)

	// Suggest to delete the file if `unset` was selected
	if opts.Unset && len(kubeconfigPath) != 0 {
		rmCmd := "rm"
		switch opts.Shell {
		case "cmd":
			rmCmd = "del"
		case "powershell":
			rmCmd = "Remove-Item –path"
		}
		fmt.Printf("# Optionally, delete the kubeconfig file executing:\n# %s %s\n", rmCmd, kubeconfigPath)
	}

	return nil
}

func getNodesRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.GetNodesGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	quiet := cmd.Flags().Lookup("quiet").Value.String() == "true"

	// DEBUG:
	config.Logger.Debugf("get nodes %s --nodes %v --pools %v\n", strings.Join(opts.ClustersName, " "), opts.Nodes, opts.Pools)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	cni, err := getNodesInfo(ctx, opts.ClustersName, opts.Nodes, opts.Pools)
	if err != nil {
		return err
	}

	format := "wide"
	if quiet {
		format = "quiet"
	}

	output, err := cni.Sprintf(format, false)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// getClusterKubeConfigPath request to the client the kubeconfig file content to
// save it and return the location
func getClusterKubeConfigPath(clusterName, kubeconfigFile string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	// Request to the server the cluster description but get only the kubeconfig file content
	output, err := config.client.Describe(ctx, []string{"kubeconfig"}, clusterName)
	if err != nil {
		return "", err
	}

	// transform the string output (JSON) to a struct to get only the kubeconfig file content
	outputStruct := struct {
		Kubeconfig string
	}{}
	if err := json.Unmarshal([]byte(output), &outputStruct); err != nil {
		return "", err
	}
	fileContent := []byte(outputStruct.Kubeconfig)

	if len(fileContent) == 0 {
		return "", fmt.Errorf("the cluster %q exists but does not have a kubeconfig file yet", clusterName)
	}

	if kubeconfigFile, err = cleanPath(kubeconfigFile, clusterName); err != nil {
		return "", err
	}

	// Save the file with the kubeconfig content
	if err := ioutil.WriteFile(kubeconfigFile, fileContent, 0644); err != nil {
		return "", err
	}

	// retun just the location of the kubeconfig file
	return kubeconfigFile, nil
}

// Identify the location to save the kubeconfig file. If no kubeconfigFile is
// pass, it will save the file to `~/.kube` using the cluster name for the
// filename and `.kconf` as extension.
func cleanPath(path, clusterName string) (string, error) {
	if len(path) == 0 {
		home, _ := homedir.Dir()
		kPath := filepath.Join(home, ".kube")
		if _, err := os.Stat(kPath); os.IsNotExist(err) {
			if err := os.MkdirAll(kPath, 0755); err != nil {
				return "", err
			}
		}
		return filepath.Join(kPath, clusterName+".kconf"), nil
	}

	// If there is a kubeconfigFile make sure to be clean, absolute and without `~`
	if strings.HasPrefix(path, "~/") {
		home, _ := homedir.Dir()
		path = filepath.Join(home, path[2:])
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(pwd, path)
	}

	return path, nil
}

func getNodesInfo(ctx context.Context, clustersName, nodes, pools []string) (cli.ClusterNodeInfo, error) {
	ci := cli.ClusterNodeInfo{}

	for _, clusterName := range clustersName {
		// Request to the server the cluster description but get only the nodes information
		output, err := config.client.Describe(ctx, []string{"nodes"}, clusterName)
		if err != nil {
			return ci, err
		}

		// transform the string output (JSON) to a struct to get only the kubeconfig file content
		type node struct {
			Name       string   `json:"name,omitempty"`
			PoolName   string   `json:"poolName,omitempty"`
			PublicIP   string   `json:"publicIp,omitempty"`
			PrivateIP  string   `json:"privateIp,omitempty"`
			OtherIps   []string `json:"otherIps,omitempty"`
			PublicDNS  string   `json:"publicDns,omitempty"`
			PrivateDNS string   `json:"privateDns,omitempty"`
			OtherDNS   []string `json:"otherDns,omitempty"`
			RoleName   string   `json:"roleName,omitempty"`
		}
		type nodePool struct {
			PoolName string `json:"poolName,omitempty"`
			Nodes    []node `json:"nodes,omitempty"`
		}
		type clusterNodes struct {
			NodePools []nodePool `json:"nodePools,omitempty"`
		}
		outputStruct := struct {
			Nodes clusterNodes `json:"nodes,omitempty"`
		}{}
		if err := json.Unmarshal([]byte(output), &outputStruct); err != nil {
			return nil, err
		}

		allHosts := configurator.Hosts{}
		for _, pool := range outputStruct.Nodes.NodePools {
			for _, node := range pool.Nodes {
				h := configurator.Host{
					PublicIP:   node.PublicIP,
					PrivateIP:  node.PrivateIP,
					PublicDNS:  node.PublicDNS,
					PrivateDNS: node.PrivateDNS,
					RoleName:   node.RoleName,
					Pool:       node.PoolName,
				}
				allHosts = append(allHosts, h)
			}
		}

		var hosts configurator.Hosts
		if len(nodes) != 0 {
			hosts = allHosts.FilterByNode(nodes...)
		} else if len(pools) != 0 {
			hosts = allHosts.FilterByRole(pools...)
		} else {
			hosts = allHosts
		}

		ci[clusterName] = hosts
	}

	return ci, nil
}
