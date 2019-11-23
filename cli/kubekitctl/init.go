package kubekitctl

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/cli"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init [cluster] NAME",
	Aliases: []string{"i"},
	Short:   "Initialize a cluster configuration",
	Long:    `Init is used to create a cluster configuration. This is usually the first command to execute.`,
	RunE:    initClusterRun,
}

// initClusterCmd represents the 'init cluster' command
var initClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Initialize a cluster configuration",
	Long: `The command init cluster is used to generate a cluster configuration using 
default values for the given platform and the provided variables in the flags or 
environment variables.`,
	RunE: initClusterRun,
}

func initAddCommands() {
	// init [cluster] NAME --platform NAME --var NAME01=VALUE01 --var NAME02=VALUE02 ...
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("platform", "p", "", "platform where the cluster going to be provisioned (Example: aws, vsphere)")
	initCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")
	// ... [credentials]
	initCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	initCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	initCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EKS credentials]
	initCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	initCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	initCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	initCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	initCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	initCmd.Flags().String("subscription_id", "", "Azure Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	initCmd.Flags().String("tenant_id", "", "Azure Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	initCmd.Flags().String("client_id", "", "Azure Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	initCmd.Flags().String("client_secret", "", "Azure Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")

	initCmd.AddCommand(initClusterCmd)
	initClusterCmd.Flags().StringP("platform", "p", "", "platform where the cluster going to be provisioned (Example: aws, vsphere)")
	initClusterCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")
	// ... [credentials]
	initClusterCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	initClusterCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	initClusterCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EKS credentials]
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
}

func initClusterRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.InitGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	// DEBUG:
	config.Logger.Debugf("init cluster %s --platform %s --var %v\n", opts.ClusterName, opts.Platform, opts.Variables)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	init, err := config.client.Init(ctx, opts.ClusterName, opts.Platform, opts.Variables, opts.Credentials)
	if err != nil {
		return err
	}

	fmt.Println(init)

	return nil
}
