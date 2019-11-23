package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Delete returns the KubeKit Server delete using HTTP/REST or gRPC
func (c *Config) Delete(ctx context.Context, clusterName string, destroyAll bool) (string, error) {
	c.Logger.Debugf("Sending parameters to server to delete the cluster %q", clusterName)

	return c.RunGRPCnRESTFunc("delete", true,
		func() (string, error) {
			return c.deleteGRPC(ctx, clusterName, destroyAll)
		},
		func() (string, error) {
			return c.deleteHTTP(clusterName, destroyAll)
		})
}

func (c *Config) deleteGRPC(ctx context.Context, clusterName string, destroyAll bool) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	reqDelete := apiv1.DeleteRequest{
		Api:         c.APIVersion,
		ClusterName: clusterName,
		DestroyAll:  destroyAll,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Getting JSON input for the `grpcurl` instruction
	variablesJSON, err := deleteVariablesInJSON(clusterName, destroyAll)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/Delete`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

	resDelete, err := c.GrpcClient.Delete(childCtx, &reqDelete)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request delete. %s", err)
	}
	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	deleteJSON, err := jsm.MarshalToString(resDelete)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received delete response: %+v. %s", resDelete, err)
	}

	return string(deleteJSON), nil
}

func (c *Config) deleteHTTP(clusterName string, destroyAll bool) (string, error) {
	deleteURL := fmt.Sprintf("%s/api/%s/cluster/%s", c.HTTPBaseURL, c.APIVersion, clusterName)

	variablesJSON, err := deleteVariablesInJSON(clusterName, destroyAll)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`curl request: curl -s -k -X DELETE -d '%s' "%s"`, string(variablesJSON), deleteURL)

	resp, err := c.HTTPClient.Post(deleteURL, "application/json", bytes.NewBuffer(variablesJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	deleteJSON, err := ioutil.ReadAll(resp.Body)
	return string(deleteJSON), err
}

func deleteVariablesInJSON(clusterName string, destroyAll bool) ([]byte, error) {
	allVariables := map[string]interface{}{
		"cluster_name": clusterName,
		"destroy_all":  destroyAll,
	}

	return json.Marshal(allVariables)
}
