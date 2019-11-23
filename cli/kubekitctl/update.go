package kubekitctl

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/cli"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update [cluster] NAME",
	Aliases: []string{"i"},
	Short:   "Update a cluster configuration",
	Long:    `Update is used to modify an existing cluster configuration.`,
	RunE:    updateClusterRun,
}

// updateClusterCmd represents the 'update cluster' command
var updateClusterCmd = &cobra.Command{
	Use:     "cluster NAME",
	Aliases: []string{"c"},
	Short:   "Update a cluster configuration",
	Long: `The command update cluster is used to modify an existing cluster configuration 
using provided variables in the flags or environment variables.`,
	RunE: updateClusterRun,
}

func updateAddCommands() {
	// update [cluster] NAME --var NAME01=VALUE01 --var NAME02=VALUE02 ...
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")
	// ... [credentials]
	updateCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	updateCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	updateCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EKS credentials]
	updateCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	updateCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	updateCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	updateCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	updateCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	updateCmd.Flags().String("subscription_id", "", "Azure Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	updateCmd.Flags().String("tenant_id", "", "Azure Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	updateCmd.Flags().String("client_id", "", "Azure Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	updateCmd.Flags().String("client_secret", "", "Azure Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")

	updateCmd.AddCommand(updateClusterCmd)
	updateClusterCmd.Flags().StringArray("var", []string{}, "KubeKit variable to be used for the cluster configuration")
	// ... [credentials]
	updateClusterCmd.Flags().String("server", "", "Provisioner Server IP or DNS. Also retrived from $KUBEKIT_<PLATFORM>_SERVER or $<PLATFORM>_SERVER, like $VSPHERE_SERVER")
	updateClusterCmd.Flags().String("username", "", "Provisioner Username. Also retrived from $KUBEKIT_<PLATFORM>_USERNAME or $<PLATFORM>_USERNAME, like $VSPHERE_USERNAME")
	updateClusterCmd.Flags().String("password", "", "Provisioner Password. Also retrived from $KUBEKIT_<PLATFORM>_PASSWORD or $<PLATFORM>_PASSWORD, like $VSPHERE_PASSWORD")
	// ... [AWS/EKS credentials]
	updateClusterCmd.Flags().String("access_key", "", "AWS Access Key Id. Also retrived from $KUBEKIT_AWS_ACCESS_KEY_ID or $AWS_ACCESS_KEY_ID")
	updateClusterCmd.Flags().String("secret_key", "", "AWS Secret Access Key. Also retrived from $KUBEKIT_AWS_SECRET_ACCESS_KEY or $AWS_SECRET_ACCESS_KEY")
	updateClusterCmd.Flags().String("session_token", "", "AWS Secret Session Token. Also retrived from $KUBEKIT_AWS_SESSION_TOKEN or $AWS_SESSION_TOKEN")
	updateClusterCmd.Flags().String("region", "", "AWS Default Region. Also retrived from $KUBEKIT_AWS_DEFAULT_REGION or $AWS_DEFAULT_REGION")
	updateClusterCmd.Flags().String("profile", "", "AWS Profile. Also retrived from $KUBEKIT_AWS_PROFILE or $AWS_PROFILE")
	// ... [Azure credentials]
	updateClusterCmd.Flags().String("subscription_id", "", "Provisioner Subscription ID. Also retrived from $KUBEKIT_AZURE_SUBSCRIPTION or $AZURE_SUBSCRIPTION")
	updateClusterCmd.Flags().String("tenant_id", "", "Provisioner Tenant ID. Also retrived from $KUBEKIT_AZURE_TENANT_ID or $AZURE_TENANT_ID")
	updateClusterCmd.Flags().String("client_id", "", "Provisioner Client ID. Also retrived from $KUBEKIT_AZURE_CLIENT_ID or $AZURE_CLIENT_ID")
	updateClusterCmd.Flags().String("client_secret", "", "Provisioner Client Secret. Also retrived from $KUBEKIT_AZURE_CLIENT_SECRET or $AZURE_CLIENT_SECRET")
}

func updateClusterRun(cmd *cobra.Command, args []string) error {
	opts, warns, err := cli.UpdateGetOpts(cmd, args)
	if err != nil {
		return err
	}
	if len(warns) != 0 {
		for _, w := range warns {
			config.Logger.Warn(w)
		}
	}

	// DEBUG:
	config.Logger.Debugf("update cluster %s --var %v\n", opts.ClusterName, opts.Variables)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	output, err := config.client.UpdateCluster(ctx, opts.ClusterName, opts.Variables, opts.Credentials)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
