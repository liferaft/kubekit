package kubekit

import (
	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the KubeKit version",
	Long:  `Print the KubeKit version and git commit SHA`,
	Example: `	kubekit version
		kubekit version --verbose
	`,
	RunE: versionRun,
}

func addVersionCmd() {
	// --version
	RootCmd.Flags().BoolVar(&versionFlag, "version", false, "Print the kubekit version")
	// version
	RootCmd.AddCommand(versionCmd)
}

func versionRun(cmd *cobra.Command, args []string) error {
	longVersion := cmd.Flags().Lookup("verbose").Changed
	printVersion(longVersion)
	return nil
}

func printVersion(longVersion bool) {
	if longVersion {
		version.LongPrintln()
	} else {
		version.Println()
	}
}
