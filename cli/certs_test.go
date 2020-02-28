package cli

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Test_getVarNames(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     string
		want1    []string
	}{
		{"Checking deprecated AWS platform", "aws", "aws", []string{"access_key", "secret_key", "session_token", "region", "profile"}},
		{"Checking EKS platform", "eks", "aws", []string{"access_key", "secret_key", "session_token", "region", "profile"}},
		{"Checking EKS platform uppercased", "EKS", "aws", []string{"access_key", "secret_key", "session_token", "region", "profile"}},
		{"Checking EC2 platform", "ec2", "aws", []string{"access_key", "secret_key", "session_token", "region", "profile"}},
		{"Checking AKS platform", "aks", "azure", []string{"subscription_id", "tenant_id", "client_id", "client_secret"}},
		{"Checking unused Azure platform", "azure", "azure", []string{"subscription_id", "tenant_id", "client_id", "client_secret"}},
		{"Checking OpenStack platform", "openstack", "openstack", []string{"server", "username", "password"}},
		{"Checking vSphere platform", "vsphere", "vsphere", []string{"server", "username", "password"}},
		{"Checking creds-less VRA platform", "vra", "vra", nil},
		{"Checking creds-less Raw platform", "raw", "raw", nil},
		{"Checking creds-less stacki platform", "stacki", "stacki", nil},
		{"Checking an unknonw platform", "foo", "foo", []string{"server", "username", "password"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getVarNames(tt.platform)
			if got != tt.want {
				t.Errorf("getVarNames() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getVarNames() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getVarValue(t *testing.T) {
	type flagArgs struct {
		name  string
		value string
	}
	type args struct {
		name      string
		envPrefix string
		flagArgs  *flagArgs
	}
	tests := []struct {
		name string
		env  map[string]string
		args args
		want string
	}{
		{"read from CLI only", nil, args{"foo", "aws", &flagArgs{"foo", "bar"}}, "bar"},
		{"read from ENV only", map[string]string{"AWS_FOO": "bar"}, args{"foo", "aws", nil}, "bar"},
		{"read from CLI with ENV", map[string]string{"AWS_FOO": "barenv"}, args{"foo", "aws", &flagArgs{"foo", "barcli"}}, "barcli"},
		{"read from ENV with CLI", map[string]string{"AWS_FOO": "barenv"}, args{"foo", "aws", &flagArgs{"foo", ""}}, "barenv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for n, v := range tt.env {
				os.Setenv(n, v)
				defer os.Unsetenv(n)
			}
			var flag *pflag.Flag
			if tt.args.flagArgs != nil {
				fs := pflag.NewFlagSet("Test", pflag.ContinueOnError)
				fs.String(tt.args.flagArgs.name, tt.args.flagArgs.value, "test flag")
				flag = fs.Lookup(tt.args.flagArgs.name)
			}

			if got := getVarValue(tt.args.name, tt.args.envPrefix, flag); got != tt.want {
				t.Errorf("getVarValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		cmdArgs  map[string]string
		env      map[string]string
		want     map[string]string
	}{
		{"EC2 from CLI only", "ec2",
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
			nil,
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
		},
		{"EC2 from ENV only", "ec2",
			nil,
			map[string]string{"AWS_ACCESS_KEY_ID": "SOME-FAKE//ACCESS_KEY", "AWS_SECRET_ACCESS_KEY": "A-FAKE_SECRET//KEY", "AWS_SESSION_TOKEN": "THIS-IS_A//FAKE||TOKEN", "AWS_DEFAULT_REGION": "us-west-2"},
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
		},
		{"EC2 from CLI and ENV", "ec2",
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": ""},
			map[string]string{"AWS_ACCESS_KEY_ID": "AN_ACCESS_KEY_YOU_WILL_NEVER_SEE", "AWS_SECRET_ACCESS_KEY": "A-FAKE_SECRET//KEY", "AWS_DEFAULT_REGION": "us-west-2"},
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
		},
		{"EKS from CLI only", "eks",
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
			nil,
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
		},
		{"AWS from ENV only", "aws",
			nil,
			map[string]string{"AWS_ACCESS_KEY_ID": "SOME-FAKE//ACCESS_KEY", "AWS_SECRET_ACCESS_KEY": "A-FAKE_SECRET//KEY", "AWS_SESSION_TOKEN": "THIS-IS_A//FAKE||TOKEN", "AWS_DEFAULT_REGION": "us-west-2"},
			map[string]string{"access_key": "SOME-FAKE//ACCESS_KEY", "secret_key": "A-FAKE_SECRET//KEY", "session_token": "THIS-IS_A//FAKE||TOKEN", "region": "us-west-2"},
		},
		{"AKS from CLI and ENV", "aks",
			map[string]string{"subscription_id": "SOME-FAKE//SUBS_ID", "client_id": "THIS-IS_A//FAKE||CLIENT-ID", "client_secret": ""},
			map[string]string{"AZURE_SUBSCRIPTION_ID": "A_SUBS_ID_YOU_WILL_NEVER_SEE", "AZURE_TENANT_ID": "A-FAKE_TENANT//ID", "AZURE_CLIENT_SECRET": "I-Have_NO_53Cr3ts"},
			map[string]string{"subscription_id": "SOME-FAKE//SUBS_ID", "tenant_id": "A-FAKE_TENANT//ID", "client_id": "THIS-IS_A//FAKE||CLIENT-ID", "client_secret": "I-Have_NO_53Cr3ts"},
		},
		{"vSphere from CLI only", "vsphere",
			map[string]string{"server": "10.10.10.10", "username": "fake", "password": "SeCret!"},
			nil,
			map[string]string{"server": "10.10.10.10", "username": "fake", "password": "SeCret!"},
		},
		{"OpenStack from ENV only", "openstack",
			nil,
			map[string]string{"OPENSTACK_SERVER": "10.10.10.10", "OPENSTACK_USERNAME": "fake", "OPENSTACK_PASSWORD": "SeCret!"},
			map[string]string{"server": "10.10.10.10", "username": "fake", "password": "SeCret!"},
		},
		{"new platform from CLI and ENV", "new",
			map[string]string{"server": "10.10.10.10", "username": "fake", "password": ""},
			map[string]string{"NEW_SERVER": "0.0.0.0", "NEW_PASSWORD": "SeCret!"},
			map[string]string{"server": "10.10.10.10", "username": "fake", "password": "SeCret!"},
		},
		{"VRA from CLI and ENV", "vra",
			map[string]string{"server": "", "username": "fake"},
			map[string]string{"VRA_SERVER": "10.10.10.10", "VRA_USERNAME": "fake", "VRA_PASSWORD": "SeCret!"},
			map[string]string{},
		},
		{"Raw from CLI and ENV", "raw",
			map[string]string{"server": "", "username": "fake"},
			map[string]string{"RAW_SERVER": "10.10.10.10", "RAW_USERNAME": "fake", "RAW_PASSWORD": "SeCret!"},
			map[string]string{},
		},
		{"Stacki from CLI and ENV", "stacki",
			map[string]string{"server": "", "username": "fake"},
			map[string]string{"STACKI_SERVER": "10.10.10.10", "STACKI_USERNAME": "fake", "STACKI_PASSWORD": "SeCret!"},
			map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for n, v := range tt.env {
				os.Setenv(n, v)
				defer os.Unsetenv(n)
				t.Logf("export %s=%s", n, v)
			}
			cmd := &cobra.Command{Use: "test"}
			if len(tt.cmdArgs) != 0 {
				args := []string{}
				for p, v := range tt.cmdArgs {
					cmd.Flags().String(p, v, "test command named: "+p)
					args = append(args, p, v)
				}
				// cmd.SetArgs(args)

				var args4log string
				for n := range tt.cmdArgs {
					v := cmd.Flags().Lookup(n).Value.String()
					args4log = args4log + " --" + n + "=" + v
				}
				t.Logf("Command: test %s", args4log)
			}

			if got := GetCredentials(tt.platform, cmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
