package v1

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/struct"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"github.com/kubekit/kubekit/pkg/aws_iam_authenticator/token"
	"github.com/kubekit/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

// Token mimics the token command from aws-iam-authenticator
func (s *KubeKitService) Token(ctx context.Context, in *apiv1.TokenRequest) (*apiv1.TokenResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	cluster, err := kluster.LoadCluster(in.ClusterName, s.clustersPath, s.ui)
	if err != nil {
		return nil, err
	}

	token, err := token.GenerateToken(cluster, in.Role)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, token.Status.ExpirationTimestamp.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	expirationTimestamp, err := ptypes.TimestampProto(t)
	if err != nil {
		return nil, err
	}
	spec := &structpb.Struct{}

	s.ui.Log.Debugf("token requested for cluster %q with role %q, returning Kind: %s, ApiVersion: %s, ExpirationTime: %s and the token", cluster.Name, in.Role, token.Kind, token.APIVersion, expirationTimestamp)

	return &apiv1.TokenResponse{
		Kind:       token.Kind,
		ApiVersion: token.APIVersion,
		Spec:       spec,
		Status: &apiv1.TokenStatus{
			ExpirationTimestamp: expirationTimestamp,
			Token:               token.Status.Token,
		},
	}, nil
}
