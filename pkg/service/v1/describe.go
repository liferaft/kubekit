package v1

import (
	"fmt"
	"strings"

	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/kluster"
	context "golang.org/x/net/context"
)

// Describe request to the server the description of one cluster with the given name
func (s *KubeKitService) Describe(ctx context.Context, in *apiv1.DescribeRequest) (*apiv1.DescribeResponse, error) {
	if err := s.checkAPIVersion(in.Api); err != nil {
		return nil, err
	}

	ci, err := kluster.GetClustersInfo(s.clustersPath, map[string]string{}, in.ClusterName)
	if err != nil {
		return nil, err
	}
	if len(ci) == 0 {
		return nil, fmt.Errorf("not found cluster %q", in.ClusterName)
	}

	ki := ci[0]

	platform := apiv1.PlatformName_value[strings.ToUpper(ki.Platform)]
	status := apiv1.Status_value[strings.ToUpper(ki.Status)]
	clusterInfo := &apiv1.Cluster{
		Name:     ki.Name,
		Platform: apiv1.PlatformName(platform),
		Nodes:    int32(ki.Nodes),
		Status:   apiv1.Status(status),
	}

	showParam := func(item string) bool {
		for _, i := range in.ShowParams {
			if strings.ToLower(i) == strings.ToLower(item) {
				return true
			}
		}
		return false
	}

	// Basic information
	response := &apiv1.DescribeResponse{
		Api:     apiVersion,
		Cluster: clusterInfo,
	}

	if len(in.ShowParams) == 0 {
		return response, nil
	}

	cluster, err := kluster.LoadCluster(in.ClusterName, s.clustersPath, s.ui)
	if err != nil {
		return nil, err
	}

	showAllParams := showParam("all")

	if showAllParams || showParam("config") {
		cfg, err := getClusterConfig(cluster)
		if err != nil {
			return nil, err
		}
		response.Config = cfg
	}

	if showAllParams || showParam("nodes") {
		response.Nodes = getClusterNodes(cluster)
	}

	if showAllParams || showParam("entrypoint") {
		response.Entrypoint = cluster.GetEntrypoint()
	}

	if showAllParams || showParam("kubeconfig") {
		response.Kubeconfig, _ = cluster.GetKubeconfig()
	}

	return response, nil
}

func getClusterConfig(cluster *kluster.Kluster) (*apiv1.ClusterConfig, error) {
	vars, err := cluster.ConfigVariables()
	if err != nil {
		return nil, err
	}

	return &apiv1.ClusterConfig{
		Variables: vars,
		Resources: cluster.Resources,
	}, nil
}

func getClusterNodes(cluster *kluster.Kluster) *apiv1.ClusterNodes {
	mapNodePools := map[string]*apiv1.NodePool{}

	platform := cluster.Platform()
	nodes := cluster.State[platform].Nodes
	for _, node := range nodes {
		var nodePool *apiv1.NodePool
		poolName := node.Pool

		var ok bool
		if nodePool, ok = mapNodePools[poolName]; !ok {
			nodePool = &apiv1.NodePool{
				PoolName: node.Pool,
				Nodes:    []*apiv1.Node{},
			}
		}
		var hostname string
		if hostname = node.PublicDNS; len(hostname) == 0 {
			if hostname = node.PrivateDNS; len(hostname) != 0 {
				hostname = strings.Split(hostname, ".")[0]
			}
		}

		node := &apiv1.Node{
			Name:       hostname,
			PublicIp:   node.PublicIP,
			PublicDns:  node.PublicDNS,
			PrivateIp:  node.PrivateIP,
			PrivateDns: node.PrivateDNS,
			RoleName:   node.RoleName,
			PoolName:   node.Pool,
		}
		nodePool.Nodes = append(nodePool.Nodes, node)

		mapNodePools[poolName] = nodePool
	}

	nodePools := []*apiv1.NodePool{}
	for _, nodePool := range mapNodePools {
		nodePools = append(nodePools, nodePool)
	}

	return &apiv1.ClusterNodes{
		NodePools: nodePools,
	}
}
