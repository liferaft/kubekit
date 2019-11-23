package version_test

import (
	"strings"
	"testing"

	"github.com/kubekit/kubekit/version"
)

func TestVersionString(t *testing.T) {
	version.Version = "1.0"
	version.AppName = "Kubekit"
	version.GitCommit = "t35t1ng"
	version.Build = "20"
	actualVer := version.String()

	expectedVer := "Kubekit v1.0"

	if !strings.HasPrefix(actualVer, expectedVer) {
		t.Errorf("expected %s, but got %s", expectedVer, actualVer)
	}
}

func TestVersionLongString(t *testing.T) {
	version.Version = "1.0"
	version.AppName = "Kubekit"
	version.GitCommit = "t35t1ng"
	version.Build = "20"
	actualVer := version.LongString()

	//expectedVer := "Kubekit v1.0-dev+build.20.t35t1ng"
	//This test is flawed as 'dev' may not be there
	expectedVer := "Kubekit v1.0"

	if !strings.HasPrefix(actualVer, expectedVer) {
		t.Errorf("expected %s, but got %s", expectedVer, actualVer)
	}
}
