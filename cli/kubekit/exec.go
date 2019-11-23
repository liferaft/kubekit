package kubekit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kubekit/kubekit/cli"

	"github.com/spf13/cobra"
	"github.com/kubekit/kubekit/pkg/kluster"
)

var (
	sudoExec bool
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:     "exec [cluster] NAME",
	Aliases: []string{"x"},
	Short:   "Executes a remote command in the clusters nodes",
	Long:    `Executes a remote command in all the nodes of the cluster or some of them.`,
	RunE:    execClusterRun,
}

// execClusterCmd represents the 'exec cluster' command
var execClusterCmd = &cobra.Command{
	Hidden:  true,
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Executes a remote command in the clusters nodes",
	Long:    `Executes a remote command in all the nodes of the cluster or some of them.`,
	RunE:    execClusterRun,
}

// execPackageCmd represents the 'exec package' command
var execPackageCmd = &cobra.Command{
	Hidden:  true,
	Use:     "package NAME",
	Aliases: []string{"pkg"},
	Short:   "Install a package previously copied to all nodes",
	Long:    `Install a previously copied package with the 'copy package' command.`,
	RunE:    execPackageRun,
}

func addExecCmd() {
	// exec [cluster] NAME
	RootCmd.AddCommand(execCmd)
	execCmd.Flags().StringP("cmd", "c", "", "command to execute")
	execCmd.Flags().StringP("file", "f", "", "script file to transfer and execute")
	execCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes where to execute the command or script")
	execCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where in such nodes execute the command or script")
	execCmd.Flags().BoolVar(&sudoExec, "sudo", false, "use sudo. The user needs to have sudo access")
	execCmd.Flags().StringP("output", "o", "yaml", "Output format. Available formats: 'json', 'yaml' and 'toml'")
	execCmd.Flags().Bool("pp", false, "Pretty print. Show the configuration in a human readable format. Applies only for 'json' format")

	execCmd.AddCommand(execClusterCmd)
	execClusterCmd.Flags().StringP("cmd", "c", "", "command to execute")
	execClusterCmd.Flags().StringP("file", "f", "", "script file to transfer and execute")
	execClusterCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes where to execute the command or script")
	execClusterCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where in such nodes execute the command or script")
	execClusterCmd.Flags().BoolVar(&sudoExec, "sudo", false, "use sudo. The user needs to have sudo access")
	execClusterCmd.Flags().StringP("output", "o", "yaml", "Output format. Available formats: 'json', 'yaml' and 'toml'")
	execClusterCmd.Flags().Bool("pp", false, "Pretty print. Show the configuration in a human readable format. Applies only for 'json' format")

	// exec package CLUSTER-NAME
	execCmd.AddCommand(execPackageCmd)
	execPackageCmd.Flags().String("package-file", "", "package filename to install. By default is 'kubekit.rpm'")
	execPackageCmd.Flags().Bool("force-pkg", false, "Force install of package")
}

func execClusterRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a cluster name")
	}
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 cluster name, received %d. %v", len(args), args)
	}
	clusterName := args[0]
	if len(clusterName) == 0 {
		return fmt.Errorf("cluster name cannot be empty")
	}

	command := cmd.Flags().Lookup("cmd").Value.String()
	script := cmd.Flags().Lookup("file").Value.String()

	if len(command) != 0 && len(script) != 0 {
		return fmt.Errorf("'cmd' and 'file' flags are mutually exclusive, use --cmd or --file but not both in the same command")
	}

	nodesStr := cmd.Flags().Lookup("nodes").Value.String()
	nodes, err := cli.StringToArray(nodesStr)
	if err != nil {
		return fmt.Errorf("failed to parse the list of nodes")
	}
	poolsStr := cmd.Flags().Lookup("pools").Value.String()
	pools, err := cli.StringToArray(poolsStr)
	if err != nil {
		return fmt.Errorf("failed to parse the list of pools")
	}
	if len(nodes) != 0 && len(pools) != 0 {
		return fmt.Errorf("'nodes' and 'pools' flags are mutually exclusive, use --nodes or --pools but not both in the same command")
	}

	output := cmd.Flags().Lookup("output").Value.String()
	pp := cmd.Flags().Lookup("pp").Value.String() == "true"

	// DEBUG:
	// var ppFlag, sudoFlag string
	// if pp {
	// 	ppFlag = " --pp"
	// }
	// if sudoFiles {
	// 	sudoFlag = " --sudo"
	// }
	// fmt.Printf("exec cluster %s --cmd %q --file %s --nodes %v --pools %v %s --output %s %s\n", clusterName, command, script, nodes, pools, sudoFlag, output, ppFlag)

	cluster, err := loadCluster(clusterName)
	if err != nil {
		return err
	}

	result, err := cluster.Exec(command, script, nodes, pools, sudoExec)
	if err != nil {
		return err
	}

	var outputResult []byte
	switch output {
	case "json":
		outputResult, err = result.JSON(pp)
	case "yaml":
		outputResult, err = result.YAML()
	case "toml":
		outputResult, err = result.TOML()
	}

	fmt.Println(string(outputResult))
	return nil
}

func execPackageRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a cluster name")
	}
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 cluster name, received %d. %v", len(args), args)
	}
	clusterName := args[0]
	if len(clusterName) == 0 {
		return fmt.Errorf("cluster name cannot be empty")
	}

	cluster, err := loadCluster(clusterName)
	if err != nil {
		return err
	}

	pkgFilename := cmd.Flags().Lookup("package-file").Value.String()
	forcePkg := cmd.Flags().Lookup("force-pkg").Value.String() == "true"

	return execPackage(cluster, pkgFilename, forcePkg)
}

func execPackage(cluster *kluster.Kluster, filename string, forcePkg bool) error {
	if len(filename) == 0 {
		filename = filepath.Join("/tmp", "kubekit.rpm")
	}

	var forceStrInfo string
	if forcePkg {
		forceStrInfo = " using force"
	}

	config.UI.Log.Infof("installing the package %s%s on every host of the cluster %s", filename, forceStrInfo, cluster.Name)

	result, failedHosts, err := cluster.InstallPackage(filename, forcePkg)
	if err != nil {
		return err
	}

	fmt.Printf("package successfuly installed in %d/%d nodes\n", result.Success, result.Success+result.Failures)

	if result.Failures != 0 {
		fmt.Printf("the package install failed in the following nodes: %s\n", strings.Join(failedHosts, ", "))
	}

	return nil
}
