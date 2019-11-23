package kubekitctl

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/kubekit/kubekit/cli"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:     "apply [cluster] NAME",
	Aliases: []string{"a"},
	Short:   "Apply the changes to the cluster configuration",
	Long: `Apply is used to create or modify a cluster using the settings in the cluster 
configuration file. The apply command is usually executed after the init command.`,
	RunE: applyClusterRun,
}

// applyClusterCmd represents the 'apply cluster' command
var applyClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Apply the changes to the cluster configuration",
	Long: `Apply is used to create or modify a cluster using the settings in the cluster 
configuration file. The apply command is usually executed after the init command.`,
	RunE: applyClusterRun,
}

func applyAddCommands() {
	// apply [cluster] NAME --provision --configure --certificates --generate-certs --export --plan --CERT-key-file FILE --CERT-cert-file FILE
	RootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolP("provision", "p", false, "only apply the provisioning. If possible for the cluster platform, creates or updates the nodes of the cluster")
	applyCmd.Flags().BoolP("configure", "c", false, "only apply the configuration. The cluster must exists. Generate the certificates (if doesn't exists), install and configure Kubernetes on the existing cluster")
	applyCmd.Flags().StringP("package-file-url", "f", "", "URL to get a package to install before configure. If not given will use the package located in the KubeKit server")
	applyCmd.Flags().Bool("force-pkg", false, "force install of package")
	cli.AddCertFlags(applyCmd)

	applyCmd.AddCommand(applyClusterCmd)
	applyClusterCmd.Flags().BoolP("provision", "p", false, "only apply the provisioning. If possible for the cluster platform, creates or updates the nodes of the cluster")
	applyClusterCmd.Flags().BoolP("configure", "c", false, "only apply the configuration. The cluster must exists. Generate the certificates (if doesn't exists), install and configure Kubernetes on the existing cluster")
	applyClusterCmd.Flags().StringP("package-file-url", "u", "", "URL to get a package to install before configure. If not given will use the package located in the KubeKit server")
	applyClusterCmd.Flags().Bool("force-pkg", false, "force install of package")
	cli.AddCertFlags(applyClusterCmd)
}

func applyClusterRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.ApplyGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	// DEBUG:
	var actionFlag string
	if opts.Action != "ALL" {
		actionFlag = " --" + strings.ToLower(actionFlag)
	}
	var pkgURLFlag string
	if opts.PackageURL != "" {
		pkgURLFlag = "--package-file-url " + opts.PackageURL
	}
	var forcePkgFlag string
	if opts.ForcePackage {
		forcePkgFlag = " --force-pkg"
	}
	var certFlags string
	for fname, kp := range opts.UserCACerts {
		if kp.KeyFile != "" {
			certFlags = fmt.Sprintf("%s --%s-key-file %s", certFlags, fname, kp.KeyFile)
		}
		if kp.CertFile != "" {
			certFlags = fmt.Sprintf("%s --%s-cert-file %s", certFlags, fname, kp.CertFile)
		}
	}
	config.Logger.Debugf("apply cluster %s%s%s%s%s\n", opts.ClusterName, actionFlag, pkgURLFlag, forcePkgFlag, certFlags)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	output, err := config.client.Apply(ctx, opts.ClusterName, opts.Action, opts.PackageURL, opts.ForcePackage, opts.UserCACerts)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
