package kubekit

import (
	"fmt"

	"github.com/spf13/cobra"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:     "describe [cluster] NAME[,NAME ...]",
	Aliases: []string{"desc"},
	Short:   "Prints information about a cluster, template or node",
	Long: `Prints information in the formats JSON, YAML or TOML about a list of given
clusters, templates or nodes in a cluster. By default is JSON`,
	RunE: describeClusterRun,
}

// describeClusterCmd represents the cluster command
var describeClusterCmd = &cobra.Command{
	Use:     "cluster NAME[,NAME ...]",
	Aliases: []string{"c"},
	Short:   "Prints information about a list of clusters name",
	Long: `Prints information about a list of clusters name in JSON, YAML or TOML format.
By default is JSON.`,
	RunE: describeClusterRun,
}

// describeTemplatesCmd represents the templates command
var describeTemplatesCmd = &cobra.Command{
	Hidden:  true,
	Use:     "templates NAME[,NAME ...]",
	Aliases: []string{"t"},
	Short:   "Prints information about a list of templates",
	Long: `Prints information about a list of templates in JSON, YAML or TOML format. By
default is JSON.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'templates' still not implemented")
		return nil
	},
}

// describeNodesCmd represents the nodes command
var describeNodesCmd = &cobra.Command{
	Use:     "nodes CLUSTER-NAME",
	Aliases: []string{"n"},
	Short:   "Prints information about a list of nodes in a cluster",
	Long: `Prints information about a list of nodes in a cluster in JSON, YAML or TOML
format. By default is JSON.`,
	RunE: describeNodesRun,
}

func addDescribeCmd() {
	// describe [cluster] NAME[,NAME ...] --output (json|yaml|toml) --pp
	RootCmd.AddCommand(describeCmd)
	describeCmd.PersistentFlags().StringP("output", "o", "yaml", "Output format. Available formats: 'json', 'yaml' and 'toml'")
	describeCmd.PersistentFlags().BoolP("pp", "p", false, "Pretty print. Show the configuration in a human readable format. Applies only for 'json' format")
	describeCmd.Flags().String("format", "", "pretty-print the cluster configuration using a Go template")

	describeCmd.AddCommand(describeClusterCmd)
	// describe templates NAME[,NAME ...] --output (json|yaml|toml) --pp
	describeCmd.AddCommand(describeTemplatesCmd)
	// describe nodes CLUSTER-NAME --output (json|yaml|toml) --pp
	describeCmd.AddCommand(describeNodesCmd)
	describeNodesCmd.Flags().StringP("cluster", "c", "", "cluster name where this node is located")
}

func describeClusterRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a cluster name")
	}
	clustersName := args
	output := cmd.Flags().Lookup("output").Value.String()
	format := cmd.Flags().Lookup("format").Value.String()
	pp := cmd.Flags().Lookup("pp").Value.String() == "true"

	switch output {
	case "json", "yaml", "toml":
	default:
		return fmt.Errorf("unknown or unsupported format %q", output)
	}

	if len(output) != 0 && len(format) != 0 {
		return fmt.Errorf("format template output (--format) cannot be used with any form of output (--output | -o %s), use only one", output)
	}

	// DEBUG:
	// var ppFlag string
	// if pp {
	// 	ppFlag = " --pp"
	// }
	// fmt.Printf("describe clusters %s--output %v %s\n", strings.Join(clustersName, " "), output, ppFlag)

	return printClustersInfo(clustersName, map[string]string{}, output, pp, format)
}

func describeNodesRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a node hostname or IP address")
	}
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 node, received %d. %v", len(args), args)
	}
	nodeName := args[0]
	if len(nodeName) == 0 {
		return fmt.Errorf("node hostname or IP cannot be empty")
	}

	// TODO: Should we remove this and search the node in every cluster?
	clusterName := cmd.Flags().Lookup("cluster").Value.String()
	if len(clusterName) == 0 {
		return fmt.Errorf("cluster name is required")
	}

	output := cmd.Flags().Lookup("output").Value.String()
	pp := cmd.Flags().Lookup("pp").Value.String() == "true"

	// DEBUG:
	// var ppFlag string
	// if pp {
	// 	ppFlag = " --pp"
	// }
	// fmt.Printf("describe node %s --cluster %s --output %s %s\n", nodeName, clusterName, output, ppFlag)

	return describeNode(clusterName, nodeName, output, pp)
}

func describeNode(clusterName, nodeName, output string, pp bool) error {
	// TODO: Improve this function in the future to print more information about this node.
	cni, err := getNodesInfo([]string{clusterName}, []string{nodeName}, nil)
	if err != nil {
		return err
	}

	result, err := cni.Sprintf(output, pp)

	fmt.Println(result)
	return err
}

// func describeKluster(output string, pp bool, clusterName ...string) error {
// 	// The default format for describe is 'yaml'
// 	if len(output) == 0 {
// 		output = "yaml"
// 	}

// 	klusterList, err := kluster.List(config.ClustersDir())
// 	if err != nil {
// 		return err
// 	}
// 	platformList := provisioner.SupportedPlatformsName()

// 	var cOutput []clustersOutput

// 	for _, klusterName := range clusterName {
// 		// Find the kluster 'klusterName' in the list of kluster
// 		var cluster *kluster.Kluster
// 		for _, k := range klusterList {
// 			if k.Name == klusterName {
// 				cluster = k
// 				break
// 			}
// 		}
// 		// If not found, go to the next one
// 		if cluster == nil {
// 			config.Logger.Errorf("cluster %q not found in the cluster directory", klusterName)
// 			continue
// 		}

// 		// Get all the kluster information and dump it into 'cOutput'
// 		statusMap := make(map[string]string, len(platformList))
// 		kubeConfigMap := make(map[string]string, len(platformList))
// 		for _, name := range platformList {
// 			statusMap[name] = cluster.State[name].Status
// 			kubeConfigMap[name] = filepath.Join(filepath.Dir(cluster.Path()), "certificates", fmt.Sprintf("%s-kubeconfig", name))
// 		}
// 		cOutput = append(cOutput, clustersOutput{
// 			Name:       cluster.Name,
// 			Nodes:      0,
// 			Status:     statusMap,
// 			Path:       cluster.Path(),
// 			Kubeconfig: kubeConfigMap,
// 		})
// 	}

// 	// Marshal 'cOutput' to the required format into 'bClusterOutput'
// 	var bClusterOutput []byte
// 	switch output {
// 	case "json":
// 		if pp {
// 			bClusterOutput, err = json.MarshalIndent(cOutput, "", "  ")
// 		} else {
// 			bClusterOutput, err = json.Marshal(cOutput)
// 		}
// 	case "yaml":
// 		bClusterOutput, err = yaml.Marshal(cOutput)
// 	case "toml":
// 		bClusterOutput, err = toml.Marshal(cOutput)
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(string(bClusterOutput))
// 	return nil
// }
