package v1

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Version returns the KubeKit Server version using HTTP/REST or gRPC
func (c *Config) Version(ctx context.Context) (string, error) {
	c.Logger.Debugf("Getting version from server")

	return c.RunGRPCnRESTFunc("init", false,
		func() (string, error) {
			return c.versionGRPC(ctx)
		},
		func() (string, error) {
			return c.versionHTTP()
		})
}

func (c *Config) versionGRPC(ctx context.Context) (string, error) {
	reqVersion := apiv1.VersionRequest{
		Api: c.APIVersion,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.Logger.Debugf(`grpcurl request: grpcurl -insecure %s:%s kubekit.%s.Kubekit/Version`, c.Host, c.GrpcPort, c.APIVersion)

	resVersion, err := c.GrpcClient.Version(childCtx, &reqVersion)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request version. %s", err)
	}
	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	versionJSON, err := jsm.MarshalToString(resVersion)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received version response: %+v. %s", resVersion, err)
	}

	return string(versionJSON), nil
}

func (c *Config) versionHTTP() (string, error) {
	versionURL := fmt.Sprintf("%s/api/%s/version", c.HTTPBaseURL, c.APIVersion)

	c.Logger.Debugf(`curl request: curl -s -k -X GET "%s"`, versionURL)

	resp, err := c.HTTPClient.Get(versionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	versionJSON, err := ioutil.ReadAll(resp.Body)
	return string(versionJSON), err
}
