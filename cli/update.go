package cli

import "github.com/spf13/cobra"

// UpdateOpts encapsulate all the CLI parameters received from the `update` command
type UpdateOpts struct {
	ClusterName  string
	Variables    map[string]string
	Credentials  map[string]string
	TemplateName string // --template flags is not functional yet
}

// UpdateGetOpts get the `update` command parameters from the cobra commands and arguments
func UpdateGetOpts(cmd *cobra.Command, args []string) (opts *UpdateOpts, warns []string, err error) {
	warns = make([]string, 0)

	// Validate cluster name
	clusterName, err := GetOneClusterName(cmd, args, false)
	if err != nil {
		return nil, warns, err
	}

	// TODO: templateName will be used later to create a cluster from a template
	var templateName string
	if templateNameFlag := cmd.Flags().Lookup("template"); templateNameFlag != nil {
		templateName = templateNameFlag.Value.String()
	}

	// Variables:
	varsStr := cmd.Flags().Lookup("var").Value.String()
	variables, warnV, errV := GetVariables(varsStr)
	if errV != nil {
		return nil, warns, err
	}
	if len(warnV) != 0 {
		warns = append(warns, warnV...)
	}

	// Credentials:
	creds := GetGenericCredentials(cmd)

	return &UpdateOpts{
		ClusterName:  clusterName,
		Variables:    variables,
		Credentials:  creds,
		TemplateName: templateName,
	}, warns, nil
}
