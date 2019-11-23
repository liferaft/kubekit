package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteOpts encapsulate all the CLI parameters received from the `delete` command
type DeleteOpts struct {
	ClusterName string
	DestroyAll  bool
	Force       bool
}

// DeleteGetOpts get the `delete` command parameters from the cobra commands and arguments
func DeleteGetOpts(cmd *cobra.Command, args []string) (opts *DeleteOpts, warns []string, err error) {
	warns = make([]string, 0)

	// cluster_name
	clusterName, err := GetOneClusterName(cmd, args, false)
	if err != nil {
		return nil, warns, err
	}

	// Get the flags `--all` and `--force`
	destroyAll := false
	destroyAllFlag := cmd.Flags().Lookup("all")
	if destroyAllFlag != nil {
		destroyAll = destroyAllFlag.Value.String() == "true"
	}
	force := false
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag != nil {
		force = forceFlag.Value.String() == "true"
	}

	opts = &DeleteOpts{
		ClusterName: clusterName,
		DestroyAll:  destroyAll,
		Force:       force,
	}

	return opts, warns, nil
}

// Confirm ask to the user to confirm to delete the cluster
func (opts *DeleteOpts) Confirm() bool {
	deleteIt := true
	if !opts.Force {
		question := fmt.Sprintf("Do you want to destroy cluster %q", opts.ClusterName)
		deleteIt = HardConfirmation(question, "yes")
	}

	return deleteIt
}
