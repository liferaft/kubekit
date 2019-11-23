package version

import (
	"bytes"
	"fmt"

	"github.com/kubekit/kubekit/pkg/manifest"
)

var (
	// GitCommit is the git commit that was compiled. This will be filled in by the compiler
	// in the Makefile from Git short SHA-1 of HEAD commit.
	GitCommit string

	// Build is the Jenkins build that compiled this version. This will be filled
	// in by the compiler in the Makefile from Jenkins build.
	Build string

	// Version is the main version number that is being run at the moment. This will be
	// filled in by the compiler in the Makefile from latest.go
	Version string

	// Prerelease is a pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	Prerelease string

	// AppName is the application name to show with the version. It may be empty
	// but looks good to have a name.
	AppName = "KubeKit"
)

func init() {
	Version = manifest.Version
	Prerelease = manifest.Prerelease
}

// Println prints the version using the output of String()
func Println() {
	fmt.Println(String())
}

// String return the version as it will be show in the terminal
func String() string {
	var version bytes.Buffer
	if AppName != "" {
		fmt.Fprintf(&version, "%s ", AppName)
	}
	fmt.Fprintf(&version, "v%s", Version)
	if Prerelease != "" {
		fmt.Fprintf(&version, "-%s", Prerelease)
	}

	return version.String()
}

// LongPrintln prints the version long format using the output of LongString()
func LongPrintln() {
	fmt.Println(LongString())
}

// LongString return the version in a long format that includes the build number
// and git commit SHA
func LongString() string {
	var version bytes.Buffer

	fmt.Fprintf(&version, "%s", String())

	if Build != "" {
		fmt.Fprintf(&version, "+build.%s", Build)
		if GitCommit != "" {
			fmt.Fprintf(&version, ".%s", GitCommit)
		}
	}

	release, ok := manifest.KubeManifest.Releases[Version]
	if !ok {
		return version.String()
	}

	if len(release.KubernetesVersion) != 0 {
		fmt.Fprintf(&version, "\nKubernetes version: %s", release.KubernetesVersion)
	}
	if len(release.DockerVersion) != 0 {
		fmt.Fprintf(&version, "\nDocker version: %s", release.DockerVersion)
	}
	if len(release.EtcdVersion) != 0 {
		fmt.Fprintf(&version, "\netcd version: %s", release.EtcdVersion)
	}

	return version.String()
}
