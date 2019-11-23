package v1

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nightlyone/lockfile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/crypto/tls"
	"github.com/liferaft/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

const pkgName = "kubekit.rpm"

// Apply creates a configuration file for the given kind (`cluster` or `template`)
func (s *KubeKitService) Apply(ctx context.Context, in *apiv1.ApplyRequest) (*apiv1.ApplyResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	cluster, err := kluster.LoadCluster(in.ClusterName, s.clustersPath, s.ui)
	if err != nil {
		return nil, err
	}

	// only apply if not dry
	if !s.dry {
		go s.doApply(ctx, cluster, in)
	}

	return &apiv1.ApplyResponse{
		Api:    apiVersion,
		Status: kluster.AbsentStatus.String(),
	}, nil
}

func (s *KubeKitService) doApply(ctx context.Context, cluster *kluster.Kluster, in *apiv1.ApplyRequest) {
	var err error
	var status string

	var lock lockfile.Lockfile
	if lock, err = cluster.Lock("apply"); err != nil {
		return
	}
	defer lock.Unlock()

	platform := cluster.Platform()

	defer func() {
		if err != nil {
			s.ui.Log.Errorf("failed to apply the cluster %s. %s", cluster.Name, err)
		}
		cluster.State[platform].Status = status
		if errS := cluster.Save(); errS != nil {
			s.ui.Log.Errorf("failed to save the cluster configuration file for %s. %s", cluster.Name, errS)
		}
	}()

	// 1. Generate the SSH keys:
	// required for the terraform templates and provisioner
	s.ui.Log.Infof("generating cluster %q ssh keys", in.ClusterName)
	if err = cluster.HandleKeys(); err != nil {
		status = kluster.FailedProvisioningStatus.String()
		return
	}

	// 2. Provisioning, Upload & Install the KubeKit package:
	if in.Action == apiv1.ApplyAction_ALL || in.Action == apiv1.ApplyAction_PROVISION {
		status = kluster.FailedProvisioningStatus.String()

		s.ui.Log.Infof("provisioning cluster %q on %s", in.ClusterName, platform)
		if err = doProvisioning(cluster, platform); err != nil {
			return
		}
		s.ui.Log.Infof("uploading and installing the package to cluster %q on %s", in.ClusterName, platform)
		if err = installPackage(cluster, s.clustersPath, platform, in.ForcePackage); err != nil {
			return
		}

		s.ui.Log.Infof("cluster %q successfully provisioned on %s", in.ClusterName, platform)
		status = kluster.ProvisionedStatus.String()
	}

	// 3. Generate certificates, Create Kubeconfig file & Configure Kubernetes
	if in.Action == apiv1.ApplyAction_ALL || in.Action == apiv1.ApplyAction_CONFIGURE {
		status = kluster.FailedConfigurationStatus.String()

		var caCertsFiles tls.KeyPairs
		caCertsFiles, err = getCACerts(cluster, in.CaCerts)
		if err != nil {
			return
		}
		s.ui.Log.Infof("generating CA certificates for cluster %q on %s", in.ClusterName, platform)
		if err = cluster.GenerateCerts(caCertsFiles, true); err != nil {
			return
		}
		if err = cluster.LoadState(); err != nil {
			return
		}
		s.ui.Log.Infof("creating the Kubeconfig file for cluster %q on %s", in.ClusterName, platform)
		if err = cluster.CreateKubeConfigFile(); err != nil {
			return
		}
		s.ui.Log.Infof("configuring Kubernetes cluster %q on %s", in.ClusterName, platform)
		if err = doConfiguration(cluster); err != nil {
			return
		}

		s.ui.Log.Infof("cluster %q successfully configured on %s", in.ClusterName, platform)
		status = kluster.RunningStatus.String()
	}
}

// doProvisioning do the provisioning of the platform
func doProvisioning(cluster *kluster.Kluster, platform string) error {
	errP := cluster.Create()
	errS := cluster.Save()
	if errP != nil && errS != nil {
		return grpc.Errorf(codes.Internal, "failed to provision the cluster %s on %s. %s\nand failed to save the cluster configuration file %s. %s", cluster.Name, platform, errP, cluster.Path(), errS)
	}
	if errP != nil {
		return grpc.Errorf(codes.Internal, "failed to provision the cluster %s on %s. %s", cluster.Name, platform, errP)
	}
	if errS != nil {
		return grpc.Errorf(codes.Internal, "the cluster %s was successfully provisioned on %s but failed to save the cluster configuration file %s. %s", cluster.Name, platform, cluster.Path(), errS)
	}
	return nil
}

// installPackage copy the default package to every node then install the
// package on every node. This is to be executed only after a successfull provisioning
func installPackage(cluster *kluster.Kluster, clustersPath, platform string, forcePackage bool) error {
	// AWS requires a forced install
	if platform == "aws" {
		forcePackage = true
	}
	// package install is not for AKS or EKS, unless force is used
	if (platform == "aks" || platform == "eks") && !forcePackage {
		return nil
	}

	pkgFilename := filepath.Join(filepath.Dir(clustersPath), pkgName)

	if _, err := os.Stat(pkgFilename); os.IsNotExist(err) {
		return grpc.Errorf(codes.Internal, "KubeKit Package not found, looks like KubeKit was incorrectly installed. Contact the KubeKit admin to download the KubeKit package and save it to the KubeKit home directory %q", pkgFilename)
	}

	if err := cluster.CopyPackage(pkgFilename, "/tmp/", true); err != nil {
		return err
	}

	filename := filepath.Base(pkgFilename)
	pkgFilepath := filepath.Join("/tmp", filename)

	if result, _, err := cluster.InstallPackage(pkgFilepath, forcePackage); err != nil || result.Failures != 0 {
		return grpc.Errorf(codes.Internal, "failed to install the package in %d/%d nodes. %s", result.Success, result.Success+result.Failures, err)
	}

	return nil
}

// getCACerts creates all the CA certificates key pairs and save the files from
// the received certificates
func getCACerts(cluster *kluster.Kluster, caCerts map[string]string) (caCertFiles tls.KeyPairs, err error) {
	caCertFiles = make(tls.KeyPairs, len(kluster.CACertNames))

	var baseCertsDir string
	if len(caCerts) != 0 {
		platform := cluster.Platform()
		if baseCertsDir, err = cluster.MakeCertDir(platform); err != nil {
			return tls.KeyPairs{}, err
		}
	}

	writeFile := func(name string, pem []byte) (filename string, err error) {
		filename = filepath.Join(baseCertsDir, name)
		err = ioutil.WriteFile(filename, pem, 0600)
		return filename, err
	}

	for name, caCertInfo := range kluster.CACertNames {
		var (
			keyFile, certFile string
			keyPEM, certPEM   []byte
		)

		if content, ok := caCerts[name+"_key"]; ok {
			keyPEM = []byte(content)
			if keyFile, err = writeFile(name+".key", keyPEM); err != nil {
				return tls.KeyPairs{}, err
			}
		}
		if content, ok := caCerts[name+"_crt"]; ok {
			certPEM = []byte(content)
			if certFile, err = writeFile(name+".crt", certPEM); err != nil {
				return tls.KeyPairs{}, err
			}
		}

		caCertFiles[name] = &tls.KeyPair{
			Name:           name,
			KeyFile:        keyFile,
			PrivateKeyPEM:  keyPEM,
			CN:             caCertInfo.CN,
			CertFile:       certFile,
			CertificatePEM: keyPEM,
			IsCA:           true,
		}
	}

	return caCertFiles, nil
}

func doConfiguration(cluster *kluster.Kluster) error {
	errC := cluster.Configure()
	errS := cluster.Save()
	if errC != nil && errS != nil {
		return grpc.Errorf(codes.Internal, "failed to configure Kubernetes for the cluster %s. %s\nand failed to save the cluster configuration file %s.%s", cluster.Name, errC, cluster.Path(), errS)
	}
	if errC != nil {
		return grpc.Errorf(codes.Internal, "failed to configure Kubernetes for the cluster %s. %s", cluster.Name, errC)
	}
	if errS != nil {
		return grpc.Errorf(codes.Internal, "Kubernetes was successfully configured for cluster %s but failed to save the cluster configuration file %s. %s", cluster.Name, cluster.Path(), errS)
	}
	return nil
}
