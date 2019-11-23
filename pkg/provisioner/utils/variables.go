package utils

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/hashicorp/hcl/json/parser"
)

// Map returns a map[string]interface type with all the variables.
// The keys in the map are the name of the json tag
func Map(v interface{}) map[string]interface{} {
	vars := make(map[string]interface{})

	tmpVar, err := json.Marshal(v)
	if err != nil {
		return vars
	}
	json.Unmarshal(tmpVar, &vars)

	return vars
}

// HCL returns the configuration in HCL format. Required to generate the
// Terraform variables file
func HCL(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	// HCL requires JSON to parse the object
	jsonVar, err := json.Marshal(v)
	if err != nil {
		return b.Bytes(), err
	}
	// Get the AST (Abstract Syntax Tree) or HCL from the JSON
	ast, err := parser.Parse(jsonVar)
	if err != nil {
		return b.Bytes(), err
	}

	// Dump the AST into the buffer
	err = printer.Fprint(&b, ast)

	return b.Bytes(), err
}

// RemoveEnv removes the environment variables with the given prefix, it returns whatever was removed
func RemoveEnv(envConfig map[string]string, prefixes ...string) (map[string]string, map[string]string) {
	rmEnvConf := make(map[string]string)
	for _, prefix := range prefixes {
		for k, v := range envConfig {
			if strings.HasPrefix(k, prefix) {
				rmEnvConf[k] = v
				delete(envConfig, k)
			}
		}
	}
	return envConfig, rmEnvConf
}

// TrimLeft removes the given prefix from the keys of the given map with environment variables
func TrimLeft(envConfig map[string]string, prefix string) map[string]string {
	result := make(map[string]string)
	for k, v := range envConfig {
		if strings.HasPrefix(k, prefix) {
			result[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return result
}
