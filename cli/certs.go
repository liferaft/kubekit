package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/liferaft/kubekit/pkg/crypto/tls"
	"github.com/liferaft/kubekit/pkg/kluster"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AddCertFlags adds the flags to the given command to receive certificate
// files: private key and certificate.
func AddCertFlags(cmd *cobra.Command) {
	for _, caCertInfo := range kluster.CACertNames {
		if len(caCertInfo.Desc) == 0 {
			continue
		}
		cmd.Flags().String(caCertInfo.CN+"-key-file", "", "CA RSA Key file "+caCertInfo.Desc+", recommended for production.")
		cmd.Flags().String(caCertInfo.CN+"-cert-file", "", "CA x509 Certificate file "+caCertInfo.Desc+", recommended for production.")
	}
}

// GetCertFlags creates the list of CA certificates with the original
// file name and CN. The original filename will be replaced when the certificate
// is saved in the certificates directory.
func GetCertFlags(cmd *cobra.Command) (tls.KeyPairs, error) {
	userCACertsFiles := make(tls.KeyPairs, len(kluster.CACertNames))

	home, err := homedir.Dir()
	if err != nil {
		panic(fmt.Errorf("cannot find home directory. %s", err))
	}

	getFileName := func(flagName string) string {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			return ""
		}

		filename := flag.Value.String()

		// Expand `~` if it's part of keyFile
		if strings.HasPrefix(filename, "~/") {
			return filepath.Join(home, filename[2:])
		}

		return filename
	}

	for name, caCertInfo := range kluster.CACertNames {
		var keyFile, certFile string

		keyFile = getFileName(caCertInfo.CN + "-key-file")
		certFile = getFileName(caCertInfo.CN + "-cert-file")

		var keyPEM, certPEM []byte
		if keyFile != "" {
			if keyPEM, err = ioutil.ReadFile(keyFile); err != nil {
				return nil, fmt.Errorf("failed to read the tls key file %q. %s", keyFile, err)
			}
		}
		if certFile != "" {
			if certPEM, err = ioutil.ReadFile(certFile); err != nil {
				return nil, fmt.Errorf("failed to read the tls certificate file %q. %s", certFile, err)
			}
		}

		userCACertsFiles[name] = &tls.KeyPair{
			Name:           name,
			KeyFile:        keyFile,
			PrivateKeyPEM:  keyPEM,
			CN:             caCertInfo.CN,
			CertFile:       certFile,
			CertificatePEM: certPEM,
			IsCA:           true,
		}
	}

	return userCACertsFiles, nil
}

func getVarNames(platform string) (string, []string) {
	platform = strings.ToLower(platform)

	var varNames []string
	switch platform {
	case "vra", "raw", "stacki":
		// These platforms do not request for credentials
		// do nothing and return an empty list of variables
	case "aws", "ec2", "eks":
		// For EC2 and EKS the platform name in environment variables is AWS
		platform = "aws"
		varNames = []string{
			"access_key",
			"secret_key",
			"session_token",
			"region",
			"profile",
		}
	case "azure", "aks":
		// For AKS the platform name in environment variables is AZURE
		platform = "azure"
		varNames = []string{
			"subscription_id",
			"tenant_id",
			"client_id",
			"client_secret",
		}
	default:
		varNames = []string{
			"server",
			"username",
			"password",
		}
	}

	return platform, varNames
}

func getVarValue(name, envPrefix string, flag *pflag.Flag) string {
	// Get the value of the flag `name` (if exists) and append it to the list of credentials
	if flag != nil {
		// The flag may be defined (not nil) but not set by the user (value == "")
		if value := flag.Value.String(); value != "" {
			// The flags has priority over the environment variable, so continue to the next parameter if found
			return value
		}
	}

	// Unfortunatelly the flag and variable name of these credentials are not like the standard AWS environment variables
	// TODO: Make these flags and variables name like the standard AWS variables
	switch name {
	case "access_key":
		name = "access_key_id"
	case "secret_key":
		name = "secret_access_key"
	case "region":
		name = "default_region"
	}

	return os.Getenv(strings.ToUpper(envPrefix + "_" + name))
}

// GetCredentials get the credentials from cobra CLI flags and insert them into
// the list of variables. Returns, as a warning, the list of variables ignored
// or replaced
func GetCredentials(platform string, cmd *cobra.Command) map[string]string {
	creds := map[string]string{}

	envPrefix, varNames := getVarNames(platform)
	for _, name := range varNames {
		if value := getVarValue(name, envPrefix, cmd.Flags().Lookup(name)); value != "" {
			// Get the value of the environment variable `PLATFORM_NAME` (if exists) and append it to the list of variables
			creds[name] = value
		}
	}

	return creds
}

// GetGenericCredentials get the credentials from cobra CLI flags for any given
// platform and insert them into the list of variables. Returns, as a warning,
// the list of variables ignored or replaced
func GetGenericCredentials(cmd *cobra.Command) map[string]string {
	credMap := map[string]string{}

	// Only check for AWS, Azure and vSphere. Any the other platform will have the
	// same credential parameters as these 3. Do not use "vra", "raw" or "stacki",
	// they do not have credentials
	for _, platform := range []string{"aws", "azure", "vsphere"} {
		cred := GetCredentials(platform, cmd)
		for k, v := range cred {
			credMap[k] = v
		}
	}

	return credMap
}
