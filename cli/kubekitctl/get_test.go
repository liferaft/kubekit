package kubekitctl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func Test_cleanPath(t *testing.T) {
	home, err := homedir.Dir()
	if err != nil {
		t.Fatalf("cannot find home directory. %s", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot find current working directory. %s", err)
	}

	// tmp, err := ioutil.TempDir("", "cleanpath_test")
	// if err != nil {
	// 	t.Fatalf("could not create temporal directory: %s", err)
	// }
	// defer func() {
	// 	t.Logf("removing temporal directory %s", tmp)
	// 	os.RemoveAll(tmp)
	// }()

	type args struct {
		path        string
		clusterName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"home", args{"~/config", ""}, filepath.Clean(home + "/config"), false},
		{"clean", args{"/some/path/dirty/../config", ""}, "/some/path/config", false},
		{"clean #2", args{"/some/path/clean/./config", ""}, "/some/path/clean/config", false},
		// In Jenkins there is no permission to create this directory
		// {"default", args{"", "foo"}, home + "/.kube/foo.kconf", false},
		{"relative", args{"./some/relative/path/config", "foo"}, pwd + "/some/relative/path/config", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkDir := len(tt.args.path) == 0

			got, err := cleanPath(tt.args.path, tt.args.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cleanPath() = %v, want %v", got, tt.want)
				return
			}
			// If not path is given, after this test, the ~/.kube should exists
			if checkDir {
				if _, err := os.Stat(home + "/.kube"); os.IsNotExist(err) {
					t.Errorf("cleanPath() : the directory %s/.kube was not created", home)
				}
			}
		})
	}
}
