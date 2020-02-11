package kubekit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/liferaft/kubekit/cli"

	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/spf13/cobra"
)

var (
	doList bool
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:     "login [cluster] NAME",
	Aliases: []string{"l"},
	Short:   "Used to enter crendetials to the platform engine or login to a node",
	Long: `Required command for clusters as it provide the credentials to login or to use
the cluster platform. Each platform will have different kind of credentials.
When used with a node, it's for login or ssh into a node.`,
	RunE: loginClusterRun,
}

// loginClusterCmd represents the 'login cluster' command
var loginClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Used to enter crendetials to the platform engine",
	Long: `Required command for clusters as it provide the credentials to login or to use
the cluster platform. Each platform will have different kind of credentials.`,
	RunE: loginClusterRun,
}

// loginNodeCmd represents the 'login node' command
var loginNodeCmd = &cobra.Command{
	Use:     "node NAME",
	Aliases: []string{"n"},
	Short:   "To login to a cluster node",
	Long:    `To login or ssh into a cluster node.`,
	RunE:    loginNodeRun,
}

func addLoginCmd() {
	// login [cluster] NAME --platform platform --list --access_key aws_access_key_id --secret_key aws_secret_access_key --region aws_default_region --server server_ip_or_dns --username username --password password
	RootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolVar(&doList, "list", false, "lists the recognized credentials. Hides or replaces with *, the sensitive information")
	// ... [credentials]
	loginCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	loginCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	loginCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EC2/EKS credentials]
	loginCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	loginCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	loginCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	loginCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	loginCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	loginCmd.Flags().String("subscription_id", "", "Provisioner Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	loginCmd.Flags().String("tenant_id", "", "Provisioner Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	loginCmd.Flags().String("client_id", "", "Provisioner Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	loginCmd.Flags().String("client_secret", "", "Provisioner Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")

	loginCmd.AddCommand(loginClusterCmd)
	loginClusterCmd.Flags().BoolVar(&doList, "list", false, "lists the recognized credentials. Hides or replaces with *, the sensitive information")
	// ... [credentials]
	loginClusterCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	loginClusterCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	loginClusterCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EC2/EKS credentials]
	loginClusterCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	loginClusterCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	loginClusterCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	loginClusterCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	loginClusterCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	loginClusterCmd.Flags().String("subscription_id", "", "Provisioner Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	loginClusterCmd.Flags().String("tenant_id", "", "Provisioner Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	loginClusterCmd.Flags().String("client_id", "", "Provisioner Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	loginClusterCmd.Flags().String("client_secret", "", "Provisioner Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")

	// login node NAME --cluster NAME
	loginCmd.AddCommand(loginNodeCmd)
	loginNodeCmd.Flags().StringP("cluster", "c", "", "cluster name where this node is located")
}

func loginClusterRun(cmd *cobra.Command, args []string) error {
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
	// var listFlag string
	// if doList {
	// 	listFlag = " --list"
	// }
	// fmt.Printf("login cluster %s%s --access_key %s --secret_key %s --region %s --server %s --username %s --password %s\n",
	// 	clusterName, listFlag, loginOpts.accessKey, loginOpts.secretKey, loginOpts.region, loginOpts.server, loginOpts.username, loginOpts.password)

	clusterList, err := kluster.List(config.ClustersDir(), clusterName)
	if err != nil {
		return err
	}
	if clusterList == nil || len(clusterList) == 0 {
		return fmt.Errorf("cluster %q not found", clusterName)
	}
	cluster := clusterList[0]
	platform := cluster.Platform()
	path := filepath.Join(filepath.Dir(cluster.Path()), ".credentials")

	if doList {
		// If --list, then list what you have in the credentials file
		credentials := kluster.NewCredentials(clusterName, platform, path)
		if err := credentials.Read(); err != nil {
			return err
		}
		// list them and exit
		credentials.List()
		return nil
	}

	// create empty credentials
	credentials := kluster.NewCredentials(cluster.Name, platform, path)

	// 1st: read credentials from the flags, if not there take it from env variable
	creds := cli.GetCredentials(platform, cmd)
	config.UI.Log.Debugf("assigning credentials from flags and environment variables, in that order")
	if err := credentials.AssignFromMap(creds); err != nil {
		return err
	}

	if credentials.Complete() {
		// write if credentials were completed from flags/env variables
		return credentials.Write()
	}

	// recreate the empty credentials so we don't get credentials from mixed sources
	credentials = kluster.NewCredentials(cluster.Name, platform, path)

	// 2nd: read the parameters from the AWS configuration, if this is AWS
	switch platform {
	case "ec2", "eks":
		getAWSCredVariables(credentials.(*kluster.AwsCredentials), creds)
		if credentials.Complete() {
			return credentials.Write()
		}
		credentials = kluster.NewCredentials(cluster.Name, platform, path)
	}

	// 3rd: read the credentials file
	if err := credentials.Read(); err != nil {
		config.UI.Log.Warnf("cannot read the credentials file: %s", err)
	}

	if !credentials.Complete() {
		// 4th: ask the user to provide the missing values
		config.UI.Log.Debugf("get credentials asking the user")
		if err := credentials.Ask(); err != nil {
			return err
		}
	}

	return credentials.Write()
}

func loginNodeRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cli.UserErrorf("requires a node hostname or IP address")
	}
	if len(args) != 1 {
		return cli.UserErrorf("accepts 1 node, received %d. %v", len(args), args)
	}
	nodeName := args[0]
	if len(nodeName) == 0 {
		return cli.UserErrorf("node hostname or IP cannot be empty")
	}

	// TODO: Should we remove this and search the node in every cluster?
	clusterName := cmd.Flags().Lookup("cluster").Value.String()
	if len(clusterName) == 0 {
		return cli.UserErrorf("cluster name is required")
	}

	// DEBUG:
	// fmt.Printf("login node %s --cluster %s\n", nodeName, clusterName)

	cluster, err := loadCluster(clusterName)
	if err != nil {
		return err
	}
	return cluster.StartShellTo(nodeName, os.Stdin, os.Stdout, os.Stderr)
}

// getAWSCredVariables read the AWS credentials from the AWS shared profile and
// configuration
func getAWSCredVariables(awsCred *kluster.AwsCredentials, variables map[string]string) {
	// return if there is everything that is needed
	if awsCred.Complete() {
		return
	}

	profile := variables["profile"]
	if len(profile) == 0 {
		profile = variables["aws_profile"]
	}
	if len(profile) == 0 {
		profile = os.Getenv("AWS_PROFILE")
	}
	if len(profile) == 0 {
		profile = "default"
	}

	config.UI.Log.Warn("AWS credentials incomplete or not provided in flags and environment, taking them from local AWS configuration")

	// Get the missing parameters from AWS configuration. Do not force/overwrite existing parameters
	errCred := awsCred.LoadSharedCredentialsFromProfile(profile, false)
	errReg := awsCred.LoadSharedRegionFromProfile(profile, false)

	if errCred == nil && errReg == nil {
		// ... then set the profile name
		awsCred.Profile = profile
	}
}
