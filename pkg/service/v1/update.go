package v1

import (
	"fmt"
	"path/filepath"

	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"github.com/kubekit/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

// UpdateCluster updates an existing cluster using the parameters received.
func (s *KubeKitService) UpdateCluster(ctx context.Context, in *apiv1.UpdateClusterRequest) (*apiv1.UpdateClusterResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	if len(in.ClusterName) == 0 {
		return nil, fmt.Errorf("cluster name is required")
	}

	cluster, err := kluster.LoadCluster(in.ClusterName, s.clustersPath, s.ui)
	if err != nil {
		return nil, err
	}

	if s.dry {
		return &apiv1.UpdateClusterResponse{
			Api: apiVersion,
		}, nil
	}

	if len(in.Variables) != 0 {
		if err := cluster.Update(in.Variables); err != nil {
			return nil, err
		}
	}

	if len(in.Resources) != 0 {
		cluster.Resources = in.Resources
	}

	if err := cluster.Save(); err != nil {
		return nil, err
	}

	if len(in.Credentials) != 0 {
		platform := cluster.Platform()
		path := filepath.Join(filepath.Dir(cluster.Path()), ".credentials")

		credentials := kluster.NewCredentials(in.ClusterName, platform, path)

		// take the parameters from the file. ignore err, the file may not be there
		s.ui.Log.Debugf("reading credentials from file")
		if rerr := credentials.Read(); rerr != nil {
			s.ui.Log.Warnf("cannot read the credentials file. %s", rerr)
		}

		s.ui.Log.Debugf("assinging new credentials")
		if err := credentials.AssignFromMap(in.Credentials); err != nil {
			return nil, err
		}

		if err := credentials.Write(); err != nil {
			return nil, err
		}
	}

	return &apiv1.UpdateClusterResponse{
		Api: apiVersion,
	}, nil
}
