package kubekit_test

import (
	"strings"
	"testing"

	"os"

	"github.com/kubekit/kubekit/cli/kubekit"
	"github.com/kubekit/kubekit/version"
)

// TestVersionCmd tests the 'version' command
func TestVersionCmd(t *testing.T) {
	version.Version = "1.0"
	version.AppName = "Kubekit"
	version.GitCommit = "t35t1ng"
	version.Build = "20"
	expectedVer := "Kubekit"

	// Even with verbose set to 'true', the long version is only printed with
	// '-v'/'--verbose' flag
	os.Setenv("KUBEKIT_VERBOSE", "true")
	kubekitCmd.SetArgs(strings.Split("version", " "))

	actualVer, err := getOutput(kubekit.Execute)
	if err != nil {
		t.Error(err)
	}

	if !strings.HasPrefix(actualVer, expectedVer) {
		t.Errorf("expected '%s', but got '%s'", expectedVer, actualVer)
	}
}

// TestVersionCmd tests the 'version' command with verbose flag
// func TestVerboseVersionCmd(t *testing.T) {
// 	version.Version = "1.0"
// 	version.AppName = "Kubekit"
// 	version.GitCommit = "t35t1ng"
// 	version.Build = "20"
// 	expectedVer := "Kubekit v1.0+build.20.t35t1ng"

// 	// If with verbose set to 'false', the long version is only printed with
// 	// '-v'/'--verbose' flag
// 	os.Setenv("KUBEKIT_VERBOSE", "false")
// 	cmd.RootCmd.SetArgs(strings.Split("version --verbose", " "))

// 	actualVer, err := getOutput(cmd.Execute)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if actualVer != expectedVer {
// 		t.Errorf("expected '%s', but got '%s'", expectedVer, actualVer)
// 	}
// }
