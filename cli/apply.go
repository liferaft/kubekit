package cli

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/kubekit/kubekit/pkg/crypto/tls"
)

// ApplyOpts encapsulate all the CLI parameters received from the `apply` command
type ApplyOpts struct {
	ClusterName  string
	Action       string
	PackageURL   string
	ForcePackage bool
	UserCACerts  tls.KeyPairs
}

// ApplyGetOpts get the `apply` command parameters from the cobra commands and arguments
func ApplyGetOpts(cmd *cobra.Command, args []string) (opts *ApplyOpts, warns []string, err error) {
	warns = make([]string, 0)

	// cluster_name
	clusterName, err := GetOneClusterName(cmd, args, false)
	if err != nil {
		return nil, warns, err
	}

	// --provision --configure
	var action string
	actions := map[string]bool{
		"provision": false,
		"configure": false,
	}
	for actionName := range actions {
		if actionFlag := cmd.Flags().Lookup(actionName); actionFlag != nil {
			actions[actionName] = actionFlag.Value.String() == "true"
		}
	}
	if (!actions["provision"] && !actions["configure"]) || (actions["provision"] && actions["configure"]) {
		// if both are flase, or both are true, then do all the actions (ALL = 0)
		action = "ALL"
	} else if actions["provision"] {
		// if only provision is true, then do provision only
		action = "PROVISION"
	} else {
		// the only left option is that configuration is set
		action = "CONFIGURE"
	}

	// --package-file --package-file-url
	var pkgURL string
	if pkgFileFlag := cmd.Flags().Lookup("package-file"); pkgFileFlag != nil {
		pkgFile := pkgFileFlag.Value.String()
		if pkgFile != "" {
			if pkgFile, err = filepath.Abs(pkgFile); err != nil {
				return nil, warns, fmt.Errorf("failed to get the absolute path of the given file path %q. %s", pkgFile, err)
			}
			pkgURL = "file://" + pkgFile
		}
	}
	if pkgURLFlag := cmd.Flags().Lookup("package-file-url"); pkgURLFlag != nil {
		if pkgURL != "" {
			return nil, warns, fmt.Errorf("flags 'package-file' and 'package-file-url' cannot be set at same time. If the file is local and working directly with KubeKit, use 'package-file'. If the file is remote use 'package-file-url'. Use 'package-file-url' if working with KubeKit as a client")
		}
		pkgURL := pkgURLFlag.Value.String()
		if pkgURL != "" {
			if _, err = url.Parse(pkgURL); err != nil {
				return nil, warns, fmt.Errorf("the value of 'package-file-url' (%q) is not a valid URL. %s", pkgURL, err)
			}
		}
	}

	// --force-pkg
	var forcePkg bool
	if forcePkgFlag := cmd.Flags().Lookup("force-pkg"); forcePkgFlag != nil {
		forcePkg = forcePkgFlag.Value.String() == "true"
	}

	userCACerts, err := GetCertFlags(cmd)
	if err != nil {
		return nil, warns, err
	}

	opts = &ApplyOpts{
		ClusterName:  clusterName,
		Action:       action,
		PackageURL:   pkgURL,
		ForcePackage: forcePkg,
		UserCACerts:  userCACerts,
	}

	return opts, warns, nil
}
