package v1

import (
	"os"
	"path/filepath"

	"github.com/nightlyone/lockfile"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"github.com/kubekit/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

// Delete deletes or terminate an existing cluster, removing all the cluster resources.
func (s *KubeKitService) Delete(ctx context.Context, in *apiv1.DeleteRequest) (*apiv1.DeleteResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	cluster, err := kluster.LoadCluster(in.ClusterName, s.clustersPath, s.ui)
	if err != nil {
		return nil, err
	}

	// don't delete if dry
	if !s.dry {
		go s.doDelete(ctx, cluster, in.DestroyAll)
	}

	platform := cluster.Platform()
	return &apiv1.DeleteResponse{
		Api:    apiVersion,
		Status: cluster.State[platform].Status,
	}, nil
}

func (s *KubeKitService) doDelete(ctx context.Context, cluster *kluster.Kluster, destroyAll bool) {
	var err error
	var status string

	var lock lockfile.Lockfile
	if lock, err = cluster.Lock("delete"); err != nil {
		return
	}
	defer lock.Unlock()

	platform := cluster.Platform()

	defer func() {
		if err != nil {
			s.ui.Log.Errorf("failed to destroy the cluster %s. %s", cluster.Name, err)
		}
		if destroyAll {
			return
		}

		cluster.State[platform].Status = status
		if errS := cluster.Save(); errS != nil {
			s.ui.Log.Errorf("failed to save the cluster configuration file for %s. %s", cluster.Name, errS)
		}
	}()

	if err = cluster.Terminate(); err != nil {
		status = kluster.FailedTerminationStatus.String()
		s.ui.Log.Errorf("failed to delete the cluster %s. %s", cluster.Name, err)
		return
	}

	// Delete certificates after destroy the cluster, they cannot be used again
	certsDir := filepath.Join(cluster.Dir(), kluster.CertificatesDirname)
	if err = os.RemoveAll(certsDir); err != nil {
		status = kluster.FailedTerminationStatus.String()
		s.ui.Log.Errorf("cluster %s destroyed but failed to delete the certificates. %s", cluster.Name, err)
		return
	}

	status = kluster.TerminatedStatus.String()
	s.ui.Log.Infof("cluster %s successfully destroyed and certificates deleted", cluster.Name)

	if !destroyAll {
		s.ui.Log.Infof("the cluster configuration still exists, the cluster %s can be re-created", cluster.Name)
		return
	}

	s.ui.Log.Infof("the cluster configuration for %q and other files are deleted", cluster.Name)

	// destroy the cluster configuration and other files
	if err := s.doDeleteClusterConfig(cluster.Path()); err != nil {
		s.ui.Log.Warnf("cluster %s successfully destroyed but failed to delete the cluster configuration files. %s", cluster.Name, err)
		return
	}
}
