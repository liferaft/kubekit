package kubekit

import (
	"fmt"
	"path/filepath"

	"github.com/liferaft/kubekit/cli"

	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Retrieve information from a given object name",
	Long: `The get command is used to retrieve information from a given object. If no
object name is specified, it will return information from all the occurrences of
such object. The get command mimics the behavior of the GET method of a REST API.`,
}

// rclustersCmd represents the 'get clusters' command
// var rclustersCmd = &cobra.Command{
// 	Hidden:  true,
// 	Use:     "[get] clusters NAME[,NAME...]",
// 	Aliases: []string{"clusters"},
// 	Short:   "Prints information about the clusters configured in the system",
// 	Long: `Prints information about the clusters configured in my system, like: cluster
// name, total number of nodes, status and platform.`,
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		fmt.Println("[ERROR] command 'get clusters' still not implemented")
// 		return nil
// 	},
// }

// getClustersCmd represents the 'get clusters' command
var getClustersCmd = &cobra.Command{
	Use:     "clusters NAME[,NAME...]",
	Aliases: []string{"c"},
	Short:   "Prints information about the clusters configured in the system",
	Long: `Prints information about the clusters configured in my system, like: cluster
name, total number of nodes, status and platform.`,
	RunE: getClustersRun,
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

// getTemplatesCmd represents the 'get templates' command
var getTemplatesCmd = &cobra.Command{
	Hidden:  true,
	Use:     "templates NAME[,NAME...]",
	Aliases: []string{"t"},
	Short:   "Prints information about the cluster templates in the system",
	Long: `Prints information about the clusters templates in my system, like: template
name, total number of nodes and supported platforms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'get templates' still not implemented")
		return nil
	},
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

func addGetCmd() {
	RootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringP("output", "o", "", "Output format. Available formats: none (regular output), 'wide', 'json', 'yaml' and 'toml'")
	getCmd.PersistentFlags().Bool("pp", false, "Pretty print. Show the configuration in a human readable format. Applies only for 'json' format")

	// [get] clusters [NAME[,NAME...]] --output (wide|json|yaml|toml) --pp
	// RootCmd.AddCommand(rclustersCmd)
	getCmd.AddCommand(getClustersCmd)
	getClustersCmd.Flags().StringArray("filter", []string{}, "filter output based on conditions provided. Each filter is a key/value pair")
	getClustersCmd.Flags().String("format", "", "pretty-print clusters configuration using a Go template")

	// [get] nodes CLUSTER-NAME NAME[,NAME...] --output (wide|json|yaml|toml) --pp --nodes NODE[,NODE] --pools POOL[,POOL]
	// RootCmd.AddCommand(getNodesCmd)
	getCmd.AddCommand(getNodesCmd)
	getNodesCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to print information")
	getNodesCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools to print information about the nodes in there")

	// [get] files CLUSTER-NAME FILENAME[,FILENAME...] --output (wide|json|yaml|toml) --pp --nodes NODE[,NODE] --pools POOL[,POOL] --path PATHT[,PATH]
	// RootCmd.AddCommand(getFilesCmd)
	// getCmd.AddCommand(getFilesCmd)
	// getFilesCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes where to locate the files")
	// getFilesCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where in such nodes locate the files")
	// getFilesCmd.Flags().StringSlice("path", nil, "path in the selected nodes where to find the given filenames")

	// [get] templates NAME[,NAME...] --output (wide|json|yaml|toml) --pp
	// RootCmd.AddCommand(getTemplatesCmd)
	getCmd.AddCommand(getTemplatesCmd)

	// [get] env NAME
	// RootCmd.AddCommand(getEnvCmd)
	getCmd.AddCommand(getEnvCmd)
	getEnvCmd.Flags().BoolP("unset", "u", false, "prints unset commands which reverse the command effect")
	getEnvCmd.Flags().String("shell", cli.DefaultShell, "use the variable set command for this shell")
}

func getClustersRun(cmd *cobra.Command, args []string) error {
	clustersName := args
	output := cmd.Flags().Lookup("output").Value.String()
	format := cmd.Flags().Lookup("format").Value.String()
	pp := cmd.Flags().Lookup("pp").Value.String() == "true"

	if config.Quiet {
		if len(output) != 0 {
			return fmt.Errorf("quiet mode cannot be used with any form of output, use only one")
		}
		output = "quiet"
	}

	if len(output) != 0 && len(format) != 0 {
		return fmt.Errorf("format template output (--format) cannot be used with any form of output (--output | -o %s), use only one", output)
	}

	filter, warns, err := cli.GetFilters(cmd)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.UI.Log.Warn(w)
		}
	}

	// DEBUG:
	// var ppFlag string
	// if pp {
	// 	ppFlag = " --pp"
	// }
	// fmt.Printf("get clusters %s--output %v %s\n", strings.Join(clustersName, " "), output, ppFlag)

	return printClustersInfo(clustersName, filter, output, pp, format)
}

func getNodesRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.GetNodesGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.UI.Log.Warn(w)
		}
	}

	if config.Quiet {
		if len(opts.Output) != 0 {
			return fmt.Errorf("cannot use an output format %q and quiet mode at same time", opts.Output)
		}
		opts.Output = "quiet"
	}

	// TODO: Should we implement this rule?
	// if config.Quiet && len(opts.ClustersName) > 1 {
	// 	return fmt.Errorf("cannot use quiet mode for multiple clusters, just one cluster is allowed")
	// }

	// DEBUG:
	// var ppFlag string
	// if pp {
	// 	ppFlag = " --pp"
	// }
	// fmt.Printf("get nodes %s --nodes %v --pools %v --output %s%s\n", strings.Join(opts.ClustersName, " "), opts.Nodes, opts.Pools, opts.Output, ppFlag)

	cni, err := getNodesInfo(opts.ClustersName, opts.Nodes, opts.Pools)
	if err != nil {
		return err
	}

	output, err := cni.Sprintf(opts.Output, opts.Pp)
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
			config.UI.Log.Warn(w)
		}
	}

	// DEBUG:
	// var unsetFlag string
	// if opts.Unset {
	// 	unsetFlag = " --unset"
	// }
	// fmt.Printf("get env %s%s --shell %s\n", opts.ClusterName, unsetFlag, opts.Shell)

	var kubeconfigPath string
	if !opts.Unset {
		klusterList, err := kluster.List(config.ClustersDir(), opts.ClusterName)
		if err != nil {
			return err
		}
		if len(klusterList) == 0 || klusterList[0] == nil {
			return fmt.Errorf("cluster %q not found, to find the cluster name try: kubekit get clusters", opts.ClusterName)
		}
		kubeconfigPath = filepath.Join(filepath.Dir(klusterList[0].Path()), "certificates", "kubeconfig")
	}

	// At this time the only environment variable to print is KUBECONFIG, if/when more include them here
	env := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	output := opts.SprintEnv(env)
	fmt.Println(output)

	return nil
}

// func getFilesRun(cmd *cobra.Command, args []string) error {
// 	if len(args) == 0 {
// 		return fmt.Errorf("requires a cluster name")
// 	}
// 	clusterName := args[0]
// 	if len(clusterName) == 0 {
// 		return fmt.Errorf("cluster name cannot be empty")
// 	}

// 	filenames := args[1:]
// 	if len(filenames) == 0 {
// 		return fmt.Errorf("requires at least a filename")
// 	}

// 	output := cmd.Flags().Lookup("output").Value.String()
// 	pp := cmd.Flags().Lookup("pp").Value.String() == "true"

// 	pathStr := cmd.Flags().Lookup("path").Value.String()
// 	paths, err := stringToArray(pathStr)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse the list of paths")
// 	}
// 	if len(paths) == 0 {
// 		paths = []string{"/"}
// 	}
// 	nodesStr := cmd.Flags().Lookup("nodes").Value.String()
// 	nodes, err := stringToArray(nodesStr)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse the list of nodes")
// 	}
// 	poolsStr := cmd.Flags().Lookup("pools").Value.String()
// 	pools, err := stringToArray(poolsStr)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse the list of pools")
// 	}
// 	if len(nodes) != 0 && len(pools) != 0 {
// 		return fmt.Errorf("'nodes' and 'pools' flags are mutually exclusive, use --nodes or --pools but not both in the same command")
// 	}

// 	// DEBUG:
// 	var ppFlag string
// 	if pp {
// 		ppFlag = " --pp"
// 	}
// 	fmt.Printf("get files %s %v --path %v --nodes %v --pools %v --output %s%s\n", clusterName, filenames, paths, nodes, pools, output, ppFlag)

// 	cluster, err := loadCluster(clusterName)
// 	if err != nil {
// 		return err
// 	}

// 	var outputResult []byte
// 	// path := paths[0]
// 	// if len(paths) > 1 {
// 	// 	path = fmt.Sprintf("{%s}", strings.Join(paths, ","))
// 	// }
// 	filename := filenames[0]
// 	if len(filenames) > 1 {
// 		filename = fmt.Sprintf("{%s}", strings.Join(filenames, ","))
// 	}
// 	command := fmt.Sprintf("ls -al %s 2>/dev/null", filename)
// 	fmt.Println(command)

// 	result, err := cluster.Exec(command, "", nodes, pools, true)
// 	if err != nil {
// 		return err
// 	}

// 	switch output {
// 	case "json":
// 		outputResult, err = result.JSON(pp)
// 	case "yaml":
// 		outputResult, err = result.YAML()
// 	case "toml":
// 		outputResult, err = result.TOML()
// 	}

// 	fmt.Println(string(outputResult))
// 	return nil
// }

func printClustersInfo(clustersName []string, filter map[string]string, output string, pp bool, format string) error {
	ci, err := kluster.GetClustersInfo(config.ClustersDir(), filter, clustersName...)
	if err != nil {
		return err
	}

	var result string
	if len(format) != 0 {
		result, err = ci.Template(format)
	} else {
		result, err = ci.Stringf(output, pp)
	}
	if err != nil {
		return err
	}

	fmt.Print(result)
	return nil
}

func getNodesInfo(clustersName, nodes, pools []string) (cli.ClusterNodeInfo, error) {
	cni := make(cli.ClusterNodeInfo, 0)
	for _, cName := range clustersName {
		cluster, err := loadCluster(cName)
		if err != nil {
			return nil, err
		}
		n := cluster.HostsFilterBy(nodes, pools)
		cni[cluster.Name] = n
	}

	return cni, nil
}
