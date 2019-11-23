package kubekit

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/pkg/aws_iam_authenticator/token"
)

// When EKS does not need 'aws-iam-authenticator' as a dependency, this command
// can be eliminated by deleting this file, removing the call to 'addTokenCmd'
// from 'cmd.go' around line 100 and delete the package from the vendors

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Hidden:  true,
	Use:     "token",
	Aliases: []string{"t"},
	Short:   "Authenticate using AWS IAM and get token for Kubernetes",
	Long: `Mimic the token command from aws-iam-authenticator until EKS doesn't need it 
anymore. This command elimiate the need to install aws-iam-authenticator.`,
	RunE: tokenRun,
}

func addTokenCmd() {
	RootCmd.AddCommand(tokenCmd)
	tokenCmd.Flags().StringP("cluster-id", "i", "", "Specify the cluster `ID`, a unique-per-cluster identifier for your aws-iam-authenticator installation.")
	tokenCmd.Flags().StringP("role", "r", "", "Assume an IAM Role ARN before signing this token")
}

func tokenRun(cmd *cobra.Command, args []string) error {
	clusterID := cmd.Flags().Lookup("cluster-id").Value.String()
	roleARN := cmd.Flags().Lookup("role").Value.String()
	// DEBUG:
	// fmt.Printf("token -i %s -r %s\n", clusterID, roleARN)

	if len(clusterID) == 0 {
		return fmt.Errorf("requires a cluster name")
	}

	cluster, err := loadCluster(clusterID)
	if err != nil {
		return err
	}

	token, err := token.GenerateToken(cluster, roleARN)
	if err != nil {
		return err
	}

	fmt.Println(token.FormatJSON())

	return nil
}
