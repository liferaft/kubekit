package kubekit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/liferaft/kubekit/cli"

	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/liferaft/kubekit/pkg/packages"
	"github.com/spf13/cobra"
)

var (
	doProvision,
	doConfigure,
	doCerts,
	genCACerts,
	doExportTF,
	doExportK8s,
	doPlan,
	doPkgBackup bool
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:     "apply [cluster] NAME",
	Aliases: []string{"a"},
	Short:   "Applies the configuration into your infrastructure",
	Long: `Apply is used to apply the configuration into your infrastructure. This means
apply will create the cluster if it doesn't exists or will update or apply the
configuration changes if it exists.`,
	RunE: applyClusterRun,
}

// applyCertificatesCmd represents the 'apply certificates' command
var applyCertificatesCmd = &cobra.Command{
	Use:     "certificates CLUSTER-NAME",
	Aliases: []string{"certs"},
	Short:   "apply the certificates for the given cluster",
	Long: `The command apply certificates applies all the certificates required by a cluster. 
The cluster certificates can be applied without running through the entire configuration process.`,
	RunE: applyCertificatesRun,
}

// applyClusterCmd represents the 'apply cluster' command
var applyClusterCmd = &cobra.Command{
	Hidden:  true,
	Aliases: []string{"c"},
	Use:     "cluster NAME",
	Short:   "Applies the configuration into your infrastructure",
	Long: `Apply is used to apply the configuration into your infrastructure. This means
apply will create the cluster if it doesn't exists or will update or apply the
configuration changes if it exists.`,
	RunE: applyClusterRun,
}

// applyPackageCmd represents the 'apply package' command
var applyPackageCmd = &cobra.Command{
	Hidden:  true,
	Aliases: []string{"pkg"},
	Use:     "package CLUSTER-NAME",
	Short:   "Applies or install the package to every cluster node",
	Long: `Apply is used to install a package to every cluster node. Useful for development
or testing.`,
	RunE: applyPackageRun,
}

func addApplyCmd() {
	// apply [cluster] NAME --provision --configure --certificates --generate-certs --export --plan --CERT-key-file FILE --CERT-cert-file FILE
	RootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVarP(&doProvision, "provision", "p", false, "only apply the provisioning. If possible for the cluster platform, creates or updates the nodes of the cluster")
	applyCmd.Flags().BoolVarP(&doConfigure, "configure", "c", false, "only apply the configuration. The cluster must exists. Generate the certificates (if doesn't exists), install and configure Kubernetes on the existing cluster")
	applyCmd.Flags().StringP("package-file", "f", "", "package to install before configure. By default will be at the cluster directory named 'kubekit.rpm' or '.deb'")
	// It's been discussed by the team if keep or remove --certificates. Kubernetes can do it with kubectl
	// applyCmd.Flags().BoolVar(&doCerts, "certificates", false, "only apply the certificates. The Kubernetes cluster must exists. If the certificates doesn't exists then will be created")
	applyCmd.Flags().BoolVar(&genCACerts, "generate-ca-certs", false, "Overwrite or force the certificates generation even if they exists")
	applyCmd.Flags().BoolVar(&doExportTF, "export-tf", false, "don't apply, just export the Terraform templates to the cluster config directory")
	applyCmd.Flags().BoolVar(&doExportK8s, "export-k8s", false, "don't apply, just export the Kubernetes manifests templates to the cluster config directory")
	applyCmd.Flags().Bool("force-pkg", false, "force install of package")
	// Advance command, do not print in help:
	// applyCmd.Flags().MarkHidden("export")
	applyCmd.Flags().BoolVar(&doPlan, "plan", false, "don't apply, just print the provisioning changes")

	// Advance command, do not print in help:
	// applyCmd.Flags().MarkHidden("plan")
	addCertFlags(applyCmd)

	applyCmd.AddCommand(applyClusterCmd)
	applyClusterCmd.Flags().BoolVarP(&doProvision, "provision", "p", false, "only apply the provisioning. If possible for the cluster platform, creates or updates the nodes of the cluster")
	applyClusterCmd.Flags().BoolVarP(&doConfigure, "configure", "c", false, "only apply the configuration. The cluster must exists. Generate the certificates (if doesn't exists), install and configure Kubernetes on the existing cluster")
	applyClusterCmd.Flags().StringP("package-file", "f", "", "package to install before configure. By default will be at the cluster directory named 'kubekit.rpm' or '.deb'")
	// It's been discussed by the team if keep or remove --certificates. Kubernetes can do it with kubectl
	// applyClusterCmd.Flags().BoolVar(&doCerts, "certificates", false, "only apply the certificates. The Kubernetes cluster must exists. If the certificates doesn't exists then will be created")
	applyClusterCmd.Flags().BoolVar(&genCACerts, "generate-ca-certs", false, "Overwrite or force the certificates generation even if they exists")
	applyClusterCmd.Flags().BoolVar(&doExportTF, "export-tf", false, "don't apply, just export the Terraform templates to the cluster config directory")
	applyClusterCmd.Flags().BoolVar(&doExportK8s, "export-k8s", false, "don't apply, just export the Kubernetes manifests templates to the cluster config directory")
	// Advance command, do not print in help:
	// applyClusterCmd.Flags().MarkHidden("export")
	applyClusterCmd.Flags().BoolVar(&doPlan, "plan", false, "don't apply, just print the provisioning changes")
	applyClusterCmd.Flags().Bool("force-pkg", false, "force install of package")
	// Advance command, do not print in help:
	// applyClusterCmd.Flags().MarkHidden("plan")
	addCertFlags(applyClusterCmd)

	applyCmd.AddCommand(applyPackageCmd)
	applyPackageCmd.Flags().BoolVar(&doPkgBackup, "backup", false, "backup the package at the cluster node if there was a previous package file")
	applyPackageCmd.Flags().StringP("package-file", "f", "", "package file to apply. By default will be at the cluster directory named 'kubekit.rpm' or '.deb'")
	applyPackageCmd.Flags().Bool("force-pkg", false, "force install of package")

	applyCmd.AddCommand(applyCertificatesCmd)
}

func applyCertificatesRun(cmd *cobra.Command, args []string) error {
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

	// TODO: implement CA rotation and apply if toggled

	return cluster.ApplyClientCertificates(true)
}

func applyClusterRun(cmd *cobra.Command, args []string) error {
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

	// the cluster config file must exists. This command should be executed after 'init' or will fail
	cluster, err := loadCluster(clusterName)
	if err != nil {
		return err
	}

	// generate (if doesn't exists) the SSH keys, required for the terraform templates and provisioner
	if err := cluster.HandleKeys(); err != nil {
		return err
	}

	// If so, that's all, export them and return
	if doExportTF {
		return cluster.ExportTF()
	}

	// If so, that's all, print the plan and return
	if doPlan {
		return cluster.Plan(false)
	}

	pkgFilename := cmd.Flags().Lookup("package-file").Value.String()
	forcePkg := cmd.Flags().Lookup("force-pkg").Value.String() == "true"
	//check to see if the rpm matches what kubekit expects, if it was passed in

	if err := packages.CheckRpmPackage(pkgFilename, forcePkg); err != nil {
		return err
	}
	// if one of these flags is set, then do not apply the entire process, just
	// the explicit actions specified by the flags
	explicitActions := doProvision || doConfigure // || doCerts

	// if not explicit action was set, then do provisioning as part of the e2e
	// process. Otherwise, do provisioning only if explicitly requested
	if (!explicitActions || doProvision) && !doExportK8s {
		if err := provision(cluster); err != nil {
			return err
		}
		// ... apply the package (if any) ...

		platform := cluster.Platform()
		if pkgFilename != "" && (platform == "aks" || platform == "eks") {
			// warn about not needing package
			msg := "It will NOT be copied and installed without --force-pkg option"
			if forcePkg {
				msg = "It will be copied and installed as --force-pkg was set"
			}
			config.UI.Log.Warnf("package file is specified and is not required for eks or aks. %s", msg)
		}

		//if (not eks AND not aks) OR we are forcing, copy and install
		if (platform != "aks" && platform != "eks") || forcePkg {
			if err := copyAndExecPackage(cluster, pkgFilename, forcePkg); err != nil {
				return err
			}
		}
	}

	// if not explicit action was set, then do configuration as part of the e2e
	// process. Otherwise, do configuration only if explicitly requested
	if !explicitActions || doConfigure {
		// for platforms that are not "eks" and "aks", generate the certificates and apply the package (if any)
		// initialize the certificates if they don't exists, unless requested with forceCerts ...
		userCACertsFiles, err := cli.GetCertFlags(cmd)
		if err != nil {
			return err
		}

		//Do not check for eks/aks as there are no packages installed
		if cluster.Platform() != "aks" && cluster.Platform() != "eks" {
			config.UI.Log.Info("Verifying the installed packages are correct")
			if err := packages.GetBaseImages(cluster, forcePkg, pkgFilename); err != nil {
				return err
			}
		}

		if err := initCertificates(clusterName, cluster, genCACerts, userCACertsFiles); err != nil {
			return err
		}

		// ... generate the kubeconfig file
		if err := cluster.LoadState(); err != nil {
			return err
		}
		if err := cluster.CreateKubeConfigFile(); err != nil {
			return err
		}

		if doExportK8s {
			return cluster.ExportK8s()
		}

		// ... and do the configuration
		return configure(cluster)
	}

	// DEBUG:
	// var provisionFlag, configureFlag, certificatesFlag, generateCertsFlag, exportFlag, planFlag string
	// if doProvision {
	// 	provisionFlag = " --provision"
	// }
	// if doConfigure {
	// 	configureFlag = " --configure"
	// }
	// if doCerts {
	// 	certificatesFlag = " --certificates"
	// }
	// if forceCerts {
	// 	generateCertsFlag = " --generate-certs"
	// }
	// if doExportTF {
	// 	exportFlagTF = " --export-tf"
	// }
	// if doExportK8s {
	// 	exportFlagK8s = " --export-k8s"
	// }
	// if doPlan {
	// 	planFlag = " --plan"
	// }
	// cmd.Printf("apply%s%s%s%s%s%s%s%s --CERT-key-file FILE --CERT-cert-file FILE\n", clusterName, provisionFlag, configureFlag, certificatesFlag, generateCertsFlag, exportFlagTF, exportFlagK8s, planFlag)

	return nil
}

func applyPackageRun(cmd *cobra.Command, args []string) error {
	return actionPackageRun(cmd, args, true)
}

func findPackage(clusterPath string) string {
	formats := []string{"rpm", "deb"}
	pkgFile := filepath.Join(clusterPath, "kubekit.")
	for _, f := range formats {
		if _, err := os.Stat(pkgFile + f); err == nil {
			return pkgFile + f
		}
	}
	return ""
}

func loadCluster(clusterName string) (*kluster.Kluster, error) {
	return kluster.LoadCluster(clusterName, config.ClustersDir(), config.UI)
}

func provision(cluster *kluster.Kluster) error {
	errP := cluster.Create()
	// TODO: Should it save the cluster if fail?
	// If there is an active state, and fail to re-provision, the state is empty and will be deleted.
	// If the provision fails but something was provisioned, should the cluster state/config be saved/updated?
	errS := cluster.Save()
	if errP != nil && errS != nil {
		return fmt.Errorf("failed to create one of the platform(s) and to save the cluster configuration file.\n%s\n%s", errP, errS)
	}
	if errP != nil {
		return errP
	}
	return errS
}

func copyAndExecPackage(cluster *kluster.Kluster, pkgFilename string, forcePkg bool) error {
	clusterPath := cluster.Dir()

	if len(pkgFilename) == 0 {
		pkgFilename = findPackage(clusterPath)
		if len(pkgFilename) == 0 {
			config.UI.Log.Warnf("cannot find any package file in the cluster directory")
			return nil
		}
	}
	if err := copyPackage(cluster, pkgFilename, true); err != nil {
		return err
	}

	filename := filepath.Base(pkgFilename)
	pkgFilepath := filepath.Join("/tmp", filename)

	return execPackage(cluster, pkgFilepath, forcePkg)
}

func configure(cluster *kluster.Kluster) error {
	errC := cluster.Configure()
	errS := cluster.Save()
	if errC != nil && errS != nil {
		return fmt.Errorf("failed to configure Kubernetes and to save the cluster configuration file.\n%s\n%s", errC, errS)
	}
	if errC != nil {
		return errC
	}
	return errS
}

func actionPackageRun(cmd *cobra.Command, args []string, doExec bool) error {
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

	pkgFilename := cmd.Flags().Lookup("package-file").Value.String()

	cluster, err := loadCluster(clusterName)
	if err != nil {
		return err
	}
	clusterPath := cluster.Dir()

	if len(pkgFilename) == 0 {
		pkgFilename = findPackage(clusterPath)
		if len(pkgFilename) == 0 {
			return fmt.Errorf("cannot find the package name for the cluster %s", clusterName)
		}
	}

	// // DEBUG:
	// var backupFlag string
	// actionStr := "copy"
	// if doPkgBackup {
	// 	backupFlag = " --backup"
	// }
	// if doExec {
	// 	actionStr = "apply"
	// }
	// cmd.Printf("%s package %s --package-file %s %s\n", actionStr, clusterName, pkgFilename, backupFlag)

	err = copyPackage(cluster, pkgFilename, doPkgBackup)
	if err != nil || !doExec {
		return err
	}

	filename := filepath.Base(pkgFilename)
	pkgFilepath := filepath.Join("/tmp", filename)
	forcePkg := cmd.Flags().Lookup("force-pkg").Value.String() == "true"

	return execPackage(cluster, pkgFilepath, forcePkg)
}
