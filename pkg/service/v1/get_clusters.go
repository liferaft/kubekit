package v1

import (
	"fmt"
	"strings"

	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

// parseStatus translate the kluster status to the api status
func parseStatus(status string) int32 {
	switch kluster.ParseStatus(status) {
	case kluster.FailedProvisioningStatus:
		return apiv1.Status_value["FAILED_PROVISIONING"]
	case kluster.FailedConfigurationStatus:
		return apiv1.Status_value["FAILED_CONFIGURATION"]
	case kluster.FailedCreationStatus:
		return apiv1.Status_value["FAILED_CREATION"]
	case kluster.FailedTerminationStatus:
		return apiv1.Status_value["FAILED_TERMINATION"]
	default:
		return apiv1.Status_value[strings.ToUpper(status)]
	}
}

// GetClusters request to the server the list of clusters
func (s *KubeKitService) GetClusters(ctx context.Context, in *apiv1.GetClustersRequest) (*apiv1.GetClustersResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	filter, err := SliceToMap(in.Filter)
	if err != nil {
		return nil, err
	}

	if invalidFilterFields := kluster.InvalidFilterParams(filter); len(invalidFilterFields) != 0 {
		var plural string
		if len(invalidFilterFields) > 1 {
			plural = "s"
		}
		return nil, fmt.Errorf("invalid filter%s: %s", plural, invalidFilterFields)
	}

	ci, err := kluster.GetClustersInfo(s.clustersPath, filter, in.Names...)
	if err != nil {
		return nil, err
	}

	clusters := []*apiv1.Cluster{}
	for _, ki := range ci {
		platform := apiv1.PlatformName_value[strings.ToUpper(ki.Platform)]
		status := parseStatus(ki.Status)

		cluster := &apiv1.Cluster{
			Name:     ki.Name,
			Platform: apiv1.PlatformName(platform),
			Nodes:    int32(ki.Nodes),
			Status:   apiv1.Status(status),
		}
		clusters = append(clusters, cluster)
	}

	return &apiv1.GetClustersResponse{
		Api:      apiVersion,
		Clusters: clusters,
	}, nil
}

// SliceToMap convert a slice of strings in the format `key=value` to a map of
// strings of strings to store the key/value pairs
func SliceToMap(slice []string) (map[string]string, error) {
	m := make(map[string]string, 0)
	for _, kv := range slice {
		pair := strings.Split(kv, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid key/value pair. %q", kv)
		}
		if len(pair[0]) == 0 {
			return nil, fmt.Errorf("invalid key, the value %q require a key", kv)
		}
		m[pair[0]] = pair[1]
	}

	return m, nil
}
