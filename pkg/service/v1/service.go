package v1

import (
	"github.com/kraken/ui"
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const apiVersion = "v1"

// KubeKitService implement the interface rpc.KubeKitServiceClient and encapsulate
// all the properties of a KubeKit service
type KubeKitService struct {
	clustersPath string
	ui           *ui.UI
	dry          bool
}

// NewKubeKitService creates a new KubeKit service
func NewKubeKitService(clustersPath string, parentUI *ui.UI, dry bool) *KubeKitService {
	if dry {
		parentUI.Log.Warn("Starting the KubeKit service in dry mode. API calls will be inert.")
	}

	return &KubeKitService{
		clustersPath: clustersPath,
		ui:           parentUI,
		dry:          dry,
	}
}

// Register registers this service to the given gRPC server
func (s *KubeKitService) Register(server *grpc.Server) {
	apiv1.RegisterKubekitServer(server, s)
}

func (s *KubeKitService) checkAPIVersion(version string) error {
	if len(version) != 0 && version != apiVersion {
		return status.Errorf(codes.Unimplemented, "API version %q is not supported. This services implements API version %q", version, apiVersion)
	}
	return nil
}
