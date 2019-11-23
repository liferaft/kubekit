package v1

import (
	"os"
	"path/filepath"

	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"github.com/kubekit/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

// DeleteClusterConfig deletes an existing cluster configuration and all related files
func (s *KubeKitService) DeleteClusterConfig(ctx context.Context, in *apiv1.DeleteClusterConfigRequest) (*apiv1.DeleteClusterConfigResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	response := apiv1.DeleteClusterConfigResponse{
		Api:         apiVersion,
		ClusterName: in.ClusterName,
	}

	path := kluster.Path(in.ClusterName, s.clustersPath)
	if len(path) == 0 {
		response.Status = apiv1.DeleteClusterConfigStatus_NOT_FOUND
		return &response, nil
	}

	// don't delete the config if dry
	if !s.dry {
		if err := s.doDeleteClusterConfig(path); err != nil {
			return nil, err
		}
	}

	response.Status = apiv1.DeleteClusterConfigStatus_DELETED
	return &response, nil
}

func (s *KubeKitService) doDeleteClusterConfig(path string) error {
	baseDir := filepath.Dir(path)
	return os.RemoveAll(baseDir)
}
