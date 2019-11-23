// This program generates 'code.go' with the Ansible code located in `templates/`
// and `templates/roles`. It is invoked by running: go generate
package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type ansibleArgs struct {
	src     string
	dst1    string
	dst2    string
	inject  []string
	exclude []string
}

type ansibleTest []struct {
	name       string
	args       ansibleArgs
	wantErr    bool
	shouldPass bool
}

var ansibleTests ansibleTest

func init() {
	ansibleTests = ansibleTest{
		{
			name: "simple",
			args: ansibleArgs{
				src:     "../../templates/ansible",
				dst1:    "code1.go",
				dst2:    "code2.go",
				inject:  []string{},
				exclude: []string{},
			},
			wantErr:    false,
			shouldPass: true,
		},
		{
			name: "with files to exclude",
			args: ansibleArgs{
				src:     "../../templates/ansible",
				dst1:    "code1.go",
				dst2:    "code2.go",
				inject:  []string{"roles/test.bak"},
				exclude: []string{"*.bak"},
			},
			wantErr:    false,
			shouldPass: true,
		},
		{
			name: "with files to exclude",
			args: ansibleArgs{
				src:     "../../templates/ansible",
				dst1:    "code1.go",
				dst2:    "code2.go",
				inject:  []string{"roles/test.bak"},
				exclude: []string{},
			},
			wantErr:    false,
			shouldPass: false,
		},
		{
			name: "with files to exclude",
			args: ansibleArgs{
				src:     "../../templates/ansible",
				dst1:    "code1.go",
				dst2:    "code2.go",
				inject:  []string{"roles/test.bak", "roles/test.yyy"},
				exclude: []string{"*.yyy", "*.bak"},
			},
			wantErr:    false,
			shouldPass: true,
		},
	}
}

func Test_shouldIgnore(t *testing.T) {
	type args struct {
		path        string
		excludeList []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "none", args: args{path: "../template/ansible/kubekit.py", excludeList: []string{}}, want: false},
		{name: "one not included", args: args{path: "../template/ansible/kubekit.py", excludeList: []string{"*.www"}}, want: false},
		{name: "two not included", args: args{path: "../template/ansible/kubekit.py", excludeList: []string{"*.www", "*.yyy"}}, want: false},
		{name: "one included", args: args{path: "../template/ansible/kubekit.py", excludeList: []string{"*.py", "*.www"}}, want: true},
		{name: "lastone included", args: args{path: "../template/ansible/kubekit.py", excludeList: []string{"*.yyy", "*.py"}}, want: true},
		{name: "for 2nd test", args: args{path: "./ansible/roles/test.bak", excludeList: []string{"*.bak"}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldIgnore(tt.args.path, tt.args.excludeList...); got != tt.want {
				t.Errorf("shouldIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genAnsibleCode(t *testing.T) {
	for _, tt := range ansibleTests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := ioutil.TempDir("", "codegen_test")
			if err != nil {
				t.Fatalf("could not create temporal directory: %s", err)
			}
			defer func() {
				t.Logf("removing temporal directory %s", tmp)
				os.RemoveAll(tmp)
			}()

			dst1 := filepath.Join(tmp, tt.args.dst1)
			dst2 := filepath.Join(tmp, tt.args.dst2)

			t.Logf("generating Ansible code from %s to %s excluding %v", tt.args.src, dst1, tt.args.exclude)
			if err := genAnsibleCode(tt.args.src, dst1, tt.args.exclude...); (err != nil) != tt.wantErr {
				t.Errorf("genAnsibleCode() error = %v, wantErr %v", err, tt.wantErr)
			}

			tmpTemplates, err := cpTemplates(t, tt.args.src, tmp, tt.args.inject...)
			if err != nil {
				t.Fatalf("error copying the source directory %s into %s, or maybe injecting the files %v. \nOutput: %s. \nError: %s", tt.args.src, tmp, tt.args.inject, tmpTemplates, err)
			}

			t.Logf("generating temporal Ansible code from %s to %s excluding %v", tmpTemplates, dst2, tt.args.exclude)
			if err := genAnsibleCode(tmpTemplates, dst2, tt.args.exclude...); (err != nil) != tt.wantErr {
				t.Errorf("genAnsibleCode() error = %v, wantErr %v", err, tt.wantErr)
			}

			ok, err := areEqualOutputFiles(t, dst1, dst2)
			if err != nil {
				t.Fatalf("error comparing the output files %s and %s: %s", dst1, dst2, err)
			}
			if ok != tt.shouldPass {
				t.Errorf("Hashes don't match for files %s and %s", dst1, dst2)
			}
		})
	}
}

func Test_cmdToGenAnsibleCode(t *testing.T) {
	for _, tt := range ansibleTests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := ioutil.TempDir("", "codegen_test")
			if err != nil {
				t.Fatalf("could not create temporal directory: %s", err)
			}
			defer func() {
				t.Logf("removing temporal directory %s", tmp)
				os.RemoveAll(tmp)
			}()

			dst1 := filepath.Join(tmp, tt.args.dst1)
			dst2 := filepath.Join(tmp, tt.args.dst2)

			t.Logf("generating Ansible code from %s to %s excluding %v", tt.args.src, dst1, tt.args.exclude)
			output, err := executeCmd(tt.args.src, dst1, true, false, tt.args.exclude...)
			t.Log(string(output))
			if (err != nil) != tt.wantErr {
				t.Errorf("command execution error = %v, wantErr %v", err, tt.wantErr)
			}

			tmpTemplates, err := cpTemplates(t, tt.args.src, tmp, tt.args.inject...)
			if err != nil {
				t.Fatalf("error copying the source directory %s into %s, or maybe injecting the files %v. \nOutput: %s. \nError: %s", tt.args.src, tmp, tt.args.inject, tmpTemplates, err)
			}

			t.Logf("generating temporal Ansible code from %s to %s excluding %v", tmpTemplates, dst2, tt.args.exclude)
			output, err = executeCmd(tmpTemplates, dst2, true, false, tt.args.exclude...)
			t.Log(string(output))
			if (err != nil) != tt.wantErr {
				t.Errorf("command execution error = %v, wantErr %v", err, tt.wantErr)
			}

			ok, err := areEqualOutputFiles(t, dst1, dst2)
			if err != nil {
				t.Fatalf("error comparing the output files %s and %s: %s", dst1, dst2, err)
			}
			if ok != tt.shouldPass {
				t.Errorf("Hashes don't match for files %s and %s", dst1, dst2)
			}
		})
	}
}

func executeCmd(src, dst string, doAnsible, doK8s bool, exclude ...string) (output []byte, err error) {
	params := []string{"run", "main.go", "-src", src, "-dst", dst}
	if len(exclude) != 0 {
		excludeList := strings.Join(exclude, ",")
		params = append(params, "-exclude", excludeList)
	}

	output, err = exec.Command("go", params...).CombinedOutput()
	return output, err
}

func cpTemplates(t *testing.T, src, tmp string, inject ...string) (string, error) {
	tmpTemplates := filepath.Join(tmp, "templates_ansible")

	// copy template files to temp location to isolate bugs related to the source path
	t.Logf("Copying the source directory %s in a temporal directory %s", src, tmpTemplates)
	output, err := exec.Command("cp", "-R", src, tmpTemplates).CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("err while running 'cp' command: %s", err)
	}

	if len(inject) != 0 {
		// add a garbage file to the templates' role dir and confirm the MD5 doesn't change when
		// the appropriate exclusion argument is provided
		for _, filename := range inject {
			f := filepath.Join(tmpTemplates, filename)
			t.Logf("injecting file %s", f)
			if err := ioutil.WriteFile(f, []byte("this should not get zipped"), 0755); err != nil {
				return "", fmt.Errorf("unable to create injected file %s. %s", f, err)
			}
		}
	}

	return tmpTemplates, nil
}

func areEqualOutputFiles(t *testing.T, filename1, filename2 string) (bool, error) {
	f1, err := ioutil.ReadFile(filename1)
	if err != nil {
		return false, fmt.Errorf("can't read file 1: %s", err)
	}
	f1md5 := md5.Sum(f1)

	f2, err := ioutil.ReadFile(filename2)
	if err != nil {
		return false, fmt.Errorf("can't read file 2: %s", err)
	}
	f2md5 := md5.Sum(f2)

	if f1md5 == f2md5 {
		return true, nil
	}

	t.Logf("Hashes don't match")
	t.Logf("Hash1 - file 1 %s : %x", filename1, f1md5)
	t.Logf("Hash2 - file 2 %s : %x", filename2, f2md5)

	// show diff so we understand what parts of the two files
	// are different
	// if output, err := exec.Command("diff", filename1, filename2).CombinedOutput(); err != nil {
	// 	t.Log(string(output))
	// 	t.Logf("err while running 'diff' command: %s", err)
	// }

	return false, nil
}
