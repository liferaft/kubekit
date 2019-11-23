package v1

import (
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/manifest"
	"github.com/liferaft/kubekit/version"
	context "golang.org/x/net/context"
)

// Version return the current version of KubeKit
func (s *KubeKitService) Version(ctx context.Context, in *apiv1.VersionRequest) (*apiv1.VersionResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	vKubekit := version.Version
	vKubernetes := "unknown"
	vDocker := "unknown"
	vEtcd := "unknown"
	if release, ok := manifest.KubeManifest.Releases[vKubekit]; ok {
		vKubernetes = release.KubernetesVersion
		vDocker = release.DockerVersion
		vEtcd = release.EtcdVersion
	}

	s.ui.Log.Debugf("version requested, returning API: %s, KubeKit: %s, Kubernetes: %s, Docker: %s, etcd: %s", apiVersion, vKubekit, vKubernetes, vDocker, vEtcd)

	return &apiv1.VersionResponse{
		Api:        apiVersion,
		Kubekit:    vKubekit,
		Kubernetes: vKubernetes,
		Docker:     vDocker,
		Etcd:       vEtcd,
	}, nil
}
