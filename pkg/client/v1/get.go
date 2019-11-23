package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
)

// GetClusters returns the KubeKit Server get using HTTP/REST or gRPC
func (c *Config) GetClusters(ctx context.Context, quiet bool, filterMap map[string]string, clustersName ...string) (string, error) {
	c.Logger.Debugf("Sending parameters to server to get the clusters %q", strings.Join(clustersName, ", "))

	filter := MapToSlice(filterMap)

	return c.RunGRPCnRESTFunc("get", true,
		func() (string, error) {
			return c.getClustersGRPC(ctx, quiet, filter, clustersName...)
		},
		func() (string, error) {
			return c.getClustersHTTP(quiet, filter, clustersName...)
		})
}

// MapToSlice returns a slice of strings in the format `key=value` from a map of
// strings of strings
func MapToSlice(m map[string]string) []string {
	s := []string{}
	for k, v := range m {
		if len(k) == 0 {
			continue
		}
		s = append(s, k+"="+v)
	}
	return s
}

func getNames(resGetClusters *apiv1.GetClustersResponse) (string, error) {
	names := []string{}
	for _, cluster := range resGetClusters.Clusters {
		names = append(names, cluster.Name)
	}
	getJSON, err := json.Marshal(names)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received get response: %+v. %s", names, err)
	}
	return string(getJSON), nil
}

func (c *Config) getClustersGRPC(ctx context.Context, quiet bool, filter []string, clustersName ...string) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	reqGetClusters := apiv1.GetClustersRequest{
		Api:    c.APIVersion,
		Filter: filter,
	}
	if len(clustersName) != 0 {
		reqGetClusters.Names = clustersName
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Getting JSON input for the `grpcurl` instruction
	variablesJSON, err := getVariablesInJSON(filter, clustersName...)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/GetClusters`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

	resGetClusters, err := c.GrpcClient.GetClusters(childCtx, &reqGetClusters)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "fail to request the clusters with the list %v. %s", clustersName, err)
	}

	if quiet {
		return getNames(resGetClusters)
	}

	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	getJSON, err := jsm.MarshalToString(resGetClusters)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received get response: %+v. %s", resGetClusters, err)
	}

	return string(getJSON), nil
}

func encodeSlice(s []string) []string {
	encoded := []string{}
	for _, str := range s {
		encoded = append(encoded, url.QueryEscape(str))
	}
	return encoded
}

func (c *Config) getClustersHTTP(quiet bool, filter []string, clustersName ...string) (string, error) {
	var query string
	if len(clustersName) > 0 {
		names := strings.Join(encodeSlice(clustersName), "&names=")
		query = "?names=" + names
	}
	if len(filter) > 0 {
		if len(query) == 0 {
			query = "?"
		} else {
			query = query + "&"
		}
		filters := strings.Join(encodeSlice(filter), "&filter=")
		query = query + "filter=" + filters
	}
	getClustersURL := fmt.Sprintf("%s/api/%s/cluster%s", c.HTTPBaseURL, c.APIVersion, query)
	c.Logger.Debugf(`curl request: curl -s -k -X GET "%s"`, getClustersURL)

	resp, err := c.HTTPClient.Get(getClustersURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !quiet {
		getJSON, err := ioutil.ReadAll(resp.Body)
		return string(getJSON), err
	}

	resGetClusters := &apiv1.GetClustersResponse{}
	jsum := jsonpb.Unmarshaler{}
	err = jsum.Unmarshal(resp.Body, resGetClusters)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to unmarshall the received get response. %s", err)
	}

	return getNames(resGetClusters)
}

func getVariablesInJSON(filter []string, clustersName ...string) ([]byte, error) {
	allVariables := map[string]interface{}{}
	if len(filter) != 0 {
		allVariables["filter"] = filter
	}
	if len(clustersName) != 0 {
		allVariables["names"] = clustersName
	}

	return json.Marshal(allVariables)
}
