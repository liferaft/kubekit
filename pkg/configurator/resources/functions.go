package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/kubekit/kubekit/pkg/provisioner/eks"
)

var tmplFuncMap template.FuncMap

func init() {
	tmplFuncMap = template.FuncMap{
		"publicKey":     publicKey,
		"privateKey":    privateKey,
		"cert":          cert,
		"readFile":      readFile,
		"getPEM":        getPEM,
		"base64Encode":  base64Encode,
		"unmarshallEFS": unmarshallEFS,
	}
}

// publicKey read the given public key located in the given certificates
// directory and platform
func publicKey(certsPath, platform, certName string) (string, error) {
	filename := filepath.Join(certsPath, platform, certName+".crt")
	return readFile(filename)
}

// privateKey read the given private key located in the given certificates
// directory and platform
func privateKey(certsPath, platform, certName string) (string, error) {
	filename := filepath.Join(certsPath, platform, certName+".key")
	return readFile(filename)
}

// readFile reads a file and returns the content
func readFile(filename string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

// getPEM returns the PEM format of a certificate joining the public and private key
func getPEM(certsPath, platform, certName string) (string, error) {
	var (
		pubKey  string
		privKey string
		err     error
	)
	if pubKey, err = publicKey(certsPath, platform, certName); err != nil {
		return "", err
	}
	if privKey, err = privateKey(certsPath, platform, certName); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", pubKey, privKey), nil
}

// base64Encode return the base64 encode of the given data
func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// cert returns the requested certificate as a PEM file encoded
func cert(certsPath, platform, certName string) (string, error) {
	pem, err := getPEM(certsPath, platform, certName)
	if err != nil {
		return "", err
	}
	return base64Encode(pem), nil
}

// unmarshallEFS returns a list of eks.ElasticFileshareData from a given
// json representation.
func unmarshallEFS(marshalled string) []eks.ElasticFileshareData {
	shares := []eks.ElasticFileshareData{}
	json.Unmarshal([]byte(marshalled), &shares)
	return shares
}
