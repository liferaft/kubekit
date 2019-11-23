package kubekit

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubekit/kubekit/cli"

	"github.com/spf13/cobra"
	"github.com/kraken/ui"
	"github.com/kubekit/kubekit/pkg/kluster"
)

var (
	forceFiles,
	backupFiles,
	doExportCC,
	doZipCC,
	sudoFiles bool
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:     "copy [cluster] NAME",
	Aliases: []string{"cp"},
	Short:   "Copies a cluster, cluster configuration file, template file, files to/from a cluster or cluster certificates",
	Long: `
Copy is used to duplicate a cluster, a cluster configuration file, a template
file, copy files to/from an existing cluster or copy cluster certificates to a
given plath or to an existing cluster.`,
}

// copyClusterCmd represents the 'copy cluster' command
var copyClusterCmd = &cobra.Command{
	Hidden:  true,
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Duplicates an existing cluster in the same platform with a different name",
	Long: `Duplicates an existing cluster in the same platform with a different name. Same
as the command 'apply cluster' after coping the cluster configuration with a new
cluster name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'copy cluster' still not implemented")
		return nil
	},
}

// copyClusterConfigCmd represents the 'copy cluster-config' command
var copyClusterConfigCmd = &cobra.Command{
	Use:     "cluster-config CLUSTER-NAME",
	Aliases: []string{"cc"},
	Short:   "Duplicates an existing cluster configuration with a new name",
	Long: `Duplicates an existing cluster configuration with a new cluster name. It may
change the configuration if the flags '--platform' or '--template' are used.`,
	RunE: copyClusterConfigRun,
}

// copyTemplatesCmd represents the 'copy templates' command
var copyTemplatesCmd = &cobra.Command{
	Hidden:  true,
	Use:     "templates NAME",
	Aliases: []string{"t"},
	Short:   "Creates a copy of an existing template with a different name",
	Long: `Creates a copy of an existing template with a different name. It may change the
configuration if the '--platforms' flag is used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'copy templates' still not implemented")
		return nil
	},
}

// copyFilesCmd represents the 'copy files' command
var copyFilesCmd = &cobra.Command{
	Use:     "files",
	Aliases: []string{"f"},
	Short:   "Transfers files from/to cluster nodes.",
	Long: `The command 'copy files' transfer files from a host, or localhost, to an
specific location at the selected nodes or viseversa.`,
	RunE: copyFilesRun,
}

// copyCertificatesCmd represents the 'copy certificates' command
var copyCertificatesCmd = &cobra.Command{
	Hidden:  true,
	Use:     "certificates",
	Aliases: []string{"certs"},
	Short:   "Used to import or export certificates to/from a cluster",
	Long: `Copy certificates command is used to import or export certificates to/from a
cluster. After the certificates are imported they will be applied into the
cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'copy certificates' still not implemented")
		return nil
	},
}

// copyPackageCmd represents the 'copy package' command
var copyPackageCmd = &cobra.Command{
	Use:     "package",
	Aliases: []string{"pkg"},
	Short:   "Used to transfer a package to every node of the cluster",
	Long: `Transfer a package file to every node of the cluster. It's similar to copy a
file but with less parameters.`,
	RunE: copyPackageRun,
}

func addCopyCmd() {
	// copy [cluster] NAME --to NEW-NAME --provision --configure --certificates --generate-certs --export --plan --CERT-key-file FILE --CERT-cert-file FILE
	RootCmd.AddCommand(copyCmd)
	copyCmd.Flags().String("to", "", "new cluster name")
	copyCmd.Flags().BoolP("provision", "p", false, "only apply the provisioning. If possible for the cluster platform, creates or updates the nodes of the cluster")
	copyCmd.Flags().BoolP("configure", "c", false, "only apply the configuration. The cluster must exists. Generate the certificates (if doesn't exists), install and configure Kubernetes on the existing cluster")
	copyCmd.Flags().Bool("certificates", false, "only apply the certificates. The Kubernetes cluster must exists. If the certificates doesn't exists then will be created")
	copyCmd.Flags().Bool("generate-certs", false, "Overwrite or force the certificates generation even if they exists")
	copyCmd.Flags().Bool("export", false, "don't apply, just export the Terraform templates to the cluster config directory")
	// Advance command, do not print in help:
	// copyCmd.Flags().MarkHidden("export")
	copyCmd.Flags().Bool("plan", false, "don't apply, just print the provisioning changes")
	// Advance command, do not print in help:
	// copyCmd.Flags().MarkHidden("plan")
	addCertFlags(copyCmd)
	copyCmd.AddCommand(copyClusterCmd)
	copyClusterCmd.Flags().String("to", "", "new cluster name")
	copyClusterCmd.Flags().BoolP("provision", "p", false, "only apply the provisioning. If possible for the cluster platform, creates or updates the nodes of the cluster")
	copyClusterCmd.Flags().BoolP("configure", "c", false, "only apply the configuration. The cluster must exists. Generate the certificates (if doesn't exists), install and configure Kubernetes on the existing cluster")
	copyClusterCmd.Flags().Bool("certificates", false, "only apply the certificates. The Kubernetes cluster must exists. If the certificates doesn't exists then will be created")
	copyClusterCmd.Flags().Bool("generate-certs", false, "Overwrite or force the certificates generation even if they exists")
	copyClusterCmd.Flags().Bool("export", false, "don't apply, just export the Terraform templates to the cluster config directory")
	// Advance command, do not print in help:
	// copyClusterCmd.Flags().MarkHidden("export")
	copyClusterCmd.Flags().Bool("plan", false, "don't apply, just print the provisioning changes")
	// Advance command, do not print in help:
	// copyClusterCmd.Flags().MarkHidden("plan")
	addCertFlags(copyClusterCmd)

	// copy cluster-config CLUSTER-NAME --export --zip --to NEW-NAME --platform NAME --path PATH --template NAME
	copyCmd.AddCommand(copyClusterConfigCmd)
	copyClusterConfigCmd.Flags().String("to", "", "new cluster name")
	copyClusterConfigCmd.Flags().StringP("platform", "p", "", "platform where the cluster going to be provisioned (Example: aws, vsphere)")
	copyClusterConfigCmd.Flags().String("path", "", "path to store the cluster configuration file if is not the default location")
	copyClusterConfigCmd.Flags().StringP("format", "f", "yaml", "cluster config file format. Available formats: 'json', 'yaml' and 'toml'")
	copyClusterConfigCmd.Flags().StringP("template", "t", "", "cluster template to create this cluster from. Could be a name or absolute location")
	copyClusterConfigCmd.Flags().BoolVar(&doExportCC, "export", false, "export the cluster with the same name to a different path or the local directory")
	copyClusterConfigCmd.Flags().BoolVar(&doZipCC, "zip", false, "export the cluster to a zip file. To be used with --export. The default filename is the cluster name")
	copyClusterConfigCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")

	// copy template NAME
	copyCmd.AddCommand(copyTemplatesCmd)
	copyTemplatesCmd.Flags().String("to", "", "new cluster name")
	copyTemplatesCmd.Flags().StringP("platform", "p", "", "platform where the cluster going to be provisioned (Example: aws, vsphere)")
	copyTemplatesCmd.Flags().String("path", "", "path to store the cluster configuration file if is not the default location")

	// copy files
	copyCmd.AddCommand(copyFilesCmd)
	copyFilesCmd.Flags().StringP("from", "f", "", "source location in the form [host:]:/path/to/files. Target location should be a cluster node")
	copyFilesCmd.Flags().StringP("to", "t", "", "target location in the form [host:]:/path/to/files. Source location should be a cluster node")
	copyFilesCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes where to locate the files")
	copyFilesCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where in such nodes locate the files")
	copyFilesCmd.Flags().BoolVar(&forceFiles, "force", false, "overwrite the file if exists")
	copyFilesCmd.Flags().BoolVarP(&backupFiles, "backup", "b", false, "create a backup if the file exists")
	copyFilesCmd.Flags().BoolVar(&sudoFiles, "sudo", false, "use sudo. The user needs to have sudo access")
	copyFilesCmd.Flags().StringP("owner", "o", "", "name of the user that will own the target file")
	copyFilesCmd.Flags().StringP("group", "g", "", "name of the group that will own the target file")
	copyFilesCmd.Flags().StringP("mode", "m", "", "mode of the target file. Could be octal number or string like chmod format")

	// copy certificates
	copyCmd.AddCommand(copyCertificatesCmd)
	copyCertificatesCmd.Flags().StringP("from", "f", "", "source location in the form [host:]:/path/to/files. The target should be a cluster name")
	copyCertificatesCmd.Flags().String("from-cluster", "", "source cluster name")
	copyCertificatesCmd.Flags().StringP("to", "t", "", "target location in the form [host:]:/path/to/files. The source should be a cluster name")
	copyCertificatesCmd.Flags().String("to-cluster", "", "target cluster name")
	copyCertificatesCmd.Flags().Bool("dry-run", false, "just copy the certificates to the cluster directory, do not apply them to the Kubernetes cluster")
	copyCertificatesCmd.Flags().Bool("zip", false, "store the certificates in a zip file named like the cluster name")
	copyCertificatesCmd.Flags().Bool("no-backup", false, "do not backup the existing certificates to replace")

	// copy package CLUSTER-NAME --package-file FILE --backup
	copyCmd.AddCommand(copyPackageCmd)
	copyPackageCmd.Flags().BoolVar(&doPkgBackup, "backup", false, "backup the package at the cluster node if there was a previous package file")
	copyPackageCmd.Flags().StringP("package-file", "f", "", "package file to transfer. By default will be at the cluster directory named 'kubekit.rpm' or '.deb'")
}

func copyClusterConfigRun(cmd *cobra.Command, args []string) error {
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

	newClusterName := cmd.Flags().Lookup("to").Value.String()
	if len(newClusterName) == 0 && !doExportCC {
		return fmt.Errorf("new cluster name not found. Use the '--to' flag to set the new name")
	}
	if len(newClusterName) != 0 && doExportCC {
		return fmt.Errorf("cannot export the cluster config with a new name. Do not use '--to' with '--export'")
	}

	newPlatform := cmd.Flags().Lookup("platform").Value.String()
	if len(newPlatform) != 0 && doExportCC {
		return fmt.Errorf("cannot export the cluster to a new platform. Do not use the `--export` flag with `--platform` flag")
	}
	path := cmd.Flags().Lookup("path").Value.String()
	format := cmd.Flags().Lookup("format").Value.String()
	templateName := cmd.Flags().Lookup("template").Value.String()
	if len(templateName) != 0 && doExportCC {
		return fmt.Errorf("cannot export the cluster using a template. Do not use the `--export` flag with `--template` flag")
	}

	if len(newClusterName) == 0 && doExportCC {
		newClusterName = clusterName
	}

	if len(path) == 0 && doExportCC {
		path = "."
	}

	varsStr := cmd.Flags().Lookup("var").Value.String()
	variables, warns, err := cli.GetVariables(varsStr)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.UI.Log.Warn(w)
		}
	}

	// DEBUG:
	// fmt.Printf("copy cluster-config %s --to %s --platform %v --path %s  --template %s --var %v\n", clusterName, newClusterName, newPlatform, path, templateName, variables)

	_, err = copyClusterConfig(clusterName, newClusterName, newPlatform, path, format, variables, templateName, doExportCC, doZipCC)
	return err
}

func copyFilesRun(cmd *cobra.Command, args []string) error {
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

	from := cmd.Flags().Lookup("from").Value.String()
	to := cmd.Flags().Lookup("to").Value.String()
	check := from[0:1] + to[0:1]
	if check == "::" || !strings.Contains(check, ":") {
		return fmt.Errorf("target and source location has to be local and the other remote. Remote locations begins with ':'")
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
	owner := cmd.Flags().Lookup("owner").Value.String()
	group := cmd.Flags().Lookup("group").Value.String()
	mode := cmd.Flags().Lookup("mode").Value.String()

	// DEBUG:
	// var forceFlag, backupFlag, sudoFlag string
	// if forceFiles {
	// 	forceFlag = " --force"
	// }
	// if backupFiles {
	// 	backupFlag = " --backup"
	// }
	// if sudoFiles {
	// 	sudoFlag = " --sudo"
	// }
	// fmt.Printf("copy files %s --from %s --to %s --nodes %v --pools %v %s%s%s  --owner %s --group %s --mode %s\n", clusterName, from, to, nodes, pools, forceFlag, backupFlag, sudoFlag, owner, group, mode)

	// the cluster config file and the hosts must exists. This command should be
	// executed after 'apply' or 'apply --provision'
	cluster, err := loadCluster(clusterName)
	if err != nil {
		return err
	}

	config.UI.Log.Infof("copying file to/from every host of cluster %s", clusterName)

	config.UI.Notify("KubeKit", "copy", "<copy>", "", ui.Upload)
	defer config.UI.Notify("KubeKit", "copy", "</copy>", "", ui.Upload)
	err = cluster.CopyFile(from, to, nodes, pools, forceFiles, backupFiles, sudoFiles, owner, group, mode)

	return err
}

func copyPackageRun(cmd *cobra.Command, args []string) error {
	return actionPackageRun(cmd, args, false)
}

func copyClusterConfig(sourceClusterName, newClusterName, platform, path, format string, variables map[string]string, templateName string, doExport, doZip bool) (*kluster.Kluster, error) {
	if len(newClusterName) == 0 {
		return nil, fmt.Errorf("the new cluster name cannot be empty")
	}
	if len(path) == 0 && !kluster.Unique(newClusterName, config.ClustersDir()) {
		return nil, fmt.Errorf("cluster name %q already exists. List all the names with the 'get clusters' command and select a unique name", newClusterName)
	}

	klusterFile := kluster.Path(sourceClusterName, config.ClustersDir())
	if len(klusterFile) == 0 {
		return nil, fmt.Errorf("failed to find the source cluster named %q", sourceClusterName)
	}
	sourceCluster, err := kluster.LoadSummary(klusterFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load the kluster config file %s. %s", klusterFile, err)
	}

	config.UI.Log.Debugf("copying cluster %q to %q", sourceClusterName, newClusterName)

	defaultClustersPath := config.ClustersDir()
	if len(path) == 0 {
		path = defaultClustersPath
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("path %q does not exists", path)
		}
	}

	if path == defaultClustersPath {
		path, err = kluster.NewPath(path)
		if err != nil {
			return nil, err
		}
	} else {
		path = filepath.Join(path, newClusterName)
		os.MkdirAll(path, 0755)
	}

	newCluster, err := sourceCluster.Copy(newClusterName, platform, path, format, config.UI, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to copy the cluster %s to %s. %s", sourceClusterName, newClusterName, err)
	}
	params, err := sourceCluster.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get the credentials from the source cluster, %s", err)
	}

	if err := newCluster.SaveCredentials(params...); err != nil {
		return nil, err
	}
	if err := newCluster.Save(); err != nil {
		return nil, err
	}

	// If it's just a copy (not an export) this is the end ...
	if !doExport {
		return newCluster, nil
	}

	// ... otherwise, get the state and copy the KubeConfig file

	config.UI.Log.Debugf("loading the state from source cluster %s", sourceCluster.Name)
	newCluster.State = sourceCluster.State
	// Save it again to save the state
	if err := newCluster.Save(); err != nil {
		return nil, err
	}

	// Copy the KubeConfig file, do NOT generate it
	kcSourceFilePath := filepath.Join(sourceCluster.CertsDir(), "kubeconfig")
	kcNewFilePath := filepath.Join(newCluster.CertsDir(), "kubeconfig")
	if err := cpFile(kcSourceFilePath, kcNewFilePath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to copy the KubeConfig file from %s to %s. %s", kcSourceFilePath, kcNewFilePath, err)
		}
		config.UI.Log.Warnf("the cluster %s does not have a KubeConfig file to export. %s", sourceCluster.Name, err)
	}
	config.UI.Log.Debugf("kubeconfig file from %s was copied to %s", sourceCluster.Name, kcNewFilePath)

	// Copy the .tfstate file
	tfSourceTFFile := sourceCluster.StateFile()
	tfNewTFFile := newCluster.StateFile()
	if err := cpFile(tfSourceTFFile, tfNewTFFile); err != nil {
		return nil, fmt.Errorf("failed to copy the Terraform state file. %s", err)
	}
	config.UI.Log.Debugf("Terraform state file from %s was copied to %s", sourceCluster.Name, tfNewTFFile)

	if !doZipCC {
		return newCluster, nil
	}

	newClusterDir := newCluster.Dir()
	newClusterZipFile := filepath.Join(".", newCluster.Name+".zip")
	if err := zipDir(newClusterDir, newClusterZipFile); err != nil {
		return newCluster, fmt.Errorf("failed to zip the cluster %s directory, the exported cluster is located at %s", newCluster.Name, newClusterDir)
	}
	config.UI.Log.Debugf("zip file created for cluster %s at %s", newCluster.Name, newClusterZipFile)

	os.RemoveAll(newClusterDir)
	config.UI.Log.Debugf("deleted temporal cluster directory %s", newClusterDir)

	return newCluster, nil
}

func cp(source, target string, infos ...os.FileInfo) (err error) {
	var info os.FileInfo
	if len(infos) > 0 {
		info = infos[0]
	} else {
		if info, err = os.Stat(source); err != nil {
			return err
		}
	}

	if info.IsDir() {
		return cpDir(source, target, info)
	}
	return cpFile(source, target, info)
}

func cpFile(source, target string, infos ...os.FileInfo) (err error) {
	var info os.FileInfo
	if len(infos) > 0 {
		info = infos[0]
	} else {
		if info, err = os.Stat(source); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
		return err
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to read the file %s", source)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create the file %s", target)
	}
	defer destinationFile.Close()

	if err = os.Chmod(destinationFile.Name(), info.Mode()); err != nil {
		return err
	}

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func cpDir(source, target string, infos ...os.FileInfo) (err error) {
	var info os.FileInfo
	if len(infos) > 0 {
		info = infos[0]
	} else {
		if info, err = os.Stat(source); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(target, info.Mode()); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	for _, f := range files {
		src := path.Join(source, f.Name())
		tgt := path.Join(target, f.Name())
		cp(src, tgt, f)
	}

	return nil
}

// ZipDir zips a directory
func zipDir(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// var zipBuffer = new(bytes.Buffer)

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	baseDir := filepath.Base(source)
	if source, err = filepath.Abs(source); err != nil {
		return err
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if header.Name, err = filepath.Rel(source, path); err != nil {
			return err
		}
		header.Name = filepath.Join(baseDir, header.Name)

		//These next line set the born on date for the files in the archive
		header.SetModTime(time.Date(2018, time.November, 10, 23, 0, 0, 0, time.UTC))
		header.ModifiedTime = 0
		header.ModifiedDate = 0
		header.Modified = time.Date(2018, time.November, 10, 23, 0, 0, 0, time.UTC)
		header.SetMode(0644) //set file permissions

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func copyPackage(cluster *kluster.Kluster, pkgFilename string, backup bool) error {
	var backupInfo string
	if backup {
		backupInfo = " backing up any previous file"
	}

	config.UI.Log.Infof("copying package %s to every host of cluster %s%s", pkgFilename, cluster.Name, backupInfo)

	config.UI.Notify("KubeKit", "packages", "<packages>", "", ui.Upload)
	defer config.UI.Notify("KubeKit", "packages", "</packages>", "", ui.Upload)
	err := cluster.CopyPackage(pkgFilename, "/tmp/", backup)

	return err
}
