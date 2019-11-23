package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
)

// Describe returns the KubeKit Server describe using HTTP/REST or gRPC
func (c *Config) Describe(ctx context.Context, showParams []string, clustersName ...string) (string, error) {
	c.Logger.Debugf("Sending parameters to server to describe the clusters %q", strings.Join(clustersName, ", "))

	return c.RunGRPCnRESTFunc("describe", true,
		func() (string, error) {
			return c.describeGRPC(ctx, showParams, clustersName...)
		},
		func() (string, error) {
			return c.describeHTTP(showParams, clustersName...)
		})
}

func describe(descClusterFunc func(string) (string, error), clustersName ...string) (string, error) {
	descriptions := []string{}
	for _, clusterName := range clustersName {
		resDescribe, err := descClusterFunc(clusterName)
		if err != nil {
			return "", err
		}
		descriptions = append(descriptions, resDescribe)
	}

	switch len(descriptions) {
	case 0:
		return "", grpc.Errorf(codes.Internal, "no description found for the clusters %q", strings.Join(clustersName, ", "))
	case 1:
		return descriptions[0], nil
	default:
		return fmt.Sprintf("[%s]", strings.Join(descriptions, ",")), nil
	}
}

func (c *Config) describeGRPC(ctx context.Context, showParams []string, clustersName ...string) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	return describe(func(clusterName string) (string, error) {
		reqDescribe := apiv1.DescribeRequest{
			Api:         c.APIVersion,
			ClusterName: clusterName,
			ShowParams:  showParams,
		}

		childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Getting JSON input for the `grpcurl` instruction
		variablesJSON, err := describeVariablesInJSON(clusterName, showParams)
		if err != nil {
			return "", err
		}
		c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/Describe`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

		resDescribe, err := c.GrpcClient.Describe(childCtx, &reqDescribe)
		if err != nil {
			return "", grpc.Errorf(codes.Internal, "fail to request description for cluster %q. %s", clusterName, err)
		}
		jsm := jsonpb.Marshaler{
			EmitDefaults: true,
		}
		describeJSON, err := jsm.MarshalToString(resDescribe)
		if err != nil {
			return "", grpc.Errorf(codes.Internal, "failed to marshall the received describe response: %+v. %s", resDescribe, err)
		}

		return string(describeJSON), nil
	}, clustersName...)
}

func (c *Config) describeHTTP(showParams []string, clustersName ...string) (string, error) {
	var paramsQuery string
	if len(showParams) > 0 {
		paramsQuery = "?"
		for _, param := range showParams {
			paramsQuery = paramsQuery + "&show_params=" + param
		}
	}
	generalDescribeURL := fmt.Sprintf("%s/api/%s/cluster/%%s%s", c.HTTPBaseURL, c.APIVersion, paramsQuery)

	return describe(func(clusterName string) (string, error) {
		describeURL := fmt.Sprintf(generalDescribeURL, clusterName)
		c.Logger.Debugf(`curl request: curl -s -k -X GET "%s"`, describeURL)

		resp, err := c.HTTPClient.Get(describeURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		describeJSON, err := ioutil.ReadAll(resp.Body)
		return string(describeJSON), err
	}, clustersName...)
}

func describeVariablesInJSON(clusterName string, showParams []string) ([]byte, error) {
	allVariables := map[string]interface{}{
		"cluster_name": clusterName,
		"show_params":  showParams,
	}

	return json.Marshal(allVariables)
}
