package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1alpha1 "k8s.io/client-go/pkg/apis/clientauthentication/v1alpha1"

	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/aws_iam_authenticator/token"
)

// Token returns the token  using HTTP/REST or gRPC
func (c *Config) Token(ctx context.Context, clusterName, roleARN string) (string, error) {
	c.Logger.Debugf("Getting token for cluster %s with role %s", clusterName, roleARN)

	return c.RunGRPCnRESTFunc("init", false,
		func() (string, error) {
			return c.tokenGRPC(ctx, clusterName, roleARN)
		},
		func() (string, error) {
			return c.tokenHTTP(clusterName, roleARN)
		})
}

func (c *Config) tokenGRPC(ctx context.Context, clusterName, roleARN string) (string, error) {
	reqToken := apiv1.TokenRequest{
		Api:         c.APIVersion,
		ClusterName: clusterName,
		Role:        roleARN,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resToken, err := c.GrpcClient.Token(childCtx, &reqToken)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request token. %s", err)
	}

	// DEBUG:
	c.Logger.Debugf("Token received: %+v", resToken)

	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '{"api": "%s", "cluster_name": "%s", "role": "%s"}' %s:%s kubekit.%s.Kubekit/Token`, c.APIVersion, clusterName, roleARN, c.Host, c.GrpcPort, c.APIVersion)

	tkn, err := createToken(resToken)
	if err != nil {
		return "", err
	}

	tokenJSON, err := json.Marshal(tkn)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received token response: %+v. %s", resToken, err)
	}

	return string(tokenJSON), nil
}

func (c *Config) tokenHTTP(clusterName, roleARN string) (string, error) {
	var roleQuery string
	if roleARN != "" {
		roleQuery = "&role=" + roleARN
	}
	tokenURL := fmt.Sprintf("%s/api/%s/cluster/%s/token?api=%s%s", c.HTTPBaseURL, c.APIVersion, clusterName, c.APIVersion, roleQuery)

	c.Logger.Debugf(`curl request: curl -s -k -X GET "%s"`, tokenURL)

	resp, err := c.HTTPClient.Get(tokenURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tokenJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(tokenJSON), nil
}

func createToken(resToken *apiv1.TokenResponse) (token.Token, error) {
	tkn := token.Token{}

	t, err := ptypes.Timestamp(resToken.Status.ExpirationTimestamp)
	if err != nil {
		return tkn, err
	}
	expirationTimestamp := metav1.NewTime(t)

	tkn.TypeMeta = metav1.TypeMeta{
		APIVersion: resToken.ApiVersion,
		Kind:       resToken.Kind,
	}

	tkn.Status = &clientauthv1alpha1.ExecCredentialStatus{
		ExpirationTimestamp: &expirationTimestamp,
		Token:               resToken.Status.Token,
	}

	return tkn, nil
}
