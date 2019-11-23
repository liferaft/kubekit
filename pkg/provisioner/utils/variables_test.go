package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

type V struct {
	ClusterName string   `json:"cluster_name,omitempty" mapstructure:"cluster_name"`
	NumMasters  int      `json:"num_masters,omitempty" mapstructure:"num_masters"`
	NumWorkers  int      `json:"num_workers,omitempty" mapstructure:"num_workers"`
	LinkedClone bool     `json:"linked_clone" mapstructure:"linked_clone"`
	DNSServers  []string `json:"dns_servers" mapstructure:"dns_servers"`
}

var v = V{
	ClusterName: "test",
	NumMasters:  1,
	NumWorkers:  1,
	LinkedClone: true,
	DNSServers:  []string{"153.64.180.100", "153.64.251.200"},
}

func TestMap(t *testing.T) {
	r := map[string]interface{}{
		"cluster_name": "test",
		"num_masters":  float64(1),
		"num_workers":  float64(1),
		"linked_clone": true,
		"dns_servers": []interface{}{
			"153.64.180.100",
			"153.64.251.200",
		},
	}

	tests := []struct {
		input  V
		result map[string]interface{}
		err    error
	}{
		{input: v, result: r, err: nil},
	}
	for _, test := range tests {
		result := utils.Map(test.input)
		assert.Equal(t, test.result, result)
	}
}

func TestHCL(t *testing.T) {
	r := `"cluster_name" = "test"

"num_masters" = 1

"num_workers" = 1

"linked_clone" = true

"dns_servers" = ["153.64.180.100", "153.64.251.200"]`

	tests := []struct {
		input  V
		result string
		err    error
	}{
		{input: v, result: r, err: nil},
	}
	for _, test := range tests {
		result, err := utils.HCL(test.input)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, string(result))
	}
}
