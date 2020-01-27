package ec2

// func TestConfig_MergeWithEnvConfig(t *testing.T) {
// 	type fields struct {
// 		clusterName       string
// 		KubeAPISSLPort    int
// 		DisableMasterHA   bool
// 		KubeVIPAPISSLPort int
// 		Username          string
// 		AwsAccessKey      string
// 		AwsSecretKey      string
// 		AwsRegion         string
// 		AwsVpcID          string
// 		PrivateKey        string
// 		PrivateKeyFile    string
// 		PublicKey         string
// 		PublicKeyFile     string
// 		DefaultNodePool   NodePool
// 		NodePools         map[string]NodePool
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
// 			name: "config merged with aws variable",
// 			fields: fields{
// 				Username:      defaultConfig.Username,
// 				PublicKeyFile: "~/.ssh/id_rsa.pub",
// 			},
// 			args: args{
// 				envConfig: map[string]string{
// 					"aws_username": "myuser",
// 					"username":     "fakeuser",
// 				},
// 			},
// 			result: &Config{
// 				Username:      "myuser",
// 				PublicKeyFile: "~/.ssh/id_rsa.pub",
// 			},
// 		},
// 		// Test #3
// 		{
// 			name: "config merged with non aws variable",
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
// 				clusterName:       tt.fields.clusterName,
// 				KubeAPISSLPort:    tt.fields.KubeAPISSLPort,
// 				DisableMasterHA:   tt.fields.DisableMasterHA,
// 				KubeVIPAPISSLPort: tt.fields.KubeVIPAPISSLPort,
// 				Username:          tt.fields.Username,
// 				AwsAccessKey:      tt.fields.AwsAccessKey,
// 				AwsSecretKey:      tt.fields.AwsSecretKey,
// 				AwsRegion:         tt.fields.AwsRegion,
// 				AwsVpcID:          tt.fields.AwsVpcID,
// 				PrivateKey:        tt.fields.PrivateKey,
// 				PrivateKeyFile:    tt.fields.PrivateKeyFile,
// 				PublicKey:         tt.fields.PublicKey,
// 				PublicKeyFile:     tt.fields.PublicKeyFile,
// 				DefaultNodePool:   tt.fields.DefaultNodePool,
// 				NodePools:         tt.fields.NodePools,
// 			}
// 			c.MergeWithEnvConfig(tt.args.envConfig)
// 			assert.Equal(t, tt.result, c)
// 		})
// 	}
// }
