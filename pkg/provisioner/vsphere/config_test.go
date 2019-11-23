package vsphere

// func TestConfig_MergeWithEnvConfig(t *testing.T) {
// 	type fields struct {
// 		clusterName             string
// 		KubeAPISSLPort          int
// 		DisableMasterHA         bool
// 		KubeVirtualIPShortname  string
// 		KubeVirtualIPApi        string
// 		KubeVIPAPISSLPort       int
// 		PublicAPIServerDNSName  string
// 		PrivateAPIServerDNSName string
// 		Username                string
// 		VsphereUsername         string
// 		VspherePassword         string
// 		VsphereServer           string
// 		Datacenter              string
// 		Datastore               string
// 		ResourcePool            string
// 		VsphereNet              string
// 		Folder                  string
// 		DNSServers              []string
// 		PrivateKey              string
// 		PrivateKeyFile          string
// 		PublicKey               string
// 		PublicKeyFile           string
// 		DefaultNodePool         NodePool
// 		NodePools               map[string]NodePool
// 	}
// 	type args struct {
// 		envConfig map[string]string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		result *Config
// 	}{
// 		// Test #1
// 		{
// 			name: "config merged with nil",
// 			fields: fields{
// 				Username: defaultConfig.Username,
// 			},
// 			args: args{
// 				envConfig: nil,
// 			},
// 			result: &Config{
// 				Username: defaultConfig.Username,
// 			},
// 		},
// 		// Test #2
// 		{
// 			name: "config merged with vsphere variable",
// 			fields: fields{
// 				Username:      defaultConfig.Username,
// 				PublicKeyFile: "~/.ssh/id_rsa.pub",
// 			},
// 			args: args{
// 				envConfig: map[string]string{
// 					"vsphere_username": "myuser",
// 					"username":         "fakeuser",
// 				},
// 			},
// 			result: &Config{
// 				Username:      "myuser",
// 				PublicKeyFile: "~/.ssh/id_rsa.pub",
// 			},
// 		},
// 		// Test #3
// 		{
// 			name: "config merged with non vsphere variable",
// 			fields: fields{
// 				Username:      defaultConfig.Username,
// 				PublicKeyFile: "~/.ssh/id_rsa.pub",
// 			},
// 			args: args{
// 				envConfig: map[string]string{
// 					"username": "myuser",
// 				},
// 			},
// 			result: &Config{
// 				Username:      "myuser",
// 				PublicKeyFile: "~/.ssh/id_rsa.pub",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &Config{
// 				clusterName:             tt.fields.clusterName,
// 				KubeVirtualIPShortname:  tt.fields.KubeVirtualIPShortname,
// 				KubeVirtualIPApi:        tt.fields.KubeVirtualIPApi,
// 				PublicAPIServerDNSName:  tt.fields.PublicAPIServerDNSName,
// 				PrivateAPIServerDNSName: tt.fields.PrivateAPIServerDNSName,
// 				VsphereUsername:         tt.fields.VsphereUsername,
// 				VspherePassword:         tt.fields.VspherePassword,
// 				VsphereServer:           tt.fields.VsphereServer,
// 				Datacenter:              tt.fields.Datacenter,
// 				Datastore:               tt.fields.Datastore,
// 				ResourcePool:            tt.fields.ResourcePool,
// 				VsphereNet:              tt.fields.VsphereNet,
// 				Folder:                  tt.fields.Folder,
// 				DNSServers:              tt.fields.DNSServers,
// 				KubeAPISSLPort:          tt.fields.KubeAPISSLPort,
// 				DisableMasterHA:         tt.fields.DisableMasterHA,
// 				KubeVIPAPISSLPort:       tt.fields.KubeVIPAPISSLPort,
// 				Username:                tt.fields.Username,
// 				PrivateKey:              tt.fields.PrivateKey,
// 				PrivateKeyFile:          tt.fields.PrivateKeyFile,
// 				PublicKey:               tt.fields.PublicKey,
// 				PublicKeyFile:           tt.fields.PublicKeyFile,
// 				DefaultNodePool:         tt.fields.DefaultNodePool,
// 				NodePools:               tt.fields.NodePools,
// 			}
// 			c.MergeWithEnvConfig(tt.args.envConfig)
// 			assert.Equal(t, tt.result, c)
// 		})
// 	}
// }
