package kubekitctl

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Authenticate using AWS IAM and get token for Kubernetes",
	Long: `Mimic the token command from aws-iam-authenticator until EKS doesn't need it 
anymore. This command elimiate the need to install aws-iam-authenticator.`,
	RunE: tokenRun,
}

func tokenAddCommands() {
	// token -i <CLUSTER-NAME> -r [ROLE]
	RootCmd.AddCommand(tokenCmd)
	tokenCmd.Flags().StringP("cluster-id", "i", "", "Specify the cluster `ID`, a unique-per-cluster identifier for your aws-iam-authenticator installation.")
	tokenCmd.Flags().StringP("role", "r", "", "Assume an IAM Role ARN before signing this token")
}

func tokenRun(cmd *cobra.Command, args []string) error {
	clusterID := cmd.Flags().Lookup("cluster-id").Value.String()
	roleARN := cmd.Flags().Lookup("role").Value.String()

	// DEBUG:
	config.Logger.Debugf("token -i %s -r %s", clusterID, roleARN)

	if len(clusterID) == 0 {
		return fmt.Errorf("requires a cluster name")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	token, err := config.client.Token(ctx, clusterID, roleARN)
	if err != nil {
		return err
	}

	fmt.Println(token)

	return nil
}
