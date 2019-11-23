package kubekitctl

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints the KubeKit Server version",
	Long:  `Prints the KubeKit Server, Kubernetes, Docker and etcd version.`,
	RunE:  versionRun,
}

func versionAddCommands() {
	// version
	RootCmd.AddCommand(versionCmd)
}

func versionRun(cmd *cobra.Command, args []string) error {
	config.Logger.Debugf("version")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	version, err := config.client.Version(ctx)
	if err != nil {
		return err
	}

	fmt.Println(version)

	return nil
}
