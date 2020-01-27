package kluster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v2"
)

const (
	emptyValue = "(none)"
	truncLen   = 10
)

// CredentialsFileName is the filename to store the credentials
const CredentialsFileName = ".credentials"

// PlatformCredentials represents the credentials for any platform but AWS
type PlatformCredentials struct {
	Platform string `json:"platform" yaml:"platform" toml:"platform" mapstructure:"platform" env:"-"`
	Server   string `json:"server" yaml:"server" toml:"server" mapstructure:"server" env:"SERVER"`
	Username string `json:"username" yaml:"username" toml:"username" mapstructure:"username" env:"USERNAME"`
	Password string `json:"password" yaml:"password" toml:"password" mapstructure:"password" env:"PASSWORD"`
	cluster  string
	path     string
}

// NewPlatformCredentials creates an struct ready for any platform credentials
func NewPlatformCredentials(clustername, platform, path string) *PlatformCredentials {
	return &PlatformCredentials{
		Platform: platform,
		cluster:  clustername,
		path:     path,
	}
}

// CredentialHandler is an interface with all the methods a credentials struct should implement
type CredentialHandler interface {
	platform() string
	clusterName() string
	clusterPath() string
	SetPath(string)
	Getenv(bool) error
	setenv() error
	List()
	Ask() error
	Read() error
	Write() error
	SetParameters(...string) error
	AssignFromMap(map[string]string) error
	parameters() []string
	asMap() map[string]string
	Empty() bool
	Complete() bool
}

// NewCredentials creates a new credentials [handler] based on the given platform
func NewCredentials(clustername, platform, path string) CredentialHandler {
	switch platform {
	case "ec2", "eks":
		return NewAWSCredentials(clustername, path)
	case "aks":
		return NewAzureCredentials(clustername, path)
	default:
		return NewPlatformCredentials(clustername, platform, path)
	}
}

func (c *PlatformCredentials) clusterName() string {
	return c.cluster
}

func (c *PlatformCredentials) clusterPath() string {
	return c.path
}

// SetPath sets the path of the credentials file be stored
func (c *PlatformCredentials) SetPath(path string) {
	c.path = path
}

func (c *PlatformCredentials) platform() string {
	return c.Platform
}

func (c *PlatformCredentials) parameters() []string {
	return []string{c.Server, c.Username, c.Password}
}

func (c *PlatformCredentials) asMap() map[string]string {
	return map[string]string{
		"server":   c.Server,
		"username": c.Username,
		"password": c.Password,
	}
}

// SetParameters sets the credentials parameters
func (c *PlatformCredentials) SetParameters(params ...string) error {
	if len(params) != 3 {
		return fmt.Errorf("incorrect number of parameters, expecting 3 and got %d", len(params))
	}
	c.Server = params[0]
	c.Username = params[1]
	c.Password = params[2]
	return nil
}

func assignFromMap(params map[string]string, v interface{}) error {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return json.Unmarshal(paramsJSON, v)
}

// AssignFromMap sets the credentials parameters from a map. The key is the name
// of the parameter as defined in the json metadata of the PlatformCredentials structure
func (c *PlatformCredentials) AssignFromMap(params map[string]string) error {
	return assignFromMap(params, c)
}

// Empty returns true if there isn't any credentials set
func (c *PlatformCredentials) Empty() bool {
	return len(c.Server) == 0 && len(c.Username) == 0 && len(c.Password) == 0
}

// Complete returns true if all credentials are set. Use it as !Complete() to
// know if there is a missing parameter. !Complete() is not Empty()
func (c *PlatformCredentials) Complete() bool {
	return len(c.Server) != 0 && len(c.Username) != 0 && len(c.Password) != 0
}

// Getenv gets the platform credentials from environment variables
func (c *PlatformCredentials) Getenv(force bool) error {
	return parseFn(c,
		func(key string) string {
			return os.Getenv(strings.ToUpper(c.platform()) + "_" + key)
		},
		func(k, v string) {},
		force,
	)
}

func (c *PlatformCredentials) setenv() error {
	return parseFn(c,
		func(key string) string {
			return ""
		},
		func(key, value string) {
			os.Setenv(strings.ToUpper(c.platform())+"_"+key, value)
		},
		true,
	)
}

// List prints to stdout the platform credentials in table format
func (c *PlatformCredentials) List() {
	header := "Name\tPlatform\tServer\tUsername\tPassword" //\tLocation"
	server := c.Server
	if len(server) == 0 {
		server = emptyValue
	}
	username := c.Username
	if len(username) == 0 {
		username = emptyValue
	}
	password := c.Password
	if len(password) == 0 {
		password = emptyValue
	} else {
		password = filter(password, 0, true)
	}
	row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", c.cluster, c.Platform, server, username, password) //, c.path)
	printCredentials(header, row)
}

// Ask to the user from stdin the platform credentials suggesting values from the environment
func (c *PlatformCredentials) Ask() error {
	return parseFn(c,
		func(key string) string {
			param := strings.Replace(strings.ToLower(key), "_", " ", -1)
			title := fmt.Sprintf("%s %s", c.Platform, param)
			env := os.Getenv(strings.ToUpper(c.platform()) + "_" + key)
			v, _ := AskDefault(title, env, true)
			return v
		},
		func(k, v string) {},
		true,
	)
}

// Read reads the platform credentials from the cluster credentials file
func (c *PlatformCredentials) Read() error {
	credentialsBytes, err := read(c.path)
	if err != nil || credentialsBytes == nil {
		return err
	}
	return yaml.Unmarshal(credentialsBytes, &c)
}

func read(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	return ioutil.ReadFile(path)
}

// Write writes the platform credentials to the cluster credentials file
func (c *PlatformCredentials) Write() error {
	credentialsBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.path, credentialsBytes, 0600)
}

func printCredentials(header, row string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, header+"\n")
	fmt.Fprintf(w, row+"\n")
	w.Flush()
}

// parseFn handles loading variables into credential structs in a generic way
// it uses reflection to determine the fields of the struct pointer
// v is the credential struct (or really just any struct pointer)
// getFn is an anonymous function that takes the name of the field and returns it's value
// setFn is an anonymous function that takes the name and value of a field to be set
// force determines if the current field should be overwritten
func parseFn(v interface{}, getFn func(string) string, setFn func(string, string), force bool) error {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Ptr {
		return fmt.Errorf("this is not a valid credential interface, it's not a struct pointer, it's a %s. %v", ptrRef.Kind().String(), ptrRef)
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return fmt.Errorf("this is not a valid credential interface, it's not a struct, it's a %s. %v", ref.Kind().String(), ref)
	}
	refType := ref.Type()

	for i := 0; i < refType.NumField(); i++ {
		refField := ref.Field(i)
		refTypeField := refType.Field(i)

		key, ok := refTypeField.Tag.Lookup("env")
		// If doesn't have the tag 'env' or is empty or is '-', next ...
		if !ok || len(key) == 0 || key == "-" {
			continue
		}

		if refField.Kind() != reflect.String {
			return fmt.Errorf("the field %q is not string type", refTypeField.Name)
		}
		// If it has a value, call the setFn ...
		if len(refField.String()) != 0 {
			setFn(key, refField.String())
		}

		value := getFn(key)
		// If the value is empty, do not assign it, next ...
		if len(value) == 0 {
			continue
		}
		if !refField.IsValid() || !refField.CanSet() {
			return fmt.Errorf("cannot assign value to field %q", refTypeField.Name)
		}
		if force || refField.Len() == 0 {
			refField.SetString(value)
		}
	}

	return nil
}

// AskDefault asks to the user a query proposing a default value
func AskDefault(title, defValue string, sensitive bool) (string, error) {
	title = fmt.Sprintf("Enter %s [%s]: ", title, defValue)
	resp, err := Ask(title, sensitive)
	if len(resp) == 0 {
		resp = defValue
	}
	return resp, err
}

// Ask asks a query to the user
func Ask(title string, sensitive bool) (string, error) {
	fmt.Print(title)
	reader := bufio.NewReader(os.Stdin)
	resp, err := reader.ReadString('\n')
	return strings.TrimRight(resp, "\r\n"), err
}

// filter returns the given string tructated to a max length, replacing the
// trucated chars with `...`. Or, returns repeated `*` if it is a sensitive
// information. Or, returns `(none)` if it's empty. The returned string
// (tructated or `*`'ed) is not smaller than 5 characters.
func filter(str string, maxLen int, sensitive bool) string {
	totalLength := len(str)

	// return `(none)` if empty
	if totalLength == 0 {
		return emptyValue
	}

	// return `*` if sensitive
	if sensitive {
		len := totalLength
		if maxLen != 0 && maxLen < totalLength {
			len = maxLen
		}
		if len < 5 {
			len = 5
		}
		return strings.Repeat("*", len)
	}

	// returns a truncated string
	if totalLength <= maxLen {
		return str
	}
	if maxLen < 5 {
		// We don't shorten to less than 5 chars
		// as that would be pointless with ... (3 chars)
		maxLen = 5
	}

	dots := "..."
	partLen := maxLen / 2

	leftStrx := partLen - 1
	leftPart := str[0:leftStrx]

	rightStrx := totalLength - partLen - 1

	overlap := maxLen - (partLen*2 + len(dots))
	if overlap < 0 {
		rightStrx -= overlap
	}

	rightPart := str[rightStrx:]

	return leftPart + dots + rightPart
}
