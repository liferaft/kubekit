package kubekitctl

import (
	"context"
	"fmt"
	"strings"

	"github.com/liferaft/kubekit/cli"
	"github.com/spf13/cobra"
)

// describeCmd represents the `describe` command
var describeCmd = &cobra.Command{
	Use:     "describe [cluster] NAME[,NAME ...]",
	Aliases: []string{"desc"},
	Short:   "Prints information about a cluster, template or node",
	Long: `Prints information in the formats JSON, YAML or TOML about a list of given
clusters, templates or nodes in a cluster. By default is JSON.`,
	RunE: describeClusterRun,
}

// describeClusterCmd represents the 'describe cluster' command
var describeClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Prints information about a list of clusters name",
	Long: `Prints information about a list of clusters name in JSON, YAML or TOML format.
By default is JSON.`,
	RunE: describeClusterRun,
}

func describeAddCommands() {
	// describe [cluster] NAME[,NAME ...] --full --show-config --show-nodes
	RootCmd.AddCommand(describeCmd)
	describeCmd.Flags().BoolP("full", "f", false, "Show everything about the cluster, this includes: basic info, configuration and nodes")
	describeCmd.Flags().Bool("show-config", false, "Include the cluster configuration in the description")
	describeCmd.Flags().Bool("show-nodes", false, "Include the nodes configuration in the description")
	describeCmd.Flags().Bool("show-entrypoint", false, "Include the cluster entrypoint in the description")
	describeCmd.Flags().Bool("show-kubeconfig", false, "Include the kubeconfig file content in the description")

	describeCmd.AddCommand(describeClusterCmd)
	describeClusterCmd.Flags().BoolP("full", "f", false, "Show everything about the cluster, this includes: basic info, configuration and nodes")
	describeClusterCmd.Flags().Bool("show-config", false, "Include the cluster configuration in the description")
	describeClusterCmd.Flags().Bool("show-nodes", false, "Include the nodes configuration in the description")
	describeClusterCmd.Flags().Bool("show-entrypoint", false, "Include the cluster entrypoint in the description")
	describeClusterCmd.Flags().Bool("show-kubeconfig", false, "Include the kubeconfig file content in the description")
}

func describeClusterRun(cmd *cobra.Command, args []string) error {
	clustersName, err := cli.GetMultipleClustersName(cmd, args)
	if err != nil {
		return err
	}

	showParams := []string{}

	showAll := cmd.Flags().Lookup("full").Value.String() == "true"
	if showAll {
		showParams = append(showParams, "all")
	} else {
		for _, name := range []string{"config", "nodes", "entrypoint", "kubeconfig"} {
			flg := cmd.Flags().Lookup("show-" + name)
			if flg == nil {
				continue
			}

			if showParamSet := flg.Value.String() == "true"; showParamSet {
				showParams = append(showParams, name)
			}
		}
	}

	// DEBUG:
	showParamsFlags := []string{}
	if showAll {
		showParamsFlags = append(showParamsFlags, " --full ")
	} else {
		for _, name := range showParams {
			showParamsFlags = append(showParamsFlags, " --show-"+name+" ")
		}
	}
	config.Logger.Debugf("describe cluster %s %s\n", strings.Join(clustersName, ", "), showParamsFlags)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	output, err := config.client.Describe(ctx, showParams, clustersName...)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
