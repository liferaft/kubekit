package kubekit

import (
	"path/filepath"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
)

func Test_validFile(t *testing.T) {
	home, err := homedir.Dir()
	if err != nil {
		t.Fatalf("cannot find home directory. %s", err)
	}
	type args struct {
		certDir  string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		errMsg  string
	}{
		{"relative file", args{"/dir", "file"}, "", true, `file not found "/dir/file"`},
		{"absolute file", args{"/dir", "/other/dir/file"}, "/other/dir/file", true, `file not found "/other/dir/file"`},
		{"home not used at begining", args{"/dir", "~/file"}, "", true, `file not found "/dir/~/file"`},
		{"at home file", args{"~/dir", "file"}, "", true, `file not found "` + filepath.Join(home, "/dir/file") + `"`},
		{"no file", args{"/some/dir", ""}, "", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validFile(tt.args.certDir, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("validFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("validFile() error = %v, errMsg %v", err, tt.errMsg)
				}
				return
			}
			if got != tt.want {
				t.Errorf("validFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
