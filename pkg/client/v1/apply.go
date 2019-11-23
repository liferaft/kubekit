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
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"github.com/kubekit/kubekit/pkg/crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Apply returns the KubeKit Server apply using HTTP/REST or gRPC
func (c *Config) Apply(ctx context.Context, clusterName string, action string, pkgURL string, forcePkg bool, userCACerts tls.KeyPairs) (string, error) {
	c.Logger.Debugf("Sending parameters to server to apply changes to the cluster %q", clusterName)

	var (
		applyAction int32
		ok          bool
	)
	if applyAction, ok = apiv1.ApplyAction_value[strings.ToUpper(action)]; !ok {
		applyAction = int32(apiv1.ApplyAction_ALL)
	}

	caCerts := map[string]string{}
	for name, kp := range userCACerts {
		if len(kp.PrivateKeyPEM)+len(kp.CertificatePEM) != 0 {
			caCerts[name+"_key"] = string(kp.PrivateKeyPEM)
			caCerts[name+"_crt"] = string(kp.CertificatePEM)
		}
	}

	return c.RunGRPCnRESTFunc("apply", true,
		func() (string, error) {
			return c.applyGRPC(ctx, clusterName, applyAction, pkgURL, forcePkg, caCerts)
		},
		func() (string, error) {
			return c.applyHTTP(clusterName, applyAction, pkgURL, forcePkg, caCerts)
		})
}

func (c *Config) applyGRPC(ctx context.Context, clusterName string, action int32, pkgURL string, forcePkg bool, caCerts map[string]string) (string, error) {
	if c.GrpcClient == nil {
		return "", nil
	}

	reqApply := apiv1.ApplyRequest{
		Api:          c.APIVersion,
		ClusterName:  clusterName,
		Action:       apiv1.ApplyAction(action),
		PackageUrl:   "",
		ForcePackage: false,
		CaCerts:      caCerts,
	}

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Getting JSON input for the `grpcurl` instruction
	variablesJSON, err := applyVariablesInJSON(clusterName, action, pkgURL, forcePkg, caCerts)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '%s' %s:%s kubekit.%s.Kubekit/Apply`, string(variablesJSON), c.Host, c.GrpcPort, c.APIVersion)

	resApply, err := c.GrpcClient.Apply(childCtx, &reqApply)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to request apply. %s", err)
	}
	jsm := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	applyJSON, err := jsm.MarshalToString(resApply)
	if err != nil {
		return "", grpc.Errorf(codes.Internal, "failed to marshall the received apply response: %+v. %s", resApply, err)
	}

	return string(applyJSON), nil
}

func (c *Config) applyHTTP(clusterName string, action int32, pkgURL string, forcePkg bool, caCerts map[string]string) (string, error) {
	applyURL := fmt.Sprintf("%s/api/%s/cluster/%s", c.HTTPBaseURL, c.APIVersion, clusterName)

	variablesJSON, err := applyVariablesInJSON(clusterName, action, pkgURL, forcePkg, caCerts)
	if err != nil {
		return "", err
	}
	c.Logger.Debugf(`curl request: curl -s -k -X POST -d '%s' "%s"`, string(variablesJSON), applyURL)

	resp, err := c.HTTPClient.Post(applyURL, "application/json", bytes.NewBuffer(variablesJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	applyJSON, err := ioutil.ReadAll(resp.Body)
	return string(applyJSON), err
}

func applyVariablesInJSON(clusterName string, action int32, pkgURL string, forcePkg bool, caCerts map[string]string) ([]byte, error) {
	allVariables := map[string]interface{}{
		"cluster_name":  clusterName,
		"action":        apiv1.ApplyAction(action),
		"package_url":   pkgURL,
		"force_package": forcePkg,
		"ca_certs":      caCerts,
	}

	return json.Marshal(allVariables)
}
