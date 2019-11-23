package kluster

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kubekit/kubekit/pkg/provisioner"

	"github.com/kubekit/kubekit/pkg/configurator"
)

func TestKluster_ConfigVariables(t *testing.T) {
	tests := []struct {
		name         string
		clusterName  string
		platform     string
		platformVars map[string]string
		configJSON   []byte
		want         map[string]string
		wantErr      bool
		version      string
	}{
		{name: "simple (empty cluster)",
			clusterName: "simple",
			platform:    "eks",
			want:        map[string]string{},
			version:     "1.0",
		},
		{name: "eks",
			clusterName: "eks01",
			platform:    "eks",
			version:     "1.0",
			platformVars: map[string]string{
				"aws_region":                             "us-west-2",
				"aws_vpc_id":                             "vpc-8d56b9e9",
				"ingress_subnets":                        "[subnet-5bddc82c,subnet-478a4123]",
				"cluster_security_groups":                "sg-502d9a37",
				"default_node_pool__worker_pool_subnets": "subnet-5bddc82c",
				"default_node_pool__security_groups":     "sg-502d9a37",
				// default_node_pool__name should normally be empty,
				// default_node_pool__placementgroup_strategy should normally be empty,
				// setting here for test issue
				"default_node_pool__name":                    "default-pool",
				"default_node_pool__placementgroup_strategy": "cluster",
			},
			want: map[string]string{
				"aws_region":                                                  "us-west-2",
				"aws_vpc_id":                                                  "vpc-8d56b9e9",
				"cluster_logs_types":                                          "[api, audit, authenticator, controllerManager, scheduler]",
				"cluster_security_groups":                                     "[sg-502d9a37]",
				"default_node_pool__aws_ami":                                  "",
				"default_node_pool__aws_instance_type":                        "",
				"default_node_pool__count":                                    "0",
				"default_node_pool__kubelet_node_labels":                      "[node-role.kubernetes.io/compute=\"\", node.kubernetes.io/compute=\"\"]",
				"default_node_pool__kubelet_node_taints":                      "[]",
				"default_node_pool__placementgroup_strategy":                  "cluster",
				"default_node_pool__root_volume_size":                         "100",
				"default_node_pool__security_groups":                          "[sg-502d9a37]",
				"default_node_pool__worker_pool_subnets":                      "[subnet-5bddc82c]",
				"endpoint_private_access":                                     "false",
				"endpoint_public_access":                                      "true",
				"ingress_subnets":                                             "[subnet-5bddc82c, subnet-478a4123]",
				"kubernetes_version":                                          "",
				"max_map_count":                                               "262144",
				"max_pods":                                                    "110",
				"node_pools__compute_fast_ephemeral__aws_ami":                 "",
				"node_pools__compute_fast_ephemeral__aws_instance_type":       "m5d.2xlarge",
				"node_pools__compute_fast_ephemeral__count":                   "1",
				"node_pools__compute_fast_ephemeral__kubelet_node_labels":     "[node-role.kubernetes.io/compute=\"\", node.kubernetes.io/compute=\"\", ephemeral-volumes=fast]",
				"node_pools__compute_fast_ephemeral__kubelet_node_taints":     "[]",
				"node_pools__compute_fast_ephemeral__placementgroup_strategy": "",
				"node_pools__compute_fast_ephemeral__root_volume_size":        "100",
				"node_pools__compute_fast_ephemeral__security_groups":         "[]",
				"node_pools__compute_fast_ephemeral__worker_pool_subnets":     "[]",
				"node_pools__compute_slow_ephemeral__aws_ami":                 "",
				"node_pools__compute_slow_ephemeral__aws_instance_type":       "m5.2xlarge",
				"node_pools__compute_slow_ephemeral__count":                   "1",
				"node_pools__compute_slow_ephemeral__kubelet_node_labels":     "[node-role.kubernetes.io/compute=\"\", node.kubernetes.io/compute=\"\", ephemeral-volumes=slow]",
				"node_pools__compute_slow_ephemeral__kubelet_node_taints":     "[]",
				"node_pools__compute_slow_ephemeral__placementgroup_strategy": "",
				"node_pools__compute_slow_ephemeral__root_volume_size":        "100",
				"node_pools__compute_slow_ephemeral__security_groups":         "[]",
				"node_pools__compute_slow_ephemeral__worker_pool_subnets":     "[]",
				"node_pools__persistent_storage__aws_ami":                     "",
				"node_pools__persistent_storage__aws_instance_type":           "i3.2xlarge",
				"node_pools__persistent_storage__count":                       "3",
				"node_pools__persistent_storage__kubelet_node_labels":         "[node-role.kubernetes.io/persistent=\"\", node.kubernetes.io/persistent=\"\", ephemeral-volumes=slow, storage=persistent]",
				"node_pools__persistent_storage__kubelet_node_taints":         "[storage=persistent:NoSchedule]",
				"node_pools__persistent_storage__placementgroup_strategy":     "spread",
				"node_pools__persistent_storage__root_volume_size":            "100",
				"node_pools__persistent_storage__security_groups":             "[]",
				"node_pools__persistent_storage__worker_pool_subnets":         "[]",
				"private_key":                          "",
				"private_key_file":                     "",
				"public_key":                           "",
				"public_key_file":                      "",
				"route_53_name":                        "[]",
				"s3_buckets":                           "[]",
				"username":                             "ec2-user",
				"clustername":                          "eks01",
				"node_pools__persistent_storage__name": "persistent_storage",
				"node_pools__compute_fast_ephemeral__name": "compute_fast_ephemeral",
				"node_pools__compute_slow_ephemeral__name": "compute_slow_ephemeral",
				"default_node_pool__name":                  "default-pool",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kluster{
				Name: tt.clusterName,
			}

			if len(tt.configJSON) != 0 {
				config := &configurator.Config{}
				json.Unmarshal(tt.configJSON, config)
				k.Config = config
			}

			if len(tt.platformVars) != 0 {
				platform, err := provisioner.New(tt.clusterName, tt.platform, tt.platformVars, nil, tt.version)
				if err != nil {
					t.Errorf("failed to create the %s provisioner for %q", tt.platform, tt.clusterName)
				}
				k.Platforms = make(map[string]interface{}, 1)
				k.provisioner = make(map[string]provisioner.Provisioner, 1)
				k.Platforms[tt.platform] = platform.Config()
				k.provisioner[tt.platform] = platform
			}

			got, err := k.ConfigVariables()
			if (err != nil) != tt.wantErr {
				t.Errorf("Kluster.ConfigVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				printDiff(t, got, tt.want)
				t.Errorf("Kluster.ConfigVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func printDiff(t *testing.T, m1, m2 map[string]string) {
	d1 := diffMap(m1, m2)
	var diff1 string
	for k, v := range d1 {
		diff1 = fmt.Sprintf("%s\n\t- %q : %s", diff1, k, v)
	}
	t.Logf("Kluster.ConfigVariables() differs from expected result in: %s", diff1)
	d2 := diffMap(m2, m1)
	var diff2 string
	for k, v := range d2 {
		diff2 = fmt.Sprintf("%s\n\t+ %q : %s", diff2, k, v)
	}
	t.Logf("Expected result differs from Kluster.ConfigVariables() in: %v", diff2)
}

func diffMap(m1, m2 map[string]string) map[string]string {
	d := map[string]string{}

	for k1, v1 := range m1 {
		found := false
		var v string
		for k2, v2 := range m2 {
			if k1 == k2 {
				v = v2
				found = true
				break
			}
		}
		if !found {
			d[k1] = v1 + " (not found)"
			continue
		}
		if v1 != v {
			d[k1] = v1 + " (differs from: " + v + ")"
		}
	}

	return d
}
