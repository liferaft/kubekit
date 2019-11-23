package kubekitctl

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/kubekit/kubekit/cli"
)

// deleteCmd represents the `delete` command
var deleteCmd = &cobra.Command{
	Use:     "delete [cluster] NAME",
	Aliases: []string{"d", "del", "destroy"},
	Short:   "Deletes or destroy a cluster and, if requested, the cluster configuration files",
	Long: `Delete is to terminate all the cluster nodes and may also delete the cluster
configuration.`,
	RunE: deleteClusterRun,
}

// deleteClusterCmd represents the 'delete cluster' command
var deleteClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Terminate a cluster",
	Long: `Delete is to terminate all the cluster nodes and may also delete the cluster
configuration.`,
	RunE: deleteClusterRun,
}

// deleteClustersConfigCmd represents the 'delete clusters-config' command
var deleteClustersConfigCmd = &cobra.Command{
	Use:     "clusters-config NAME",
	Aliases: []string{"cc"},
	Short:   "Deletes a cluster configuration file",
	Long: `Deletes the cluster configuration file for the given cluster name. It will
delete it only if the cluster does not exists, unless '--force' is used.`,
	RunE: deleteClustersConfigRun,
}

func deleteAddCommands() {
	// delete [cluster] NAME --force --all
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().Bool("force", false, "do not confirm or ask to the user before delete the resource")

	deleteCmd.Flags().Bool("all", false, "delete all the cluster resources such as configuration files, certificates and state")

	deleteCmd.AddCommand(deleteClusterCmd)
	deleteClusterCmd.Flags().Bool("all", false, "delete all the cluster resources such as configuration files, certificates and state")

	deleteCmd.AddCommand(deleteClustersConfigCmd)
}

func deleteClusterRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.DeleteGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	// DEBUG:
	var allFlag string
	if opts.DestroyAll {
		allFlag = " --all"
	}
	var forceFlag string
	if opts.Force {
		forceFlag = " --force"
	}
	config.Logger.Debugf("delete cluster %s%s%s\n", opts.ClusterName, allFlag, forceFlag)

	if ok := opts.Confirm(); !ok {
		fmt.Printf("cluster %q was not destroyed\n", opts.ClusterName)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	output, err := config.client.Delete(ctx, opts.ClusterName, opts.DestroyAll)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}

func deleteClustersConfigRun(cmd *cobra.Command, args []string) error {
	clustersName, err := cli.GetMultipleClustersName(cmd, args)
	if err != nil {
		return err
	}

	force := false
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag != nil {
		force = forceFlag.Value.String() == "true"
	}

	// DEBUG:
	var forceFlagStr string
	if force {
		forceFlagStr = " --force"
	}
	config.Logger.Debugf("delete cluster-config %s%s\n", clustersName, forceFlagStr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	for _, clusterName := range clustersName {
		opts := cli.DeleteOpts{
			ClusterName: clusterName,
			Force:       force,
		}
		if ok := opts.Confirm(); !ok {
			fmt.Printf("cluster %q was not destroyed\n", opts.ClusterName)
			continue
		}

		output, err := config.client.DeleteClusterConfig(ctx, clusterName)
		if err != nil {
			errMsg := fmt.Sprintf("fail to delete configuration of cluster %q", clusterName)
			config.Logger.Error(errMsg)
			fmt.Fprintf(os.Stderr, "ERROR: %s", errMsg)
			continue
		}

		fmt.Println(output)
	}

	return nil
}
