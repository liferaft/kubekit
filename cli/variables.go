package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// EnvConfigPrefix is the prefix for the variables to pass to KubeKit through
// environment variables
const EnvConfigPrefix = "KUBEKIT_VAR_"

// GetVariables return a map for variable name and its value from the CLI
// parmeters and environment variables
func GetVariables(varsStr string) (variables map[string]string, warns []string, err error) {
	warns = []string{}

	// Get the variables from CLI
	variables, warn, err := getCMDVariables(varsStr)
	if err != nil {
		return nil, []string{}, err
	}
	if len(warn) != 0 {
		warns = append(warns, warn)
	}

	// Get/merge the variables and the variables from environment
	variables, warn = getEnvVariables(variables)
	if len(warn) != 0 {
		warns = append(warns, warn)
	}

	return variables, warns, nil
}

// getCMDVariables return a map for variable name and its value from the CLI parmeters
func getCMDVariables(varsStr string) (map[string]string, string, error) {
	var warn string

	variables, wrongNames, err := StringToMap(varsStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse the variables. %s", err)
	}

	if len(wrongNames) != 0 {
		warn = fmt.Sprintf("wrong definition of input variable(s): \n\t%s\nThe correct form is: --var VARIABLE01=VALUE01 --var VARIABLE02=VALUE02", strings.Join(wrongNames, "\n\t"))
	}

	return variables, warn, nil
}

// getEnvVariables return a map for variable name and its value from the environment variables
func getEnvVariables(variables map[string]string) (map[string]string, string) {
	var (
		warn   string
		ignore string
	)

	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		if strings.HasPrefix(pair[0], EnvConfigPrefix) {
			k := strings.ToLower(strings.TrimPrefix(pair[0], EnvConfigPrefix))
			if v2, ok := variables[k]; ok {
				ignore = fmt.Sprintf("%s\n\t%s = %q\tinstead, will use value: %q", ignore, strings.ToUpper(k), pair[1], v2)
				continue
			}
			variables[k] = pair[1]
		}
	}

	if len(ignore) != 0 {
		warn = fmt.Sprintf("the following environment variables were found but their values are ignored because are set in the command flags: %s", ignore)
	}

	return variables, warn
}

// StringToArray convert the cobra flag of type StringArray to an array of strings
func StringToArray(val string) ([]string, error) {
	if val == "" || val == "[]" {
		return []string{}, nil
	}
	list := strings.Trim(val, "[]")
	stringReader := strings.NewReader(list)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

// StringToMap convert the cobra flag of type StringArray in the form of
// 'key=value' to an map of strings of strings
func StringToMap(str string) (map[string]string, []string, error) {
	if len(str) == 0 {
		return map[string]string{}, nil, nil
	}

	arr, err := StringToArray(str)
	if err != nil {
		return nil, nil, err
	}

	var orphanKeys []string
	kvMap := make(map[string]string, 0)

	for _, v := range arr {
		pair := strings.SplitN(v, "=", 2)
		if len(pair) != 2 {
			orphanKeys = append(orphanKeys, v)
			continue
		}
		kvMap[strings.ToLower(pair[0])] = pair[1]
	}

	return kvMap, orphanKeys, nil
}
