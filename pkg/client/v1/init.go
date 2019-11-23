package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Init returns the KubeKit Server init using HTTP/REST or gRPC
func (c *Config) Init(ctx context.Context, clusterName, platform string, variables, credentials map[string]string) (string, error) {
	c.Logger.Debugf("Sending parameters to server to initialize the cluster %q for platform %s", clusterName, platform)

	platform = strings.ToUpper(platform)
	var (
		platformName int32
		ok           bool
	)
	if platformName, ok = apiv1.PlatformName_value[platform]; !ok {
		return "", grpc.Errorf(codes.Internal, "Unknown platform %q", platform)
	}

	return c.RunGRPCnRESTFunc("init", true,
		func() (string, error) {
			return c.initGRPC(ctx, clusterName, platformName, variables, credentials)
		},
		func() (string, error) {
			return c.initHTTP(clusterName, platformName, variables, credentials)
		})
}

func (c *Config) initGRPC(ctx context.Context, clusterName string, platformName int32, variables, credentials map[string]string) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	reqInit := apiv1.InitRequest{
		Api:         c.APIVersion,
		ClusterName: clusterName,
		Platform:    apiv1.PlatformName(platformName),
		Variables:   variables,
		Credentials: credentials,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Getting JSON input for the `grpcurl` instruction
	variablesJSON, err := initVariablesInJSON(clusterName, apiv1.PlatformName_name[platformName], variables, credentials)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/Init`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

	resInit, err := c.GrpcClient.Init(childCtx, &reqInit)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request init. %s", err)
	}
	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	initJSON, err := jsm.MarshalToString(resInit)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received init response: %+v. %s", resInit, err)
	}

	return string(initJSON), nil
}

func (c *Config) initHTTP(clusterName string, platformName int32, variables, credentials map[string]string) (string, error) {
	initURL := fmt.Sprintf("%s/api/%s/cluster", c.HTTPBaseURL, c.APIVersion)

	variablesJSON, err := initVariablesInJSON(clusterName, apiv1.PlatformName_name[platformName], variables, credentials)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`curl request: curl -s -k -X POST -d '%s' "%s"`, string(variablesJSON), initURL)

	resp, err := c.HTTPClient.Post(initURL, "application/json", bytes.NewBuffer(variablesJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	initJSON, err := ioutil.ReadAll(resp.Body)
	return string(initJSON), err
}

func initVariablesInJSON(clusterName, platformName string, variables, credentials map[string]string) ([]byte, error) {
	allVariables := map[string]interface{}{
		"cluster_name": clusterName,
		"platform":     platformName,
		"variables":    variables,
		"credentials":  credentials,
	}

	return json.Marshal(allVariables)
}
