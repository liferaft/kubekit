package kubekit

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/cli"
	"github.com/liferaft/kubekit/pkg/kluster"
	"os"
	"path/filepath"
	"strings"
)

var (
	deleteForce bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete [cluster] NAME",
	Aliases: []string{"d", "del", "destroy"},
	Short:   "Deletes or destroy a cluster, a cluster configuration file, a template or file in a cluster node",
	Long: `Delete is to terminate all the cluster nodes and may also delete the cluster
configuration. It deletes a cluster configuration file, a template file or a
file located in a cluster node.`,
	RunE: deleteClusterRun,
}

// deleteClusterCmd represents the 'delete cluster' command
var deleteClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Terminate a cluster",
	Long: `Delete is to terminate all the cluster nodes and may also delete the cluster
configuration file.`,
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

// deleteTemplatesCmd represents the 'delete templates' command
var deleteTemplatesCmd = &cobra.Command{
	Hidden:  true,
	Use:     "templates NAME[,NAME...]",
	Aliases: []string{"t"},
	Short:   "Deletes the given template files",
	Long:    `Deletes the given template files located in the default templates directory.`,
	RunE:    deleteTemplatesRun,
}

// deleteFilesCmd represents the 'delete files' command
var deleteFilesCmd = &cobra.Command{
	Hidden:  true,
	Use:     "files CLUSTER_NAME FILE[,FILE...]",
	Aliases: []string{"f"},
	Short:   "A brief description of your command",
	Long: `Deletes files on every node of the cluster or on specific nodes listed with
'--nodes' flag or on the specific pools listed on '--pools' flag. Filenames can
include wildcards (*,?) or ranges (i.e. [1-3])`,
	RunE: deleteFilesRun,
}

func addDeleteCmd() {
	// delete [cluster] NAME --force --all
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().BoolVar(&deleteForce, "force", false, "do not confirm or ask to the user before delete the resource")

	deleteCmd.Flags().Bool("all", false, "delete all the cluster resources such as configuration files, certificates and state")
	deleteCmd.Flags().BoolVar(&doPlan, "plan", false, "don't delete the cluster, just print the changes to apply")

	deleteCmd.AddCommand(deleteClusterCmd)
	deleteClusterCmd.Flags().Bool("all", false, "delete all the cluster resources such as configuration files, certificates and state")
	deleteClusterCmd.Flags().BoolVar(&doPlan, "plan", false, "don't delete the cluster, just print the changes to apply")

	// delete clusters-config NAME --force
	deleteCmd.AddCommand(deleteClustersConfigCmd)

	// delete templates NAME[,NAME...] --force
	deleteCmd.AddCommand(deleteTemplatesCmd)

	// delete files CLUSTER-NAME FILE[,FILE...] --force --nodes NODE[,NODE] --pools POOL[,POOL]
	deleteCmd.AddCommand(deleteFilesCmd)
	deleteFilesCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes where to delete the files")
	deleteFilesCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where in such nodes the files will be delete")
}

func deleteClusterRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.DeleteGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.UI.Log.Warn(w)
		}
	}

	if doPlan {
		opts.Force = true
	}

	// DEBUG:
	// cmd.Printf("delete cluster %s\n", opts.ClusterName)

	if ok := opts.Confirm(); !ok {
		fmt.Printf("cluster %q was not destroyed\n", opts.ClusterName)
		return nil
	}

	// the cluster config file must exists. This command should be executed after 'init' or 'apply' otherwise will fail
	cluster, err := loadCluster(opts.ClusterName)
	if err != nil {
		return err
	}

	// If so, that's all, print the plan and return
	if doPlan {
		return cluster.Plan(true)
	}

	var errS error
	errT := cluster.Terminate()

	updated := make(map[string]string)
	if cluster.Platform() != "aks" {
		updated["private_key"] = ""
		updated["private_key_file"] = ""
	}
	updated["public_key_file"] = ""
	updated["public_key"] = ""

	cluster.Update(updated)

	if !opts.DestroyAll {
		errS = cluster.Save()
	}
	if errT != nil && errS != nil {
		return fmt.Errorf("failed to destroy the cluster and to save the cluster configuration file.\n%s\n%s", errT, errS)
	}
	if errT != nil {
		return errT
	}
	if errS != nil {
		return errS
	}

	if err := deleteCerts(opts.ClusterName, cluster); err != nil {
		return err
	}

	if opts.DestroyAll {
		fmt.Printf("cluster %q was destroyed and certificates deleted\n", opts.ClusterName)
		return deleteKluster(deleteForce, opts.ClusterName)
	}

	fmt.Printf("cluster %q was destroyed and certificates deleted. The cluster configuration still exists\n", opts.ClusterName)

	return nil
}

func deleteClustersConfigRun(cmd *cobra.Command, args []string) error {
	clustersName, err := cli.GetMultipleClustersName(cmd, args)
	if err != nil {
		return err
	}

	// DEBUG:
	// cmd.Printf("delete clusters-config %v\n", strings.Join(clustersName, " "))

	return deleteKluster(deleteForce, clustersName...)
}

func deleteTemplatesRun(cmd *cobra.Command, args []string) error {
	fmt.Println("[ERROR] command 'delete templates' still not implemented")
	return nil
}

func deleteFilesRun(cmd *cobra.Command, args []string) error {
	fmt.Println("[ERROR] command 'delete files' still not implemented")
	return nil
}

func deleteCerts(clusterName string, cluster *kluster.Kluster) (err error) {
	if cluster == nil {
		if cluster, err = loadCluster(clusterName); err != nil {
			return err
		}
	}
	certsDir := filepath.Join(cluster.Dir(), kluster.CertificatesDirname)
	return os.RemoveAll(certsDir)
}

func deleteKluster(force bool, clustersName ...string) error {
	errs := []error{}

	for _, clusterName := range clustersName {
		clusterConfigFilePath := kluster.Path(clusterName, config.ClustersDir())
		// If not found, go to the next one
		if len(clusterConfigFilePath) == 0 {
			if !force {
				err := fmt.Errorf("cluster %q not found in the clusters directory", clusterName)
				errs = append(errs, err)
			}
			continue
		}

		baseDir := filepath.Dir(clusterConfigFilePath)

		deleteIt := true
		if !force {
			deleteIt = false
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Do you want to delete cluster %q located in %s [type 'yes']?: ", clusterName, baseDir)
			answer, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSuffix(answer, "\n")) == "yes" {
				deleteIt = true
			}
		}
		if deleteIt {
			err := os.RemoveAll(baseDir)
			if err != nil {
				if !force {
					errs = append(errs, err)
				}
				continue
			}
			fmt.Printf("deleted cluster %q from %s\n", clusterName, baseDir)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	errStr := "failed to delete the following custers:"
	for _, err := range errs {
		errStr = errStr + "\n" + err.Error()
	}
	return fmt.Errorf(errStr)
}
