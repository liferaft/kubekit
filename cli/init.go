package cli

import (
	"fmt"
	"strings"

	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/spf13/cobra"
)

// InitOpts encapsulate all the CLI parameters received from the `init` command
type InitOpts struct {
	ClusterName  string
	Platform     string
	Path         string
	Format       string
	Variables    map[string]string
	Credentials  map[string]string
	TemplateName string // --template flags is not functional yet
	Update       bool   // Deprecate --update flag when `update` command exists
}

// GetMultipleClustersName retrive from the CLI (cobra) arguments one or more clusters name
func GetMultipleClustersName(cmd *cobra.Command, args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, UserErrorf("requires a cluster name")
	}
	return args, nil
}

// GetOneClusterName retrive from the CLI (cobra) arguments one cluster name which could be valid or not
func GetOneClusterName(cmd *cobra.Command, args []string, validate bool) (clusterName string, err error) {
	if len(args) == 0 {
		return "", UserErrorf("requires a cluster name")
	}
	if len(args) != 1 {
		return "", UserErrorf("accepts 1 cluster name, received %d. %v", len(args), args)
	}
	if len(args[0]) == 0 {
		return "", UserErrorf("cluster name cannot be empty")
	}
	if validate {
		return kluster.ValidClusterName(args[0])
	}
	return args[0], nil
}

// InitGetOpts get the `init` command parameters from the cobra commands and arguments
func InitGetOpts(cmd *cobra.Command, args []string) (opts *InitOpts, warns []string, err error) {
	warns = make([]string, 0)

	// Validate cluster name
	clusterName, err := GetOneClusterName(cmd, args, true)
	if err != nil {
		if clusterName == "" {
			return nil, warns, err
		}
		warns = append(warns, fmt.Sprintf("%s. They were replaced and the new cluster name is: %q", err, clusterName))
	}

	// The `--update` flag will be deprecated and replaced by the `update` command
	update := false
	updateFlag := cmd.Flags().Lookup("update")
	if updateFlag != nil {
		update = updateFlag.Value.String() == "true"
	}

	// Validate platform (required unless it's an update)
	platform := cmd.Flags().Lookup("platform").Value.String()
	if len(platform) == 0 && !update {
		return nil, warns, UserErrorf("platform is required")
	}
	platform = strings.ToLower(platform)

	// The `--path` and `--format` flags are only part of the `kubekit` binary
	var path string
	if pathFlag := cmd.Flags().Lookup("path"); pathFlag != nil {
		path = pathFlag.Value.String()
	}
	var format string
	if formatFlag := cmd.Flags().Lookup("format"); formatFlag != nil {
		format = formatFlag.Value.String()
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
	creds := GetCredentials(platform, cmd)

	return &InitOpts{
		ClusterName:  clusterName,
		Platform:     platform,
		Path:         path,
		Format:       format,
		Variables:    variables,
		Credentials:  creds,
		TemplateName: templateName,
		Update:       update,
	}, warns, nil
}
