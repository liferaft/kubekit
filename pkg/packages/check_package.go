package packages

import (
	"fmt"
	"github.com/cavaliercoder/go-rpm"

	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/liferaft/kubekit/pkg/manifest"
)

//GetBaseImages checks what is installed on the remote machine
func GetBaseImages(cluster *kluster.Kluster, forcePkg bool, pkgFilename string) error {
	//No need to run as the rpm was already checked
	if len(pkgFilename) > 0 || forcePkg {
		return nil
	}

	failure := false

	check := getPackages()

	for file := range check {
		cmd := fmt.Sprintf("ls %s", file)
		result, err := cluster.Exec(cmd, "", nil, nil, true)
		if result.Failures != 0 || err != nil {
			failure = true
			break
		}
	}

	if failure && !forcePkg {
		return fmt.Errorf("checking for kubekit required images failed . Verify the installed kubekit images are the correct version")
	}
	return nil
}

//CheckRpmPackage opens the contents of the RPM package and validates against
//the manifest
func CheckRpmPackage(pkgFilename string, forcePkg bool) error {

	if len(pkgFilename) == 0 || forcePkg {
		return nil
	}
	check := getPackages()

	p, err := rpm.OpenPackageFile(pkgFilename)
	if err != nil {
		return fmt.Errorf("error with rpm file. %q", err)
	}

	failure := false
	for _, fi := range p.Files() {
		if _, ok := check[fi.Name()]; ok {
			delete(check, fi.Name())
			continue
		} else {
			failure = true
			break
		}
	}

	if failure && !forcePkg {
		return fmt.Errorf("rpm contents didn't match what kubekit was expecting and --force-pkg was not set. ")
	}
	return nil
}

func getPackages() map[string]string {
	check := make(map[string]string)
	for k := range manifest.KubeManifest.Releases[manifest.Version].Dependencies.ControlPlane {
		check[manifest.KubeManifest.Releases[manifest.Version].Dependencies.ControlPlane[k].PrebakePath] = "NOT FOUND"
	}
	for k := range manifest.KubeManifest.Releases[manifest.Version].Dependencies.Core {
		check[manifest.KubeManifest.Releases[manifest.Version].Dependencies.Core[k].PrebakePath] = "NOT FOUND"
	}

	return check
}
