package config

import (
	"fmt"
	"reflect"
	"strings"
)

const findByTagKey = "yaml"

// GetListFromInterface extracts a list from an interface
// if the interface is actually a list of string or a csv string
func GetListFromInterface(i interface{}) []string {
	l := []string{}
	switch t := i.(type) {
	case []interface{}:
		for _, id := range t {
			l = append(l, id.(string))
		}
	default:
		l = strings.Split(t.(string), ",")
	}
	return l
}

// SetField extracts a field content from an interface
func SetField(c interface{}, name string, value interface{}) {
	sValue := reflect.ValueOf(c).Elem()
	sFieldValue := sValue.FieldByName(name)
	if !sFieldValue.IsValid() {
		sFieldValue = fieldByTag(sValue, name)
	}
	if !sFieldValue.IsValid() {
		panic(fmt.Errorf("field %q not found", name))
	}
	if !sFieldValue.CanSet() {
		panic(fmt.Errorf("cannot set value to field %q", name))
	}

	sFieldType := sFieldValue.Type()
	if sFieldType.Kind() == reflect.String && value == nil {
		sFieldValue.SetString("")
		return
	}

	v := reflect.ValueOf(value)

	if !v.IsValid() {
		v = reflect.Zero(sFieldType)
	}

	if sFieldType != v.Type() {
		panic(fmt.Errorf("value type of field %q does not match with the config field type (%s != %s)", name, sFieldType.Kind().String(), v.Type().Kind().String()))
	}

	sFieldValue.Set(v)
}

func fieldByTag(value reflect.Value, name string) reflect.Value {
	for i := 0; i < value.NumField(); i++ {
		retValue := value.Field(i)
		tag := value.Type().Field(i).Tag
		allTagValues := tag.Get(findByTagKey)
		// the tag value may have more than one value, take the first one
		tagValue := strings.Split(allTagValues, ",")[0]
		if tagValue == name {
			return retValue
		}
	}
	panic(fmt.Errorf("not found field with %s tag named %q", findByTagKey, name))
}
