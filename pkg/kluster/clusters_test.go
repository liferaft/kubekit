package kluster

import (
	"io/ioutil"
	"testing"

	"github.com/johandry/log"
	"github.com/sirupsen/logrus"
	"github.com/kraken/ui"
)

var (
	parentUI *ui.UI
)

func init() {
	l := log.NewDefault()
	l.SetLevel(logrus.DebugLevel)
	parentUI = ui.New(false, l)
}

func TestValidClusterName(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		want        string
		wantErr     bool
		errMsg      string
	}{
		{"empty cluster name", "", "", true, "cluster name cannot be empty"},
		{"uppercase char", "SomeThing", "something", true, "cluster cannot have uppercase characters"},
		{"underscore char", "some_thing", "some-thing", true, "cluster name cannot contain underscore ('_')"},
		{"uppercase and underscore chars 1", "Some_ThingHere", "some-thinghere", true, "cluster name cannot contain uppercase characters or underscore ('_')"},
		{"uppercase and underscore chars 2", "Something_Is.Wrong09", "something-is.wrong09", true, "cluster name cannot contain uppercase characters or underscore ('_')"},
		{"uppercase and underscore chars 3", "100Something_Is.Wrong", "100something-is.wrong", true, "cluster name cannot contain uppercase characters or underscore ('_')"},
		{"uppercase and underscore chars 4", "Something_Is.Wrong", "something-is.wrong", true, "cluster name cannot contain uppercase characters or underscore ('_')"},
		{"valid name 1", "agoodname", "agoodname", false, ""},
		{"valid name 2", "kkdemo", "kkdemo", false, ""},
		{"valid name 3", "kkdemo01", "kkdemo01", false, ""},
		{"valid name 4", "00kkdemo", "00kkdemo", false, ""},
		{"valid name 5", "kk.demo", "kk.demo", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidClusterName(tt.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidClusterName(%q) error = %v, wantErr %v", tt.clusterName, err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidClusterName(%q) error = %v, errMsg %v", tt.clusterName, err, tt.errMsg)
				return
			}
			if got != tt.want {
				t.Errorf("ValidClusterName(%q) = %v, want %v", tt.clusterName, got, tt.want)
			}
		})
	}
}

func TestCreateCluster(t *testing.T) {
	path, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("CreateCluster() failed to create a temporal directory. %v", err)
	}
	// This is the default format
	format := "yaml"
	requiredValueTxt := "# Required value. Example: "

	type args struct {
		clusterName string
		platform    string
		variables   map[string]string
	}
	type kluster struct {
		name               string
		awsVpcID           string
		awsSecurityGroupID string
		awsSubnetID        string
	}
	tests := []struct {
		name        string
		args        args
		wantCluster kluster
		wantErr     bool
	}{
		{"default cluster on aws", args{"kkdemo01", "aws", map[string]string{}}, kluster{name: "kkdemo01",
			awsVpcID:           requiredValueTxt + "vpc-8d56b9e9",
			awsSecurityGroupID: requiredValueTxt + "sg-502d9a37",
			awsSubnetID:        requiredValueTxt + "subnet-5bddc82c",
		}, false},
		{"simple cluster on aws", args{"kkdemo02", "aws", map[string]string{
			"aws_subnet_id": "vpc-8d56b9e9",
			"default_node_pool__aws_security_group_id": "sg-502d9a37",
			"default_node_pool__aws_subnet_id":         "subnet-5bddc82c",
		}}, kluster{
			name:               "kkdemo02",
			awsVpcID:           "vpc-8d56b9e9",
			awsSecurityGroupID: "sg-502d9a37",
			awsSubnetID:        "subnet-5bddc82c",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCluster, err := CreateCluster(tt.args.clusterName, tt.args.platform, path, format, tt.args.variables, parentUI)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCluster.Name != tt.wantCluster.name {
				t.Errorf("CreateCluster() Name = %v, want %v", gotCluster.Name, tt.wantCluster.name)
			}
		})
	}
}
