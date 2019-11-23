package kubekit

import (
	"fmt"

	"github.com/spf13/cobra"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Hidden: true,
	Use:    "scale [cluster] NAME POOL-NAME=[+|-]NUMBER",
	Short:  "scales up or down a cluster",
	Long: `Scales by increasing or decreasing the number of nodes in a cluster from a given
pool. The number could be possitive (to add), negative (to remove) or unsigned
number (assign).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'scale' still not implemented")
		return nil
	},
}

// scaleClusterCmd represents the cluster command
var scaleClusterCmd = &cobra.Command{
	Hidden: true,
	Use:    "cluster NAME POOL-NAME=[+|-]NUMBER",
	Short:  "scales up or down a cluster",
	Long: `Scales by increasing or decreasing the number of nodes in a cluster from a given
pool. The number could be possitive (to add), negative (to remove) or unsigned
number (assign).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'scale cluster' still not implemented")
		return nil
	},
}

func addScaleCmd() {
	// scale [cluster] NAME POOL-NAME=[+|-]N
	RootCmd.AddCommand(scaleCmd)
	scaleCmd.AddCommand(scaleClusterCmd)
}
