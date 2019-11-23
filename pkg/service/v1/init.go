package v1

import (
	"path/filepath"
	"strings"

	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const defaultClusterConfigFormat = "yaml"

// Init creates a configuration file for the given kind (`cluster` or `template`)
func (s *KubeKitService) Init(ctx context.Context, in *apiv1.InitRequest) (*apiv1.InitResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	if _, err := kluster.ValidClusterName(in.ClusterName); err != nil {
		return nil, err
	}
	if in.Platform == apiv1.PlatformName_UNKNOWN {
		return nil, grpc.Errorf(codes.Internal, "the platform name is required")
	}

	if s.dry {
		return &apiv1.InitResponse{
			Api:  apiVersion,
			Kind: in.Kind,
			Name: in.ClusterName,
		}, nil
	}

	var cluster *kluster.Kluster
	var initResponse *apiv1.InitResponse

	platformName := strings.ToLower(in.Platform.String())

	createClusterConfig := func() error {
		s.ui.Log.Debugf("initializing cluster %q configuration", in.ClusterName)

		var err error
		cluster, err = kluster.CreateCluster(in.ClusterName, platformName, s.clustersPath, defaultClusterConfigFormat, in.Variables, s.ui)
		if err != nil {
			return grpc.Errorf(codes.Internal, "failed to initialize the cluster %s. %s", in.ClusterName, err)
		}

		initResponse = &apiv1.InitResponse{
			Api:  apiVersion,
			Kind: cluster.Kind,
			Name: cluster.Name,
		}

		return nil
	}

	// If this is a cluster for the following platforms, do not process the
	// credentials. They do not have them
	switch platformName {
	case "vra", "raw", "stacki":
		return initResponse, createClusterConfig()
	}

	s.ui.Log.Debugf("initializing cluster %q credentials", in.ClusterName)

	credentials := kluster.NewCredentials(in.ClusterName, platformName, "")
	if err := credentials.AssignFromMap(in.Credentials); err != nil {
		return nil, err
	}

	if !credentials.Complete() {
		return nil, grpc.Errorf(codes.Internal, "credentials for the platform %q are incomplete or missing in the given variables", platformName)
	}

	if err := createClusterConfig(); err != nil {
		return nil, err
	}

	credentialsPath := filepath.Join(filepath.Dir(cluster.Path()), ".credentials")
	credentials.SetPath(credentialsPath)

	if err := credentials.Write(); err != nil {
		return nil, grpc.Errorf(codes.Internal, "the cluster configuration was created but failed to write the credentials files. %s", err)
	}

	return initResponse, nil
}
