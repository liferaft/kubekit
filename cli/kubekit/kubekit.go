package kubekit

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubekit",
	Short: "KubeKit is a Kubernetes toolkit",
	Long: `KubeKit is a toolkit for setting up a Kubernetes-powered cluster. It's the
easiest way to get a Kubernetes cluster on different platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		longVersion := cmd.Flags().Lookup("verbose").Changed
		if versionFlag {
			printVersion(longVersion)
			return
		}

		cmd.HelpFunc()(cmd, args)
	},
}
