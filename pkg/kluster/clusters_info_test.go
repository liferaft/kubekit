package kluster

import (
	"reflect"
	"sort"
	"testing"
)

func TestClusterInfo_ContainsAll(t *testing.T) {
	testClustersInfo := ClustersInfo{
		ClusterInfo{"demo01", 3, "aws", "running", "1.0", "/home/user/.kubekit.d/clusters/UID01", "http://fake.com/entrypoint:8080", "/home/user/.kubekit.d/clusters/UID01/certificates/kubeconfig"},
		ClusterInfo{"demo03", 0, "vsphere", "absent", "1.0", "/home/user/.kubekit.d/clusters/UID02", "None", ""},
	}

	tests := []struct {
		name   string
		ci     ClusterInfo
		params map[string]string
		want   bool
	}{
		{"no params", testClustersInfo[0], map[string]string{}, true},
		{"simple", testClustersInfo[1], map[string]string{"name": "demo03"}, true},
		{"non lowercase", testClustersInfo[1], map[string]string{"NaMe": "demo03"}, true},
		{"case in value matters", testClustersInfo[1], map[string]string{"NaMe": "Demo03"}, false},
		{"invalid parameter", testClustersInfo[1], map[string]string{"fake": "param"}, false},
		{"contain entrypoint", testClustersInfo[0], map[string]string{"entrypoint": "http://fake.com/entrypoint:8080"}, true},
		{"contain empty entrypoint", testClustersInfo[1], map[string]string{"entrypoint": ""}, true},
		{"not contains all", testClustersInfo[1], map[string]string{"name": "demo03", "platform": "eks"}, false},
		{"check non string, default value", testClustersInfo[1], map[string]string{"nodes": "0"}, true},
		{"check non string, correct value", testClustersInfo[0], map[string]string{"nodes": "3"}, true},
		{"check non string, incorrect value", testClustersInfo[0], map[string]string{"nodes": "4"}, false},
		{"contains all", testClustersInfo[1], map[string]string{"name": "demo03", "nodes": "0", "platform": "vsphere", "status": "absent", "version": "1.0", "path": "/home/user/.kubekit.d/clusters/UID02", "url": "None", "kubeconfig": ""}, true},
		{"contains all", testClustersInfo[0], map[string]string{"name": "demo01", "nodes": "3", "platform": "aws", "status": "running", "version": "1.0", "path": "/home/user/.kubekit.d/clusters/UID01", "url": "http://fake.com/entrypoint:8080", "kubeconfig": "/home/user/.kubekit.d/clusters/UID01/certificates/kubeconfig"}, true},
		{"contains all but one", testClustersInfo[0], map[string]string{"name": "demo01", "nodes": "4", "platform": "aws", "status": "running", "version": "1.0", "path": "/home/user/.kubekit.d/clusters/UID01", "url": "http://fake.com/entrypoint:8080", "kubeconfig": "/home/user/.kubekit.d/clusters/UID01/certificates/kubeconfig"}, false},
		{"contains all but one invalid", testClustersInfo[0], map[string]string{"fake": "param", "name": "demo01", "nodes": "3", "platform": "aws", "status": "running", "version": "1.0", "path": "/home/user/.kubekit.d/clusters/UID01", "url": "http://fake.com/entrypoint:8080", "kubeconfig": "/home/user/.kubekit.d/clusters/UID01/certificates/kubeconfig"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ci.ContainsAll(tt.params); got != tt.want {
				t.Errorf("ClusterInfo.ContainsAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClustersInfo_FilterBy(t *testing.T) {
	clustersInfoTestData := ClustersInfo{
		{Name: "demo01", Platform: "aws", Nodes: 0},
		{Name: "demo02", Platform: "aws", Nodes: 3},
		{Name: "demo03", Platform: "eks", Nodes: 3},
		{Name: "demo04", Platform: "aks", Nodes: 0},
	}
	tests := []struct {
		name   string
		ci     ClustersInfo
		params map[string]string
		want   ClustersInfo
	}{
		{"filter by nothing", clustersInfoTestData, map[string]string{}, clustersInfoTestData},
		{"nothing to filter by", ClustersInfo{}, map[string]string{"name": "demo01"}, ClustersInfo{}},
		{"filter to get one item", clustersInfoTestData, map[string]string{"name": "demo01"}, ClustersInfo{{Name: "demo01", Platform: "aws", Nodes: 0}}},
		{"filter to get multimple items", clustersInfoTestData, map[string]string{"platform": "aws"}, ClustersInfo{{Name: "demo01", Platform: "aws", Nodes: 0}, {Name: "demo02", Platform: "aws", Nodes: 3}}},
		{"filter by non-string", clustersInfoTestData, map[string]string{"nodes": "3"}, ClustersInfo{{Name: "demo02", Platform: "aws", Nodes: 3}, {Name: "demo03", Platform: "eks", Nodes: 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ci.FilterBy(tt.params)
			if !reflect.DeepEqual(tt.ci, tt.want) {
				t.Errorf("ClusterInfo.FilterBy() => %v, want %v", tt.ci, tt.want)
			}
		})
	}
}

var testsFilterParams = []struct {
	name            string
	params          map[string]string
	wantIsValid     bool
	wantInvalidList []string
}{
	{"empty list", map[string]string{}, true, []string{}},
	{"valid", map[string]string{"name": "somename", "nodes": "3"}, true, []string{}},
	{"no lowercase", map[string]string{"nAme": "somename", "noDes": "3"}, true, []string{}},
	{"one invalid", map[string]string{"name": "somename", "nods": "3"}, false, []string{"nods"}},
	{"all invalid", map[string]string{"nme": "somename", "nods": "3"}, false, []string{"nme", "nods"}},
}

func TestIsValidFilter(t *testing.T) {
	for _, tt := range testsFilterParams {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidFilter(tt.params); got != tt.wantIsValid {
				t.Errorf("IsValidFilter() = %v, want %v", got, tt.wantIsValid)
			}
		})
	}
}

func TestInvalidFilterParams(t *testing.T) {
	for _, tt := range testsFilterParams {
		t.Run(tt.name, func(t *testing.T) {
			if got := InvalidFilterParams(tt.params); !arraySortedEqual(got, tt.wantInvalidList) {
				t.Errorf("InvalidFilterParams() = %v, want %v", got, tt.wantInvalidList)
			}
		})
	}
}

// test if the string arrays are equal when sorted
// useful in this case because map iteration doesn't guarantee order
func arraySortedEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))

	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	return reflect.DeepEqual(aCopy, bCopy)
}

func TestClustersInfo_Template(t *testing.T) {
	testClustersInfo := ClustersInfo{
		ClusterInfo{"demo01", 3, "aws", "running", "1.0", "/home/user/.kubekit.d/clusters/UID01", "http://fake.com/entrypoint:8080", "/home/user/.kubekit.d/clusters/UID01/certificates/kubeconfig"},
		ClusterInfo{"demo03", 0, "vsphere", "absent", "1.0", "/home/user/.kubekit.d/clusters/UID02", "None", ""},
	}
	tests := []struct {
		name    string
		ci      ClustersInfo
		format  string
		want    string
		wantErr bool
	}{
		{"empty ci & template", ClustersInfo{}, "", "", false},
		{"empty ci, render root", ClustersInfo{}, "{{.}}", "", false},
		{"empty template", testClustersInfo, "", "", false},
		{"just a text", testClustersInfo, "foo", "foo\nfoo\n", false},
		{"just names", testClustersInfo, "{{.Name}}", "demo01\ndemo03\n", false},
		{"names and platform", testClustersInfo, "{{.Name}}\t{{.Platform}}", "demo01\taws\ndemo03\tvsphere\n", false},
		{"table, names and platform", testClustersInfo, "table {{.Name}}\t{{.Platform}}", "NAME     PLATFORM\ndemo01   aws\ndemo03   vsphere\n", false},
		{"table, names and nodes", testClustersInfo, "table {{.Name}}\t{{.Nodes}}", "NAME     NODES\ndemo01   3\ndemo03   0\n", false},
		{"table, names and entrypoint with URL", testClustersInfo, "table {{.Name}}\t{{.URL}}", "NAME     ENTRYPOINT\ndemo01   http://fake.com/entrypoint:8080\ndemo03   None\n", false},
		{"table, names and entrypoint", testClustersInfo, "table {{.Name}}\t{{.Entrypoint}}", "NAME     ENTRYPOINT\ndemo01   http://fake.com/entrypoint:8080\ndemo03   None\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ci.Template(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClustersInfo.Template() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ClustersInfo.Template() = '%v', want %v", got, tt.want)
			}
		})
	}
}
