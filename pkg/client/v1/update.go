package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// UpdateCluster returns the KubeKit Server update using HTTP/REST or gRPC
func (c *Config) UpdateCluster(ctx context.Context, clusterName string, variables, credentials map[string]string) (string, error) {
	c.Logger.Debugf("Sending parameters to server to update the cluster %q", clusterName)

	return c.RunGRPCnRESTFunc("update", true,
		func() (string, error) {
			return c.updateGRPC(ctx, clusterName, variables, credentials)
		},
		func() (string, error) {
			return c.updateHTTP(clusterName, variables, credentials)
		})
}

func (c *Config) updateGRPC(ctx context.Context, clusterName string, variables, credentials map[string]string) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	reqUpdate := apiv1.UpdateClusterRequest{
		Api:         c.APIVersion,
		ClusterName: clusterName,
		Variables:   variables,
		Credentials: credentials,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Getting JSON input for the `grpcurl` instruction
	variablesJSON, err := updateVariablesInJSON(clusterName, variables, credentials)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/Update`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

	resUpdate, err := c.GrpcClient.UpdateCluster(childCtx, &reqUpdate)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request update. %s", err)
	}
	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	updateJSON, err := jsm.MarshalToString(resUpdate)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received update response: %+v. %s", resUpdate, err)
	}

	return string(updateJSON), nil
}

func (c *Config) updateHTTP(clusterName string, variables, credentials map[string]string) (string, error) {
	updateURL := fmt.Sprintf("%s/api/%s/cluster", c.HTTPBaseURL, c.APIVersion)

	variablesJSON, err := updateVariablesInJSON(clusterName, variables, credentials)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`curl request: curl -s -k -X PUT -d '%s' "%s"`, string(variablesJSON), updateURL)

	req, err := http.NewRequest(http.MethodPut, updateURL, bytes.NewBuffer(variablesJSON))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	updateJSON, err := ioutil.ReadAll(resp.Body)
	return string(updateJSON), err
}

func updateVariablesInJSON(clusterName string, variables, credentials map[string]string) ([]byte, error) {
	allVariables := map[string]interface{}{
		"cluster_name": clusterName,
		"variables":    variables,
		"credentials":  credentials,
	}

	return json.Marshal(allVariables)
}
