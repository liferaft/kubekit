// This program generates 'code.go' with the Ansible code located in `templates/`
// and `templates/roles`. It is invoked by running: go generate
package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Data contain the data used to render the templates
type Data struct {
	Source   string
	Data     string
	Config   string
	Callback string
	Playbook string
	Version  string
}

var codeTemplate = `// Code generated automatically by 'go run codegen/ansible/main.go --src ./templates/ansible --dst ./code.go'; DO NOT EDIT.
package configurator

// Roles contain the zipped data of the Ansible roles to execute at the hosts
func init() {
	AnsibleCfg = ` + "`" + `{{ .Config }}` + "`" + `
	Callback = ` + "`" + `{{ .Callback }}` + "`" + `
	Playbook = ` + "`" + `{{ .Playbook }}` + "`" + `
	Data = "{{ .Data }}"
}
`

var (
	dst     string
	src     string
	exclude string
)

func init() {
	flag.StringVar(&dst, "dst", "./code.go", "destination to store generated code")
	flag.StringVar(&src, "src", "./templates/ansible", "location of various template files")
	flag.StringVar(&exclude, "exclude", "", "file pattern to exclude from roles directory (e.g. *.bak)")
}

func zipRoles(source string, excludeList ...string) []byte {
	var zipBuffer = new(bytes.Buffer)

	archive := zip.NewWriter(zipBuffer)

	var err error
	source, err = filepath.Abs(source)
	if err != nil {
		panic(err)
	}

	//baseDir := filepath.Base(source)

	if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip directories and file pattern exclusions
		if info.IsDir() || shouldIgnore(info.Name(), excludeList...) {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = filepath.Join("roles", header.Name)

		//These next line set the born on date for the files in the archive
		header.SetModTime(time.Date(2018, time.November, 10, 23, 0, 0, 0, time.UTC))
		header.ModifiedTime = 0
		header.ModifiedDate = 0
		header.Modified = time.Date(2018, time.November, 10, 23, 0, 0, 0, time.UTC)
		header.SetMode(0644) //set file permissions

		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	}); err != nil {
		panic(err)
	}
	if err := archive.Close(); err != nil {
		panic(err)
	}

	return zipToData(zipBuffer.Bytes())
}

func match(path, rule string) bool {
	ok, err := filepath.Match(rule, path)
	if err != nil {
		panic(err)
	}
	if ok {
		log.Printf("ignoring %s from rule %s", path, rule)
		return true
	}

	return false
}

func shouldIgnore(path string, excludeList ...string) bool {
	if len(excludeList) == 0 {
		return false
	}

	for _, rule := range excludeList {
		// Check just the filename
		if match(filepath.Base(path), rule) {
			return true
		}
		// Check the entire path
		if match(path, rule) {
			return true
		}
	}

	return false
}

func zipToData(zipData []byte) []byte {
	var buffer = new(bytes.Buffer)

	for _, b := range zipData {
		switch {
		case b == '\n':
			buffer.WriteString(`\n`)
		case b == '\\':
			buffer.WriteString(`\\`)
		case b == '"':
			buffer.WriteString(`\"`)
		case (b >= 32 && b <= 126) || b == '\t':
			buffer.WriteByte(b)
		default:
			fmt.Fprintf(buffer, "\\x%02x", b)
		}
	}

	return buffer.Bytes()
}

func getFileContent(source string) []byte {
	content, err := ioutil.ReadFile(source)
	if err != nil {
		panic(err)
	}
	return content
}

func genAnsibleCode(src, dst string, exclude ...string) error {
	log.Printf("Generating Ansible code")
	rolesTemplate := template.Must(template.New(dst).Parse(codeTemplate))

	log.Printf("Getting content from files in: %s", src)
	data := zipRoles(filepath.Join(src, "roles"), exclude...)
	config := getFileContent(filepath.Join(src, "ansible.cfg"))
	callback := getFileContent(filepath.Join(src, "kubekit.py"))
	playbook := getFileContent(filepath.Join(src, "kubekit.yml"))

	log.Printf("Creating destination file: %s", dst)
	goFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		goFile.Close()
		log.Printf("Done with Ansible code")
	}()

	tmplData := Data{
		Source:   src,
		Data:     string(data),
		Config:   string(config),
		Callback: string(callback),
		Playbook: string(playbook),
	}

	return rolesTemplate.Execute(goFile, tmplData)
}

func main() {
	flag.Parse()

	excludeList := strings.Split(exclude, ",")

	if err := genAnsibleCode(src, dst, excludeList...); err != nil {
		panic(err)
	}

}
