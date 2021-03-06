package main

// This program generates 'code.go' with the Ansible code located in `templates/`
// and `templates/roles`. It is invoked by running: go generate

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

// Data contain the data used to render the templates
type Data struct {
	Package      string
	ResourcesMap string
	Templates    string
	Expressions  string
}

// ResourceTemplates would contain the content of the generated code and is used to
var ResourceTemplates map[string]string

var codeTemplate = `package {{ .Package }}

// Code generated automatically by 'go run ../codegen/kubernetes/main.go --src ../templates/resources --dst code.go'; DO NOT EDIT THIS FILE.

func init() {
	ResourceTemplates = map[string]string{
		{{ .ResourcesMap }}
	}
}

// Expressions in the templates
/**
{{ .Expressions }}
**/

{{ .Templates }}
`

var (
	dst      string
	src      string
	exclude  string
	reverse  bool
	internal bool
)

var re *regexp.Regexp

func init() {
	flag.StringVar(&dst, "dst", "code.go", "destination to store generated code")
	flag.StringVar(&src, "src", "../templates/resources", "location of various template files")
	flag.StringVar(&exclude, "exclude", "", "file pattern to exclude from roles directory (e.g. *.bak)")
	flag.BoolVar(&reverse, "reverse", false, "generates the templates from a code.go file. When set the '--dst' is the source and '--src' is the destination")
	flag.BoolVar(&internal, "internal", false, "generates the code.go file to be used by the code generator. Used mostly before use '--reverse'")

	re = regexp.MustCompile(`{{.*?}}`)
}

func genKubernetesCode(src, dst string, exclude ...string) error {
	log.Printf("Generating Kubernetes manifests code")
	rolesTemplate := template.Must(template.New(dst).Parse(codeTemplate))

	log.Printf("Getting content from files in: %s", src)
	resources := []string{}
	resourcesName := []string{}
	templates := []string{}
	expressions := []string{}

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// skip directories and file pattern exclusions. Stop process if an error was found
		if err != nil || info.IsDir() || shouldIgnore(info.Name(), exclude...) {
			return err
		}

		log.Printf("\tprocessing %s", info.Name())

		filename := toFilename(path)
		if len(filename) == 0 {
			return fmt.Errorf("could not get filename of %q", path)
		}
		resource := toResource(filename)
		if existsResource(resource, resourcesName) {
			return fmt.Errorf("there are two files with the same name causing to have the same resource name %q, rename the file %q or the other with name %q, or ignore one of them", resource, info.Name(), filename)
		}
		template := toTemplate(filename)

		resourcesName = append(resourcesName, resource)
		resources = append(resources, fmt.Sprintf(`"%s": %s,`, resource, template))

		fileContent := getFileContent(path)

		for _, exp := range re.FindAllString(fileContent, -1) {
			if strings.HasPrefix(exp, `{{"{{`) {
				continue
			}
			expressions = append(expressions, filename+" : "+exp)
		}

		templateContent := fmt.Sprintf("const %s = `%s`", template, fileContent)
		templates = append(templates, templateContent)

		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("Creating destination file: %s", dst)

	pkg := "resources"
	if internal {
		pkg = "main"
	}

	tmplData := Data{
		Package:      pkg,
		ResourcesMap: strings.Join(resources, "\n"),
		Templates:    strings.Join(templates, "\n\n"),
		Expressions:  strings.Join(expressions, "\n"),
	}

	goFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		goFile.Close()
		log.Printf("Done with Kubernetes manifest code")
	}()

	if err = rolesTemplate.Execute(goFile, tmplData); err != nil {
		return err
	}

	output, err := exec.Command("gofmt", goFile.Name()).CombinedOutput()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(goFile.Name(), output, 0644)
}

func toFile(str string) string {
	return str + ".yaml"
}

func toFilename(str string) string {
	// str may be a path or filename, if so, the filename is the file name without extension
	_, filename := filepath.Split(str)
	filename = strings.Split(filename, ".")[0]

	if len(filename) == 0 {
		filename = str
	}
	return strcase.ToKebab(filename)
}

func toResource(str string) string {
	return strcase.ToKebab(str)
}

func toTemplate(str string) string {
	return strcase.ToLowerCamel(str) + "Tpl"
}

func getFileContent(source string) string {
	content, err := ioutil.ReadFile(source)
	if err != nil {
		panic(err)
	}
	return strings.Replace(string(content), "`", "'", -1)
}

func existsResource(res string, resources []string) bool {
	for _, r := range resources {
		if res == r {
			return true
		}
	}
	return false
}

func match(path, rule string) bool {
	ok, err := filepath.Match(rule, path)
	if err != nil {
		panic(err)
	}
	if ok {
		log.Printf("\tignoring %s from rule %s", path, rule)
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

func getKubernetesTemplates(src, dst string, exclude ...string) error {
	if len(ResourceTemplates) == 0 {
		log.Printf("Run me again changing --src with the location to store the generated manifests")
		return nil
	}
	log.Printf("Generating the Kubernetes manifests")

	for resource, content := range ResourceTemplates {
		filename := toFile(resource)
		path := filepath.Join(dst, filename)

		log.Printf("\tGenerating %s", filename)
		if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("unable to create manifest file %s. %s", filename, err)
		}
	}

	log.Printf("Done generating the Kubernetes manifests")
	return nil
}

func main() {
	flag.Parse()

	excludeList := strings.Split(exclude, ",")

	if len(ResourceTemplates) == 0 {
		if err := genKubernetesCode(src, dst, excludeList...); err != nil {
			panic(err)
		}
	}

	if reverse {
		if err := getKubernetesTemplates(dst, src, excludeList...); err != nil {
			panic(err)
		}
	}
}
