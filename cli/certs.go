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

// GetCredentials get the credentials from cobra CLI flags and insert them into
// the list of variables. Returns, as a warning, the list of variables ignored
// or replaced
func GetCredentials(platform string, cmd *cobra.Command) map[string]string {
	platform = strings.ToLower(platform)

	creds := map[string]string{}

	switch platform {
	case "vra", "raw", "stacki":
		// These platforms do not request for credentials
		return creds
	}

	getCredentialNamed := func(name string) {
		// Get the value of the flag `name` (if exists) and append it to the list of credentials
		if flag := cmd.Flags().Lookup(name); flag != nil {
			// The flag may be defined (not nil) but not set by the user (value == "")
			if value := flag.Value.String(); value != "" {
				creds[name] = value
				// The flags has priority over the environment variable, so return if found
				return
			}
		}

		// Unfortunatelly the flag and variable name of these credentials are not like the standard AWS environment variables
		// TODO: Make these flags and variables name like the standard AWS variables
		var envName string
		switch name {
		case "access_key":
			envName = "access_key_id"
		case "secret_key":
			envName = "secret_access_key"
		case "region":
			envName = "default_region"
		default:
			envName = name
		}

		// For EKS and AKS the platform is AWS and AZURE respectivelly
		switch platform {
		case "aws", "eks":
			platform = "aws"
		case "azure", "aks":
			platform = "azure"
		}

		// Get the value of the environment variable `PLATFORM_NAME` (if exists) and append it to the list of variables
		if envValue := os.Getenv(strings.ToUpper(platform + "_" + envName)); envValue != "" {
			creds[name] = envValue
		}
	}

	// In nameOfVarAndFlag the key is the name of the variable and the value is
	// the name of the flag. Most of the time is the same, except for the generic/server credentials
	var varNames []string
	switch platform {
	case "aws", "eks":
		varNames = []string{
			"access_key",
			"secret_key",
			"session_token",
			"region",
			"profile",
		}
	case "azure", "aks":
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

	for _, name := range varNames {
		getCredentialNamed(name)
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
