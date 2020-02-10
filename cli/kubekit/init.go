package kubekit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/liferaft/kubekit/cli"
	"github.com/liferaft/kubekit/pkg/crypto/tls"
	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/spf13/cobra"
)

const defPackageFormat = "rpm"

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init [cluster] NAME",
	Aliases: []string{"i"},
	Short:   "Initialize a cluster configuration file, template or certificates",
	Long: `Init is used to initialize, generate or create a cluster configuration file,
template or certificates. This is usually the first command to execute. When
it's initializing a cluster the noun cluster is optional, as it's the default
noun.`,
	RunE: initClusterRun,
}

// initClusterCmd represents the 'init cluster' command
var initClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Initialize a cluster configuration file",
	Long: `The command init cluster is used to generate a cluster configuration file with
the default values for the given platform or template.`,
	RunE: initClusterRun,
}

// initTemplateCmd represents the 'init template' command
var initTemplateCmd = &cobra.Command{
	Use:     "template NAME",
	Aliases: []string{"t"},
	Short:   "Initialize a template",
	Long: `The command init template is used to generate a template file with a cluster
configuration for different platforms. A template is used to initialize a
cluster based on the template configuration for one of the template platforms`,
	RunE: initTemplateRun,
}

// initCertificatesCmd represents the 'init certificates' command
var initCertificatesCmd = &cobra.Command{
	Use:     "certificates CLUSTER-NAME",
	Aliases: []string{"certs"},
	Short:   "Initialize or create the certificates for the given cluster",
	Long: `The command init certificates creates all the certificates required by a cluster.
The cluster certificates can be created in advance of the cluster creation or
before the configuration, if not they will be created as part of the cluster
creation process with the 'apply' command.`,
	RunE: initCertificatesRun,
}

// initPackageCmd represents the 'init certificates' command
var initPackageCmd = &cobra.Command{
	Use:     "package CLUSTER-NAME",
	Aliases: []string{"pkg"},
	Short:   "Creates a package for the given cluster",
	Long: `The command init package creates a package (RPM or DEB) for the existing
cluster. The file will be located in the cluster directory.`,
	RunE: initPackageRun,
}

func addInitCmd() {
	// init [cluster] NAME --platform NAME --path PATH --format FORMAT --template NAME --update --var NAME01=VALUE01 --var NAME02=VALUE02 ...
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("platform", "p", "", "platform where the cluster going to be provisioned (Example: aws, vsphere)")
	initCmd.Flags().String("path", "", "path to store the cluster configuration file if is not the default location")
	initCmd.Flags().StringP("format", "f", "yaml", "cluster config file format. Available formats: 'json', 'yaml' and 'toml'")
	initCmd.Flags().StringP("template", "t", "", "cluster template to create this cluster from. Could be a name or absolute location")
	initCmd.Flags().BoolP("update", "u", false, "allows to update an existing cluster configuration file")
	initCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")
	// ... [credentials]
	initCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	initCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	initCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EC2/EKS credentials]
	initCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	initCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	initCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	initCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	initCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	initCmd.Flags().String("subscription_id", "", "Provisioner Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	initCmd.Flags().String("tenant_id", "", "Provisioner Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	initCmd.Flags().String("client_id", "", "Provisioner Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	initCmd.Flags().String("client_secret", "", "Provisioner Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")

	initCmd.AddCommand(initClusterCmd)
	initClusterCmd.Flags().StringP("platform", "p", "", "platform where the cluster going to be provisioned (Example: aws, vsphere)")
	initClusterCmd.Flags().String("path", "", "path to store the cluster configuration file if is not the default location")
	initClusterCmd.Flags().StringP("format", "f", "yaml", "cluster config file format. Available formats: 'json', 'yaml' and 'toml'")
	initClusterCmd.Flags().StringP("template", "t", "", "cluster template to create this cluster from. Could be a name or absolute location")
	initClusterCmd.Flags().BoolP("update", "u", false, "allows to update an existing cluster configuration file")
	initClusterCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")
	// ... [credentials]
	initClusterCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	initClusterCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	initClusterCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EC2/EKS credentials]
	initClusterCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	initClusterCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	initClusterCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	initClusterCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	initClusterCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	initClusterCmd.Flags().String("subscription_id", "", "Provisioner Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	initClusterCmd.Flags().String("tenant_id", "", "Provisioner Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	initClusterCmd.Flags().String("client_id", "", "Provisioner Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	initClusterCmd.Flags().String("client_secret", "", "Provisioner Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")

	// init template NAME --platform NAME --path PATH --format FORMAT --update
	initCmd.AddCommand(initTemplateCmd)
	initTemplateCmd.Flags().StringSliceP("platform", "p", nil, "list of platforms where the cluster may be provisioned (Example: aws, vsphere)")
	initTemplateCmd.Flags().String("path", "", "path to store the cluster configuration file if is not the default location")
	initTemplateCmd.Flags().StringP("format", "f", "yaml", "cluster config file format. Available formats: 'json', 'yaml' and 'toml'")
	// initTemplateCmd.Flags().BoolP("update", "u", false, "allows to update an existing template file")
	initTemplateCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")

	// init certificates CLUSTER-NAME --CERT-key-file FILE --CERT-cert-file FILE --update
	initCmd.AddCommand(initCertificatesCmd)
	addCertFlags(initCertificatesCmd)
	// initCertificatesCmd.Flags().BoolP("update", "u", false, "allows to update the existing certificate files")

	// init package CLUSTER-NAME --update
	initCmd.AddCommand(initPackageCmd)
	initPackageCmd.Flags().StringP("format", "f", "rpm", "package format. Available formats: 'rpm' 'deb'")
	initPackageCmd.Flags().StringP("target", "t", "", "filename or directory to store the package")
	// initPackageCmd.Flags().BoolP("update", "u", false, "allows to update the existing package file")
}

func initClusterRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.InitGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.UI.Log.Warn(w)
		}
	}

	// DEBUG:
	// var updateFlag string
	// if opts.updateInit {
	// 	updateFlag = "--update"
	// }
	// cmd.Printf("init cluster %s --platform %v --path %v --format %v --template %v %s --var %v\n", opts.clusterName, opts.platform, opts.path, opts.format, opts.templateName, updateFlag, opts.variables)

	var cluster *kluster.Kluster

	if opts.Update {
		config.UI.Log.Debugf("updating cluster %q", opts.ClusterName)

		if cluster, err = loadCluster(opts.ClusterName); err != nil {
			return err
		}

		if err := cluster.Update(opts.Variables); err != nil {
			return err
		}

		if err := cluster.Save(); err != nil {
			return err
		}

		return nil
	}

	if len(opts.Path) == 0 {
		opts.Path = config.ClustersDir()
	} else {
		if _, err := os.Stat(opts.Path); os.IsNotExist(err) {
			return cli.UserErrorf("path %q does not exists", opts.Path)
		}
	}

	createClusterConfig := func() error {
		config.UI.Log.Debugf("initializing cluster %q configuration", opts.ClusterName)

		if cluster, err = kluster.CreateCluster(opts.ClusterName, opts.Platform, opts.Path, opts.Format, opts.Variables, config.UI); err != nil {
			return fmt.Errorf("failed to initialize the cluster %s. %s", opts.ClusterName, err)
		}

		return nil
	}

	// If this is a cluster for the following platforms, do not process the
	// credentials. They do not have them
	switch opts.Platform {
	case "vra", "raw", "stacki":
		return createClusterConfig()

	case "aws":
		// Special case, "aws" is no longer a concrete platform, instead use ec2
		// Alternate option would be to accept aws but silently remap it to ec2
		return cli.UserErrorf("'aws' is no longer a supported platform name, use 'ec2' instead")
	}

	config.UI.Log.Debugf("initializing cluster %q credentials", opts.ClusterName)

	// Create credentials reading values from:
	// 1st: from the variables that have creds from flags, environment and KubeKit
	// variables, in that order of priority
	credentials := kluster.NewCredentials(opts.ClusterName, opts.Platform, "")
	if err := credentials.AssignFromMap(opts.Credentials); err != nil {
		return err
	}

	if !credentials.Complete() {
		// 2nd: if this is AWS, read from the parameters in the AWS configuration
		switch opts.Platform {
		case "ec2", "eks":
			getAWSCredVariables(credentials.(*kluster.AwsCredentials), opts.Variables)
		}
	}

	if !credentials.Complete() {
		// 3rd: ask the missing values to the user
		config.UI.Log.Debugf("get credentials asking the user")
		if err := credentials.Ask(); err != nil {
			return err
		}
	}

	if err := createClusterConfig(); err != nil {
		return err
	}

	credentialsPath := filepath.Join(filepath.Dir(cluster.Path()), ".credentials")
	credentials.SetPath(credentialsPath)

	return credentials.Write()
}

func initTemplateRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cli.UserErrorf("requires a template name")
	}
	if len(args) != 1 {
		return cli.UserErrorf("accepts 1 template name, received %d. %v", len(args), args)
	}
	templateName := args[0]
	platformStr := cmd.Flags().Lookup("platform").Value.String()
	platforms, err := cli.StringToArray(platformStr)
	if err != nil {
		return cli.UserErrorf("failed to parse the list of platform")
	}
	path := cmd.Flags().Lookup("path").Value.String()
	format := cmd.Flags().Lookup("format").Value.String()

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
	// cmd.Printf("init template %s --platform %v --path %v --format %v --var %v\n", templateName, platforms, path, format, variables)

	_, err = initTemplate(templateName, platforms, path, format, variables)
	return err
}

func initCertificatesRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cli.UserErrorf("requires a cluster name")
	}
	if len(args) != 1 {
		return cli.UserErrorf("accepts 1 cluster name, received %d. %v", len(args), args)
	}
	clusterName := args[0]
	if len(clusterName) == 0 {
		return cli.UserErrorf("cluster name cannot be empty")
	}

	// DEBUG:
	// cmd.Printf("init certificates %s \n", clusterName)

	userCACertsFiles, err := cli.GetCertFlags(cmd)
	if err != nil {
		return err
	}

	forceGenerateCA := false
	//forceGenerateCA := cmd.Flags().Lookup("generate-ca-certs").Value.String() == "true"
	return initCertificates(clusterName, nil, forceGenerateCA, userCACertsFiles)
}

func initPackageRun(cmd *cobra.Command, args []string) error {
	targetFilename := cmd.Flags().Lookup("target").Value.String()

	// cluster name is optional if target flag is used (either cluster name or target flag is required)
	if len(args) == 0 && len(targetFilename) == 0 {
		return cli.UserErrorf("requires a cluster name or target path to store the package")
	}
	if len(args) > 1 {
		return cli.UserErrorf("accepts 1 cluster name, received %d. %v", len(args), args)
	}
	clusterName := args[0]
	if len(clusterName) == 0 && len(targetFilename) == 0 {
		return cli.UserErrorf("cluster name cannot be empty or use the targe flag to indicate packge location")
	}

	format := cmd.Flags().Lookup("format").Value.String()
	format = strings.ToLower(format)

	// get format from filename or error if they are different
	if len(targetFilename) != 0 {
		formatFile := filepath.Ext(targetFilename)
		if len(formatFile) != 0 && len(format) != 0 && formatFile != format {
			return cli.UserErrorf("target file format (%s) and requested package format (%s) are different", formatFile, format)
		}
		if len(formatFile) != 0 && len(format) == 0 {
			format = formatFile
		}
	}

	// validate format or use default if empty
	switch format {
	case "rpm", "deb":
	case "":
		format = defPackageFormat
	default:
		return cli.UserErrorf("unkwnow package format %q, use 'rpm' or 'deb'", format)
	}

	// if target and cluster name are in use, use target
	// if target is a directory, use the default filename: kubekit.$format
	// if cluster name, use cluster directory (if cluster exists)
	if len(targetFilename) == 0 {
		clusterDir := kluster.Path(clusterName, config.ClustersDir())
		if len(clusterDir) == 0 {
			return cli.UserErrorf("cluster name %q not found on your system", clusterName)
		}
		targetFilename = filepath.Join(clusterDir, "kubekit."+format)
	} else {
		targetStat, err := os.Stat(targetFilename)

		// if err != nil, file does not exists which is ok, but if it's a directory then error
		// if no extension, then should be a directory, so error if doesn't exists
		if os.IsNotExist(err) && filepath.Ext(targetFilename) == "" {
			return cli.UserErrorf("looks like target %q is a directory that doesn't exists", targetFilename)
		}
		if err == nil && targetStat.IsDir() {
			targetFilename = filepath.Join(targetFilename, "kubekit."+format)
		}
		// if err == nil and is a file, it will be overwriten ... is that ok? or should it err?
	}

	// DEBUG:
	cmd.Printf("init package %s --format %s --target %s\n", clusterName, format, targetFilename)

	return initPackage(targetFilename, format)
}

func initTemplate(templateName string, platforms []string, path, format string, variables map[string]string) (*kluster.Kluster, error) {
	if len(templateName) == 0 {
		return nil, cli.UserErrorf("template name cannot be empty")
	}
	if len(path) == 0 {
		tplDir := filepath.Join(config.ClustersDir(), "..", "templates")
		path = filepath.Clean(tplDir)
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, cli.UserErrorf("path %q does not exists", path)
		}
	}
	if strings.Contains(templateName, "_") {
		templateName = strings.Replace(templateName, "_", "-", -1)
		config.UI.Log.Warnf("template name cannot contain dashed ('-'), they were replaced by underscore ('_'). New template name: %q ", templateName)
	}

	template, err := kluster.NewTemplate(templateName, platforms, path, format, config.UI, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the template %s. %s", templateName, err)
	}
	template.Save()

	return template, nil
}

func initCertificates(clusterName string, cluster *kluster.Kluster, forceGenCA bool, userCACertsFiles tls.KeyPairs) (err error) {
	if cluster == nil {
		cluster, err = loadCluster(clusterName)
		if err != nil {
			return err
		}
	}
	return cluster.GenerateCerts(userCACertsFiles, forceGenCA)
}

func initPackage(filename, format string) error {
	return nil
}
