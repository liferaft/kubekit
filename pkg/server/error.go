package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v2"
)

// APIError represents an error to return to the HTTP/REST API, usually in JSON format
type APIError struct {
	StatusCode int    `json:"code" yaml:"code" toml:"code"`
	Status     string `json:"status" yaml:"status" toml:"status"`
	Message    string `json:"message" yaml:"message" toml:"message"`
	Target     string `json:"target" yaml:"target" toml:"target"`
}

// Error creates an API Error
func Error(code int, target, message string) *APIError {
	status := http.StatusText(code)
	return &APIError{
		StatusCode: code,
		Status:     status,
		Message:    message,
		Target:     target,
	}
}

// Errorf creates an error with a formatted error message
func Errorf(code int, target, message string, a ...interface{}) *APIError {
	msg := fmt.Sprintf(message, a...)
	return Error(code, target, msg)
}

func (e *APIError) String() string {
	data, err := e.JSON(false)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// Stringf returns the API Error in the given format: json, yaml or toml
func (e *APIError) Stringf(format string) string {
	var data []byte
	var err error

	switch format {
	case "yaml":
		data, err = e.YAML()
	case "json":
		data, err = e.JSON(false)
	case "toml":
		data, err = e.TOML()
	default:
		err = fmt.Errorf("can't stringify the API Error, unknown format %q", format)
	}

	if err != nil {
		return err.Error()
	}
	return string(data)
}

// YAML returns the API Error in YAML format
func (e *APIError) YAML() ([]byte, error) {
	return yaml.Marshal(e)
}

// JSON returns the API Error in JSON format
func (e *APIError) JSON(pp bool) ([]byte, error) {
	if pp {
		return json.MarshalIndent(e, "", "  ")
	}
	return json.Marshal(e)
}

// TOML returns the API Error in TOML format
func (e *APIError) TOML() ([]byte, error) {
	return toml.Marshal(e)
}

// Write writes the error to the given http response writer
func (e *APIError) Write(w http.ResponseWriter) {
	http.Error(w, e.String(), e.StatusCode)
}
