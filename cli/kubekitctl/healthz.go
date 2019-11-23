package kubekitctl

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// healthzCmd represents the healthz command
var healthzCmd = &cobra.Command{
	Use:   "healthz",
	Short: "check health status of the service",
	Long:  `Use it to get the health check status of the KubeKit service.`,
	RunE:  healthzRun,
}

func healthzAddCommands() {
	// healthz
	RootCmd.AddCommand(healthzCmd)
}

func healthzRun(cmd *cobra.Command, args []string) error {
	config.Logger.Debugf("healthz --host %s --healthz-port %s", config.Host, config.PortHealthz)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if config.client.GrpcConn != nil {
		defer config.client.GrpcConn.Close()
	}

	status, err := config.client.Healthz(ctx, apiVersion+".Kubekit")
	if err != nil {
		return err
	}

	fmt.Printf("Healthz Status: %s\n", status)

	return nil
}
