package cli

import (
	"fmt"
	"testing"
)

func TestGetEnvOpts_SprintEnv(t *testing.T) {
	type fields struct {
		ClusterName    string
		Shell          string
		Unset          bool
		KubeconfigFile string
	}
	tests := []struct {
		name   string
		fields fields
		env    map[string]string
		want   string
	}{
		{"bash", fields{"kkdemo01", "bash", false, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "export KUBECONFIG=/this/good/path/config" + makeComments("kkdemo01", "", "")},
		{"bash unset", fields{"kkdemo01", "bash", true, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "unset KUBECONFIG" + makeComments("kkdemo01", " -u", "")},

		{"fish", fields{"kkdemo01", "fish", false, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "set -x KUBECONFIG /this/good/path/config;" + makeComments("kkdemo01", "", " --shell fish")},
		{"fish unset", fields{"kkdemo01", "fish", true, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "unset KUBECONFIG;" + makeComments("kkdemo01", " -u", " --shell fish")},

		{"cmd", fields{"kkdemo01", "cmd", false, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "set KUBECONFIG=/this/good/path/config" + makeComments("kkdemo01", "", " --shell cmd")},
		{"cmd unset", fields{"kkdemo01", "cmd", true, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "unset KUBECONFIG" + makeComments("kkdemo01", " -u", " --shell cmd")},

		{"powershell", fields{"kkdemo01", "powershell", false, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "$Env:KUBECONFIG = \"/this/good/path/config\"" + makeComments("kkdemo01", "", " --shell powershell")},
		{"powershell unset", fields{"kkdemo01", "powershell", true, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config"}, "unset KUBECONFIG" + makeComments("kkdemo01", " -u", " --shell powershell")},

		// {"more env", fields{"kkdemo01", "bash", false, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config", "OTHER": "foo"}, "export KUBECONFIG=/this/good/path/config\nexport OTHER=foo" + makeComments("kkdemo01", "", "")},
		// {"more env unset", fields{"kkdemo01", "bash", true, "/this/fake/path/config"}, map[string]string{"KUBECONFIG": "/this/good/path/config", "OTHER": "foo"}, "unset KUBECONFIG\nunset OTHER" + makeComments("kkdemo01", " -u", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &GetEnvOpts{
				ClusterName:    tt.fields.ClusterName,
				Shell:          tt.fields.Shell,
				Unset:          tt.fields.Unset,
				KubeconfigFile: tt.fields.KubeconfigFile,
			}
			if got := o.SprintEnv(tt.env); got != tt.want {
				t.Errorf("GetEnvOpts.SprintEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeComments(clustername, unset, shell string) string {
	return fmt.Sprintf("\n# Run this command to configure your shell:\n# eval \"$(kubekit get env %s%s%s)\"\n", clustername, unset, shell)
}
