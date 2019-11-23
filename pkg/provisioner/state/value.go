package state

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/states"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var (
	errEmptyOutputValues = errors.New("outputvalues is empty")
)

// OutputKeysValueAsStringDefault returns the given OutputValue as a string in JSON format for
// the OutputValues map, however an error will make it return the defaultVal
// A check to outputvalues being empty should be done outside this function if you are not
// expecting the defaultVal to be returned when empty
func OutputKeysValueAsStringDefault(outputVals map[string]*states.OutputValue, key, defaultVal string) string {
	if outputVals == nil {
		return defaultVal
	}
	val, err := OutputKeysValueAsString(outputVals, key)
	if err != nil {
		return defaultVal
	}
	return val
}

// OutputKeysValueAsString returns the given OutputValue as a string in JSON format for
// the OutputValues map
func OutputKeysValueAsString(outputVals map[string]*states.OutputValue, key string) (string, error) {
	if outputVals == nil {
		return "", errEmptyOutputValues
	}
	if outputVal, ok := outputVals[key]; ok {
		return ValueAsString(outputVal)
	}
	return "", fmt.Errorf("key not found: %s", key)
}

// ValueAsString returns the given OutputValue as a string in JSON format.
// Examples: `15`, `Hello`, ``, `true`, `["hello", true]`
func ValueAsString(v *states.OutputValue) (s string, err error) {
	if v == nil {
		return s, nil
	}

	var b []byte
	b, err = ctyjson.Marshal(v.Value, v.Value.Type())
	if err != nil {
		return s, err
	}
	s = string(b)

	s = strings.Trim(s, `"`)
	s = strings.Replace(s, `\n`, "\n", -1)

	return s, nil
}
