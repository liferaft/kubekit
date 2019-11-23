package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// DeleteClusterConfig returns the KubeKit Server delete cluster config using HTTP/REST or gRPC
func (c *Config) DeleteClusterConfig(ctx context.Context, clusterName string) (string, error) {
	c.Logger.Debugf("Sending parameters to server to delete the configuration of cluster %q", clusterName)

	return c.RunGRPCnRESTFunc("delete cluster config", true,
		func() (string, error) {
			return c.deleteClusterConfigGRPC(ctx, clusterName)
		},
		func() (string, error) {
			return c.deleteClusterConfigHTTP(clusterName)
		})
}

func (c *Config) deleteClusterConfigGRPC(ctx context.Context, clusterName string) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	reqDeleteClusterConfig := apiv1.DeleteClusterConfigRequest{
		Api:         c.APIVersion,
		ClusterName: clusterName,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Getting JSON input for the `grpcurl` instruction
	variablesJSON, err := deleteClusterConfigVariablesInJSON(clusterName)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/DeleteClusterConfig`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

	resDeleteClusterConfig, err := c.GrpcClient.DeleteClusterConfig(childCtx, &reqDeleteClusterConfig)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request delete cluster config. %s", err)
	}
	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	deleteClusterConfigJSON, err := jsm.MarshalToString(resDeleteClusterConfig)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received delete cluster config response: %+v. %s", resDeleteClusterConfig, err)
	}

	return string(deleteClusterConfigJSON), nil
}

func (c *Config) deleteClusterConfigHTTP(clusterName string) (string, error) {
	deleteClusterConfigURL := fmt.Sprintf("%s/api/%s/cluster/%s/config", c.HTTPBaseURL, c.APIVersion, clusterName)

	variablesJSON, err := deleteClusterConfigVariablesInJSON(clusterName)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`curl request: curl -s -k -X DELETE -d '%s' "%s"`, string(variablesJSON), deleteClusterConfigURL)

	resp, err := c.HTTPClient.Post(deleteClusterConfigURL, "application/json", bytes.NewBuffer(variablesJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	deleteClusterConfigJSON, err := ioutil.ReadAll(resp.Body)
	return string(deleteClusterConfigJSON), err
}

func deleteClusterConfigVariablesInJSON(clusterName string) ([]byte, error) {
	allVariables := map[string]interface{}{
		"cluster_name": clusterName,
	}

	return json.Marshal(allVariables)
}
